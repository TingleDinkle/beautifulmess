package main

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
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

	// Level Transition state
	IsTransitioning bool
	TransitionTime  float64 // Current progress (0.0 to 1.0)
	TargetLevel     int
	ReunionPoint    core.Vector2

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
	// A predictable but varied seed ensures unique level layouts across different sessions
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

	g.SpriteRunner = generateAstroSprite()
	
	// Pre-loading audio samples avoids frame-stuttering during combat
	g.World.Audio.LoadFile("shoot", "assets/shoot.wav")
	g.World.Audio.LoadFile("boom", "assets/boom.wav")

	// Base color initialization prevents flickering before the first entropy cycle
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

		if idx >= len(g.Levels) { idx = len(g.Levels) - 1 }

		g.CurrentLevel = idx

		lvl := g.Levels[idx]

		w := g.World

	

		// Total state reset prevents entity leakage and memory bloat between level transitions

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

		

		w.Particles.Reset()

		

		g.spawnLevelEntities(lvl)

	}

	

	func (g *Game) spawnLevelEntities(lvl level.Level) {

		w := g.World

	

		for _, well := range lvl.Wells {

			id := w.CreateEntity()

			w.Transforms[id] = &components.Transform{Position: well.Position}

			w.GravityWells[id] = &components.GravityWell{Radius: well.Radius, Mass: well.Mass}

			w.Tags[id] = &components.Tag{Name: "gravity_well"}

		}

	

		for _, wall := range lvl.Walls {

			spawnWall(w, wall.X, wall.Y, wall.Destructible)

		}

	

				// Entity archetypes are composed through data rather than rigid inheritance

	

				g.SpectreID = w.CreateEntity()

	

				w.Tags[g.SpectreID] = &components.Tag{Name: "spectre"}

	

				w.Transforms[g.SpectreID] = &components.Transform{Position: lvl.StartP2}

	

				w.Physics[g.SpectreID] = &components.Physics{MaxSpeed: 6.0, Friction: 0.96, Mass: 1.0, GravityMultiplier: 3.5}

	

				w.Renders[g.SpectreID] = &components.Render{Sprite: g.SpriteSpectre, Color: color.RGBA{255, 50, 50, 255}, Glow: true}

	

				w.AIs[g.SpectreID] = &components.AI{ScriptName: "spectre.lua"}

	

			

	

				g.RunnerID = w.CreateEntity()

		w.Tags[g.RunnerID] = &components.Tag{Name: "runner"}

		w.Transforms[g.RunnerID] = &components.Transform{Position: lvl.StartP1}

		w.Physics[g.RunnerID] = &components.Physics{MaxSpeed: 7.5, Friction: 0.92, Mass: 1.0}

		w.Renders[g.RunnerID] = &components.Render{Sprite: g.SpriteRunner, Color: color.RGBA{0, 255, 255, 255}, Glow: true, Scale: 1.0}

		w.AIs[g.RunnerID] = &components.AI{ScriptName: "runner.lua"}

		w.InputControlleds[g.RunnerID] = &components.InputControlled{}

		w.ProjectileEmitters[g.RunnerID] = &components.ProjectileEmitter{Interval: 1.0}

	

		w.AIs[g.SpectreID].TargetID = int(g.RunnerID)

		w.AIs[g.RunnerID].TargetID = int(g.SpectreID)

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

// ======================================================================================
// GAME LOOP
// ======================================================================================

func (g *Game) Update() error {
	g.handleInput()

	if g.IsTransitioning {
		return g.updateTransition()
	}

	if g.IsPaused {
		return g.updatePaused()
	}

	return g.updateActive()
}

func (g *Game) updateTransition() error {
	// A deliberate transition speed allows the visual bloom to reach its emotional peak before the scene reset
	g.TransitionTime += 1.0 / 120.0 
	
	if g.TransitionTime >= 1.0 {
		g.IsTransitioning = false
		g.LoadLevel(g.TargetLevel)
		g.TransitionTime = 0
	}
	return nil
}

func (g *Game) handleInput() {
	// Manual toggles allow users to control session flow independently of game events
	if inpututil.IsKeyJustPressed(ebiten.KeyP) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.IsPaused = !g.IsPaused
		if !g.IsPaused {
			g.Popup = nil
		}
	}
}

