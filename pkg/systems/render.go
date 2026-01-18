package systems

import (
	"image/color"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/level"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func DrawLevel(screen *ebiten.Image, lvl *level.Level, spectrePos core.Vector2) {
	// Draw Gravity Wells (Black Holes)
	for _, well := range lvl.Wells {
		// Event Horizon (Black)
		DrawWrappedCircle(screen, well.Position, well.Radius, color.RGBA{0, 0, 0, 255}, true)
		// Accretion Disk (Purple)
		DrawWrappedCircle(screen, well.Position, well.Radius, color.RGBA{100, 0, 100, 255}, false)
	}

	// Draw Memory Node
	// It glows Red if the Spectre is currently trapped inside it
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
		DrawWrappedSprite(screen, r.Sprite, trans.Position, trans.Rotation)
	}
}

func DrawWrappedCircle(screen *ebiten.Image, pos core.Vector2, r float64, c color.RGBA, fill bool) {
	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			x := float32(pos.X + ox*core.ScreenWidth)
			y := float32(pos.Y + oy*core.ScreenHeight)

			// Optimization: Check visibility
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

func DrawWrappedSprite(screen *ebiten.Image, img *ebiten.Image, pos core.Vector2, rot float64) {
	w, h := img.Size()
	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			x := pos.X + ox*core.ScreenWidth
			y := pos.Y + oy*core.ScreenHeight

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			op.GeoM.Rotate(rot)
			op.GeoM.Translate(x, y)
			screen.DrawImage(img, op)
		}
	}
}