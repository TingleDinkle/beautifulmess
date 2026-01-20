package systems

import (
	"image/color"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"
	
	"github.com/hajimehoshi/ebiten/v2"
)

type SpectreVisualState struct {
	State string
	Timer float64
}

func SystemSpectreVisuals(w *world.World, gState *SpectreVisualState, spectreID core.Entity, sprites map[string]*ebiten.Image) {
	if int(spectreID) >= len(w.Renders) || w.Renders[spectreID] == nil { return }
	render := w.Renders[spectreID]
	trans := w.Transforms[spectreID]
	phys := w.Physics[spectreID]
	if trans == nil { return }

	targetState := "normal"
	
	// Optimization: Use Squared Distance to avoid math.Sqrt in the hot loop
	const kewtRangeSq = 30.0 * 30.0
	for id, well := range w.GravityWells {
		if well == nil { continue }
		wellTrans := w.Transforms[id]
		if wellTrans == nil { continue }
		
		if core.DistSqWrapped(trans.Position, wellTrans.Position) < (well.Radius*well.Radius + kewtRangeSq) {
			targetState = "kewt"
			break
		}
	}

	if targetState == "normal" && phys != nil {
		// Use squared velocity for the check
		velSq := phys.Velocity.X*phys.Velocity.X + phys.Velocity.Y*phys.Velocity.Y
		accSq := phys.Acceleration.X*phys.Acceleration.X + phys.Acceleration.Y*phys.Acceleration.Y
		if velSq > 56.25 || accSq > 1.21 { // 7.5^2 and 1.1^2
			targetState = "angy"
		}
	}

	// Hysteresis logic
	if targetState != gState.State {
		if gState.Timer <= 0 {
			gState.State = targetState
			gState.Timer = 0.5 
		}
	}
	if gState.Timer > 0 { gState.Timer -= 1.0 / 60.0 }

	// Apply visuals
	render.Sprite = sprites[gState.State]
	
	specW, _ := render.Sprite.Size()
	baseScale := 80.0 / float64(specW)
	targetScale := baseScale
	targetColor := color.RGBA{255, 255, 255, 255}
	
	switch gState.State {
	case "angy":
		targetColor = color.RGBA{255, 230, 230, 255}
		render.Glow = true
	case "kewt":
		targetScale = baseScale * 0.95
		render.Glow = false
	default:
		render.Glow = true
	}

	// Professional-grade lerping for smooth transitions
	const lerpSpeed = 0.1
	render.Scale += (targetScale - render.Scale) * lerpSpeed
	
	r, g, b, _ := render.Color.RGBA()
	tr, tg, tb, _ := targetColor.RGBA()
	render.Color.R = uint8(float64(uint8(r>>8)) + (float64(uint8(tr>>8))-float64(uint8(r>>8)))*lerpSpeed)
	render.Color.G = uint8(float64(uint8(g>>8)) + (float64(uint8(tg>>8))-float64(uint8(g>>8)))*lerpSpeed)
	render.Color.B = uint8(float64(uint8(b>>8)) + (float64(uint8(tb>>8))-float64(uint8(b>>8)))*lerpSpeed)
	render.Color.A = 255
}