func (g *Game) updatePaused() error {
	if g.Popup != nil {
		// Pagination controls allow narrative data to be explored at the user's pace
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			g.PopupPhotoIndex = (g.PopupPhotoIndex + 1) % len(g.Popup.Photos)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			g.PopupPhotoIndex = (g.PopupPhotoIndex - 1 + len(g.Popup.Photos)) % len(g.Popup.Photos)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			// Closing the popup initiates the 'Reunion' transition phase
			g.IsPaused = false
			g.Popup = nil
			g.IsTransitioning = true
			g.TransitionTime = 0
			g.TargetLevel = g.CurrentLevel + 1
		}
	}
	return nil
}

func (g *Game) updateActive() error {
	// Frame-skipping during hit-stop emphasizes impact weight through rhythmic disruption
	if g.HitStop > 0 {
		g.HitStop -= 1.0 / 60.0
		g.World.ScreenShake *= 0.9
		return nil
	}

	g.World.ScreenShake *= 0.9
	if g.World.ScreenShake < 0.5 { g.World.ScreenShake = 0 }

	lvl := &g.Levels[g.CurrentLevel]
	systems.SystemInput(g.World)
	systems.SystemAI(g.World, lvl)
	systems.SystemPhysics(g.World)
	systems.SystemEntropy(g.World, g.FrostMask)
	systems.SystemProjectileEmitter(g.World)
	systems.SystemLifetime(g.World)
	g.World.Particles.Update()

	return g.checkWinCondition(lvl)
}

func (g *Game) checkWinCondition(lvl *level.Level) error {
	pSpec, okS := g.World.Transforms[g.SpectreID]
	pRun, okR := g.World.Transforms[g.RunnerID]
	if !okS || !okR { return nil }

	// Spatial proximity check for win condition triggers the 'Reunion' state
	if core.DistWrapped(pSpec.Position, pRun.Position) < 80 {
		for id, well := range g.World.GravityWells {
			wellTrans, ok := g.World.Transforms[id]
			if !ok || well == nil { continue }
			
			// Captured by the singularity: The ultimate endpoint of their shared trajectory
			if core.DistWrapped(pSpec.Position, wellTrans.Position) < well.Radius + 15 {
				g.IsPaused = true
				g.Popup = &lvl.Memory
				g.PopupTime = time.Now()
				g.PopupPhotoIndex = 0
				g.ReunionPoint = pSpec.Position // Persistent tracking anchors the bloom effect to the narrative climax
				g.World.Audio.Play("chime")
				return nil
			}
		}
	}
	return nil
}



func (g *Game) Draw(screen *ebiten.Image) {
	// Screen shake is calculated once per frame to ensure visual consistency across all layers
	shake := core.Vector2{}
	if g.World.ScreenShake > 0 {
		shake.X = (rand.Float64() - 0.5) * g.World.ScreenShake * 2
		shake.Y = (rand.Float64() - 0.5) * g.World.ScreenShake * 2
	}

	g.drawWorld(screen, shake)
	
	if g.IsTransitioning {
		g.drawTransition(screen)
	}
	
	g.drawUI(screen)
}

