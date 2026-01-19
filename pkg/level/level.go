package level

import (
	"image/color"
	"math"
	"math/rand"
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
	Photos      []string
}

type WallDef struct {
	X, Y         float64
	Destructible bool
}

type Level struct {
	Name    string
	Wells   []GravityWell
	Walls   []WallDef
	Memory  MemoryNode
	StartP1 core.Vector2
	StartP2 core.Vector2
}

func InitLevels() []Level {
	// Procedural generation helpers reduce boilerplate and ensure grid-alignment
	genGrid := func(startX, startY, w, h, step int, dest bool) []WallDef {
		var walls []WallDef
		for i := 0; i < w; i++ {
			for j := 0; j < h; j++ {
				walls = append(walls, WallDef{X: float64(startX + i*step), Y: float64(startY + j*step), Destructible: dest})
			}
		}
		return walls
	}

	genLine := func(x1, y1, x2, y2 int, dest bool) []WallDef {
		var walls []WallDef
		steps := int(math.Max(math.Abs(float64(x2-x1)), math.Abs(float64(y2-y1))) / 10)
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(steps)
			walls = append(walls, WallDef{X: float64(x1) + t*float64(x2-x1), Y: float64(y1) + t*float64(y2-y1), Destructible: dest})
		}
		return walls
	}

	return []Level{
		// 1. Terminal Velocity: Focuses on orbital mechanics and momentum conservation
		{
			Name: "Terminal Velocity",
			Wells: []GravityWell{{Position: core.Vector2{X: 640, Y: 360}, Radius: 150, Mass: 5.0}},
			Memory: MemoryNode{
				Position:    core.Vector2{X: 640, Y: 360},
				Title:       "The Singularity",
				Description: "We fell too fast.\nGravity took the wheel.",
				Color:       color.RGBA{100, 100, 255, 255},
				Photos:      []string{"p1_1", "photo1_2"},
			},
			StartP1: core.Vector2{X: 100, Y: 360},
			StartP2: core.Vector2{X: 1100, Y: 360},
		},
		// 2. Data Fragmentation: Introduces terraforming through destruction
		{
			Name: "Data Fragmentation",
			Wells: []GravityWell{{Position: core.Vector2{X: 640, Y: 360}, Radius: 60, Mass: 2.0}},
			Walls: func() []WallDef {
				w := genGrid(200, 100, 88, 52, 10, true)
				var filtered []WallDef
				for _, wall := range w {
					if core.DistWrapped(core.Vector2{X: wall.X, Y: wall.Y}, core.Vector2{X: 640, Y: 360}) > 120 {
						filtered = append(filtered, wall)
					}
				}
				return filtered
			}(),
			Memory: MemoryNode{
				Position:    core.Vector2{X: 640, Y: 360},
				Title:       "Fragmentation",
				Description: "Pieces of us scattered.\nI had to break everything to find you.",
				Color:       color.RGBA{255, 150, 50, 255},
				Photos:      []string{"p2_1"},
			},
			StartP1: core.Vector2{X: 50, Y: 360},
			StartP2: core.Vector2{X: 1230, Y: 360},
		},
		// 3. Packet Loss: Linear precision challenge
		{
			Name: "Packet Loss",
			Wells: []GravityWell{
				{Position: core.Vector2{X: 300, Y: 90}, Radius: 40, Mass: 1.5},
				{Position: core.Vector2{X: 900, Y: 630}, Radius: 40, Mass: 1.5},
			},
			Walls: append(genLine(0, 240, 1280, 240, false), genLine(0, 480, 1280, 480, false)...),
			Memory: MemoryNode{
				Position:    core.Vector2{X: 640, Y: 360},
				Title:       "Parallel Lines",
				Description: "We were running parallel.\nNever crossing.",
				Color:       color.RGBA{50, 200, 50, 255},
				Photos:      []string{"p3_1"},
			},
			StartP1: core.Vector2{X: 50, Y: 120},
			StartP2: core.Vector2{X: 1230, Y: 600},
		},
		// 4. Static Field: Unpredictable vector field navigation
		{
			Name: "Static Field",
			Wells: func() []GravityWell {
				var wells []GravityWell
				for i := 0; i < 12; i++ {
					wells = append(wells, GravityWell{
						Position: core.Vector2{X: float64(100 + rand.Intn(1080)), Y: float64(100 + rand.Intn(520))},
						Radius:   30, Mass: 1.0,
					})
				}
				return wells
			}(),
			Memory: MemoryNode{
				Position:    core.Vector2{X: 640, Y: 360},
				Title:       "Noise",
				Description: "Too much noise.\nI couldn't hear you calling.",
				Color:       color.RGBA{200, 200, 200, 255},
				Photos:      []string{"p4_1"},
			},
			StartP1: core.Vector2{X: 640, Y: 50},
			StartP2: core.Vector2{X: 640, Y: 670},
		},
		// 5. Firewall Breach: Puzzle-based environmental manipulation
		{
			Name: "Firewall Breach",
			Wells: []GravityWell{{Position: core.Vector2{X: 300, Y: 360}, Radius: 100, Mass: 3.0}},
			Walls: append(append(genLine(540, 260, 740, 260, true), genLine(540, 460, 740, 460, true)...),
				append(genLine(540, 260, 540, 460, true), genLine(740, 260, 740, 460, true)...)...),
			Memory: MemoryNode{
				Position:    core.Vector2{X: 640, Y: 360},
				Title:       "Firewall",
				Description: "You built walls to keep me out.\nI tore them down to keep you in.",
				Color:       color.RGBA{255, 50, 50, 255},
				Photos:      []string{"p5_1"},
			},
			StartP1: core.Vector2{X: 100, Y: 360},
			StartP2: core.Vector2{X: 640, Y: 360},
		},
		// 6. Eclipse: Physics exploitation using Lagrange points
		{
			Name: "Eclipse",
			Wells: []GravityWell{
				{Position: core.Vector2{X: 440, Y: 360}, Radius: 80, Mass: 2.5},
				{Position: core.Vector2{X: 840, Y: 360}, Radius: 80, Mass: 2.5},
			},
			Walls: append(genLine(640, 260, 640, 460, false), genLine(600, 360, 680, 360, false)...),
			Memory: MemoryNode{
				Position:    core.Vector2{X: 640, Y: 360},
				Title:       "Eclipse",
				Description: "Caught between two stars.\nBurnt by both.",
				Color:       color.RGBA{255, 255, 100, 255},
				Photos:      []string{"p6_1"},
			},
			StartP1: core.Vector2{X: 640, Y: 100},
			StartP2: core.Vector2{X: 640, Y: 620},
		},
		// 7. System Failure: Inverts safety zones by weaponizing the wrap mechanic
		{
			Name: "System Failure",
			Wells: []GravityWell{
				{Position: core.Vector2{X: 0, Y: 0}, Radius: 100, Mass: 3.0},
				{Position: core.Vector2{X: 1280, Y: 0}, Radius: 100, Mass: 3.0},
				{Position: core.Vector2{X: 0, Y: 720}, Radius: 100, Mass: 3.0},
				{Position: core.Vector2{X: 1280, Y: 720}, Radius: 100, Mass: 3.0},
			},
			Walls: append(genLine(300, 0, 300, 720, false), genLine(980, 0, 980, 720, false)...),
			Memory: MemoryNode{
				Position:    core.Vector2{X: 640, Y: 360},
				Title:       "Edge Case",
				Description: "There was no way out.\nThe edges were fraying.",
				Color:       color.RGBA{200, 50, 200, 255},
				Photos:      []string{"p7_1"},
			},
			StartP1: core.Vector2{X: 640, Y: 300},
			StartP2: core.Vector2{X: 640, Y: 420},
		},
		// 8. The Void: Final confrontation in a decaying grid
		{
			Name: "The Void",
			Wells: []GravityWell{{Position: core.Vector2{X: 640, Y: 360}, Radius: 50, Mass: 1.0}},
			Walls: genGrid(100, 100, 10, 5, 100, true),
			Memory: MemoryNode{
				Position:    core.Vector2{X: 640, Y: 360},
				Title:       "Zero State",
				Description: "Silence at last.\nWe drift together.",
				Color:       color.RGBA{255, 255, 255, 255},
				Photos:      []string{"p8_1"},
			},
			StartP1: core.Vector2{X: 200, Y: 360},
			StartP2: core.Vector2{X: 1080, Y: 360},
		},
	}
}

