package audio

import (
	"io"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const SampleRate = 44100

type AudioSystem struct {
	Context *audio.Context
	Sounds  map[string]*audio.Player
}

func NewAudioSystem() *AudioSystem {
	// A shared audio context prevents resource contention across multiple sound sources
	ctx := audio.NewContext(SampleRate)
	as := &AudioSystem{
		Context: ctx,
		Sounds:  make(map[string]*audio.Player),
	}

	as.generateInternalSounds()
	return as
}

func (as *AudioSystem) LoadFile(name, path string) {
	// Pre-loading assets into memory avoids disk I/O latency during time-critical game events
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Audio Load Error: %v", err)
		return
	}
	defer f.Close()

	d, err := wav.DecodeWithSampleRate(SampleRate, f)
	if err != nil {
		log.Printf("WAV Decode Error: %v", err)
		return
	}

	b, err := io.ReadAll(d)
	if err != nil {
		log.Printf("Audio Read Error: %v", err)
		return
	}

	as.Sounds[name] = as.Context.NewPlayerFromBytes(b)
}

func (as *AudioSystem) Play(name string) {
	if p, ok := as.Sounds[name]; ok {
		if !p.IsPlaying() {
			p.Rewind()
			p.Play()
		}
	}
}

func (as *AudioSystem) generateInternalSounds() {
	// Synthetic sound generation provides a zero-dependency fallback for core game feedback
	as.Sounds["boost"] = as.createPlayerFromBytes(genNoise(0.2))
	as.Sounds["chime"] = as.createPlayerFromBytes(genSine(880, 0.5))
	as.Sounds["drone"] = as.createPlayerFromBytes(genSine(110, 2.0))
	as.Sounds["spectre_dash"] = as.createPlayerFromBytes(genBreathyNoise(0.5))
}

func (as *AudioSystem) createPlayerFromBytes(b []byte) *audio.Player {
	return as.Context.NewPlayerFromBytes(b)
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

