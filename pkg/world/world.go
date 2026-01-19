package world

import (
	"beautifulmess/pkg/audio"
	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/particles"

	lua "github.com/yuin/gopher-lua"
)

type World struct {
	// Pointer-slices provide O(1) access with safe nil-validation for stable IDs
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
	
	// Active lists allow systems to skip empty slots, maintaining high ALU throughput
	ActiveEntities []core.Entity 
	ActiveWalls    []core.Entity

	// A Spatial Hash Grid optimizes static geometry queries to O(1) neighborhood checks
	Grid [13][8][]core.Entity 

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
	// Slice truncation retains capacity to eliminate heap churn during level resets
	w.Transforms, w.Physics, w.Renders = w.Transforms[:0], w.Physics[:0], w.Renders[:0]
	w.AIs, w.Tags, w.GravityWells = w.AIs[:0], w.Tags[:0], w.GravityWells[:0]
	w.InputControlleds, w.Walls = w.InputControlleds[:0], w.Walls[:0]
	w.ProjectileEmitters, w.Lifetimes = w.ProjectileEmitters[:0], w.Lifetimes[:0]
	
	w.ActiveEntities = w.ActiveEntities[:0]
	w.ActiveWalls = w.ActiveWalls[:0]
	
	for x := 0; x < 13; x++ {
		for y := 0; y < 8; y++ {
			w.Grid[x][y] = w.Grid[x][y][:0]
		}
	}
	
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
	w.ActiveWalls = append(w.ActiveWalls, id)
}

func (w *World) DestroyEntity(id core.Entity) {
	idx := int(id)
	if idx >= len(w.Transforms) { return }
	
	w.Transforms[idx], w.Physics[idx], w.Renders[idx] = nil, nil, nil
	w.AIs[idx], w.Tags[idx], w.GravityWells[idx] = nil, nil, nil
	w.InputControlleds[idx], w.Walls[idx] = nil, nil
	w.ProjectileEmitters[idx], w.Lifetimes[idx] = nil, nil

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

func (w *World) UpdateGrid() {
	for x := 0; x < 13; x++ {
		for y := 0; y < 8; y++ {
			w.Grid[x][y] = w.Grid[x][y][:0]
		}
	}

	for _, id := range w.ActiveWalls {
		trans := w.Transforms[id]
		if trans == nil { continue }
		
		gx, gy := int(trans.Position.X/100), int(trans.Position.Y/100)
		if gx >= 0 && gx < 13 && gy >= 0 && gy < 8 {
			w.Grid[gx][gy] = append(w.Grid[gx][gy], id)
		}
	}
}
