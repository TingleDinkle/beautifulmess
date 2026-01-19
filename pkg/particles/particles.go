package particles

import (
	"image/color"
	"math/rand"

	"beautifulmess/pkg/core"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Particle struct {
	Position core.Vector2
	Velocity core.Vector2
	Life     float64 // 0.0 to 1.0
	Decay    float64
	Color    color.RGBA
	Size     float64
}

type ParticleSystem struct {
	particles []*Particle
}

func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		particles: make([]*Particle, 0, 1000),
	}
}

func (ps *ParticleSystem) Reset() {
	ps.particles = ps.particles[:0]
}

func (ps *ParticleSystem) Emit(pos core.Vector2, vel core.Vector2, col color.RGBA, decay float64) {
	ps.particles = append(ps.particles, &Particle{
		Position: pos,
		Velocity: vel,
		Life:     1.0,
		Decay:    decay,
		Color:    col,
		Size:     rand.Float64()*2 + 1,
	})
}

func (ps *ParticleSystem) Update() {
	// Filter in place
	n := 0
	for _, p := range ps.particles {
		p.Life -= p.Decay
		if p.Life > 0 {
			p.Position.X += p.Velocity.X
			p.Position.Y += p.Velocity.Y
			
			// Optional: Drag
			p.Velocity.X *= 0.95
			p.Velocity.Y *= 0.95
			
			ps.particles[n] = p
			n++
		}
	}
	ps.particles = ps.particles[:n]
}

func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	for _, p := range ps.particles {
		// Alpha fade
		c := p.Color
		c.A = uint8(float64(c.A) * p.Life)
		
		DrawWrappedParticle(screen, p.Position, p.Size*p.Life, c)
	}
}

// DrawWrappedParticle is a simplified version of DrawWrappedCircle for particles
func DrawWrappedParticle(screen *ebiten.Image, pos core.Vector2, size float64, c color.RGBA) {
	x, y := float32(pos.X), float32(pos.Y)
	
	// Fast wrap check
	if x > 0 && x < float32(core.ScreenWidth) && y > 0 && y < float32(core.ScreenHeight) {
		vector.DrawFilledRect(screen, x, y, float32(size), float32(size), c, false)
		return
	}
	
	// Simplified wrapping (only checking immediate neighbors)
	for ox := -1.0; ox <= 1.0; ox++ {
		for oy := -1.0; oy <= 1.0; oy++ {
			wx := x + float32(ox*core.ScreenWidth)
			wy := y + float32(oy*core.ScreenHeight)
			
			if wx > -10 && wx < float32(core.ScreenWidth)+10 && wy > -10 && wy < float32(core.ScreenHeight)+10 {
				vector.DrawFilledRect(screen, wx, wy, float32(size), float32(size), c, false)
			}
		}
	}
}
