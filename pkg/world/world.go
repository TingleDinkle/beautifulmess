package world

import (
	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"

	lua "github.com/yuin/gopher-lua"
)

type World struct {
	Transforms map[core.Entity]*components.Transform
	Physics    map[core.Entity]*components.Physics
	Renders    map[core.Entity]*components.Render
	AIs        map[core.Entity]*components.AI
	Tags       map[core.Entity]*components.Tag
	GravityWells map[core.Entity]*components.GravityWell
	InputControlleds map[core.Entity]*components.InputControlled
	LState     *lua.LState
	nextID     core.Entity
}

func NewWorld() *World {
	return &World{
		Transforms: make(map[core.Entity]*components.Transform),
		Physics:    make(map[core.Entity]*components.Physics),
		Renders:    make(map[core.Entity]*components.Render),
		AIs:        make(map[core.Entity]*components.AI),
		Tags:       make(map[core.Entity]*components.Tag),
		GravityWells: make(map[core.Entity]*components.GravityWell),
		InputControlleds: make(map[core.Entity]*components.InputControlled),
		LState:     lua.NewState(),
	}
}

func (w *World) CreateEntity() core.Entity {
	id := w.nextID
	w.nextID++
	return id
}
