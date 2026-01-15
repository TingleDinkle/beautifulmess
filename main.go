package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	lua "github.com/yuin/gopher-lua"
)

// ======================================================================================
// THE VIBE (Shaders)
// ======================================================================================

// This shader generates the cosmic background. 
// It's cheaper to calculate a nebula on the GPU than to load a 4K PNG.
var shaderNebula = []byte(`
package main

var Cursor vec3 // .z is Time

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	pos := position.xy / imageDstTextureSize()
	t := Cursor.z * 0.2 // Slow time down, space is big. 
	
	// Fractional Brownian Motion (FBM) for cloud fluffiness
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
	
	// Vignette: Darken the corners to focus attention
	dist := distance(pos, vec2(0.5))
	finalColor *= (1.0 - dist*0.8)
	
	return vec4(finalColor, 1.0)
}
`)

// ======================================================================================
// THE CONFIG
// ======================================================================================

const (
	ScreenWidth   = 1280
	ScreenHeight  = 720
	EntitySize    = 8.0
	MemoryRadius  = 70.0 // The size of the "Trap" / "Goal" 
	MistWidth     = 320  // Low-res buffer for that retro feel
	MistHeight    = 180
)

// ======================================================================================
// THE ARCHITECTURE (ECS)
// We treat entities as bags of data. Systems provide the behavior.
// ======================================================================================

// Components (Data)
type Vector2 struct { X, Y float64 }
type Transform struct { Position Vector2; Rotation float64 }
type Physics struct { Velocity, Acceleration Vector2; MaxSpeed, Friction, Mass float64 }
type Render struct { Sprite *ebiten.Image; Color color.RGBA; Glow bool }
type AI struct { LState *lua.LState; ScriptPath string; TargetID int }
type Tag struct { Name string }

// Entity (ID)
type Entity int

// World (The Database)
// Stores all components in contiguous memory maps for cache friendliness.
type World struct {
	Transforms map[Entity]*Transform
	Physics    map[Entity]*Physics
	Renders    map[Entity]*Render
	AIs        map[Entity]*AI
	Tags       map[Entity]*Tag
	nextID     Entity
}

func NewWorld() *World {
	return &World{
		Transforms: make(map[Entity]*Transform),
		Physics:    make(map[Entity]*Physics),
		Renders:    make(map[Entity]*Render),
		AIs:        make(map[Entity]*AI),
		Tags:       make(map[Entity]*Tag),
	}
}

func (w *World) CreateEntity() Entity {
	id := w.nextID
	w.nextID++
	return id
}

// ======================================================================================
// GAME STATE
// ======================================================================================

type GravityWell struct { Position Vector2; Radius float64; Mass float64 }
type MemoryNode struct { Position Vector2; Title, Description string; Color color.RGBA }
type Level struct { Name string; Wells []GravityWell; Memory MemoryNode; StartP1, StartP2 Vector2 }

type Game struct {
	World        *World
	RunnerID     Entity
	SpectreID    Entity
	
	Levels       []Level
	CurrentLevel int
	IsPaused     bool
	Popup        *MemoryNode
	
	// Graphics Assets
	FrostMask    *image.RGBA
	FrostImg     *ebiten.Image
	NebulaShader *ebiten.Shader
	SpriteSpectre *ebiten.Image
	SpriteRunner  *ebiten.Image
	
	StartTime    time.Time
}

// ======================================================================================
// SYSTEMS (The Brains)
// ======================================================================================

// SystemInput: Reads keyboard state and shoves the Runner around.
func SystemInput(w *World, e Entity) {
	phys := w.Physics[e]
	if phys == nil { return }
	
	// "Shift" to burn fuel/glitch
	accel := 1.5
	if ebiten.IsKeyPressed(ebiten.KeyShift) { accel = 3.5 }
	
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft)  || ebiten.IsKeyPressed(ebiten.KeyA) { phys.Acceleration.X -= accel }
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) { phys.Acceleration.X += accel }
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp)    || ebiten.IsKeyPressed(ebiten.KeyW) { phys.Acceleration.Y -= accel }
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown)  || ebiten.IsKeyPressed(ebiten.KeyS) { phys.Acceleration.Y += accel }
	
	// Rotate sprite to face movement
	trans := w.Transforms[e]
	if phys.Velocity.X != 0 || phys.Velocity.Y != 0 {
		trans.Rotation = math.Atan2(phys.Velocity.Y, phys.Velocity.X)
	}
}

