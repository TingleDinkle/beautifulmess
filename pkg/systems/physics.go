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
		entSize := 10.0
		for wallID, wall := range w.Walls {
			if wall.IsDestroyed {
				continue
			}
			wallTrans := w.Transforms[wallID]
			if wallTrans == nil {
				continue
			}
			
			// Simple collision
			if math.Abs(trans.Position.X - wallTrans.Position.X) < (entSize/2 + wall.Size/2) &&
			   math.Abs(trans.Position.Y - wallTrans.Position.Y) < (entSize/2 + wall.Size/2) {
				
				// If this is a bullet, destroy it silently
				if tag, ok := w.Tags[id]; ok && tag.Name == "bullet" {
					delete(w.Physics, id)
					delete(w.Renders, id)
					delete(w.Transforms, id)
					delete(w.Tags, id)
					return // Stop processing this entity
				}

				phys.Velocity.X = 0
				phys.Velocity.Y = 0
				
				if wall.Destructible {
					wall.IsDestroyed = true
				}
			}
		}
		
		// Bullet vs Spectre Collision
		if tag, ok := w.Tags[id]; ok && tag.Name == "bullet" {
			for specID, specTag := range w.Tags {
				if specTag.Name == "spectre" {
					specTrans := w.Transforms[specID]
					specPhys := w.Physics[specID]
					if specTrans != nil && specPhys != nil {
						dist := core.DistWrapped(trans.Position, specTrans.Position)
						if dist < 20 { 
							specPhys.GravityMultiplier += 0.5 
							w.Audio.Play("boom") 
							
							delete(w.Physics, id)
							delete(w.Renders, id)
							delete(w.Transforms, id)
							delete(w.Tags, id)
							return // Stop processing this entity
						}
					}
				}
			}
		}
	}
}
