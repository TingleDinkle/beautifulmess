package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"time"
	"fmt"
	"strings"

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

var shaderNebula = []byte(`
package main
var Cursor vec3
func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	pos := position.xy / imageDstTextureSize()
	t := Cursor.z * 0.2
	val := 0.0
	scale := 3.0
	for i := 0; i < 3; i++ {
		val += sin(pos.x*scale + t) + sin(pos.y*scale - t*0.5)
		scale *= 2.0
	}
	val /= 6.0
	base := vec3(0.1, 0.0, 0.2)
	light := vec3(0.0, 0.4, 0.5)
	finalColor := base + light * (val + 0.5)
	dist := distance(pos, vec2(0.5))
	finalColor *= (1.0 - dist*0.8)
	return vec4(finalColor, 1.0)
}
`)

type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StatePaused
	StateTransitioning
)

type Game struct {
	World     *world.World
	State     GameState
	EasyMode  bool
	MasterVolume float64
	MenuIndex    int
	RunnerID  core.Entity
	SpectreID core.Entity
	Levels       []level.Level
	CurrentLevel int
	Popup        *level.MemoryNode
	PopupTime    time.Time
	PopupPhotoIndex int
	PopupAutoMode   bool
	PopupWaitTimer  float64
	TransitionTime  float64
	TargetLevel     int
	ReunionPoint    core.Vector2
	FrostMask     *image.RGBA
	FrostImg      *ebiten.Image
	NebulaShader  *ebiten.Shader
	ShaderOptions ebiten.DrawRectShaderOptions 
	PopupRNG      *rand.Rand
	PhotoCache    map[string]*ebiten.Image
	SpectreSprites map[string]*ebiten.Image
	SpectreState   systems.SpectreVisualState
	SpriteRunner  *ebiten.Image
	StartTime time.Time
	HitStop     float64
	TitleTimer  float64
	StartAnimation float64
	TypewriterChars int
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	s, err := ebiten.NewShader(shaderNebula)
	if err != nil { log.Fatal(err) }
	g := &Game{
		World:         world.NewWorld(),
		State:         StateTitle,
		MasterVolume:  0.5,
		Levels:        level.InitLevels(),
		FrostMask:     image.NewRGBA(image.Rect(0, 0, core.MistWidth, core.MistHeight)),
		NebulaShader:  s,
		ShaderOptions: ebiten.DrawRectShaderOptions{Uniforms: make(map[string]interface{})},
		PopupRNG:      rand.New(rand.NewSource(0)),
		PhotoCache:    make(map[string]*ebiten.Image),
		StartTime:     time.Now(),
		SpectreState:  systems.SpectreVisualState{State: "normal"},
	}
	g.SpectreSprites = systems.LoadSpectreSet("assets/normal.png", "assets/angy.png", "assets/kewt.png")
	g.SpriteRunner = generateAstroSprite()
	g.World.Audio.LoadFile("shoot", "assets/shoot.wav")
	g.World.Audio.LoadFile("boom", "assets/boom.wav")
	g.World.Audio.SetVolume(g.MasterVolume)
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
	g.World.Reset()
	g.World.Particles.Reset()
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
	for _, wall := range lvl.Walls { spawnWall(w, wall.X, wall.Y, wall.Destructible) }
	
	g.SpectreID = w.CreateEntity()
	w.Tags[g.SpectreID] = &components.Tag{Name: "spectre"}
	w.Transforms[g.SpectreID] = &components.Transform{Position: lvl.StartP2}
	w.Physics[g.SpectreID] = &components.Physics{MaxSpeed: 6.0, Friction: 0.96, Mass: 1.0, GravityMultiplier: 3.5}
	
	// Dynamic scaling to maintain photo integrity while fitting the world
	specW, _ := g.SpectreSprites["normal"].Size()
	sScale := 80.0 / float64(specW)
	if sScale > 1.5 { sScale = 1.5 }
	
	w.Renders[g.SpectreID] = &components.Render{Sprite: g.SpectreSprites["normal"], Color: color.RGBA{255, 255, 255, 255}, Glow: true, Scale: sScale}
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
	w.AddToActiveWalls(id)
	w.Transforms[id] = &components.Transform{Position: core.Vector2{X: x, Y: y}}
	w.Walls[id] = &components.Wall{Size: 10, Destructible: destructible}
	img := ebiten.NewImage(10, 10)
	c := color.RGBA{0, 255, 255, 255}
	if destructible { c = color.RGBA{255, 150, 50, 255} }
	img.Fill(c)
	w.Renders[id] = &components.Render{Sprite: img, Color: color.RGBA{255, 255, 255, 255}, Scale: 1.0}
}

