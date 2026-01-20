package systems

import (
	"image/color"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"
	
	"github.com/hajimehoshi/ebiten/v2"
)

// SpectreVisualState tracks the high-level emotional state of the Spectre entity.
// This is separated from the Physics state to allow for visual smoothing and transitions
// that don't interfere with the deterministic simulation.
type SpectreVisualState struct {
	State string
	Timer float64
}

// SystemSpectreVisuals manages the emotional reactivity of the Spectre.
// It bridges the gap between raw physics data (velocity/acceleration) and 
// narrative-driven visual feedback (the 3 source photos).
func SystemSpectreVisuals(w *world.World, gState *SpectreVisualState, spectreID core.Entity, sprites map[string]*ebiten.Image) {
	if int(spectreID) >= len(w.Renders) || w.Renders[spectreID] == nil { return }
	render := w.Renders[spectreID]
	trans := w.Transforms[spectreID]
	phys := w.Physics[spectreID]
	if trans == nil { return }

	targetState := "normal"
	
	// We prioritize the "Kewt" state as it represents a total loss of agency within a well.
	// 30px buffer provides a 'gravity-well' event horizon for the visual transition.
	const kewtRangeSq = 30.0 * 30.0
	for id, well := range w.GravityWells {
		if well == nil { continue }
		wellTrans := w.Transforms[id]
		if wellTrans == nil { continue }
		
		// Squared distance avoids the computationally expensive math.Sqrt in the hot update path.
		if core.DistSqWrapped(trans.Position, wellTrans.Position) < (well.Radius*well.Radius + kewtRangeSq) {
			targetState = "kewt"
			break
		}
	}

	// "Angy" state is a byproduct of extreme physical exertion (the Swish/Dodging behavior).
	if targetState == "normal" && phys != nil {
		velSq := phys.Velocity.X*phys.Velocity.X + phys.Velocity.Y*phys.Velocity.Y
		accSq := phys.Acceleration.X*phys.Acceleration.X + phys.Acceleration.Y*phys.Acceleration.Y
		// Thresholds (7.5 for vel, 1.1 for acc) are tuned to catch active player-evasion maneuvers.
		if velSq > 56.25 || accSq > 1.21 { 
			targetState = "angy"
		}
	}

	// State Hysteresis: We force a minimum dwell-time in "angy" and "kewt" to prevent
	// sprite flickering when physics values oscillate rapidly around the thresholds.
	if targetState != gState.State {
		if gState.Timer <= 0 {
			gState.State = targetState
			gState.Timer = 0.5 
		}
	}
	if gState.Timer > 0 { gState.Timer -= 1.0 / 60.0 }

	// Sprite selection is now stable thanks to the uniform 128x128 resolution forced at load-time.
	render.Sprite = sprites[gState.State]
	
	specW, _ := render.Sprite.Size()
	baseScale := 80.0 / float64(specW)
	targetScale := baseScale
	targetColor := color.RGBA{255, 255, 255, 255}
	
	switch gState.State {
	case "angy":
		// Red tinting is subtle to preserve the integrity of the original photo.
		targetColor = color.RGBA{255, 230, 230, 255}
		render.Glow = true
	case "kewt":
		// Shrinking the scale slightly simulates the compression of being 'trapped'.
		targetScale = baseScale * 0.95
		render.Glow = false
	default:
		render.Glow = true
	}

	// Exponential smoothing (lerp) ensures that state transitions feel like organic
	// emotional shifts rather than binary code swaps.
	const lerpSpeed = 0.1
	render.Scale += (targetScale - render.Scale) * lerpSpeed
	
	r, g, b, _ := render.Color.RGBA()
	tr, tg, tb, _ := targetColor.RGBA()
	render.Color.R = uint8(float64(uint8(r>>8)) + (float64(uint8(tr>>8))-float64(uint8(r>>8)))*lerpSpeed)
	render.Color.G = uint8(float64(uint8(g>>8)) + (float64(uint8(tg>>8))-float64(uint8(g>>8)))*lerpSpeed)
	render.Color.B = uint8(float64(uint8(b>>8)) + (float64(uint8(tb>>8))-float64(uint8(b>>8)))*lerpSpeed)
	render.Color.A = 255
}