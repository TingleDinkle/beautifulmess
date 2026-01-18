package systems

import (
	"image/color"
	"math"
	"math/rand"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func SystemInput(w *world.World) {
	for e := range w.InputControlleds {
		phys := w.Physics[e]
		if phys == nil {
			continue
		}

		// Boost acceleration for dash mechanic
		accel := 1.5
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			accel = 3.5
			w.Audio.Play("boost")
			
			// Emit thruster particles
			if trans := w.Transforms[e]; trans != nil {
				// Reverse velocity vector for exhaust
				exhaustX := -phys.Velocity.X * 0.5
				exhaustY := -phys.Velocity.Y * 0.5
				// Jitter
				exhaustX += (rand.Float64() - 0.5) * 2
				exhaustY += (rand.Float64() - 0.5) * 2
				
				w.Particles.Emit(
					trans.Position,
					core.Vector2{X: exhaustX, Y: exhaustY},
					color.RGBA{0, 255, 255, 255},
					0.05,
				)
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {


			phys.Acceleration.X -= accel

		}

		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {

			phys.Acceleration.X += accel

		}

		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {

			phys.Acceleration.Y -= accel

		}

		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {

			phys.Acceleration.Y += accel

		}



		// Align orientation with velocity vector for visual feedback

		trans := w.Transforms[e]

		if trans != nil && (phys.Velocity.X != 0 || phys.Velocity.Y != 0) {

			trans.Rotation = math.Atan2(phys.Velocity.Y, phys.Velocity.X)

		}

	}

}


