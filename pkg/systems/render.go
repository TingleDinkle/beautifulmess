package systems

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"os"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/level"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func LoadSpectreSet(normal, angy, kewt string) map[string]*ebiten.Image {
	set := make(map[string]*ebiten.Image)
	set["normal"] = LoadAndProcessSpectre(normal)
	set["angy"] = LoadAndProcessSpectre(angy)
	set["kewt"] = LoadAndProcessSpectre(kewt)
	return set
}

func LoadAndProcessSpectre(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("failed to open %s: %v", path, err)
		return nil
	}
	defer f.Close()

	srcImg, _, err := image.Decode(f)
	if err != nil {
		log.Printf("failed to decode %s: %v", path, err)
		return nil
	}

	const targetRes = 128
	res := image.NewRGBA(image.Rect(0, 0, targetRes, targetRes))
	srcB := srcImg.Bounds()
	sw, sh := float64(srcB.Dx()), float64(srcB.Dy())
	
	const referenceSize = 600.0
	fitScale := float64(targetRes) / referenceSize
	destW, destH := sw*fitScale, sh*fitScale
	offsetX, offsetY := (float64(targetRes)-destW)/2.0, (float64(targetRes)-destH)/2.0

	for y := 0; y < targetRes; y++ {
		glitchOffset := 0.0
		if rand.Float64() < 0.005 { glitchOffset = rand.NormFloat64() * 2.0 }

		for x := 0; x < targetRes; x++ {
			srcX := (float64(x) - offsetX + glitchOffset) / fitScale
			srcY := (float64(y) - offsetY) / fitScale

			if srcX < 0 || srcY < 0 || srcX >= sw || srcY >= sh { continue }

			c := color.RGBAModel.Convert(srcImg.At(int(srcX)+srcB.Min.X, int(srcY)+srcB.Min.Y)).(color.RGBA)
			if c.A < 10 { continue }
			
			lum := 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
			r, g, bl := float64(c.R), float64(c.G), float64(c.B)
			
			if lum < 128 { r *= 1.1; g *= 0.8; bl *= 0.8 }
			if y%2 == 0 { r *= 0.92; g *= 0.92; bl *= 0.92 }

			nx := (float64(x) - targetRes/2.0) / (targetRes / 2.0)
			ny := (float64(y) - targetRes/2.0) / (targetRes / 2.0)
			dist := math.Sqrt(nx*nx + ny*ny)
			
			alpha := float64(c.A)
			if dist > 0.4 {
				fade := 1.0 - (dist-0.4)/0.6
				if fade < 0 { fade = 0 }
				alpha *= fade * fade
			}

			res.Set(x, y, color.RGBA{
				R: uint8(math.Min(255, r)),
				G: uint8(math.Min(255, g)),
				B: uint8(math.Min(255, bl)),
				A: uint8(alpha),
			})
		}
	}
	return ebiten.NewImageFromImage(res)
}

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
	
	const scrW, scrH = float32(core.ScreenWidth), float32(core.ScreenHeight)
	sizeW, sizeH := float32(float64(w)*scale), float32(float64(h)*scale)

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest 
	op.ColorScale.ScaleWithColor(clr)

	// If the entity is far from the world edges, we bypass the 9-way wrap loop to save CPU/GPU cycles
	const margin = 100.0
	if pos.X > margin && pos.X < float64(core.ScreenWidth)-margin && pos.Y > margin && pos.Y < float64(core.ScreenHeight)-margin {
		op.GeoM.Translate(-halfW, -halfH)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Rotate(rot)
		op.GeoM.Translate(pos.X, pos.Y)
		screen.DrawImage(img, op)
		return
	}

	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			x, y := float32(pos.X)+float32(ox)*scrW, float32(pos.Y)+float32(oy)*scrH

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





