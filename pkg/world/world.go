package world

import (
	"beautifulmess/pkg/audio"
	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/particles"

	lua "github.com/yuin/gopher-lua"
)

type World struct {
	// Slices replace maps to provide cache-aligned, contiguous memory access
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
	
	// Active lists allow systems to skip empty 'nil' slots in component slices, maximizing CPU throughput
	ActiveEntities []core.Entity 
	ActiveWalls    []core.Entity

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
	w.Reset()
	return w
}

func (w *World) Reset() {
	// Truncating while keeping capacity minimizes future heap allocations
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
	
	w.ActiveEntities = w.ActiveEntities[:0]
	w.ActiveWalls = w.ActiveWalls[:0]
	
	w.nextID = 0
	w.ScreenShake = 0
}

func (w *World) CreateEntity() core.Entity {
	id := w.nextID
	w.nextID++
	
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
	
	w.ActiveEntities = append(w.ActiveEntities, id)
	return id
}

func (w *World) AddToActiveWalls(id core.Entity) {
	// Specialized lists for static geometry optimize collision and rendering subsystems
	w.ActiveWalls = append(w.ActiveWalls, id)
}

func (w *World) DestroyEntity(id core.Entity) {
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

	// Removal from active lists prevents systems from visiting nullified component slots
	for i, eid := range w.ActiveEntities {
		if eid == id {
			w.ActiveEntities[i] = w.ActiveEntities[len(w.ActiveEntities)-1]
			w.ActiveEntities = w.ActiveEntities[:len(w.ActiveEntities)-1]
			break
		}
	}
	for i, eid := range w.ActiveWalls {
		if eid == id {
			w.ActiveWalls[i] = w.ActiveWalls[len(w.ActiveWalls)-1]
			w.ActiveWalls = w.ActiveWalls[:len(w.ActiveWalls)-1]
			break
		}
	}
}