// SystemPhysics: The heavy lifter. Integrates velocity, handles gravity, and wraps the world.
func SystemPhysics(w *World, g *Game) {
	lvl := g.Levels[g.CurrentLevel]
	
	for id, phys := range w.Physics {
		trans := w.Transforms[id]
		if trans == nil { continue }

		// 1. Gravity Logic
		for _, well := range lvl.Wells {
			dx := well.Position.X - trans.Position.X
			dy := well.Position.Y - trans.Position.Y
			
			// Toroidal Distance: Gravity wraps around the screen too!
			if dx > ScreenWidth/2  { dx -= ScreenWidth }
			if dx < -ScreenWidth/2 { dx += ScreenWidth }
			if dy > ScreenHeight/2 { dy -= ScreenHeight }
			if dy < -ScreenHeight/2 { dy += ScreenHeight }

			d := math.Sqrt(dx*dx + dy*dy)
			if d < 10 { d = 10 } // Prevent singularities (div by zero)
			
			// Gravity falls off with distance squared
			force := (well.Mass * 500) / (d*d)
			if force > 2.0 { force = 2.0 } // Cap force to prevent glitchy teleporting
			
			// THE TRAP: If it's the Spectre and she's inside the event horizon...
			if w.Tags[id].Name == "spectre" && d < well.Radius {
				force *= 3.5 // Crush her
				// Note: We used to apply drag here, but keeping momentum creates a cool "Orbit" effect
			}
			
			phys.Acceleration.X += (dx/d) * force
			phys.Acceleration.Y += (dy/d) * force
		}

		// 2. Integration (Velocity Verlet lite)
		phys.Velocity.X += phys.Acceleration.X
		phys.Velocity.Y += phys.Acceleration.Y
		
		// 3. Friction (The "Soup" factor)
		phys.Velocity.X *= phys.Friction
		phys.Velocity.Y *= phys.Friction
		
		// 4. Speed Limit (Safety)
		speed := math.Sqrt(phys.Velocity.X*phys.Velocity.X + phys.Velocity.Y*phys.Velocity.Y)
		if speed > phys.MaxSpeed {
			scale := phys.MaxSpeed / speed
			phys.Velocity.X *= scale
			phys.Velocity.Y *= scale
		}
		
		// 5. Update Position
		trans.Position.X += phys.Velocity.X
		trans.Position.Y += phys.Velocity.Y
		
		// 6. Toroidal Wrap (The Infinite Loop)
		if trans.Position.X < 0 { trans.Position.X += ScreenWidth }
		if trans.Position.X >= ScreenWidth { trans.Position.X -= ScreenWidth }
		if trans.Position.Y < 0 { trans.Position.Y += ScreenHeight }
		if trans.Position.Y >= ScreenHeight { trans.Position.Y -= ScreenHeight }
		
		// Reset for next frame
		phys.Acceleration.X = 0
		phys.Acceleration.Y = 0
	}
}