func generateGothicSprite() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	cRed, cDark := color.RGBA{180, 20, 40, 255}, color.RGBA{80, 10, 20, 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if (x-8)*(x-8) + (y-8)*(y-8)*2 < 40 { img.Set(x, y, cRed) } else if (x-8)*(x-8) + (y-8)*(y-8)*2 < 50 && rand.Float64() < 0.5 { img.Set(x, y, cDark) }
		}
	}
	img.Set(6, 6, color.White); img.Set(9, 6, color.White)
	return ebiten.NewImageFromImage(img)
}

func generateAstroSprite() *ebiten.Image {
	img := ebiten.NewImage(16, 16)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			fx, fy := float64(x), float64(y)
			if (fy >= 0.5*fx) && (fy <= -0.5*fx+16.0) && !((fx < 4) && (fy >= 2.0*fx) && (fy <= -2.0*fx+16.0)) { img.Set(x, y, color.White) }
		}
	}
	return img
}

func (g *Game) Update() error {
	g.handleInput()

	switch g.State {
	case StateTitle:
		return g.updateTitleState()
	case StatePlaying:
		return g.updateActiveState()
	case StatePaused:
		return g.updatePausedState()
	case StateTransitioning:
		return g.updateTransitionState()
	}
	return nil
}

func (g *Game) updateTitleState() error {
	g.TitleTimer += 1.0 / 60.0

	// Typewriter audio feedback logic
	const revealSpeed = 22.0
	text1 := "Dear Stella, I made this game for you."
	text2 := "Enjoy, our love."

	chars1 := int(g.TitleTimer * revealSpeed)
	if chars1 > len(text1) {
		chars1 = len(text1)
	}
	chars2 := 0
	if g.TitleTimer > float64(len(text1))/revealSpeed+0.8 {
		chars2 = int((g.TitleTimer - (float64(len(text1))/revealSpeed + 0.8)) * revealSpeed)
		if chars2 > len(text2) {
			chars2 = len(text2)
		}
	}

	if chars1+chars2 > g.TypewriterChars {
		g.World.Audio.Play("tick")
		g.TypewriterChars = chars1 + chars2
	}

	// Menu Navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.MenuIndex = (g.MenuIndex - 1 + 5) % 5
		g.World.Audio.Play("blip")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.MenuIndex = (g.MenuIndex + 1) % 5
		g.World.Audio.Play("blip")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		if g.MenuIndex == 2 { // Volume adjustment
			g.MasterVolume = math.Max(0, g.MasterVolume-0.05)
			g.World.Audio.SetVolume(g.MasterVolume)
			g.World.Audio.Play("blip")
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		if g.MenuIndex == 2 { // Volume adjustment
			g.MasterVolume = math.Min(1.0, g.MasterVolume+0.05)
			g.World.Audio.SetVolume(g.MasterVolume)
			g.World.Audio.Play("blip")
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		switch g.MenuIndex {
		case 0: // Start
			g.State = StatePlaying
			g.LoadLevel(0)
			g.triggerSpitOut()
		case 1: // Difficulty Mode
			g.EasyMode = !g.EasyMode
			g.World.Audio.Play("blip")
		case 2: // Volume indicator
			g.World.Audio.Play("chime")
		case 3: // Display toggle
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
			g.World.Audio.Play("blip")
		case 4: // Exit
			return ebiten.Termination
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	return nil
}

func (g *Game) triggerSpitOut() {
	lvl := g.Levels[g.CurrentLevel]
	wellPos := core.Vector2{X: 640, Y: 360}
	if len(lvl.Wells) > 0 {
		wellPos = lvl.Wells[0].Position
	}

	// Reset positions to well center with slight offset to prevent immediate collision/win
	if t := g.World.Transforms[g.RunnerID]; t != nil {
		t.Position = core.Vector2{X: wellPos.X - 5, Y: wellPos.Y}
	}
	if t := g.World.Transforms[g.SpectreID]; t != nil {
		t.Position = core.Vector2{X: wellPos.X + 5, Y: wellPos.Y}
	}

	// Explosive outward velocity - bypassing MaxSpeed in SystemPhysics
	if p := g.World.Physics[g.RunnerID]; p != nil {
		p.Velocity = core.Vector2{X: -35, Y: -15}
	}
	if p := g.World.Physics[g.SpectreID]; p != nil {
		p.Velocity = core.Vector2{X: 35, Y: 15}
	}

	g.StartAnimation = 1.5
	g.World.ScreenShake = 15.0
	g.World.Audio.Play("boom")
}

func (g *Game) handleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	
	// ESC Logic
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.Popup != nil || g.State == StatePaused {
			// If we're in a popup or already paused, ESC takes us all the way back to Title
			g.State = StateTitle
			g.TitleTimer = 0
			g.TypewriterChars = 0
			g.Popup = nil
			return
		} else if g.State == StatePlaying {
			// If we're just playing, ESC pauses the game to show the menu
			g.State = StatePaused
			return
		}
	}

	if (g.State == StatePlaying || g.State == StatePaused) && g.Popup == nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			if g.State == StatePaused {
				g.State = StatePlaying
			} else {
				g.State = StatePaused
			}
		}
	}
}

