package systems

import (
	"math"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"
)

func SystemPhysics(w *world.World) {
	for id, phys := range w.Physics {
		trans := w.Transforms[id]
		if trans == nil {
			continue
		}

		// Calculate gravitational pull from all wells
		for wellID, well := range w.GravityWells {
			wellTrans := w.Transforms[wellID]
			if wellTrans == nil {
				continue
			}

			// Compute force vector across wrapped boundaries so gravity affects entities
			// from the "other side" of the screen.
			delta := core.VecToWrapped(trans.Position, wellTrans.Position)
			dx, dy := delta.X, delta.Y

			d := math.Sqrt(dx*dx + dy*dy)
			if d < 10 {
				d = 10
			} // Prevent division-by-zero singularities

			// Inverse-square law for gravity
			force := (well.Mass * 500) / (d * d)
			if force > 2.0 {
				force = 2.0
			} // Clamp force to avoid instability at close range

			// Apply entity-specific gravity susceptibility
			if phys.GravityMultiplier > 0 {
				force *= phys.GravityMultiplier
			}

			phys.Acceleration.X += (dx / d) * force
			phys.Acceleration.Y += (dy / d) * force
		}

		// Semi-implicit Euler integration
		phys.Velocity.X += phys.Acceleration.X
		phys.Velocity.Y += phys.Acceleration.Y

		// Apply friction to simulate medium resistance
		phys.Velocity.X *= phys.Friction
		phys.Velocity.Y *= phys.Friction

		// Clamp velocity to maintain control and prevent tunneling
		speed := math.Sqrt(phys.Velocity.X*phys.Velocity.X + phys.Velocity.Y*phys.Velocity.Y)
		if speed > phys.MaxSpeed {
			scale := phys.MaxSpeed / speed
			phys.Velocity.X *= scale
			phys.Velocity.Y *= scale
		}

		// Update Position
		trans.Position.X += phys.Velocity.X
		trans.Position.Y += phys.Velocity.Y

		// Wrap position to enforce toroidal topology
		if trans.Position.X < 0 {
			trans.Position.X += core.ScreenWidth
		}
		if trans.Position.X >= core.ScreenWidth {
			trans.Position.X -= core.ScreenWidth
		}
		if trans.Position.Y < 0 {
			trans.Position.Y += core.ScreenHeight
		}
		if trans.Position.Y >= core.ScreenHeight {
			trans.Position.Y -= core.ScreenHeight
		}

		// Reset for next frame
		phys.Acceleration.X = 0
		phys.Acceleration.Y = 0
	}
}
