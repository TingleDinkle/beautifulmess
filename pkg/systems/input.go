package systems

import (
	"math"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func SystemInput(w *world.World, e core.Entity) {
	phys := w.Physics[e]
	if phys == nil {
		return
	}

	// "Shift" to burn fuel/glitch
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

	// Rotate sprite to face movement
	trans := w.Transforms[e]
	if phys.Velocity.X != 0 || phys.Velocity.Y != 0 {
		trans.Rotation = math.Atan2(phys.Velocity.Y, phys.Velocity.X)
	}
}