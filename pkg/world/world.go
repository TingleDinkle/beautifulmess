package world

import (
	"beautifulmess/pkg/components"
	"beautifulmess/pkg/core"
)

type World struct {
	Transforms map[core.Entity]*components.Transform
	Physics    map[core.Entity]*components.Physics
	Renders    map[core.Entity]*components.Render
	AIs        map[core.Entity]*components.AI
	Tags       map[core.Entity]*components.Tag
	nextID     core.Entity
}

func NewWorld() *World {
	return &World{
		Transforms: make(map[core.Entity]*components.Transform),
		Physics:    make(map[core.Entity]*components.Physics),
		Renders:    make(map[core.Entity]*components.Render),
		AIs:        make(map[core.Entity]*components.AI),
		Tags:       make(map[core.Entity]*components.Tag),
	}
}

func (w *World) CreateEntity() core.Entity {
	id := w.nextID
	w.nextID++
	return id
}
