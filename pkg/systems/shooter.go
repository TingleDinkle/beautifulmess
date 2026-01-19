package systems

import (
	"image/color"
	"math"
	"time"

	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func SystemProjectileEmitter(w *world.World) {
	now := float64(time.Now().UnixNano()) / 1e9

	for id, emitter := range w.ProjectileEmitters {
		if now > emitter.LastTime+emitter.Interval {
			emitter.LastTime = now

			w.Audio.Play("shoot")
			
			// Visual and physical feedback reinforces the gun's power scale
			w.ScreenShake += 2.0

			trans, ok := w.Transforms[id]
			if !ok { continue }

			dirX, dirY := math.Cos(trans.Rotation), math.Sin(trans.Rotation)
			
			// Newton's third law: Recoil provides a tactile penalty for blind-firing
			if phys, ok := w.Physics[id]; ok {
				phys.Velocity.X -= dirX * 1.5
				phys.Velocity.Y -= dirY * 1.5
			}

			spawnBullet(w, trans.Position, trans.Rotation, dirX, dirY)
		}
	}
}

func spawnBullet(w *world.World, pos core.Vector2, rot, dx, dy float64) {
	id := w.CreateEntity()

	w.Transforms[id] = &components.Transform{
		Position: core.Vector2{X: pos.X + dx*20, Y: pos.Y + dy*20},
		Rotation: rot,
	}

	w.Physics[id] = &components.Physics{
		Velocity: core.Vector2{X: dx * 8.0, Y: dy * 8.0},
		MaxSpeed: 20.0,
		Mass:     5.0,
		Friction: 1.0, 
	}

	w.Renders[id] = &components.Render{
		Sprite: generateBulletSprite(),
		Color:  color.RGBA{255, 255, 255, 255},
		Scale:  0.5,
	}
	
	// Temporary lifespan prevents memory leaks from missed projectiles
	w.Lifetimes[id] = &components.Lifetime{TimeRemaining: 2.0}
	w.Tags[id] = &components.Tag{Name: "bullet"}
}

func generateBulletSprite() *ebiten.Image {
	// Simple square geometry fits the low-resolution arcade aesthetic
	img := ebiten.NewImage(8, 8)
	img.Fill(color.White)
	return img
}

