package systems

import (
	"image"
	"math"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"
)

func SystemEntropy(w *world.World, frostMask *image.RGBA) {
	// Periodic alpha increment simulates the universe 'healing' or forgetting entity paths
	pix := frostMask.Pix
	for i := 3; i < len(pix); i += 4 {
		if pix[i] < 240 {
			pix[i] = uint8(int(pix[i]) + 2)
		}
	}

	for id, trans := range w.Transforms {
		if w.Renders[id] == nil { continue }

		// Coordinate remapping translates world-space motion into low-res texture memory
		sx := int(trans.Position.X * (float64(core.MistWidth) / core.ScreenWidth))
		sy := int(trans.Position.Y * (float64(core.MistHeight) / core.ScreenHeight))

		rad := 2
		if tag, ok := w.Tags[id]; ok && tag.Name == "spectre" {
			rad = 3 // Larger trails emphasize the spectre's heavy presence
		}

		meltRetro(frostMask, sx, sy, rad, 100)
	}
}

func meltRetro(img *image.RGBA, cx, cy, r, val int) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	
	// 9-way neighbor check ensures trails wrap seamlessly across toroidal boundaries
	offsets := [][2]int{
		{0, 0}, {w, 0}, {-w, 0}, {0, h}, {0, -h},
		{w, h}, {w, -h}, {-w, h}, {-w, -h},
	}

	for _, off := range offsets {
		tcx, tcy := cx+off[0], cy+off[1]
		if tcx+r < 0 || tcx-r >= w || tcy+r < 0 || tcy-r >= h { continue }

		for y := tcy - r; y <= tcy+r; y++ {
			for x := tcx - r; x <= tcx+r; x++ {
				if x < 0 || x >= w || y < 0 || y >= h { continue }
				
				if (x-tcx)*(x-tcx)+(y-tcy)*(y-tcy) <= r*r {
					idx := img.PixOffset(x, y) + 3
					// Alpha reduction creates the 'melt' effect against the background
					img.Pix[idx] = uint8(math.Max(0, float64(int(img.Pix[idx])-val)))
				}
			}
		}
	}
}
