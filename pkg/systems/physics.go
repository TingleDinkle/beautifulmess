package systems

import (
	"image/color"
	"math"
	"math/rand"

	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/particles"
	"beautifulmess/pkg/world"
)

func SystemPhysics(w *world.World) {
	// Cache-friendly sequential iteration over component slices minimizes CPU pipeline stalls
	for id, phys := range w.Physics {
		if phys == nil { continue }
		trans := w.Transforms[id]
		if trans == nil { continue }

		applyForces(core.Entity(id), w)
		integrate(phys, trans)
		core.WrapPosition(&trans.Position)
		handleCollisions(core.Entity(id), w)
	}
}

func applyForces(id core.Entity, w *world.World) {
	phys := w.Physics[id]
	trans := w.Transforms[id]

	// Bullets move at high velocities and effectively ignore gravitational curvature
	if tag := w.Tags[id]; tag != nil && tag.Name == "bullet" {
		return
	}

	for wellID, well := range w.GravityWells {
		if well == nil || int(id) == wellID { continue }
		wellTrans := w.Transforms[wellID]
		if wellTrans == nil { continue }

		delta := core.VecToWrapped(trans.Position, wellTrans.Position)
		d := math.Max(10, math.Sqrt(delta.X*delta.X+delta.Y*delta.Y))

		multiplier := phys.GravityMultiplier
		if multiplier <= 0 { multiplier = 1.0 }
		
		fRatio := math.Min(5.0, (well.Mass*500)/(d*d)) * multiplier / d

		phys.Acceleration.X += delta.X * fRatio
		phys.Acceleration.Y += delta.Y * fRatio
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
	trans := w.Transforms[id]
	phys := w.Physics[id]
	if trans == nil || phys == nil { return }
	
	tag := w.Tags[id]

	// Querying the spatial grid neighbor-cells handles toroidal-wrapped collisions with zero overhead
	gx, gy := int(trans.Position.X/100), int(trans.Position.Y/100)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			// Wrapping cell indices ensures consistent spatial awareness at the universe edges
			tx, ty := (gx+dx+13)%13, (gy+dy+8)%8
			
			for _, wallID := range w.Grid[tx][ty] {
				wall := w.Walls[wallID]
				if wall == nil || wall.IsDestroyed { continue }
				wallTrans := w.Transforms[wallID]
				if wallTrans == nil { continue }

				delta := core.VecToWrapped(trans.Position, wallTrans.Position)
				adx, ady := math.Abs(delta.X), math.Abs(delta.Y)

				if adx < (5.0 + wall.Size/2) && ady < (5.0 + wall.Size/2) {
					if tag != nil && tag.Name == "bullet" {
						if adx > ady {
							phys.Velocity.X *= -1.0
							if delta.X > 0 { trans.Position.X = wallTrans.Position.X - (5.1 + wall.Size/2) } else { trans.Position.X = wallTrans.Position.X + (5.1 + wall.Size/2) }
						} else {
							phys.Velocity.Y *= -1.0
							if delta.Y > 0 { trans.Position.Y = wallTrans.Position.Y - (5.1 + wall.Size/2) } else { trans.Position.Y = wallTrans.Position.Y + (5.1 + wall.Size/2) }
						}

						if wall.Destructible {
							shatterEntity(w, wallID, phys.Velocity)
							w.Audio.Play("boom") 
							w.ScreenShake += 4.0
							w.DestroyEntity(wallID)
						} else {
							emitImpactFeedback(w, trans.Position)
							w.ScreenShake += 1.0
						}
						return
					}
					phys.Velocity.X, phys.Velocity.Y = 0, 0
				}
			}
		}
	}

	if tag != nil && tag.Name == "bullet" {
		for _, specID := range w.ActiveEntities {
			specTag := w.Tags[specID]
			if specTag == nil || specTag.Name != "spectre" { continue }
			
			specTrans, specPhys := w.Transforms[specID], w.Physics[specID]
			if specTrans != nil && specPhys != nil {
				// 400 represents the squared radius (20^2), providing a zero-sqrt hit-detection path
				if core.DistSqWrapped(trans.Position, specTrans.Position) < 400 {
					specPhys.GravityMultiplier += 1.0
					w.Audio.Play("boom")
					w.ScreenShake += 8.0
					shatterEntity(w, specID, phys.Velocity)
					w.DestroyEntity(id)
					return
				}
			}
		}
	}
}





func shatterEntity(w *world.World, id core.Entity, impactVel core.Vector2) {


	// Specialized particle quirks provide a high-fidelity 'Nintendo-grade' destruction feel


	trans := w.Transforms[id]


	render := w.Renders[id]


	if trans == nil || render == nil { return }





	// Core explosion with inherited momentum


	biasVel := core.Vector2{X: impactVel.X * 0.2, Y: impactVel.Y * 0.2}


	spawnDebrisQuirky(w, trans.Position, render.Color, 15, 4.0, biasVel, particles.QuirkStandard)


	


	// 'Orphaned' data fragments that orbit the blast center create visual complexity


	spawnDebrisQuirky(w, trans.Position, color.RGBA{255, 255, 200, 255}, 6, 6.0, biasVel, particles.QuirkOrbit)


	


	// Flickering sparks simulate energetic discharge


	spawnDebrisQuirky(w, trans.Position, render.Color, 10, 3.0, biasVel, particles.QuirkFlicker)


}

func emitImpactFeedback(w *world.World, pos core.Vector2) {
	// High-frequency flickering sparks convey the hardness of indestructible surfaces
	spawnDebrisQuirky(w, pos, color.RGBA{200, 200, 255, 255}, 5, 2.0, core.Vector2{}, particles.QuirkFlicker)
}

func spawnDebrisQuirky(w *world.World, pos core.Vector2, col color.RGBA, count int, maxSpeed float64, bias core.Vector2, quirk particles.ParticleQuirk) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := rand.Float64() * maxSpeed
		
		vel := core.Vector2{
			X: math.Cos(angle)*speed + bias.X,
			Y: math.Sin(angle)*speed + bias.Y,
		}
		
		w.Particles.EmitAdvanced(
			pos,
			vel,
			col,
			0.01 + rand.Float64()*0.04,
			quirk,
		)
	}
}		        
