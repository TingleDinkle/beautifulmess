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
	for id, well := range w.GravityWells {
		if well == nil { continue }
		trans, ok := w.Transforms[id]
		if !ok { continue }
		
		pos := core.Vector2{X: trans.Position.X + shake.X, Y: trans.Position.Y + shake.Y}
		drawGravityWell(screen, pos, well.Radius)
	}

	// Memory node highlight relies on proximity-based feedback to indicate goal viability
	mc := color.RGBA{200, 200, 255, 100}
	if core.DistWrapped(spectrePos, lvl.Memory.Position) < core.MemoryRadius {
		mc = color.RGBA{255, 50, 50, 255}
	}
	
	pos := core.Vector2{X: lvl.Memory.Position.X + shake.X, Y: lvl.Memory.Position.Y + shake.Y}
	DrawWrappedCircle(screen, pos, core.MemoryRadius, mc, false)
}

func drawGravityWell(screen *ebiten.Image, pos core.Vector2, r float64) {
	// Dual-circle rendering conveys the 'Event Horizon' vs 'Singularity' distinction
	DrawWrappedCircle(screen, pos, r, color.RGBA{0, 0, 0, 255}, true)
	DrawWrappedCircle(screen, pos, r, color.RGBA{100, 0, 100, 255}, false)
}

func DrawEntities(screen *ebiten.Image, w *world.World, shake core.Vector2) {
	for id, r := range w.Renders {
		if r == nil || r.Sprite == nil { continue }
		trans, ok := w.Transforms[id]
		if !ok { continue }

		scale := r.Scale
		if scale == 0 { scale = 1.0 }
		
		pos := core.Vector2{X: trans.Position.X + shake.X, Y: trans.Position.Y + shake.Y}
		DrawWrappedSprite(screen, r.Sprite, pos, trans.Rotation, scale, r.Color)
	}
}

func DrawWrappedCircle(screen *ebiten.Image, pos core.Vector2, r float64, c color.RGBA, fill bool) {
	// Rendering multiple instances per entity preserves visual continuity across toroidal seams
	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			x := float32(pos.X + ox*core.ScreenWidth)
			y := float32(pos.Y + oy*core.ScreenHeight)

			if !isVisible(x, y, float32(r*2), float32(r*2)) { continue }

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
	
	sw, sh := float32(float64(w)*scale), float32(float64(h)*scale)

	// Reusing DrawImageOptions across neighbor instances prevents redundant heap allocations per frame
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest 
	op.ColorScale.ScaleWithColor(clr)

	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			x, y := pos.X+ox*core.ScreenWidth, pos.Y+oy*core.ScreenHeight

			if !isVisible(float32(x), float32(y), sw, sh) { continue }

			op.GeoM.Reset()
			op.GeoM.Translate(-halfW, -halfH)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Rotate(rot)
			op.GeoM.Translate(x, y)
			
			screen.DrawImage(img, op)
		}
	}
}


func isVisible(x, y, w, h float32) bool {
	// Conservative bounds checking prevents off-screen draw calls from reaching the GPU
	return x+w/2 >= 0 && x-w/2 <= core.ScreenWidth && y+h/2 >= 0 && y-h/2 <= core.ScreenHeight
}




