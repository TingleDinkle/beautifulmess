package main

import (
	"image"
	"image/color"
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

	// Graphics Assets
	FrostMask     *image.RGBA
	FrostImg      *ebiten.Image
	NebulaShader  *ebiten.Shader
	SpriteSpectre *ebiten.Image
	SpriteRunner  *ebiten.Image

	StartTime time.Time
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
		SpriteRunner:  generateCyberSprite(),
	}

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
	// Reset physics and logic states
	w.Physics = make(map[core.Entity]*components.Physics)
	w.GravityWells = make(map[core.Entity]*components.GravityWell)
	w.InputControlleds = make(map[core.Entity]*components.InputControlled)

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
	w.Renders[g.RunnerID] = &components.Render{Sprite: g.SpriteRunner, Color: color.RGBA{0, 255, 255, 255}, Glow: true}
	w.AIs[g.RunnerID] = &components.AI{ScriptName: "runner.lua"}
	w.InputControlleds[g.RunnerID] = &components.InputControlled{}

	// Bind Lua (Now handled in systems, needs to be called after entities are created)
	// We pass targetID so they know who to chase/avoid
	// systems.BindLua(w, g.SpectreID, g.RunnerID)
	// systems.BindLua(w, g.RunnerID, g.SpectreID)

	// Link them (so they can find each other) - Wait, this was direct assignment in old code.
	// w.AIs[g.SpectreID].TargetID = int(g.RunnerID)
	// In new BindLua, we pass the targetID directly to the closure, so we don't strictly need .TargetID in struct unless Lua reads it from struct.
	// But let's set it for consistency if we kept the field.
	w.AIs[g.SpectreID].TargetID = int(g.RunnerID)
	w.AIs[g.RunnerID].TargetID = int(g.SpectreID)
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
	if g.IsPaused {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.IsPaused = false
			g.Popup = nil
			g.LoadLevel(g.CurrentLevel + 1)
		}
		return nil
	}

	// Tick Systems
	lvl := &g.Levels[g.CurrentLevel] // Pass pointer to current level
	systems.SystemInput(g.World)
	systems.SystemAI(g.World, lvl)
	systems.SystemPhysics(g.World)
	systems.SystemEntropy(g.World, g.FrostMask)
	
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
	// Draw Nebula (Shader)
	w, h := screen.Size()
	t := float32(time.Since(g.StartTime).Seconds())
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]interface{}{"Cursor": []float32{0, 0, t}}
	screen.DrawRectShader(w, h, g.NebulaShader, op)

	// Draw Particles (Background layer for depth)
	g.World.Particles.Draw(screen)

	// Draw Gravity Wells (Black Holes)
	lvl := &g.Levels[g.CurrentLevel]
	systems.DrawLevel(screen, g.World, lvl, g.World.Transforms[g.SpectreID].Position)

	// Draw Mist Layer
	g.FrostImg.WritePixels(g.FrostMask.Pix)
	mistOp := &ebiten.DrawImageOptions{}
	mistOp.GeoM.Scale(float64(core.ScreenWidth)/core.MistWidth, float64(core.ScreenHeight)/core.MistHeight)
	screen.DrawImage(g.FrostImg, mistOp)

	// Draw Entities
	systems.DrawEntities(screen, g.World)

	// 5. Draw UI Modal
	if g.IsPaused && g.Popup != nil {
		// Animation: Pop in
		dt := float64(time.Since(g.PopupTime).Seconds())
		scale := dt * 5.0
		if scale > 1.0 {
			scale = 1.0
		}
		// Elastic bounce
		scale = scale * (1.0 + 0.3*(1.0-scale))

		bx, by := 300.0, 200.0
		bw, bh := 680.0, 320.0
		
		// Apply scale from center
		cx, cy := bx+bw/2, by+bh/2
		bw *= scale
		bh *= scale
		bx = cx - bw/2
		by = cy - bh/2

		// Modal Background
		vector.DrawFilledRect(screen, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{10, 0, 10, 240}, false)

		// Border
		vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 4, color.RGBA{180, 20, 40, 255}, false)

		if scale > 0.9 {
			// Photo Placeholder (Static Noise)
			photoW, photoH := 300.0, 200.0
			px, py := bx+(bw-photoW)/2, by+40
			for i := 0; i < 100; i++ {
				rx := rand.Float64() * photoW
				ry := rand.Float64() * photoH
				rw := rand.Float64() * 20
				c := uint8(rand.Intn(255))
				vector.DrawFilledRect(screen, float32(px+rx), float32(py+ry), float32(rw), 2, color.RGBA{c, c, c, 255}, false)
			}
			vector.StrokeRect(screen, float32(px), float32(py), float32(photoW), float32(photoH), 2, color.RGBA{100, 100, 100, 255}, false)

			// Text
			ebitenutil.DebugPrintAt(screen, "[ MEMORY CORRUPTED ]", int(px)+20, int(py)+int(photoH)/2)
			ebitenutil.DebugPrintAt(screen, g.Popup.Title, int(bx)+30, int(by)+260)
			ebitenutil.DebugPrintAt(screen, g.Popup.Description, int(bx)+30, int(by)+290)
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