func (g *Game) updatePausedState() error {
	if g.Popup != nil {
		if g.PopupAutoMode {
			// Auto-advance logic
			descIdx := g.getDescIndex()
			fullText := g.Popup.Descriptions[descIdx]
			
			// Typewriter speed
			const popupRevealSpeed = 40.0
			if g.TypewriterChars < len(fullText) {
				g.TypewriterChars++
				if g.TypewriterChars % 2 == 0 {
					g.World.Audio.Play("tick")
				}
			} else {
				// Text finished, wait then advance
				g.PopupWaitTimer += 1.0 / 60.0
				if g.PopupWaitTimer > 4.5 { 
					g.PopupWaitTimer = 0
					g.PopupPhotoIndex++
					if g.PopupPhotoIndex >= len(g.Popup.Photos) {
						g.PopupAutoMode = false 
						g.PopupPhotoIndex = len(g.Popup.Photos) - 1 
						g.TypewriterChars = 9999
					} else {
						// Only reset typewriter if the description actually changes
						newDescIdx := g.getDescIndex()
						if newDescIdx != descIdx {
							g.TypewriterChars = 0
						}
					}
				}
			}
		} else {
			// Manual control restored after sequence finishes
			if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
				g.PopupPhotoIndex = (g.PopupPhotoIndex + 1) % len(g.Popup.Photos)
				g.TypewriterChars = 9999 
				g.World.Audio.Play("blip")
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
				g.PopupPhotoIndex = (g.PopupPhotoIndex - 1 + len(g.Popup.Photos)) % len(g.Popup.Photos)
				g.TypewriterChars = 9999
				g.World.Audio.Play("blip")
			}
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
				g.State, g.Popup, g.TargetLevel = StateTransitioning, nil, g.CurrentLevel+1
				g.TransitionTime = 0
			}
		}
	} else {
		if inpututil.IsKeyJustPressed(ebiten.KeyM) {
			g.State = StateTitle
			g.TitleTimer = 0
			g.TypewriterChars = 0
		}
	}
	return nil
}

