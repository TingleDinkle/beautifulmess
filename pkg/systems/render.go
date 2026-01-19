package systems

import (
	"image/color"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/level"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func DrawLevel(screen *ebiten.Image, w *world.World, lvl *level.Level, spectrePos core.Vector2, shake core.Vector2) {
	// Abstracting level-layer rendering ensures that environmental mechanics (like gravity) are visually prioritized
	for id, well := range w.GravityWells {
		if well == nil { continue }
		trans := w.Transforms[id]
		if trans == nil { continue }
		
		pos := core.Vector2{X: trans.Position.X + shake.X, Y: trans.Position.Y + shake.Y}
		drawGravityWell(screen, pos, well.Radius)
	}

	// Dynamic goal highlighting provides the primary feedback loop for win-state proximity
	mc := color.RGBA{200, 200, 255, 100}
	if core.DistWrapped(spectrePos, lvl.Memory.Position) < core.MemoryRadius {
		mc = color.RGBA{255, 50, 50, 255}
	}
	
	pos := core.Vector2{X: lvl.Memory.Position.X + shake.X, Y: lvl.Memory.Position.Y + shake.Y}
	DrawWrappedCircle(screen, pos, core.MemoryRadius, mc, false)
}

func drawGravityWell(screen *ebiten.Image, pos core.Vector2, r float64) {
	// A high-contrast 'black hole' aesthetic visually communicates the lethal nature of the singularity
	DrawWrappedCircle(screen, pos, r, color.RGBA{0, 0, 0, 255}, true)
	DrawWrappedCircle(screen, pos, r, color.RGBA{100, 0, 100, 255}, false)
}

func DrawEntities(screen *ebiten.Image, w *world.World, shake core.Vector2) {
	// Batching draw calls by component presence maintains a predictable visual hierarchy
	for id, r := range w.Renders {
		if r == nil || r.Sprite == nil { continue }
		trans := w.Transforms[id]
		if trans == nil { continue }

		scale := r.Scale
		if scale == 0 { scale = 1.0 }
		
		pos := core.Vector2{X: trans.Position.X + shake.X, Y: trans.Position.Y + shake.Y}
		DrawWrappedSprite(screen, r.Sprite, pos, trans.Rotation, scale, r.Color)
	}
}

func DrawWrappedCircle(screen *ebiten.Image, pos core.Vector2, r float64, c color.RGBA, fill bool) {
	// Pre-calculating constants outside the loop minimizes redundant type-conversion overhead in the hot rendering path
	const sw, sh = float32(core.ScreenWidth), float32(core.ScreenHeight)
	rad := float32(r)

	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			x := float32(pos.X) + float32(ox)*sw
			y := float32(pos.Y) + float32(oy)*sh

			if !isVisible(x, y, rad*2, rad*2) { continue }

			if fill {
				vector.DrawFilledCircle(screen, x, y, rad, c, true)
			} else {
				vector.StrokeCircle(screen, x, y, rad, 2, c, true)
			}
		}
	}
}

func DrawWrappedSprite(screen *ebiten.Image, img *ebiten.Image, pos core.Vector2, rot float64, scale float64, clr color.RGBA) {
	w, h := img.Size()
	halfW, halfH := float64(w)/2, float64(h)/2
	
	const sw, sh = float32(core.ScreenWidth), float32(core.ScreenHeight)
	sizeW, sizeH := float32(float64(w)*scale), float32(float64(h)*scale)

	// Pre-configuring DrawOptions outside the wrap-loop minimizes heap allocations per frame
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest 
	op.ColorScale.ScaleWithColor(clr)

	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			x := float32(pos.X) + float32(ox)*sw
			y := float32(pos.Y) + float32(oy)*sh

			if !isVisible(x, y, sizeW, sizeH) { continue }

			op.GeoM.Reset()
			op.GeoM.Translate(-halfW, -halfH)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Rotate(rot)
			op.GeoM.Translate(float64(x), float64(y))
			
			screen.DrawImage(img, op)
		}
	}
}

func isVisible(x, y, w, h float32) bool {
	// Tight bounds checking eliminates wasted draw calls for entities outside the immediate viewport
	const sw, sh = float32(core.ScreenWidth), float32(core.ScreenHeight)
	return x+w/2 >= 0 && x-w/2 <= sw && y+h/2 >= 0 && y-h/2 <= sh
}





