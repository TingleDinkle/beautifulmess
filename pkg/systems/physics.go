package systems

import (
	"math"

	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"
)

func SystemPhysics(w *world.World) {
	for id, phys := range w.Physics {
		trans, ok := w.Transforms[id]
		if !ok {
			continue
		}

		applyForces(id, w)
		integrate(phys, trans)
		enforceBoundaries(trans)
		handleCollisions(id, w)
	}
}

func applyForces(id core.Entity, w *world.World) {
	phys := w.Physics[id]
	trans := w.Transforms[id]

	// Bullets are modeled as high-energy particles unaffected by gravitational curvature
	if tag, ok := w.Tags[id]; ok && tag.Name == "bullet" {
		return
	}

	for wellID, well := range w.GravityWells {
		wellTrans, ok := w.Transforms[wellID]
		if !ok || well == nil {
			continue
		}

		delta := core.VecToWrapped(trans.Position, wellTrans.Position)
		d := math.Max(10, math.Sqrt(delta.X*delta.X+delta.Y*delta.Y))

		// Standard inverse-square law for orbital mechanics
		force := (well.Mass * 500) / (d * d)
		force = math.Min(2.0, force)

		if phys.GravityMultiplier > 0 {
			force *= phys.GravityMultiplier
		}

		phys.Acceleration.X += (delta.X / d) * force
		phys.Acceleration.Y += (delta.Y / d) * force
	}
}

func integrate(phys *components.Physics, trans *components.Transform) {
	// Semi-implicit Euler provides better energy conservation than standard Euler
	phys.Velocity.X += phys.Acceleration.X
	phys.Velocity.Y += phys.Acceleration.Y
	phys.Velocity.X *= phys.Friction
	phys.Velocity.Y *= phys.Friction

	// Speed clamping prevents 'tunnelling' through thin geometry at high velocities
	speed := math.Sqrt(phys.Velocity.X*phys.Velocity.X + phys.Velocity.Y*phys.Velocity.Y)
	if speed > phys.MaxSpeed {
		scale := phys.MaxSpeed / speed
		phys.Velocity.X *= scale
		phys.Velocity.Y *= scale
	}

	trans.Position.X += phys.Velocity.X
	trans.Position.Y += phys.Velocity.Y
	
	phys.Acceleration.X = 0
	phys.Acceleration.Y = 0
}

func enforceBoundaries(trans *components.Transform) {
	// Toroidal topology ensures a boundless play area without edge-case logic
	if trans.Position.X < 0 { trans.Position.X += core.ScreenWidth }
	if trans.Position.X >= core.ScreenWidth { trans.Position.X -= core.ScreenWidth }
	if trans.Position.Y < 0 { trans.Position.Y += core.ScreenHeight }
	if trans.Position.Y >= core.ScreenHeight { trans.Position.Y -= core.ScreenHeight }
}

func handleCollisions(id core.Entity, w *world.World) {
	trans, okT := w.Transforms[id]
	phys, okP := w.Physics[id]
	if !okT || !okP { return }
	
	tag, hasTag := w.Tags[id]

	// Proximity-based filtering reduces the computational cost of toroidal AABB checks
	entSize := 10.0
	for wallID, wall := range w.Walls {
		if wall == nil || wall.IsDestroyed { continue }
		wallTrans, ok := w.Transforms[wallID]
		if !ok { continue }

		if math.Abs(trans.Position.X-wallTrans.Position.X) < (entSize/2+wall.Size/2) &&
			math.Abs(trans.Position.Y-wallTrans.Position.Y) < (entSize/2+wall.Size/2) {

			if hasTag && tag.Name == "bullet" {
				w.DestroyEntity(id)
				return
			}

			// Momentum cancellation simulates inelastic energy transfer on impact
			phys.Velocity.X, phys.Velocity.Y = 0, 0
			if wall.Destructible { wall.IsDestroyed = true }
		}
	}

	// Dynamic entity interaction logic
	if hasTag && tag.Name == "bullet" {
		for specID, specTag := range w.Tags {
			if specTag.Name == "spectre" {
				specTrans, okST := w.Transforms[specID]
				specPhys, okSP := w.Physics[specID]
				if okST && okSP {
					if core.DistWrapped(trans.Position, specTrans.Position) < 20 {
						// Increasing susceptibility to gravity forces the spectre toward environmental hazards
						specPhys.GravityMultiplier += 0.5
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

// destroyEntity removed in favor of World.DestroyEntity