func (g *Game) getDescIndex() int {
	descIdx := g.PopupPhotoIndex
	if len(g.Popup.Photos) == 4 && len(g.Popup.Descriptions) == 3 {
		if g.PopupPhotoIndex == 1 || g.PopupPhotoIndex == 2 {
			descIdx = 1
		} else if g.PopupPhotoIndex == 3 {
			descIdx = 2
		} else {
			descIdx = 0
		}
	} else if descIdx >= len(g.Popup.Descriptions) {
		descIdx = len(g.Popup.Descriptions) - 1
	}
	return descIdx
}

func (g *Game) updateTransitionState() error {
	g.TransitionTime += 1.0 / 120.0
	if g.TransitionTime >= 1.0 {
		g.State = StatePlaying
		g.LoadLevel(g.TargetLevel)
		g.TransitionTime = 0
	}
	return nil
}

func (g *Game) updateActiveState() error {
	if g.HitStop > 0 {
		g.HitStop -= 1.0 / 60.0
		g.World.ScreenShake *= 0.9
		return nil
	}

	g.World.ScreenShake *= 0.9
	if g.World.ScreenShake < 0.5 {
		g.World.ScreenShake = 0
	}

	g.World.UpdateGrid()
	lvl := &g.Levels[g.CurrentLevel]

	if g.StartAnimation > 0 {
		g.StartAnimation -= 1.0 / 60.0
	} else {
		systems.SystemInput(g.World)
	}

	systems.SystemAI(g.World, lvl)
	systems.SystemSpectreVisuals(g.World, &g.SpectreState, g.SpectreID, g.SpectreSprites)
	systems.SystemPhysics(g.World, g.EasyMode, g.StartAnimation > 0)
	systems.SystemEntropy(g.World, g.FrostMask)
	systems.SystemProjectileEmitter(g.World)
	systems.SystemLifetime(g.World)

	g.World.Particles.Update()

	if g.StartAnimation <= 0 {
		return g.checkWinCondition(lvl)
	}
	return nil
}