// SystemAI: Bridges the Go world to the Lua brain.
func SystemAI(w *World, e Entity, g *Game) {
	ai := w.AIs[e]
	if ai == nil || ai.LState == nil { return }
	L := ai.LState
	
	// Feed the AI the current world state
	lvl := g.Levels[g.CurrentLevel]
	L.SetGlobal("mem_x", lua.LNumber(lvl.Memory.Position.X))
	L.SetGlobal("mem_y", lua.LNumber(lvl.Memory.Position.Y))
	L.SetGlobal("mem_radius", lua.LNumber(MemoryRadius))
	
	// Help AI find the nearest gravity well
	bestDist := 99999.0
	var bestWell GravityWell
	pos := w.Transforms[e].Position
	for _, well := range lvl.Wells {
		// Use wrapped distance logic so AI doesn't get confused by edges
		dx := well.Position.X - pos.X
		dy := well.Position.Y - pos.Y
		if dx > ScreenWidth/2  { dx -= ScreenWidth }
		if dx < -ScreenWidth/2 { dx += ScreenWidth }
		if dy > ScreenHeight/2 { dy -= ScreenHeight }
		if dy < -ScreenHeight/2 { dy += ScreenHeight }
		
		d := math.Sqrt(dx*dx + dy*dy)
		if d < bestDist { bestDist = d; bestWell = well }
	}
	L.SetGlobal("well_x", lua.LNumber(bestWell.Position.X))
	L.SetGlobal("well_y", lua.LNumber(bestWell.Position.Y))

	// Let Lua take the wheel
	if err := L.CallByParam(lua.P{Fn: L.GetGlobal("update_state"), NRet: 0, Protect: true}); err != nil {
		fmt.Printf("Lua Error (Brain Freeze): %v\n", err)
	}
}

// SystemEntropy: Handles the fading trails.
func SystemEntropy(g *Game) {
	// Heal the world (fade trails)
	pix := g.FrostMask.Pix
	for i := 3; i < len(pix); i += 4 {
		a := int(pix[i])
		if a < 240 { pix[i] = uint8(a + 2) } // +2 Alpha per frame
	}
	
	// Create new scars (trails)
	for id, trans := range g.World.Transforms {
		if g.World.Renders[id] == nil { continue }
		
		// Scale world pos to mist resolution
		sx := int(trans.Position.X * (float64(MistWidth) / ScreenWidth))
		sy := int(trans.Position.Y * (float64(MistHeight) / ScreenHeight))
		
		rad := 2
		if g.World.Tags[id].Name == "spectre" { rad = 3 } // She leaves a bigger mark
		
		meltRetro(g.FrostMask, sx, sy, rad, 100)
	}
}

// ======================================================================================
// BOOTSTRAP
// ======================================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	
	s, err := ebiten.NewShader(shaderNebula)
	if err != nil { log.Fatal(err) }

	g := &Game{
		World:        NewWorld(),
		Levels:       initLevels(),
		FrostMask:    image.NewRGBA(image.Rect(0, 0, MistWidth, MistHeight)),
		NebulaShader: s,
		StartTime:    time.Now(),
		SpriteSpectre: generateGothicSprite(),
		SpriteRunner:  generateCyberSprite(),
	}
	
	// Pre-fill the frost mask
	c := color.RGBA{10, 5, 20, 240}
	for i := 0; i < len(g.FrostMask.Pix); i+=4 {
		g.FrostMask.Pix[i], g.FrostMask.Pix[i+1], g.FrostMask.Pix[i+2], g.FrostMask.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	g.FrostImg = ebiten.NewImageFromImage(g.FrostMask)

	g.LoadLevel(0)
	return g
}

func (g *Game) LoadLevel(idx int) {
	if idx >= len(g.Levels) { idx = len(g.Levels)-1 }
	g.CurrentLevel = idx
	lvl := g.Levels[idx]
	w := g.World
	w.Physics = make(map[Entity]*Physics) // Reset physics state
	
	// Spawn Spectre
	g.SpectreID = w.CreateEntity()
	w.Tags[g.SpectreID] = &Tag{Name: "spectre"}
	w.Transforms[g.SpectreID] = &Transform{Position: lvl.StartP2}
	w.Physics[g.SpectreID] = &Physics{MaxSpeed: 6.0, Friction: 0.96, Mass: 1.0}
	w.Renders[g.SpectreID] = &Render{Sprite: g.SpriteSpectre, Color: color.RGBA{255,50,50,255}, Glow: true}
	w.AIs[g.SpectreID] = &AI{ScriptPath: "spectre.lua"}
	bindLua(g, g.SpectreID)

	// Spawn Runner
	g.RunnerID = w.CreateEntity()
	w.Tags[g.RunnerID] = &Tag{Name: "runner"}
	w.Transforms[g.RunnerID] = &Transform{Position: lvl.StartP1}
	w.Physics[g.RunnerID] = &Physics{MaxSpeed: 7.5, Friction: 0.92, Mass: 1.0}
	w.Renders[g.RunnerID] = &Render{Sprite: g.SpriteRunner, Color: color.RGBA{0,255,255,255}, Glow: true}
	w.AIs[g.RunnerID] = &AI{ScriptPath: "runner.lua"}
	bindLua(g, g.RunnerID)
	
	// Link them (so they can find each other)
	w.AIs[g.SpectreID].TargetID = int(g.RunnerID)
}

