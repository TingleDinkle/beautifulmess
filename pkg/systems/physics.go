package systems

import (
	"math"

	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"
)

func SystemPhysics(w *world.World) {
	for id, phys := range w.Physics {
		if phys == nil { continue }
		trans, ok := w.Transforms[id]
		if !ok { continue }

		applyForces(id, w)
		integrate(phys, trans)
		core.WrapPosition(&trans.Position)
		handleCollisions(id, w)
	}
}

func applyForces(id core.Entity, w *world.World) {
	phys := w.Physics[id]
	trans := w.Transforms[id]

	// Bullets move at high velocities and effectively ignore gravitational curvature
	if tag, ok := w.Tags[id]; ok && tag.Name == "bullet" {
		return
	}

	for wellID, well := range w.GravityWells {
		if well == nil || id == wellID { continue } // Self-gravitation is physically impossible in this model
		wellTrans, ok := w.Transforms[wellID]
		if !ok { continue }

		delta := core.VecToWrapped(trans.Position, wellTrans.Position)
		d := math.Max(10, math.Sqrt(delta.X*delta.X+delta.Y*delta.Y))

		// A dynamic gravity multiplier allows the environment to become increasingly hostile as the narrative progresses
		multiplier := phys.GravityMultiplier
		if multiplier <= 0 { multiplier = 1.0 }
		
		force := ((well.Mass * 500) / (d * d)) * multiplier
		force = math.Min(5.0, force) // Increased clamp to allow for extreme gravitational 'trap' states

		phys.Acceleration.X += (delta.X / d) * force
		phys.Acceleration.Y += (delta.Y / d) * force
	}
}

func integrate(phys *components.Physics, trans *components.Transform) {
	// Semi-implicit Euler integration preserves system energy better than standard Euler
	phys.Velocity.X += phys.Acceleration.X
	phys.Velocity.Y += phys.Acceleration.Y
	phys.Velocity.X *= phys.Friction
	phys.Velocity.Y *= phys.Friction

	speed := math.Sqrt(phys.Velocity.X*phys.Velocity.X + phys.Velocity.Y*phys.Velocity.Y)
	if speed > phys.MaxSpeed {
		scale := phys.MaxSpeed / speed
		phys.Velocity.X *= scale
		phys.Velocity.Y *= scale
	}

	trans.Position.X += phys.Velocity.X
	trans.Position.Y += phys.Velocity.Y
	
	phys.Acceleration.X, phys.Acceleration.Y = 0, 0
}


func handleCollisions(id core.Entity, w *world.World) {
	trans, okT := w.Transforms[id]
	phys, okP := w.Physics[id]
	if !okT || !okP { return }
	
	tag, hasTag := w.Tags[id]

	// Proximity checks against static geometry maintain high framerates in dense levels
	const wallSize = 10.0
	for wallID, wall := range w.Walls {
		if wall == nil || wall.IsDestroyed { continue }
		wallTrans, ok := w.Transforms[wallID]
		if !ok { continue }

		dx := math.Abs(trans.Position.X - wallTrans.Position.X)
		dy := math.Abs(trans.Position.Y - wallTrans.Position.Y)

		if dx < (5.0 + wall.Size/2) && dy < (5.0 + wall.Size/2) {
			if hasTag && tag.Name == "bullet" {
				w.DestroyEntity(id)
				return
			}

			// Inelastic impact logic simulates realistic data-packet collisions
			phys.Velocity.X, phys.Velocity.Y = 0, 0
			if wall.Destructible { wall.IsDestroyed = true }
		}
	}

	// High-priority entity interactions are processed separately to allow complex behaviors
	if hasTag && tag.Name == "bullet" {
		for specID, specTag := range w.Tags {
			if specTag.Name != "spectre" { continue }
			
			if specTrans, okST := w.Transforms[specID]; okST {
				if specPhys, okSP := w.Physics[specID]; okSP {
					if core.DistWrapped(trans.Position, specTrans.Position) < 20 {
						// Absorbing high-velocity projectiles significantly increases local mass-susceptibility
						specPhys.GravityMultiplier += 1.0
						w.Audio.Play("boom")
						w.ScreenShake += 8.0
						w.DestroyEntity(id)
						return
					}
				}
			}
		}
	}
}


