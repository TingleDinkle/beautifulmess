package main

import (
	"image"
	"image/color"
	_ "image/png" // Ensure PNG support for completeness
	"log"
	"math/rand"
	"time"

	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/level"
	"beautifulmess/pkg/systems"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ======================================================================================
// THE VIBE (Shaders)
// ======================================================================================

// This shader generates the cosmic background.
// It is significantly cheaper to compute this procedurally on the GPU than to manage large texture assets.
var shaderNebula = []byte(`
package main

var Cursor vec3 // .z is Time

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	pos := position.xy / imageDstTextureSize()
	t := Cursor.z * 0.2 // Slow time down to convey scale.
	
	// Fractional Brownian Motion (FBM) creates organic, cloud-like structures
	val := 0.0
	scale := 3.0
	for i := 0; i < 3; i++ {
		val += sin(pos.x*scale + t) + sin(pos.y*scale - t*0.5)
		scale *= 2.0
	}
	val /= 6.0
	
	// Gothic Void Palette
	base := vec3(0.1, 0.0, 0.2)   // Deep Purple
	light := vec3(0.0, 0.4, 0.5)  // Cyan glow
	finalColor := base + light * (val + 0.5)
	
	// Darken corners to focus player attention on the center
	dist := distance(pos, vec2(0.5))
	finalColor *= (1.0 - dist*0.8)
	
	return vec4(finalColor, 1.0)
}
`)

// ======================================================================================
// GAME STATE
// ======================================================================================

type Game struct {
	World     *world.World
	RunnerID  core.Entity
	SpectreID core.Entity

	Levels       []level.Level
	CurrentLevel int
	IsPaused     bool
	Popup        *level.MemoryNode
	PopupTime    time.Time
	PopupPhotoIndex int

	// Graphics Assets
	FrostMask     *image.RGBA
	FrostImg      *ebiten.Image
	NebulaShader  *ebiten.Shader
	SpriteSpectre *ebiten.Image
	SpriteRunner  *ebiten.Image

	StartTime time.Time
	
	// Game Feel "Juice"
	HitStop     float64 // Timer in seconds
}

// ======================================================================================
// BOOTSTRAP
// ======================================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	s, err := ebiten.NewShader(shaderNebula)
	if err != nil {
		log.Fatal(err)
	}

	g := &Game{
		World:         world.NewWorld(),
		Levels:        level.InitLevels(),
		FrostMask:     image.NewRGBA(image.Rect(0, 0, core.MistWidth, core.MistHeight)),
		NebulaShader:  s,
		StartTime:     time.Now(),
		SpriteSpectre: generateGothicSprite(),
	}

	// Load Assets
	g.SpriteRunner = generateAstroSprite()
	
	g.World.Audio.LoadFile("shoot", "assets/shoot.wav")
	g.World.Audio.LoadFile("boom", "assets/boom.wav")

	// Initialize frost mask with base color to avoid initial transparency artifacts
	c := color.RGBA{10, 5, 20, 240}
	for i := 0; i < len(g.FrostMask.Pix); i += 4 {
		g.FrostMask.Pix[i], g.FrostMask.Pix[i+1], g.FrostMask.Pix[i+2], g.FrostMask.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	g.FrostImg = ebiten.NewImageFromImage(g.FrostMask)

	systems.InitLua(g.World)

	g.LoadLevel(0)
	return g
}

func (g *Game) LoadLevel(idx int) {
	if idx >= len(g.Levels) {
		idx = len(g.Levels) - 1
	}
	g.CurrentLevel = idx
	lvl := g.Levels[idx]
	w := g.World
	// Reset ALL entity components to clear previous level artifacts
	w.Transforms = make(map[core.Entity]*components.Transform)
	w.Physics = make(map[core.Entity]*components.Physics)
	w.Renders = make(map[core.Entity]*components.Render)
	w.AIs = make(map[core.Entity]*components.AI)
	w.Tags = make(map[core.Entity]*components.Tag)
	w.GravityWells = make(map[core.Entity]*components.GravityWell)
	w.InputControlleds = make(map[core.Entity]*components.InputControlled)
	w.Walls = make(map[core.Entity]*components.Wall)
	w.ProjectileEmitters = make(map[core.Entity]*components.ProjectileEmitter)
	w.Lifetimes = make(map[core.Entity]*components.Lifetime)
	
	// Clear Particles
	w.Particles.Reset()
	
	// Spawn Gravity Wells
	for _, well := range lvl.Wells {
		id := w.CreateEntity()
		w.Transforms[id] = &components.Transform{Position: well.Position}
		w.GravityWells[id] = &components.GravityWell{
			Radius: well.Radius,
			Mass:   well.Mass,
		}
		w.Tags[id] = &components.Tag{Name: "gravity_well"}
	}

	// Spawn Spectre
	g.SpectreID = w.CreateEntity()
	w.Tags[g.SpectreID] = &components.Tag{Name: "spectre"}
	w.Transforms[g.SpectreID] = &components.Transform{Position: lvl.StartP2}
	w.Physics[g.SpectreID] = &components.Physics{MaxSpeed: 6.0, Friction: 0.96, Mass: 1.0, GravityMultiplier: 3.5}
	w.Renders[g.SpectreID] = &components.Render{Sprite: g.SpriteSpectre, Color: color.RGBA{255, 50, 50, 255}, Glow: true}
	w.AIs[g.SpectreID] = &components.AI{ScriptName: "spectre.lua"}

	// Spawn Runner
	g.RunnerID = w.CreateEntity()
	w.Tags[g.RunnerID] = &components.Tag{Name: "runner"}
	w.Transforms[g.RunnerID] = &components.Transform{Position: lvl.StartP1}
	w.Physics[g.RunnerID] = &components.Physics{MaxSpeed: 7.5, Friction: 0.92, Mass: 1.0}
	w.Renders[g.RunnerID] = &components.Render{
		Sprite: g.SpriteRunner, 
		Color: color.RGBA{0, 255, 255, 255}, 
		Glow: true,
		Scale: 1.0, // Scale 1.0 because we are now generating a 16x16 vector sprite, not a tiny 8x8 BMP
	}
	w.AIs[g.RunnerID] = &components.AI{ScriptName: "runner.lua"}
	w.InputControlleds[g.RunnerID] = &components.InputControlled{}
	w.ProjectileEmitters[g.RunnerID] = &components.ProjectileEmitter{Interval: 1.0}

	w.AIs[g.SpectreID].TargetID = int(g.RunnerID)
	w.AIs[g.RunnerID].TargetID = int(g.SpectreID)
	
	g.generateMap(w)
}

func (g *Game) generateMap(w *world.World) {
	// Map generation disabled per user request.
	// Returning to empty/boundless void.
}

func spawnWall(w *world.World, x, y float64, destructible bool) {
	id := w.CreateEntity()
	w.Transforms[id] = &components.Transform{Position: core.Vector2{X: x, Y: y}}
	w.Walls[id] = &components.Wall{
		Size: 10,
		Destructible: destructible,
	}
	
	// Create visual for wall: Solid Neon Line Segment
	img := ebiten.NewImage(10, 10)
	
	if destructible {
		// Destructible: Orange "Data Block"
		// A slightly smaller block to look less permanent
		c := color.RGBA{255, 150, 50, 255}
		img.Fill(color.Transparent)
		// Draw a centered box 8x8
		for py := 1; py < 9; py++ {
			for px := 1; px < 9; px++ {
				img.Set(px, py, c)
			}
		}
	} else {
		// Indestructible: Pure Neon Blue Block
		// When placed in a row, these merge into a seamless line.
		c := color.RGBA{0, 255, 255, 255} // Cyan/Neon Blue
		img.Fill(c)
	}
	
	w.Renders[id] = &components.Render{
		Sprite: img,
		Color:  color.RGBA{255, 255, 255, 255},
		Scale:  1.0,
	}
}

// Procedural Art Generators
func generateGothicSprite() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	cRed := color.RGBA{180, 20, 40, 255}
	cDark := color.RGBA{80, 10, 20, 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			dx, dy := x-8, y-8
			dist := dx*dx + dy*dy*2
			if dist < 40 {
				img.Set(x, y, cRed)
			} else if dist < 50 && rand.Float64() < 0.5 {
				img.Set(x, y, cDark)
			}
		}
	}
	img.Set(6, 6, color.White)
	img.Set(9, 6, color.White)
	return ebiten.NewImageFromImage(img)
}

