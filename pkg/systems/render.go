package systems

import (
	"image/color"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/level"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func DrawLevel(screen *ebiten.Image, w *world.World, lvl *level.Level, spectrePos core.Vector2) {

	for id, well := range w.GravityWells {

		trans := w.Transforms[id]

		if trans == nil {

			continue

		}

		DrawWrappedCircle(screen, trans.Position, well.Radius, color.RGBA{0, 0, 0, 255}, true)

		DrawWrappedCircle(screen, trans.Position, well.Radius, color.RGBA{100, 0, 100, 255}, false)

	}



	// Highlight memory node when the Spectre is vulnerable/trapped

	mc := color.RGBA{200, 200, 255, 100}

	if core.DistWrapped(spectrePos, lvl.Memory.Position) < core.MemoryRadius {

		mc = color.RGBA{255, 50, 50, 255}

	}

	DrawWrappedCircle(screen, lvl.Memory.Position, core.MemoryRadius, mc, false)

}



func DrawEntities(screen *ebiten.Image, w *world.World) {

	for id, r := range w.Renders {

		if r.Sprite == nil {

			continue

		}

		trans := w.Transforms[id]

		scale := r.Scale
		if scale == 0 {
			scale = 1.0
		}

		DrawWrappedSprite(screen, r.Sprite, trans.Position, trans.Rotation, scale, r.Color)

	}

}



func DrawWrappedCircle(screen *ebiten.Image, pos core.Vector2, r float64, c color.RGBA, fill bool) {

	for ox := -1.0; ox <= 1.0; ox++ {

		for oy := -1.0; oy <= 1.0; oy++ {

			x := float32(pos.X + ox*core.ScreenWidth)

			y := float32(pos.Y + oy*core.ScreenHeight)



			// Skip drawing instances that fall outside the viewport to reduce fill rate

			if x+float32(r) < 0 || x-float32(r) > float32(core.ScreenWidth) || y+float32(r) < 0 || y-float32(r) > float32(core.ScreenHeight) {

				continue

			}



			if fill {

				vector.DrawFilledCircle(screen, x, y, float32(r), c, true)

			} else {

				vector.StrokeCircle(screen, x, y, float32(r), 2, c, true)

			}

		}

	}

}



func DrawWrappedSprite(screen *ebiten.Image, img *ebiten.Image, pos core.Vector2, rot float64, scale float64, clr color.RGBA) {

	w, h := img.Size()

	halfW, halfH := float64(w)/2, float64(h)/2

	

	for ox := -1.0; ox <= 1.0; ox++ {

		for oy := -1.0; oy <= 1.0; oy++ {

			x := pos.X + ox*core.ScreenWidth

			y := pos.Y + oy*core.ScreenHeight



			// Skip drawing instances that fall outside the viewport to reduce fill rate

			// Box check needs to account for scale
			scaledW := float64(w) * scale
			scaledH := float64(h) * scale
			if x+scaledW/2 < 0 || x-scaledW/2 > core.ScreenWidth || y+scaledH/2 < 0 || y-scaledH/2 > core.ScreenHeight {
				continue
			}



			op := &ebiten.DrawImageOptions{}

			op.Filter = ebiten.FilterNearest

			op.GeoM.Translate(-halfW, -halfH)
			op.GeoM.Scale(scale, scale)

			op.GeoM.Rotate(rot)

			op.GeoM.Translate(x, y)
			
			op.ColorScale.ScaleWithColor(clr)

			screen.DrawImage(img, op)

		}

	}

}


