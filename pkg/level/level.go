package level

import (
	"image/color"
	"beautifulmess/pkg/core"
)

type GravityWell struct {
	Position core.Vector2
	Radius   float64
	Mass     float64
}

type MemoryNode struct {
	Position    core.Vector2
	Title       string
	Description string
	Color       color.RGBA
}

type Level struct {
	Name    string
	Wells   []GravityWell
	Memory  MemoryNode
	StartP1 core.Vector2
	StartP2 core.Vector2
}

func InitLevels() []Level {
	return []Level{
		// Level 1: Singularity
		{
			Name: "Event Horizon",
			Wells: []GravityWell{
				{Position: core.Vector2{X: 640, Y: 360}, Radius: 70, Mass: 2.0},
			},
			Memory: MemoryNode{
				Position:    core.Vector2{X: 640, Y: 360},
				Title:       "The Singularity",
				Description: `We were crushed together.
Finally one.`,
				Color:       color.RGBA{100, 100, 255, 255},
			},
			StartP1: core.Vector2{X: 100, Y: 360},
			StartP2: core.Vector2{X: 1100, Y: 360},
		},

		// Level 2: Binary
		{
			Name: "Binary Star",
			Wells: []GravityWell{
				{Position: core.Vector2{X: 300, Y: 360}, Radius: 80, Mass: 2.0},
				{Position: core.Vector2{X: 980, Y: 360}, Radius: 80, Mass: 2.0},
			},
			Memory: MemoryNode{
				Position:    core.Vector2{X: 300, Y: 360},
				Title:       "Orbit Decay",
				Description: "Spinning until we crash.",
				Color:       color.RGBA{255, 50, 50, 255},
			},
			StartP1: core.Vector2{X: 640, Y: 600},
			StartP2: core.Vector2{X: 640, Y: 100},
		},
	}
}
