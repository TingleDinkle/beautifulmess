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

			shooterTrans := w.Transforms[id]
			if shooterTrans == nil {
				continue
			}

			// Spawn Entity: Create a "Gravity Well Bullet" at Position + ForwardVector * 20
			dirX := math.Cos(shooterTrans.Rotation)
			dirY := math.Sin(shooterTrans.Rotation)

			spawnPos := core.Vector2{
				X: shooterTrans.Position.X + dirX*20,
				Y: shooterTrans.Position.Y + dirY*20,
			}

			bulletID := w.CreateEntity()

			w.Transforms[bulletID] = &components.Transform{
				Position: spawnPos,
				Rotation: shooterTrans.Rotation,
			}

			w.Physics[bulletID] = &components.Physics{
				Velocity: core.Vector2{X: dirX * 8.0, Y: dirY * 8.0},
				MaxSpeed: 20.0,
				Mass:     5.0,
				Friction: 1.0, 
			}

			w.Renders[bulletID] = &components.Render{
				Sprite: generateBulletSprite(),
				Color:  color.RGBA{255, 255, 255, 255},
				Scale:  0.5,
			}
			
			w.Lifetimes[bulletID] = &components.Lifetime{TimeRemaining: 2.0}
			
			// Add Tag to identify it
			w.Tags[bulletID] = &components.Tag{Name: "bullet"}
		}
	}
}

func generateBulletSprite() *ebiten.Image {
	img := ebiten.NewImage(8, 8)
	img.Fill(color.White) // Simple white square
	return img
}