func (g *Game) drawTransition(screen *ebiten.Image) {
	// A procedural bloom effect translates the narrative theme of growth and union into a physical visual event
	t := g.TransitionTime
	
	// 1. Reunion Bloom: A specialized floral burst at the point of character convergence
	g.drawBloom(screen, g.ReunionPoint, 400.0*t, color.RGBA{255, 200, 255, uint8(255 * (1.0 - t))})

	// 2. Gravity Well Bloom: Every singularity becomes a source of radiating light
	for id, well := range g.World.GravityWells {
		if well == nil { continue }
		trans, ok := g.World.Transforms[id]
		if !ok { continue }
		
		g.drawBloom(screen, trans.Position, 300.0*t, color.RGBA{255, 255, 200, uint8(200 * (1.0 - t))})
		
		// Central core flare emphasizes the transformation of the trap into a beacon
		coreAlpha := uint8(200 * t)
		vector.DrawFilledCircle(screen, float32(trans.Position.X), float32(trans.Position.Y), float32(well.Radius * (1.0 + t*2)), color.RGBA{255, 255, 255, coreAlpha}, true)
	}

	// 3. Sunshine Pixelizing: A grid-based transition enforces the digital/arcade aesthetic
	const cellSize = 40
	for y := 0; y < core.ScreenHeight; y += cellSize {
		for x := 0; x < core.ScreenWidth; x += cellSize {
			// Offset based on time and position creates a 'sweeping' sunshine effect
			prog := t*2.0 - (float64(x)/core.ScreenWidth)*0.5 - (float64(y)/core.ScreenHeight)*0.5
			if prog < 0 { prog = 0 }
			if prog > 1 { prog = 1 }
			
			size := float32(cellSize) * float32(prog)
			if size > 0 {
				vector.DrawFilledRect(screen, float32(x), float32(y), size, size, color.RGBA{255, 255, 200, uint8(255*prog)}, false)
			}
		}
	}
}

func (g *Game) drawBloom(screen *ebiten.Image, center core.Vector2, radius float64, clr color.RGBA) {
	// Circular 'petal' distribution mimics organic flowering patterns
	numPetals := 12
	for i := 0; i < numPetals; i++ {
		angle := (float64(i) / float64(numPetals)) * 2 * math.Pi
		px := center.X + math.Cos(angle)*radius
		py := center.Y + math.Sin(angle)*radius
		vector.DrawFilledCircle(screen, float32(px), float32(py), float32(radius*0.2), clr, true)
	}
}



func (g *Game) drawWorld(screen *ebiten.Image, shake core.Vector2) {
	// Render order establishes visual depth: Background -> Particles -> World -> Mist -> Entities
	g.drawBackground(screen)
	g.World.Particles.Draw(screen)
	
	lvl := &g.Levels[g.CurrentLevel]
	
	// Safe entity tracking ensures the renderer remains stable even during state transitions
	spectrePos := core.Vector2{}
	if trans, ok := g.World.Transforms[g.SpectreID]; ok {
		spectrePos = trans.Position
	}
	
	systems.DrawLevel(screen, g.World, lvl, spectrePos, shake)
	
	g.drawMist(screen)
	systems.DrawEntities(screen, g.World, shake)
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	w, h := screen.Size()
	t := float32(time.Since(g.StartTime).Seconds())
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]interface{}{"Cursor": []float32{0, 0, t}}
	screen.DrawRectShader(w, h, g.NebulaShader, op)
}

func (g *Game) drawMist(screen *ebiten.Image) {
	// Mist buffer is scaled to lower resolution to achieve a dithering effect without high GPU overhead
	g.FrostImg.WritePixels(g.FrostMask.Pix)
	mistOp := &ebiten.DrawImageOptions{}
	mistOp.GeoM.Scale(float64(core.ScreenWidth)/core.MistWidth, float64(core.ScreenHeight)/core.MistHeight)
	screen.DrawImage(g.FrostImg, mistOp)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	if !g.IsPaused {
		return
	}

	if g.Popup != nil {
		g.drawPopup(screen)
	} else {
		g.drawPauseMenu(screen)
	}
}

