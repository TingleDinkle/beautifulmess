package systems

import (
	"math"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/level"
	"beautifulmess/pkg/world"
)

func SystemPhysics(w *world.World, lvl *level.Level) {
	for id, phys := range w.Physics {
		trans := w.Transforms[id]
		if trans == nil {
			continue
		}

		// 1. Gravity Logic
		for _, well := range lvl.Wells {
			dx := well.Position.X - trans.Position.X
			dy := well.Position.Y - trans.Position.Y

			// Toroidal Distance: Gravity wraps around the screen too!
			if dx > core.ScreenWidth/2 {
				dx -= core.ScreenWidth
			}
			if dx < -core.ScreenWidth/2 {
				dx += core.ScreenWidth
			}
			if dy > core.ScreenHeight/2 {
				dy -= core.ScreenHeight
			}
			if dy < -core.ScreenHeight/2 {
				dy += core.ScreenHeight
			}

			d := math.Sqrt(dx*dx + dy*dy)
			if d < 10 {
				d = 10
			} // Prevent singularities (div by zero)

			// Gravity falls off with distance squared
			force := (well.Mass * 500) / (d * d)
			if force > 2.0 {
				force = 2.0
			} // Cap force to prevent glitchy teleporting

			// THE TRAP: If it's the Spectre and she's inside the event horizon...
			if tag, ok := w.Tags[id]; ok && tag.Name == "spectre" && d < well.Radius {
				force *= 3.5 // Crush her
				// Note: We used to apply drag here, but keeping momentum creates a cool "Orbit" effect
			}

			phys.Acceleration.X += (dx / d) * force
			phys.Acceleration.Y += (dy / d) * force
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