// bindLua connects Go functions to Lua scripts
func bindLua(g *Game, e Entity) {
	L := lua.NewState()
	g.World.AIs[e].LState = L
	
	// Physics Hooks
	L.SetGlobal("apply_force", L.NewFunction(func(L *lua.LState) int {
		g.World.Physics[e].Acceleration.X += float64(L.CheckNumber(1))
		g.World.Physics[e].Acceleration.Y += float64(L.CheckNumber(2))
		return 0
	}))
	
	L.SetGlobal("set_max_speed", L.NewFunction(func(L *lua.LState) int {
		g.World.Physics[e].MaxSpeed = float64(L.CheckNumber(1))
		return 0
	}))
	
	// Sensor Hooks
	L.SetGlobal("get_self", L.NewFunction(func(L *lua.LState) int {
		p := g.World.Transforms[e].Position
		v := g.World.Physics[e].Velocity
		L.Push(lua.LNumber(p.X)); L.Push(lua.LNumber(p.Y))
		L.Push(lua.LNumber(v.X)); L.Push(lua.LNumber(v.Y))
		return 4
	}))
	
	L.SetGlobal("get_vec_to", L.NewFunction(func(L *lua.LState) int {
		tx, ty := float64(L.CheckNumber(1)), float64(L.CheckNumber(2))
		mx, my := g.World.Transforms[e].Position.X, g.World.Transforms[e].Position.Y
		dx, dy := tx-mx, ty-my
		if dx > ScreenWidth/2  { dx -= ScreenWidth }
		if dx < -ScreenWidth/2 { dx += ScreenWidth }
		if dy > ScreenHeight/2 { dy -= ScreenHeight }
		if dy < -ScreenHeight/2 { dy += ScreenHeight }
		
		d := math.Sqrt(dx*dx + dy*dy)
		if d < 0.01 { d = 0.01 }
		L.Push(lua.LNumber(dx/d)); L.Push(lua.LNumber(dy/d)); L.Push(lua.LNumber(d))
		return 3
	}))
	
	L.SetGlobal("get_target", L.NewFunction(func(L *lua.LState) int {
		tid := g.RunnerID
		if e == g.RunnerID { tid = g.SpectreID }
		p := g.World.Transforms[tid].Position
		L.Push(lua.LNumber(p.X)); L.Push(lua.LNumber(p.Y))
		return 2
	}))
	
	L.SetGlobal("get_input_dir", L.NewFunction(func(L *lua.LState) int {
		dx, dy := 0.0, 0.0
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft)  || ebiten.IsKeyPressed(ebiten.KeyA) { dx = -1 }
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) { dx = 1 }
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp)    || ebiten.IsKeyPressed(ebiten.KeyW) { dy = -1 }
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown)  || ebiten.IsKeyPressed(ebiten.KeyS) { dy = 1 }
		L.Push(lua.LNumber(dx)); L.Push(lua.LNumber(dy))
		return 2
	}))
	
	L.SetGlobal("cast_ray", L.NewFunction(func(L *lua.LState) int {
		// In the void, raycasts see infinity.
		dist := float64(L.CheckNumber(2))
		L.Push(lua.LNumber(dist))
		return 1
	}))
	
	if err := L.DoFile(g.World.AIs[e].ScriptPath); err != nil {
		log.Printf("Lua Script Error: %v", err)
	}
}

