package audio

import (
	"io"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const SampleRate = 44100

type AudioSystem struct {
	Context *audio.Context
	Pools   map[string][]*audio.Player // Player pooling prevents heap fragmentation and resource exhaustion
}

func NewAudioSystem() *AudioSystem {
	// A shared audio context prevents resource contention across multiple sound sources
	ctx := audio.NewContext(SampleRate)
	as := &AudioSystem{
		Context: ctx,
	}
	as.init()
	return as
}

func (as *AudioSystem) init() {
	as.Pools = make(map[string][]*audio.Player)
	// Internal sounds are synthesized once and pooled to ensure low-latency feedback fallbacks
	as.addPool("boost", genBlitz(0.3))
	as.addPool("chime", genSine(880, 0.5))
	as.addPool("drone", genSine(110, 2.0))
	as.addPool("spectre_dash", genBreathyNoise(0.5))
}

func (as *AudioSystem) addPool(name string, b []byte) {
	// A pool size of 8 provides sufficient polyphony for overlapping arcade effects without bloat
	const poolSize = 8
	pool := make([]*audio.Player, poolSize)
	for i := 0; i < poolSize; i++ {
		pool[i] = as.Context.NewPlayerFromBytes(b)
	}
	as.Pools[name] = pool
}

func (as *AudioSystem) LoadFile(name, path string) {
	f, err := os.Open(path)
	if err != nil { return }
	defer f.Close()

	d, err := wav.DecodeWithSampleRate(SampleRate, f)
	if err != nil { return }

	b, err := io.ReadAll(d)
	if err != nil { return }

	as.addPool(name, b)
}

func (as *AudioSystem) Play(name string) {
	// Round-robin selection provides a simple, lock-free way to achieve polyphony
	if pool, ok := as.Pools[name]; ok {
		for _, p := range pool {
			if !p.IsPlaying() {
				p.Rewind()
				p.Play()
				return
			}
		}
		// If all players are busy, stealing the oldest (0) ensures feedback continuity
		pool[0].Rewind()
		pool[0].Play()
	}
}

func (as *AudioSystem) SetVolume(v float64) {
	for _, pool := range as.Pools {
		for _, p := range pool {
			p.SetVolume(v)
		}
	}
}


func genBlitz(duration float64) []byte {
	// A sub-bass 'thump' combined with a clean aerodynamic sweep simulates extreme velocity displacement
	length := int(duration * SampleRate)
	b := make([]byte, length*2)
	
	for i := 0; i < length; i++ {
		t := float64(i) / SampleRate
		
		// The 'thump': A low-frequency sine wave provides the physical impact felt at the start of the dash
		thump := math.Sin(2 * math.Pi * 60.0 * t) * math.Exp(-t * 20.0)
		
		// The 'sweep': A clean frequency descent simulates the sound of air being cut at high speed
		sweepFreq := 800.0 * math.Exp(-t * 10.0) + 100.0
		sweep := math.Sin(2 * math.Pi * sweepFreq * t)
		
		// A highly smoothed envelope prevents auditory fatigue and ensures a 'pleasant' arcade finish
		env := math.Exp(-t * 6.0)
		if t < 0.03 { env = t / 0.03 } 
		
		// Mixing the thump and sweep creates a multi-layered, professional-grade 'blitz' effect
		val := (thump * 0.6) + (sweep * 0.4)
		s := int16(val * env * 0.3 * 32767)
		b[2*i], b[2*i+1] = byte(s), byte(s >> 8)
	}
	return b
}

func genNoise(duration float64) []byte {
	length := int(duration * SampleRate)
	b := make([]byte, length*2)
	lp, alpha := 0.0, 0.15 
	for i := 0; i < length; i++ {
		white := (rand.Float64()*2 - 1)
		lp = lp + alpha*(white-lp)
		v := lp * 0.15 * math.Exp(-float64(i) / (float64(length) * 0.3))
		s := int16(v * 32767)
		b[2*i], b[2*i+1] = byte(s), byte(s >> 8)
	}
	return b
}

func genBreathyNoise(duration float64) []byte {
	length := int(duration * SampleRate)
	b := make([]byte, length*2)
	lp, alpha := 0.0, 0.05
	for i := 0; i < length; i++ {
		white := (rand.Float64()*2 - 1)
		lp = lp + alpha*(white-lp)
		t := float64(i) / SampleRate
		// Linear envelope provides a softer, more organic attack/decay for spirit-like sounds
		env := 0.0
		if t < duration*0.2 { env = t / (duration * 0.2) } else { env = 1.0 - (t-duration*0.2)/(duration*0.8) }
		s := int16(lp * 0.3 * env * 32767)
		b[2*i], b[2*i+1] = byte(s), byte(s >> 8)
	}
	return b
}

func genSine(freq float64, duration float64) []byte {
	length := int(duration * SampleRate)
	b := make([]byte, length*2)
	for i := 0; i < length; i++ {
		t := float64(i) / SampleRate
		v := (math.Sin(2*math.Pi*freq*t) + 0.3*math.Sin(4*math.Pi*freq*t)) * 0.2
		// Exponential decay simulates natural acoustic damping
		env := 1.0
		if t < 0.02 { env = t / 0.02 } else { env = math.Exp(-(t - 0.02) / (duration * 0.2)) }
		s := int16(v * env * 32767)
		b[2*i], b[2*i+1] = byte(s), byte(s >> 8)
	}
	return b
}

