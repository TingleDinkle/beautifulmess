package particles

import (
	"image/color"
	"math"
	"math/rand"

	"beautifulmess/pkg/core"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ParticleQuirk int

const (
	QuirkStandard ParticleQuirk = iota
	QuirkOrbit                  // Particles that spiral around their center point
	QuirkFlicker                // Particles that change size/alpha rapidly
)

type Particle struct {
	Position core.Vector2
	Velocity core.Vector2
	Life     float64 
	Decay    float64
	Color    color.RGBA
	Size     float64
	Quirk    ParticleQuirk
	Angle    float64 // Used for orbital or rotational quirks
}

type ParticleSystem struct {
	particles []*Particle
	pool      []*Particle // A managed pool prevents allocation-heavy GC spikes during massive shatters
}

func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		particles: make([]*Particle, 0, 2000),
		pool:      make([]*Particle, 0, 2000),
	}
}

func (ps *ParticleSystem) Reset() {
	// Returning active particles to the pool maintains zero-alloc level transitions
	for _, p := range ps.particles {
		ps.pool = append(ps.pool, p)
	}
	ps.particles = ps.particles[:0]
}

func (ps *ParticleSystem) Emit(pos core.Vector2, vel core.Vector2, col color.RGBA, decay float64) {
	ps.EmitAdvanced(pos, vel, col, decay, QuirkStandard)
}

func (ps *ParticleSystem) EmitAdvanced(pos core.Vector2, vel core.Vector2, col color.RGBA, decay float64, quirk ParticleQuirk) {
	var p *Particle
	if len(ps.pool) > 0 {
		p = ps.pool[len(ps.pool)-1]
		ps.pool = ps.pool[:len(ps.pool)-1]
	} else {
		p = &Particle{}
	}

	p.Position = pos
	p.Velocity = vel
	p.Life = 1.0
	p.Decay = decay
	p.Color = col
	p.Size = rand.Float64()*2 + 1
	p.Quirk = quirk
	p.Angle = rand.Float64() * math.Pi * 2

	ps.particles = append(ps.particles, p)
}

func (ps *ParticleSystem) Update() {
	n := 0
	for _, p := range ps.particles {
		p.Life -= p.Decay
		if p.Life > 0 {
			// Quirky movement patterns provide an organic, 'hand-crafted' feel to the digital debris
			switch p.Quirk {
			case QuirkOrbit:
				p.Angle += 0.2
				p.Position.X += p.Velocity.X + math.Cos(p.Angle)*2.0
				p.Position.Y += p.Velocity.Y + math.Sin(p.Angle)*2.0
			case QuirkFlicker:
				if rand.Float64() < 0.3 { p.Size = rand.Float64() * 4.0 }
				fallthrough
			default:
				p.Position.X += p.Velocity.X
				p.Position.Y += p.Velocity.Y
			}
			
			// Aerodynamic drag prevents particles from drifting infinitely
			p.Velocity.X *= 0.96
			p.Velocity.Y *= 0.96
			
			ps.particles[n] = p
			n++
		} else {
			// Recycling dead particles eliminates the need for fresh heap allocations
			ps.pool = append(ps.pool, p)
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