// Procedural Art Generators
func generateGothicSprite() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	cRed := color.RGBA{180, 20, 40, 255}
	cDark := color.RGBA{80, 10, 20, 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			dx, dy := x - 8, y - 8
			dist := dx*dx + dy*dy * 2
			if dist < 40 { img.Set(x, y, cRed) } else if dist < 50 && rand.Float64() < 0.5 { img.Set(x, y, cDark) }
		}
	}
	img.Set(6, 6, color.White); img.Set(9, 6, color.White)
	return ebiten.NewImageFromImage(img)
}

func generateCyberSprite() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	cCyan := color.RGBA{0, 255, 255, 255}
	cBlue := color.RGBA{0, 100, 200, 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if x > y && x > (15-y) { img.Set(x, y, cCyan) }
			if y == 8 || x == 8 { img.Set(x, y, cBlue) }
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
			g.LoadLevel(g.CurrentLevel+1) 
		}
		return nil
	}
	
	// Tick Systems
	SystemInput(g.World, g.RunnerID)
	SystemAI(g.World, g.RunnerID, g)
	SystemAI(g.World, g.SpectreID, g)
	SystemPhysics(g.World, g)
	SystemEntropy(g)
	
	// Trap / End Level Logic
	lvl := g.Levels[g.CurrentLevel]
	pSpec := g.World.Transforms[g.SpectreID].Position
	pRun := g.World.Transforms[g.RunnerID].Position
	
	dSpecMem := distWrapped(pSpec, lvl.Memory.Position)
	dRunSpec := distWrapped(pSpec, pRun)
	
	// Condition: She is trapped (close to memory) AND you are close to her
	if dSpecMem < MemoryRadius && dRunSpec < 80 {
		g.IsPaused = true
		g.Popup = &lvl.Memory
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 1. Draw Nebula (Shader)
	w, h := screen.Size()
	t := float32(time.Since(g.StartTime).Seconds())
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]interface{}{ "Cursor": []float32{0, 0, t} }
	screen.DrawRectShader(w, h, g.NebulaShader, op)
	
	// 2. Draw Gravity Wells (Black Holes)
	lvl := g.Levels[g.CurrentLevel]
	for _, well := range lvl.Wells {
		// Event Horizon (Black)
		drawWrappedCircle(screen, well.Position, well.Radius, color.RGBA{0,0,0,255}, true)
		// Accretion Disk (Purple)
		drawWrappedCircle(screen, well.Position, well.Radius, color.RGBA{100,0,100,255}, false)
	}

	// 3. Draw Memory Node
	// It glows Red if the Spectre is currently trapped inside it
	mc := color.RGBA{200, 200, 255, 100}
	if distWrapped(g.World.Transforms[g.SpectreID].Position, lvl.Memory.Position) < MemoryRadius {
		mc = color.RGBA{255, 50, 50, 255} 
	}
	drawWrappedCircle(screen, lvl.Memory.Position, MemoryRadius, mc, false)

	// 4. Draw Mist Layer
	g.FrostImg.WritePixels(g.FrostMask.Pix)
	mistOp := &ebiten.DrawImageOptions{}
	mistOp.GeoM.Scale(float64(ScreenWidth)/MistWidth, float64(ScreenHeight)/MistHeight)
	screen.DrawImage(g.FrostImg, mistOp)

	// 5. Draw Entities
	for id, r := range g.World.Renders {
		if r.Sprite == nil { continue }
		trans := g.World.Transforms[id]
		drawWrappedSprite(screen, r.Sprite, trans.Position, trans.Rotation)
	}

	// 6. Draw UI Modal
	if g.IsPaused && g.Popup != nil {
		bx, by := 300.0, 200.0
		bw, bh := 680.0, 320.0
		
		// Modal Background
		vector.DrawFilledRect(screen, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{10,0,10,240}, false)
		
		// Border
		vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 4, color.RGBA{180,20,40,255}, false)
		
		// Photo Placeholder (Static Noise)
		photoW, photoH := 300.0, 200.0
		px, py := bx + (bw-photoW)/2, by + 40
		for i := 0; i < 100; i++ {
			rx := rand.Float64() * photoW
			ry := rand.Float64() * photoH
			rw := rand.Float64() * 20
			c := uint8(rand.Intn(255))
			vector.DrawFilledRect(screen, float32(px+rx), float32(py+ry), float32(rw), 2, color.RGBA{c,c,c,255}, false)
		}
		vector.StrokeRect(screen, float32(px), float32(py), float32(photoW), float32(photoH), 2, color.RGBA{100,100,100,255}, false)
		
		// Text
		ebitenutil.DebugPrintAt(screen, "[ MEMORY CORRUPTED ]", int(px)+20, int(py)+int(photoH)/2)
		ebitenutil.DebugPrintAt(screen, g.Popup.Title, int(bx)+30, int(by)+260)
		ebitenutil.DebugPrintAt(screen, g.Popup.Description, int(bx)+30, int(by)+290)
	}
}

