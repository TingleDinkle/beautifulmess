package systems

import (
	"image/color"
	"math"
	"math/rand"

	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func SystemInput(w *world.World) {
	for e := range w.InputControlleds {
		phys, okP := w.Physics[e]
		trans, okT := w.Transforms[e]
		if !okP || !okT { continue }

		// Dynamic speed limits enable the 'overdrive' mechanic, providing physical gratification for skill-based timing
		baseMaxSpeed := 7.5
		accel, input := 1.5, core.Vector2{}
		
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			accel = 4.5 // Increased from 3.5 for more immediate responsiveness
			phys.MaxSpeed = baseMaxSpeed * 2.0 
			w.Audio.Play("boost")
			emitThrusterParticles(w, trans, phys)
		} else {
			// Decelerating back to base speed maintains the game's core physical balance
			phys.MaxSpeed = baseMaxSpeed
		}

		// Calculating an explicit input vector separates player intent from physical momentum
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) { input.X -= 1 }
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) { input.X += 1 }
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) { input.Y -= 1 }
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) { input.Y += 1 }

		phys.Acceleration.X += input.X * accel
		phys.Acceleration.Y += input.Y * accel

		// Updating rotation based on input intent prevents visual 'flipping' during recoil or external impacts
		if input.X != 0 || input.Y != 0 {
			trans.Rotation = math.Atan2(input.Y, input.X)
		}
	}
}

func emitThrusterParticles(w *world.World, trans *components.Transform, phys *components.Physics) {
	// Particle emission provides immediate visual feedback for high-velocity state changes
	exhaustX, exhaustY := -phys.Velocity.X*0.5, -phys.Velocity.Y*0.5
	exhaustX += (rand.Float64() - 0.5) * 2
	exhaustY += (rand.Float64() - 0.5) * 2
	
	w.Particles.Emit(
		trans.Position,
		core.Vector2{X: exhaustX, Y: exhaustY},
		color.RGBA{0, 255, 255, 255},
		0.05,
	)
}