func (g *Game) checkWinCondition(lvl *level.Level) error {
	pSpec, pRun := g.World.Transforms[g.SpectreID], g.World.Transforms[g.RunnerID]
	if pSpec == nil || pRun == nil { return nil }
	if core.DistWrapped(pSpec.Position, pRun.Position) < 80 {
		for id, well := range g.World.GravityWells {
			if well == nil { continue }
			wellTrans := g.World.Transforms[id]
			if wellTrans == nil { continue }
			if core.DistWrapped(pSpec.Position, wellTrans.Position) < well.Radius+15 {
				g.State, g.Popup, g.PopupTime, g.PopupPhotoIndex = StatePaused, &lvl.Memory, time.Now(), 0
				g.PopupAutoMode = true
				g.PopupWaitTimer = 0
				g.TypewriterChars = 0
				g.ReunionPoint = pSpec.Position
				g.World.Audio.Play("chime")
				return nil
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.State {
	case StateTitle:
		g.drawTitleScreen(screen)
	default:
		shake := core.Vector2{}
		if g.World.ScreenShake > 0 {
			shake.X = (rand.Float64() - 0.5) * g.World.ScreenShake * 2
			shake.Y = (rand.Float64() - 0.5) * g.World.ScreenShake * 2
		}
		g.drawWorld(screen, shake)
		if g.State == StateTransitioning { g.drawTransition(screen) }
		g.drawUI(screen)
	}
}

func (g *Game) drawWorld(screen *ebiten.Image, shake core.Vector2) {
	g.drawBackground(screen)
	g.World.Particles.Draw(screen)
	lvl := &g.Levels[g.CurrentLevel]
	spectrePos := core.Vector2{}
	if trans := g.World.Transforms[g.SpectreID]; trans != nil { spectrePos = trans.Position }
	systems.DrawLevel(screen, g.World, lvl, spectrePos, shake)
	g.drawMist(screen)
	systems.DrawEntities(screen, g.World, shake)
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	w, h := screen.Size()
	t := float32(time.Since(g.StartTime).Seconds())
	g.ShaderOptions.Uniforms["Cursor"] = []float32{0, 0, t}
	screen.DrawRectShader(w, h, g.NebulaShader, &g.ShaderOptions)
}

func (g *Game) drawMist(screen *ebiten.Image) {
	g.FrostImg.WritePixels(g.FrostMask.Pix)
	mistOp := &ebiten.DrawImageOptions{}
	mistOp.GeoM.Scale(float64(core.ScreenWidth)/core.MistWidth, float64(core.ScreenHeight)/core.MistHeight)
	screen.DrawImage(g.FrostImg, mistOp)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	switch g.State {
	case StateTitle:
		g.drawTitleScreen(screen)
	case StatePaused:
		if g.Popup != nil {
			g.drawPopup(screen)
		} else {
			g.drawPauseMenu(screen)
		}
	}
}

func (g *Game) drawTitleScreen(screen *ebiten.Image) {
	// Pure black background
	screen.Fill(color.Black)

	bw, bh := 860.0, 480.0
	bx, by := (float64(core.ScreenWidth)-bw)/2, (float64(core.ScreenHeight)-bh)/2
	
	// Colored border (Amber for warmth)
	amber := color.RGBA{255, 176, 0, 255}
	vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 1, amber, false)

	// Typewriter logic
	text1 := "Dear Stella, I made this game for you."
	text2 := "Enjoy, our love."
	
	speed := 22.0 
	charsToShow1 := int(g.TitleTimer * speed)
	if charsToShow1 > len(text1) { charsToShow1 = len(text1) }
	
	ebitenutil.DebugPrintAt(screen, text1[:charsToShow1], int(bx)+140, int(by)+80)
	
	if g.TitleTimer > float64(len(text1))/speed + 0.8 {
		charsToShow2 := int((g.TitleTimer - (float64(len(text1))/speed + 0.8)) * speed)
		if charsToShow2 > len(text2) { charsToShow2 = len(text2) }
		if charsToShow2 > 0 {
			ebitenutil.DebugPrintAt(screen, text2[:charsToShow2], int(bx)+360, int(by)+130)
		}
	}

	// Menu options
	if g.TitleTimer > 3.0 {
		options := []string{
			"START JOURNEY",
			"MODE: NORMAL",
			"VOLUME: [..........]",
			"FULLSCREEN",
			"QUIT TO DESKTOP",
		}
		if g.EasyMode { options[1] = "MODE: EASY (HOMING)" }
		
		// Update volume bar
		volDots := int(g.MasterVolume * 10)
		bar := ""
		for i := 0; i < 10; i++ {
			if i < volDots { bar += "#" } else { bar += "." }
		}
		options[2] = "VOLUME: [" + bar + "]"

		for i, opt := range options {
			y := int(by) + 240 + (i * 35)
			prefix := "  "
			if g.MenuIndex == i {
				prefix = "> "
				// Subtle blink for selection
				if math.Sin(g.TitleTimer*12) > 0 {
					ebitenutil.DebugPrintAt(screen, prefix+opt, int(bx)+320, y)
				}
			} else {
				ebitenutil.DebugPrintAt(screen, prefix+opt, int(bx)+320, y)
			}
		}
		
		ebitenutil.DebugPrintAt(screen, "(W/S) NAVIGATE  (A/D) ADJUST  (SPACE) SELECT", int(bx)+240, int(by)+430)
	}
}

func (g *Game) drawVignette(screen *ebiten.Image) {
	w, h := float32(core.ScreenWidth), float32(core.ScreenHeight)
	vector.StrokeRect(screen, 0, 0, w, h, 100, color.RGBA{0, 0, 0, 180}, false)
}

func (g *Game) drawPopup(screen *ebiten.Image) {
	dt := float64(time.Since(g.PopupTime).Seconds())
	scale := math.Min(1.0, dt*5.0)
	scale = scale * (1.0 + 0.3*(1.0-scale))
	
	// Dynamic Height Calculation
	baseHeight := 500.0
	extraHeight := 0.0
	if g.Popup != nil {
		descIdx := g.getDescIndex()
		if descIdx >= 0 && descIdx < len(g.Popup.Descriptions) {
			lines := splitLines(g.Popup.Descriptions[descIdx], 85)
			textHeight := float64(len(lines) * 14)
			if textHeight > 100 {
				extraHeight = textHeight - 80 
			}
		}
	}
	
	bx, by, bw, bh := 300.0, 100.0, 680.0, baseHeight + extraHeight
	cx, cy := bx+bw/2, by+bh/2
	bw, bh = bw*scale, bh*scale
	bx, by = cx-bw/2, cy-bh/2
	
	vector.DrawFilledRect(screen, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{10, 0, 10, 245}, false)
	vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 4, color.RGBA{180, 20, 40, 255}, false)
	if scale > 0.9 { g.renderPopupContent(screen, bx, by, bw, bh) }
}

func (g *Game) renderPopupContent(screen *ebiten.Image, bx, by, bw, bh float64) {
	// 1. Top Navigation Bar
	ebitenutil.DebugPrintAt(screen, "[ESC] QUIT TO TITLE", int(bx)+20, int(by)+15)
	ebitenutil.DebugPrintAt(screen, "--- MEMORY FRAGMENT ---", int(bx)+255, int(by)+15)

	// 2. Metadata Section
	ebitenutil.DebugPrintAt(screen, "SUBJECT: "+g.Popup.Title, int(bx)+30, int(by)+45)
	if len(g.Popup.Photos) > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("PHOTO RECORD: %d / %d", g.PopupPhotoIndex+1, len(g.Popup.Photos)), int(bx)+460, int(by)+45)
	}

	photoW, photoH := 400.0, 275.0
	px, py := bx+(bw-photoW)/2, by+85

	// Photo rendering
	if g.PopupPhotoIndex < len(g.Popup.Photos) {
		path := g.Popup.Photos[g.PopupPhotoIndex]
		img, ok := g.PhotoCache[path]
		if !ok {
			var err error
			img, _, err = ebitenutil.NewImageFromFile(path)
			if err != nil {
				log.Printf("failed to load photo %s: %v", path, err)
			} else {
				g.PhotoCache[path] = img
			}
		}

		if img != nil {
			iw, ih := img.Size()
			sw := photoW / float64(iw)
			sh := photoH / float64(ih)
			s := math.Min(sw, sh)
			
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(s, s)
			op.GeoM.Translate(px+(photoW-float64(iw)*s)/2, py+(photoH-float64(ih)*s)/2)
			screen.DrawImage(img, op)
		}
	}

	vector.StrokeRect(screen, float32(px), float32(py), float32(photoW), float32(photoH), 2, color.RGBA{100, 100, 100, 255}, false)

	// 3. Side Navigation
	if len(g.Popup.Photos) > 1 && !g.PopupAutoMode {
		ebitenutil.DebugPrintAt(screen, "< [A] PREV", int(px)-90, int(py)+int(photoH)/2)
		ebitenutil.DebugPrintAt(screen, "[D] NEXT >", int(px)+int(photoW)+10, int(py)+int(photoH)/2)
	}

	// 4. Description Section
	descIdx := g.getDescIndex()
	if descIdx >= 0 && descIdx < len(g.Popup.Descriptions) {
		fullText := g.Popup.Descriptions[descIdx]
		revealLen := g.TypewriterChars
		if revealLen > len(fullText) { revealLen = len(fullText) }
		
		displayedText := fullText[:revealLen]
		lines := splitLines(displayedText, 85) 
		for i, line := range lines {
			ebitenutil.DebugPrintAt(screen, line, int(bx)+30, int(by)+375+i*14) 
		}
	}

	// 5. Footer Prompts
	if !g.PopupAutoMode {
		ebitenutil.DebugPrintAt(screen, "[ SPACE ] RECOVER FRAGMENT", int(bx)+245, int(by)+int(bh)-25)
	} else {
		ebitenutil.DebugPrintAt(screen, "( STABILIZING MEMORY DATA... )", int(bx)+235, int(by)+int(bh)-25)
	}
}