// ======================================================================================
// UTILS
// ======================================================================================

func distWrapped(a, b Vector2) float64 {
	dx := math.Abs(a.X - b.X); dy := math.Abs(a.Y - b.Y)
	if dx > ScreenWidth/2 { dx = ScreenWidth - dx }
	if dy > ScreenHeight/2 { dy = ScreenHeight - dy }
	return math.Sqrt(dx*dx + dy*dy)
}

func drawWrappedCircle(screen *ebiten.Image, pos Vector2, r float64, c color.RGBA, fill bool) {
	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			x := float32(pos.X + ox*ScreenWidth)
			y := float32(pos.Y + oy*ScreenHeight)
			
			// Optimization: Check visibility
			if x+float32(r) < 0 || x-float32(r) > float32(ScreenWidth) || y+float32(r) < 0 || y-float32(r) > float32(ScreenHeight) { continue }
			
			if fill { vector.DrawFilledCircle(screen, x, y, float32(r), c, true) } else { vector.StrokeCircle(screen, x, y, float32(r), 2, c, true) }
		}
	}
}

func drawWrappedSprite(screen *ebiten.Image, img *ebiten.Image, pos Vector2, rot float64) {
	w, h := img.Size()
	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			x := pos.X + ox*ScreenWidth
			y := pos.Y + oy*ScreenHeight
			
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			op.GeoM.Rotate(rot)
			op.GeoM.Translate(x, y)
			screen.DrawImage(img, op)
		}
	}
}

func meltRetro(img *image.RGBA, cx, cy, r, val int) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	offsets := [][2]int{{0,0}, {w,0}, {-w,0}, {0,h}, {0,-h}}
	
	for _, off := range offsets {
		tcx, tcy := cx + off[0], cy + off[1]
		if tcx+r < 0 || tcx-r >= w || tcy+r < 0 || tcy-r >= h { continue }
		
		for y := tcy - r; y <= tcy + r; y++ {
			for x := tcx - r; x <= tcx + r; x++ {
				if x < 0 || x >= w || y < 0 || y >= h { continue }
				if (x-tcx)*(x-tcx)+(y-tcy)*(y-tcy) <= r*r {
					i := img.PixOffset(x, y) + 3
					old := int(img.Pix[i])
					if old > 0 { img.Pix[i] = uint8(math.Max(0, float64(old-val))) }
				}
			}
		}
	}
}

func initLevels() []Level {
	return []Level{
		// Level 1: Singularity
		// Note: The Memory is INSIDE the Black Hole (640, 360)
		{"Event Horizon", 
		 []GravityWell{{Vector2{640,360}, 70, 2.0}},
		 MemoryNode{Vector2{640,360}, "The Singularity", "We were crushed together.\nFinally one.", color.RGBA{100,100,255,255}},
		 Vector2{100,360}, Vector2{1100,360}},
		
		// Level 2: Binary
		{"Binary Star", 
		 []GravityWell{{Vector2{300,360}, 80, 2.0}, {Vector2{980,360}, 80, 2.0}},
		 MemoryNode{Vector2{300,360}, "Orbit Decay", "Spinning until we crash.", color.RGBA{255,50,50,255}},
		 Vector2{640,600}, Vector2{640,100}},
	}
}

func (g *Game) Layout(w, h int) (int, int) { return ScreenWidth, ScreenHeight }

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Beautiful Mess: The Final Code")
	if err := ebiten.RunGame(NewGame()); err != nil { panic(err) }
}