func (g *Game) drawPopup(screen *ebiten.Image) {
	dt := float64(time.Since(g.PopupTime).Seconds())
	scale := math.Min(1.0, dt*5.0)
	scale = scale * (1.0 + 0.3*(1.0-scale)) // Elastic scaling creates a 'tactile' UI feel

	bx, by := 300.0, 100.0
	bw, bh := 680.0, 500.0
	cx, cy := bx+bw/2, by+bh/2
	bw, bh = bw*scale, bh*scale
	bx, by = cx-bw/2, cy-bh/2

	vector.DrawFilledRect(screen, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{10, 0, 10, 245}, false)
	vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 4, color.RGBA{180, 20, 40, 255}, false)

	if scale > 0.9 {
		g.renderPopupContent(screen, bx, by, bw, bh)
	}
}

func (g *Game) renderPopupContent(screen *ebiten.Image, bx, by, bw, bh float64) {
	ebitenutil.DebugPrintAt(screen, "[ MEMORY FRAGMENT ]", int(bx)+280, int(by)+20)
	ebitenutil.DebugPrintAt(screen, g.Popup.Title, int(bx)+30, int(by)+50)

	photoW, photoH := 400.0, 300.0
	px, py := bx+(bw-photoW)/2, by+80
	
	// Dynamic noise pattern acts as a placeholder for fragmented visual data
	seed := int64(g.PopupPhotoIndex * 100)
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + seed))
	for i := 0; i < 200; i++ {
		rx, ry := rng.Float64()*photoW, rng.Float64()*photoH
		c := uint8(rng.Intn(255))
		cr, cg, cb := c, c, c
		// Visual tinting provides immediate feedback for pagination state
		if g.PopupPhotoIndex == 0 { cr = 255; cg /= 2; cb /= 2 }
		if g.PopupPhotoIndex == 1 { cr /= 2; cg = 255; cb /= 2 }
		if g.PopupPhotoIndex == 2 { cr /= 2; cg /= 2; cb = 255 }
		vector.DrawFilledRect(screen, float32(px+rx), float32(py+ry), float32(rng.Float64()*30), 2, color.RGBA{cr, cg, cb, 255}, false)
	}
	vector.StrokeRect(screen, float32(px), float32(py), float32(photoW), float32(photoH), 2, color.RGBA{100, 100, 100, 255}, false)

	if len(g.Popup.Photos) > 1 {
		ebitenutil.DebugPrintAt(screen, "< PREV (A)", int(px)-80, int(py)+int(photoH)/2)
		ebitenutil.DebugPrintAt(screen, "(D) NEXT >", int(px)+int(photoW)+10, int(py)+int(photoH)/2)
		ebitenutil.DebugPrintAt(screen, "IMG "+string(rune('1'+g.PopupPhotoIndex))+"/"+string(rune('1'+len(g.Popup.Photos)-1)), int(px)+int(photoW)-60, int(py)+int(photoH)+10)
	}

	ebitenutil.DebugPrintAt(screen, g.Popup.Description, int(bx)+30, int(by)+400)
	ebitenutil.DebugPrintAt(screen, "[ SPACE TO RECOVER ]", int(bx)+260, int(by)+470)
}

func (g *Game) drawPauseMenu(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, float32(core.ScreenWidth), float32(core.ScreenHeight), color.RGBA{0, 0, 0, 180}, false)
	
	bw, bh := 400.0, 120.0
	bx, by := float64(core.ScreenWidth-bw)/2, float64(core.ScreenHeight-bh)/2
	
	vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 2, color.RGBA{33, 33, 255, 255}, false)
	ebitenutil.DebugPrintAt(screen, "--- PAUSED ---", int(bx)+145, int(by)+40)
	ebitenutil.DebugPrintAt(screen, "RESUME: PRESS P", int(bx)+145, int(by)+70)
	
	// Symmetrical arcade-inspired accents reinforce the retro theme
	for px := 0; px < 5; px++ {
		vector.DrawFilledCircle(screen, float32(bx)+float32(px*80)+40, float32(by)+15, 2, color.RGBA{255, 255, 0, 255}, true)
		vector.DrawFilledCircle(screen, float32(bx)+float32(px*80)+40, float32(by)+float32(bh)-15, 2, color.RGBA{255, 255, 0, 255}, true)
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