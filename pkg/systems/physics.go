package systems

import (
	"math"

	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"
)

func SystemPhysics(w *world.World) {
	for id, phys := range w.Physics {
		trans := w.Transforms[id]
		if trans == nil {
			continue
		}

		// AstroParty Physics for Player
		if _, isPlayer := w.InputControlleds[id]; isPlayer {
			applyAstroPhysics(phys, trans)
		} else {
			// Standard Physics for others (Spectre, etc)
			// Calculate gravitational pull from all wells
			for wellID, well := range w.GravityWells {
				wellTrans := w.Transforms[wellID]
				if wellTrans == nil {
					continue
				}

				delta := core.VecToWrapped(trans.Position, wellTrans.Position)
				dx, dy := delta.X, delta.Y

				d := math.Sqrt(dx*dx + dy*dy)
				if d < 10 {
					d = 10
				}

				force := (well.Mass * 500) / (d * d)
				if force > 2.0 {
					force = 2.0
				}

				if phys.GravityMultiplier > 0 {
					force *= phys.GravityMultiplier
				}

				phys.Acceleration.X += (dx / d) * force
				phys.Acceleration.Y += (dy / d) * force
			}

			phys.Velocity.X += phys.Acceleration.X
			phys.Velocity.Y += phys.Acceleration.Y
			phys.Velocity.X *= phys.Friction
			phys.Velocity.Y *= phys.Friction

			// Clamp velocity
			speed := math.Sqrt(phys.Velocity.X*phys.Velocity.X + phys.Velocity.Y*phys.Velocity.Y)
			if speed > phys.MaxSpeed {
				scale := phys.MaxSpeed / speed
				phys.Velocity.X *= scale
				phys.Velocity.Y *= scale
			}
			
			// Reset acceleration
			phys.Acceleration.X = 0
			phys.Acceleration.Y = 0
		}

		// Update Position
		trans.Position.X += phys.Velocity.X
		trans.Position.Y += phys.Velocity.Y

		// Wrap position
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
		
		// Wall Collision
		// Simplistic AABB (Entity is ~8-10px, Wall is ~10-40px?)
		// Prompt says "Simplistic AABB collision for Walls (10x10 size)"
		// We'll treat entity as 10x10 too for simplicity
		entSize := 10.0
		
		// Ranging over map to get ID
		for wallID, wall := range w.Walls {
			if wall.IsDestroyed {
				continue
			}
			wallTrans := w.Transforms[wallID]
			if wallTrans == nil {
				continue
			}
			
			// Check AABB
			// Simple non-wrapped collision for walls (assuming walls don't wrap or are handled simply)
			if math.Abs(trans.Position.X - wallTrans.Position.X) < (entSize/2 + wall.Size/2) &&
			   math.Abs(trans.Position.Y - wallTrans.Position.Y) < (entSize/2 + wall.Size/2) {
				
				// Stop Velocity
				phys.Velocity.X = 0
				phys.Velocity.Y = 0
				
				if wall.Destructible {
					wall.IsDestroyed = true
				}
			}
		}
	}
}

// Apply "Tank" physics: The ship always accelerates in the direction it faces
func applyAstroPhysics(phys *components.Physics, trans *components.Transform) {
	const dt = 1.0 / 60.0
	const accel = 2.5 // High acceleration for snappy feel

	// 0 Rotation = Right (+X)
	targetDx := math.Cos(trans.Rotation) * phys.MaxSpeed
	targetDy := math.Sin(trans.Rotation) * phys.MaxSpeed

	// Linear Interpolation towards target velocity
	phys.Velocity.X = lerp(phys.Velocity.X, targetDx, dt*accel)
	phys.Velocity.Y = lerp(phys.Velocity.Y, targetDy, dt*accel)
}

func lerp(current, target, step float64) float64 {
	if math.Abs(target-current) < step {
		return target
	}
	if target > current {
		return current + step
	}
	return current - step
}
