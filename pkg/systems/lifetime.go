package systems

import (
	"beautifulmess/pkg/core"
	"beautifulmess/pkg/world"
)

func SystemLifetime(w *world.World) {
	dt := 1.0 / 60.0 

	for id, life := range w.Lifetimes {
		if life == nil { continue }
		
		life.TimeRemaining -= dt
		if life.TimeRemaining <= 0 {
			// Automated cleanup prevents memory fragmentation and logic leaks over time
			w.DestroyEntity(core.Entity(id))
		}
	}
}

