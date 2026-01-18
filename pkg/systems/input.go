package systems

import (
	"math"

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