func splitLines(s string, limit int) []string {
	var lines []string
	words := strings.Fields(s)
	if len(words) == 0 { return nil }
	current := words[0]
	for _, w := range words[1:] {
		if len(current)+1+len(w) > limit {
			lines = append(lines, current)
			current = w
		} else {
			current += " " + w
		}
	}
	lines = append(lines, current)
	return lines
}

func (g *Game) drawPauseMenu(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, float32(core.ScreenWidth), float32(core.ScreenHeight), color.RGBA{0, 0, 0, 180}, false)
	bw, bh := 450.0, 160.0
	bx, by := (float64(core.ScreenWidth)-bw)/2, (float64(core.ScreenHeight)-bh)/2
	vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 2, color.RGBA{33, 33, 255, 255}, false)
	ebitenutil.DebugPrintAt(screen, "--- PAUSED ---", int(bx)+160, int(by)+30)
	ebitenutil.DebugPrintAt(screen, "RESUME: PRESS P", int(bx)+160, int(by)+60)
	ebitenutil.DebugPrintAt(screen, "RETURN TO MENU: PRESS M or ESC", int(bx)+100, int(by)+90)
	for px := 0; px < 5; px++ {
		vector.DrawFilledCircle(screen, float32(bx)+float32(px*90)+45, float32(by)+15, 2, color.RGBA{255, 255, 0, 255}, true)
		vector.DrawFilledCircle(screen, float32(bx)+float32(px*90)+45, float32(by)+float32(bh)-15, 2, color.RGBA{255, 255, 0, 255}, true)
	}
}

