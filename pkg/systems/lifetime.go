package systems

import (
	"beautifulmess/pkg/world"
)

func SystemLifetime(w *world.World) {
	dt := 1.0 / 60.0 // Fixed time step assumption

	for id, life := range w.Lifetimes {
		life.TimeRemaining -= dt
		if life.TimeRemaining <= 0 {
			// Destroy Entity
			delete(w.Lifetimes, id)
			delete(w.Transforms, id)
			delete(w.Physics, id)
			delete(w.Renders, id)
			delete(w.Tags, id)
			delete(w.GravityWells, id)
			// Remove from other maps if necessary
		}
	}
}
