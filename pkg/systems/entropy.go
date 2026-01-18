package systems

import (
	"image"
	"math"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"
)

// SystemEntropy: Handles the fading trails.
func SystemEntropy(w *world.World, frostMask *image.RGBA) {
	// Heal the world (fade trails)
	pix := frostMask.Pix
	for i := 3; i < len(pix); i += 4 {
		a := int(pix[i])
		if a < 240 {
			pix[i] = uint8(a + 2)
		} // +2 Alpha per frame
	}

	// Create new scars (trails)
	for id, trans := range w.Transforms {
		if w.Renders[id] == nil {
			continue
		}

		// Scale world pos to mist resolution
		sx := int(trans.Position.X * (float64(core.MistWidth) / core.ScreenWidth))
		sy := int(trans.Position.Y * (float64(core.MistHeight) / core.ScreenHeight))

		rad := 2
		if tag, ok := w.Tags[id]; ok && tag.Name == "spectre" {
			rad = 3
		} // She leaves a bigger mark

		meltRetro(frostMask, sx, sy, rad, 100)
	}
}

func meltRetro(img *image.RGBA, cx, cy, r, val int) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	offsets := [][2]int{{0, 0}, {w, 0}, {-w, 0}, {0, h}, {0, -h}}

	for _, off := range offsets {
		tcx, tcy := cx+off[0], cy+off[1]
		if tcx+r < 0 || tcx-r >= w || tcy+r < 0 || tcy-r >= h {
			continue
		}

		for y := tcy - r; y <= tcy+r; y++ {
			for x := tcx - r; x <= tcx+r; x++ {
				if x < 0 || x >= w || y < 0 || y >= h {
					continue
				}
				if (x-tcx)*(x-tcx)+(y-tcy)*(y-tcy) <= r*r {
					i := img.PixOffset(x, y) + 3
					old := int(img.Pix[i])
					if old > 0 {
						img.Pix[i] = uint8(math.Max(0, float64(old-val)))
					}
				}
			}
		}
	}
}
