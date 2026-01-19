package systems

import (
	"image/color"
	"math"
	"math/rand"

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
		
		// Pre-calculating the force-to-distance ratio minimizes redundant division operations in the physics loop
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

		// Wrapped distance calculation ensures that fixed geometry behaves consistently with the toroidal universe
		delta := core.VecToWrapped(trans.Position, wallTrans.Position)
		dx, dy := math.Abs(delta.X), math.Abs(delta.Y)

		        		// AABB collision detection with inelastic response simulates structural impact
		        		if dx < (5.0 + wall.Size/2) && dy < (5.0 + wall.Size/2) {
		        			if hasTag && tag.Name == "bullet" {
		        				if wall.Destructible {
		        					// Triggering a shatter event provides a visceral sense of structural failure
		        					shatterEntity(w, wallID)
		        					w.Audio.Play("boom") 
		        					w.ScreenShake += 4.0
		        					
		        					// Removing the entity from the simulation prevents redundant collision processing
		        					w.DestroyEntity(wallID)
		        				} else {
		        					// Physical feedback on static geometry prevents the world from feeling 'unresponsive'
		        					emitImpactFeedback(w, trans.Position)
		        				}
		        				w.DestroyEntity(id)
		        				return
		        			}
		        
		        			// Inelastic collision response prevents entities from passing through solid data structures
		        			phys.Velocity.X, phys.Velocity.Y = 0, 0
		        		}
		        	}
		        
		        	// Dynamic entity interaction logic
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
		        						
		        						// Spectre shattering (visual only) reinforces the impact weight
		        						shatterEntity(w, specID)
		        						
		        						w.DestroyEntity(id)
		        						return
		        					}
		        				}
		        			}
		        		}
		        	}
		        }
		        
		        func shatterEntity(w *world.World, id core.Entity) {
		        	// A multi-layered pixel-burst simulates the violent deconstruction of data objects
		        	trans, okT := w.Transforms[id]
		        	render, okR := w.Renders[id]
		        	if !okT || !okR { return }
		        
		        	// Primary debris cloud
		        	spawnDebris(w, trans.Position, render.Color, 20, 5.0)
		        	
		        	// Secondary 'glow' shards emphasize the energetic nature of the destruction
		        	glowColor := color.RGBA{255, 255, 200, 255}
		        	spawnDebris(w, trans.Position, glowColor, 8, 8.0)
		        }
		        
		        func emitImpactFeedback(w *world.World, pos core.Vector2) {
		        	// Subtle particle emission on indestructible surfaces conveys physical hardness
		        	spawnDebris(w, pos, color.RGBA{200, 200, 255, 255}, 5, 2.0)
		        }
		        
		        func spawnDebris(w *world.World, pos core.Vector2, col color.RGBA, count int, maxSpeed float64) {
		        	for i := 0; i < count; i++ {
		        		angle := rand.Float64() * 2 * math.Pi
		        		speed := rand.Float64() * maxSpeed
		        		
		        		w.Particles.Emit(
		        			pos,
		        			core.Vector2{X: math.Cos(angle) * speed, Y: math.Sin(angle) * speed},
		        			col,
		        			0.01 + rand.Float64()*0.04, // Varied decay rates create a lingering debris field
		        		)
		        	}
		        }
		        