func generateAstroSprite() *ebiten.Image {
	img := ebiten.NewImage(16, 16)
	
	// Draw a white triangle pointing right
	// Tip at (16, 8)
	// Back at x=0
	
	// Manual rasterization for robustness
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			// Triangle logic
			// (0,0) -> (16,8)  => y = 0.5x
			// (0,16) -> (16,8) => y = -0.5x + 16
			
			fx, fy := float64(x), float64(y)
			
			inOuter := (fy >= 0.5*fx) && (fy <= -0.5*fx+16.0)
			
			// Cutout logic
			// (0,0) -> (4,8)   => y = 2x
			// (0,16) -> (4,8)  => y = -2x + 16
			inCutout := (fx < 4) && (fy >= 2.0*fx) && (fy <= -2.0*fx+16.0)
			
			if inOuter && !inCutout {
				img.Set(x, y, color.White)
			}
		}
	}
	
	// Add a simple 1px black border for "retro" feel?
	// The background is dark, so a black border might be invisible unless over particles.
	// But let's keep it simple: just the white shape.
	
	return img
}

func generateCyberSprite() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	cCyan := color.RGBA{0, 255, 255, 255}
	cBlue := color.RGBA{0, 100, 200, 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if x > y && x > (15-y) {
				img.Set(x, y, cCyan)
			}
			if y == 8 || x == 8 {
				img.Set(x, y, cBlue)
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

// ======================================================================================
// RUN LOOP
// ======================================================================================

func (g *Game) Update() error {
	// Manual Pause Toggle
	if inpututil.IsKeyJustPressed(ebiten.KeyP) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.IsPaused = !g.IsPaused
		// If manually pausing, ensure no Popup is set (unless we want to keep it)
		// If we unpause, clear popup
		if !g.IsPaused {
			g.Popup = nil
		}
	}

	if g.IsPaused {
		// Allow unpausing via Space if it was a story popup, or P/Esc handled above
		if g.Popup != nil && ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.IsPaused = false
			g.Popup = nil
			g.LoadLevel(g.CurrentLevel + 1)
		}
		return nil
	}

	// Tick Systems
	
	// Game Feel: HitStop
	if g.HitStop > 0 {
		g.HitStop -= 1.0 / 60.0
		// Shake decay
		g.World.ScreenShake *= 0.9
		// Render visuals but skip physics/logic updates
		g.World.Particles.Update() 
		return nil
	}
	
	// Game Feel: Shake Decay during normal play
	g.World.ScreenShake *= 0.9
	if g.World.ScreenShake < 0.5 { g.World.ScreenShake = 0 }

	lvl := &g.Levels[g.CurrentLevel] // Pass pointer to current level
	systems.SystemInput(g.World)
	systems.SystemAI(g.World, lvl)
	systems.SystemPhysics(g.World)
	systems.SystemEntropy(g.World, g.FrostMask)
	systems.SystemProjectileEmitter(g.World)
	systems.SystemLifetime(g.World)
	
	// Update Particles (Visuals only, no game logic impact)
	g.World.Particles.Update()

	// Check win condition: Spectre trapped near memory node while runner is present
	pSpec := g.World.Transforms[g.SpectreID].Position
	pRun := g.World.Transforms[g.RunnerID].Position

	dSpecMem := core.DistWrapped(pSpec, lvl.Memory.Position)
	dRunSpec := core.DistWrapped(pSpec, pRun)

	if dSpecMem < core.MemoryRadius && dRunSpec < 80 {
		g.IsPaused = true
		g.Popup = &lvl.Memory
		g.PopupTime = time.Now()
		g.World.Audio.Play("chime")
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Game Feel: Screen Shake
	shake := core.Vector2{}
	if g.World.ScreenShake > 0 {
		shake.X = (rand.Float64() - 0.5) * g.World.ScreenShake * 2 // Multiplier already high enough if value is 5-10
		shake.Y = (rand.Float64() - 0.5) * g.World.ScreenShake * 2
	}
	
	// Draw Nebula (Shader) - Background usually doesn't shake or shakes differently (parallax). Let's keep it static.
	w, h := screen.Size()
	t := float32(time.Since(g.StartTime).Seconds())
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]interface{}{"Cursor": []float32{0, 0, t}}
	screen.DrawRectShader(w, h, g.NebulaShader, op)

	// Draw Particles (Background layer for depth)
	g.World.Particles.Draw(screen)

	// Draw Gravity Wells (Black Holes)
	lvl := &g.Levels[g.CurrentLevel]
	systems.DrawLevel(screen, g.World, lvl, g.World.Transforms[g.SpectreID].Position, shake)

	// Draw Mist Layer
	g.FrostImg.WritePixels(g.FrostMask.Pix)
	mistOp := &ebiten.DrawImageOptions{}
	mistOp.GeoM.Scale(float64(core.ScreenWidth)/core.MistWidth, float64(core.ScreenHeight)/core.MistHeight)
	screen.DrawImage(g.FrostImg, mistOp)

	// Draw Entities
	systems.DrawEntities(screen, g.World, shake)

	// 5. Draw UI Modal
	if g.IsPaused {
		// Calculate Center
		cx, cy := float32(core.ScreenWidth/2), float32(core.ScreenHeight/2)

					if g.Popup != nil {
						// Story Popup
						dt := float64(time.Since(g.PopupTime).Seconds())
						scale := dt * 5.0
						if scale > 1.0 { scale = 1.0 }
						scale = scale * (1.0 + 0.3*(1.0-scale))
		
						// Larger box for Photos
						bx, by := 300.0, 100.0
						bw, bh := 680.0, 500.0
						
						// Apply scale
						cxModal, cyModal := bx+bw/2, by+bh/2
						bw *= scale
						bh *= scale
						bx = cxModal - bw/2
						by = cyModal - bh/2
		
						// Modal Background
						vector.DrawFilledRect(screen, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{10, 0, 10, 245}, false)
						vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 4, color.RGBA{180, 20, 40, 255}, false)
		
						if scale > 0.9 {
							// Title
							ebitenutil.DebugPrintAt(screen, "[ MEMORY FRAGMENT ]", int(bx)+280, int(by)+20)
							ebitenutil.DebugPrintAt(screen, g.Popup.Title, int(bx)+30, int(by)+50);
		
							// Photo Area
							photoW, photoH := 400.0, 300.0
							px, py := bx+(bw-photoW)/2, by+80
							
							// Placeholder Photo Art (seeded by index)
							// We use the index to change the color/pattern to show navigation works
							seed := int64(g.PopupPhotoIndex * 100)
							rng := rand.New(rand.NewSource(time.Now().UnixNano() + seed))
							
							for i := 0; i < 200; i++ {
								rx := rng.Float64() * photoW
								ry := rng.Float64() * photoH
								rw := rng.Float64() * 30
													c := uint8(rng.Intn(255))
													// Tint based on index
													cr, cg, cb := c, c, c
													if g.PopupPhotoIndex == 0 { cr = 255; cg = c/2; cb = c/2 } // Reddish
													if g.PopupPhotoIndex == 1 { cr = c/2; cg = 255; cb = c/2 } // Greenish
													if g.PopupPhotoIndex == 2 { cr = c/2; cg = c/2; cb = 255 } // Bluish
													
													vector.DrawFilledRect(screen, float32(px+rx), float32(py+ry), float32(rw), 2, color.RGBA{uint8(cr), uint8(cg), uint8(cb), 255}, false)							}
							vector.StrokeRect(screen, float32(px), float32(py), float32(photoW), float32(photoH), 2, color.RGBA{100, 100, 100, 255}, false)
		
							// Navigation Arrows
							if len(g.Popup.Photos) > 1 {
								// Left Arrow
								ebitenutil.DebugPrintAt(screen, "< PREV (A)", int(px)-80, int(py)+int(photoH)/2)
								// Right Arrow
								ebitenutil.DebugPrintAt(screen, "(D) NEXT >", int(px)+int(photoW)+10, int(py)+int(photoH)/2)
								
								// Photo Counter
								ebitenutil.DebugPrintAt(screen, "IMG "+string(rune('1'+g.PopupPhotoIndex))+"/"+string(rune('1'+len(g.Popup.Photos)-1)), int(px)+int(photoW)-60, int(py)+int(photoH)+10)
							}
		
							// Description
							ebitenutil.DebugPrintAt(screen, g.Popup.Description, int(bx)+30, int(by)+400)
							
							ebitenutil.DebugPrintAt(screen, "[ SPACE TO RECOVER ]", int(bx)+260, int(by)+470)
						}
					} else {			// Manual Pause - Retro Pacman Style
			
			// Dim background
			vector.DrawFilledRect(screen, 0, 0, float32(core.ScreenWidth), float32(core.ScreenHeight), color.RGBA{0, 0, 0, 180}, false)
			
			bw, bh := 400.0, 120.0
			bx := float64(cx) - bw/2
			by := float64(cy) - bh/2
			
			pacBlue := color.RGBA{33, 33, 255, 255}
			pacYellow := color.RGBA{255, 255, 0, 255}

			// Clean Grid Border
			vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 2, pacBlue, false)
			
			// "PAUSED" Text in Yellow
			// Since DebugPrint is always white, we can't easily change it without a font.
			// However, we can draw a yellow rect behind it or just use white and accent it.
			// Let's draw a yellow bar and put black text? 
			// No, let's just stick to the iconic white debug text but surround it with yellow pellets.
			
			ebitenutil.DebugPrintAt(screen, "--- PAUSED ---", int(bx)+145, int(by)+40)
			ebitenutil.DebugPrintAt(screen, "RESUME: PRESS P", int(bx)+145, int(by)+70)
			
			// Draw simple yellow "Pellet" grid in background or corners
			for px := 0; px < 5; px++ {
				vector.DrawFilledCircle(screen, float32(bx)+float32(px*80)+40, float32(by)+15, 2, pacYellow, true)
				vector.DrawFilledCircle(screen, float32(bx)+float32(px*80)+40, float32(by)+float32(bh)-15, 2, pacYellow, true)
			}
		}
	}
}

func (g *Game) Layout(w, h int) (int, int) { return core.ScreenWidth, core.ScreenHeight }

func main() {
	ebiten.SetWindowSize(core.ScreenWidth, core.ScreenHeight)
	ebiten.SetWindowTitle("Beautiful Mess: The Final Code")
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}