func (g *Game) drawTransition(screen *ebiten.Image) {
	t := g.TransitionTime
	g.drawBloom(screen, g.ReunionPoint, 400.0*t, color.RGBA{255, 200, 255, uint8(255 * (1.0 - t))})
	for id, wellValue := range g.World.GravityWells {
		if wellValue == nil { continue }
		wellTrans := g.World.Transforms[id]
		if wellTrans == nil { continue }
		g.drawBloom(screen, wellTrans.Position, 300.0*t, color.RGBA{255, 255, 200, uint8(200 * (1.0 - t))})
		coreAlpha := uint8(200 * t)
		vector.DrawFilledCircle(screen, float32(wellTrans.Position.X), float32(wellTrans.Position.Y), float32(wellValue.Radius * (1.0 + t*2)), color.RGBA{255, 255, 255, coreAlpha}, true)
	}
	const cellSize = 40
	for y := 0; y < core.ScreenHeight; y += cellSize {
		for x := 0; x < core.ScreenWidth; x += cellSize {
			prog := t*2.0 - (float64(x)/core.ScreenWidth)*0.5 - (float64(y)/core.ScreenHeight)*0.5
			if prog < 0 { prog = 0 } else if prog > 1 { prog = 1 }
			size := float32(cellSize) * float32(prog)
			if size > 0 { vector.DrawFilledRect(screen, float32(x), float32(y), size, size, color.RGBA{255, 255, 200, uint8(255*prog)}, false) }
		}
	}
}

func (g *Game) drawBloom(screen *ebiten.Image, center core.Vector2, radius float64, clr color.RGBA) {
	numPetals := 12
	for i := 0; i < numPetals; i++ {
		angle := (float64(i) / float64(numPetals)) * 2 * math.Pi
		px, py := center.X + math.Cos(angle)*radius, center.Y + math.Sin(angle)*radius
		vector.DrawFilledCircle(screen, float32(px), float32(py), float32(radius*0.2), clr, true)
	}
}

func (g *Game) Layout(w, h int) (int, int) { return core.ScreenWidth, core.ScreenHeight }

func main() {
	ebiten.SetWindowSize(core.ScreenWidth, core.ScreenHeight)
	ebiten.SetWindowTitle("Beautiful Mess: The Final Code")
	ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(NewGame()); err != nil { panic(err) }
}
