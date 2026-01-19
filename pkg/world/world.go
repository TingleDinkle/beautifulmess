package world

import (
	"beautifulmess/pkg/audio"
	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/particles"

	lua "github.com/yuin/gopher-lua"
)

type World struct {
	// Slices replace maps to provide cache-aligned, contiguous memory access, which is critical for high-performance ECS architectures
	Transforms       []*components.Transform
	Physics          []*components.Physics
	Renders          []*components.Render
	AIs              []*components.AI
	Tags             []*components.Tag
	GravityWells     []*components.GravityWell
	InputControlleds []*components.InputControlled
	Walls            []*components.Wall
	ProjectileEmitters []*components.ProjectileEmitter
	Lifetimes        []*components.Lifetime
	
	Particles *particles.ParticleSystem
	Audio     *audio.AudioSystem
	
	ScreenShake float64
	LState      *lua.LState
	nextID      core.Entity
}

func NewWorld() *World {
	w := &World{
		Particles: particles.NewParticleSystem(),
		Audio:     audio.NewAudioSystem(),
		LState:    lua.NewState(),
	}
	// Pre-allocating capacity reduces the frequency of heap re-allocations during the initial entity burst
	w.Reset()
	return w
}

func (w *World) Reset() {
	// Truncating slices while retaining underlying capacity allows for zero-allocation level resets
	w.Transforms = w.Transforms[:0]
	w.Physics = w.Physics[:0]
	w.Renders = w.Renders[:0]
	w.AIs = w.AIs[:0]
	w.Tags = w.Tags[:0]
	w.GravityWells = w.GravityWells[:0]
	w.InputControlleds = w.InputControlleds[:0]
	w.Walls = w.Walls[:0]
	w.ProjectileEmitters = w.ProjectileEmitters[:0]
	w.Lifetimes = w.Lifetimes[:0]
	
	w.nextID = 0
	w.ScreenShake = 0
}

func (w *World) CreateEntity() core.Entity {
	// Entity IDs map directly to slice indices, providing the fastest possible component lookup path
	id := w.nextID
	w.nextID++
	
	// Growing all slices in tandem maintains a uniform data-structure width
	w.Transforms = append(w.Transforms, nil)
	w.Physics = append(w.Physics, nil)
	w.Renders = append(w.Renders, nil)
	w.AIs = append(w.AIs, nil)
	w.Tags = append(w.Tags, nil)
	w.GravityWells = append(w.GravityWells, nil)
	w.InputControlleds = append(w.InputControlleds, nil)
	w.Walls = append(w.Walls, nil)
	w.ProjectileEmitters = append(w.ProjectileEmitters, nil)
	w.Lifetimes = append(w.Lifetimes, nil)
	
	return id
}

func (w *World) DestroyEntity(id core.Entity) {
	// Nullifying entries instead of resizing slices preserves the index-to-ID mapping stability
	idx := int(id)
	if idx >= len(w.Transforms) { return }
	
	w.Transforms[idx] = nil
	w.Physics[idx] = nil
	w.Renders[idx] = nil
	w.AIs[idx] = nil
	w.Tags[idx] = nil
	w.GravityWells[idx] = nil
	w.InputControlleds[idx] = nil
	w.Walls[idx] = nil
	w.ProjectileEmitters[idx] = nil
	w.Lifetimes[idx] = nil
}



