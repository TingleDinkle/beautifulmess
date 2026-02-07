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
	Position     core.Vector2
	Title        string
	Descriptions []string
	Color        color.RGBA
	Photos       []string
}

type WallDef struct {
	X, Y         float64
	Destructible bool
}

type Level struct {
	Name     string
	Wells    []GravityWell
	Walls    []WallDef
	Memory   MemoryNode
	StartP1  core.Vector2
	StartP2  core.Vector2
	Friction float64 // Friction override for specialized gameplay feel
}

func InitLevels() []Level {
	// Procedural generation helpers reduce boilerplate and ensure grid-alignment
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
		// 1. Where It All Began: The Spark (Normal Mechanics)
		{
			Name: "The Spark",
			Wells: []GravityWell{{Position: core.Vector2{X: 640, Y: 360}, Radius: 150, Mass: 5.0}},
			Memory: MemoryNode{
				Position: core.Vector2{X: 640, Y: 360},
				Title:    "Where It All Began",
				Descriptions: []string{
					"We met in high school during that mandatory military bootcamp trip. You told me later that you saw me three times before deciding it was fate. Funnily enough, the moment you actually came up to ask for my info was because I'd wandered into the wrong building after the ceremony. That's where you saw me for the fourth time and finally told your friend to approach me because you were too shy. We started talking and realized we have so much in common in our tastes, even if our hobbies couldn't be more different!",
					"The Yagi Storm was raging strong, but we hung out quite a lot and got to know each other way better. There was this unspoken chemistry that just worked despite us being so different—total opposites, really. Looking back, I realize the beauty was that we try to love each other in our own love languages, and that made us perfect for each other. Not perfect like matching puzzle pieces, but perfect because we are exactly what was missing in each other's life. I needed your warmth and nurturing nature, and you needed my devotion to you as my goddess. And as if blessed by fate, while the storm was raging, this beautiful piece of nature landed on me and gave us life where destruction was everywhere.",
					"This was the first time I introduced you to my friends at that birthday party. Everything went so well, and honestly, your beauty and scent just completely rocked me. It was the first real mark of a serious relationship for us. I felt like a new chapter in our life had finally started on such a positive note.",
				},
				Color:  color.RGBA{255, 100, 150, 255},
				Photos: []string{"assets/FIRSTMEET.jpg", "assets/LoveBug1.jpg", "assets/LoveBug2.jpg", "assets/firstintro1.jpg"},
			},
			StartP1:  core.Vector2{X: 100, Y: 360},
			StartP2:  core.Vector2{X: 1100, Y: 360},
			Friction: 0.94,
		},
		// 2. The Color of Your Soul: Kaleidoscope Twist (Many tiny wells)
		{
			Name: "Color of Your Soul",
			Wells: func() []GravityWell {
				var wells []GravityWell
				for i := 0; i < 15; i++ {
					wells = append(wells, GravityWell{
						Position: core.Vector2{X: 200 + rand.Float64()*880, Y: 100 + rand.Float64()*520},
						Radius:   25, Mass: 0.8,
					})
				}
				return wells
			}(),
			Walls: append(genLine(300, 100, 980, 100, false), genLine(300, 620, 980, 620, false)...),
			Memory: MemoryNode{
				Position: core.Vector2{X: 640, Y: 360},
				Title:    "Discovery",
				Descriptions: []string{
					"After that initial spark, I started discovering the world through your eyes. You aren't just 'a girl I met'; you're an artist of life. From your aesthetic to the way even a simple tea or a fresh day feels different with you. It’s when I realized your beauty wasn't just physical, but a whole vibe that started coloring my grey world.",
				},
				Color:  color.RGBA{150, 255, 150, 255},
				Photos: []string{"assets/floweigirl.jpg", "assets/floweigirlteainspo.jpg", "assets/mint.jpg"},
			},
			StartP1:  core.Vector2{X: 100, Y: 100},
			StartP2:  core.Vector2{X: 1180, Y: 620},
			Friction: 0.92,
		},
		// 3. The Muffin Chapter: Catnip Twist (Wells in the face)
		{
			Name: "The Muffin Chapter",
			Wells: []GravityWell{
				{Position: core.Vector2{X: 520, Y: 350}, Radius: 40, Mass: 1.5}, // Left Eye Well
				{Position: core.Vector2{X: 760, Y: 350}, Radius: 40, Mass: 1.5}, // Right Eye Well
				{Position: core.Vector2{X: 640, Y: 420}, Radius: 30, Mass: 1.0}, // Nose Well
			},
			Walls: func() []WallDef {
				var walls []WallDef
				// Left Ear
				walls = append(walls, genLine(450, 250, 500, 150, false)...)
				walls = append(walls, genLine(500, 150, 550, 250, false)...)
				// Right Ear
				walls = append(walls, genLine(730, 250, 780, 150, false)...)
				walls = append(walls, genLine(780, 150, 830, 250, false)...)
				// Head Outline (Top)
				walls = append(walls, genLine(550, 250, 730, 250, false)...)
				// Cheeks
				walls = append(walls, genLine(450, 250, 400, 400, false)...)
				walls = append(walls, genLine(830, 250, 880, 400, false)...)
				// Chin
				walls = append(walls, genLine(400, 400, 640, 600, false)...)
				walls = append(walls, genLine(880, 400, 640, 600, false)...)
				// Whiskers
				walls = append(walls, genLine(350, 380, 250, 350, true)...)
				walls = append(walls, genLine(350, 400, 250, 400, true)...)
				walls = append(walls, genLine(930, 380, 1030, 350, true)...)
				walls = append(walls, genLine(930, 400, 1030, 400, true)...)
				return walls
			}(),
			Memory: MemoryNode{
				Position: core.Vector2{X: 640, Y: 420},
				Title:    "Our Little Family",
				Descriptions: []string{
					"Our first step into 'forever' wasn't a contract; it was a cat. Adopting our little muffin, Tonton. Seeing you nurture this tiny creature made me realize how big your heart is. We weren't just two people anymore; we were a little family. You became a mom to this fluffball, and I realized I wanted to protect this home we were building together.",
				},
				Color:  color.RGBA{255, 200, 100, 255},
				Photos: []string{"assets/tonton1.jpg", "assets/tonton2.jpg", "assets/tonton3.jpg", "assets/tonton4.jpg", "assets/tonton5.jpg"},
			},
			StartP1:  core.Vector2{X: 640, Y: 100},
			StartP2:  core.Vector2{X: 640, Y: 650},
			Friction: 0.95,
		},
		// 4. The Beautiful Mess: Explosive Chaos (Checkerboard grid)
		{
			Name:  "The Beautiful Mess",
			Wells: []GravityWell{{Position: core.Vector2{X: 640, Y: 360}, Radius: 100, Mass: 3.0}},
			Walls: func() []WallDef {
				var walls []WallDef
				// Checkerboard pattern reduces entity count by 50% while looking "messy"
				for x := 0; x < 42; x++ {
					for y := 0; y < 24; y++ {
						if (x+y)%2 == 0 {
							walls = append(walls, WallDef{X: float64(x*30 + 15), Y: float64(y*30 + 15), Destructible: true})
						}
					}
				}
				return walls
			}(),
			Memory: MemoryNode{
				Position: core.Vector2{X: 640, Y: 360},
				Title:    "Pure Comfort",
				Descriptions: []string{
					"This is the 'Mess' part of us. The unfiltered, goofy, and sometimes 'gross' comfort of a real relationship. From the inside jokes to those massive food comas. It’s the beauty of being able to be our absolute weirdest selves without a single drop of judgment. I love our mess.",
				},
				Color:  color.RGBA{255, 100, 100, 255},
				Photos: []string{"assets/forkU.jpg", "assets/burrito.jpg", "assets/ToeSuckah.jpg", "assets/passedawazoo.jpg"},
			},
			StartP1:  core.Vector2{X: 100, Y: 100},
			StartP2:  core.Vector2{X: 1180, Y: 620},
			Friction: 0.90, // Heavy feel
		},
		// 5. Grounded in the Storm: Hurricane Twist (Corner wells pushing in)
		{
			Name: "Grounded in the Storm",
			Wells: []GravityWell{
				{Position: core.Vector2{X: 0, Y: 0}, Radius: 200, Mass: 4.0},
				{Position: core.Vector2{X: 1280, Y: 0}, Radius: 200, Mass: 4.0},
				{Position: core.Vector2{X: 0, Y: 720}, Radius: 200, Mass: 4.0},
				{Position: core.Vector2{X: 1280, Y: 720}, Radius: 200, Mass: 4.0},
			},
			Walls: append(genLine(640, 0, 640, 300, false), genLine(640, 420, 640, 720, false)...),
			Memory: MemoryNode{
				Position: core.Vector2{X: 640, Y: 360},
				Title:    "Our Sanctuary",
				Descriptions: []string{
					"The world can be destructive, but we found warmth in the cold. Whether it was literally hugging a tree or creating a magical reality to escape into, we became each other's sanctuary. You are my safe place when things get hard.",
				},
				Color:  color.RGBA{100, 100, 255, 255},
				Photos: []string{"assets/hugtree.jpg", "assets/warmth.jpg", "assets/unicorn.jpg"},
			},
			StartP1:  core.Vector2{X: 100, Y: 360},
			StartP2:  core.Vector2{X: 1180, Y: 360},
			Friction: 0.94,
		},
		// 6. The Constant Duo: Orbits Twist (Low friction spinning)
		{
			Name: "The Constant Duo",
			Wells: []GravityWell{
				{Position: core.Vector2{X: 640, Y: 360}, Radius: 100, Mass: 5.0}, // Center anchor
			},
			Memory: MemoryNode{
				Position: core.Vector2{X: 640, Y: 360},
				Title:    "The One Constant",
				Descriptions: []string{
					"Different dates, different outfits, different years—but the same 'Duo.' We’ve changed, grown, and prospered, but every day reinforces that we are the one constant in each other's lives. No matter where we go, we go together.",
				},
				Color:  color.RGBA{200, 200, 255, 255},
				Photos: []string{"assets/duo.jpg", "assets/duo2.jpg", "assets/duo3.jpg", "assets/duo4.jpg"},
			},
			StartP1:  core.Vector2{X: 640, Y: 100},
			StartP2:  core.Vector2{X: 640, Y: 620},
			Friction: 0.99, // Orbital feel
		},
		// 7. The Magnum Opus: Event Horizon Twist (Indestructible wall gap)
		{
			Name:  "The Magnum Opus",
			Wells: []GravityWell{{Position: core.Vector2{X: 640, Y: 360}, Radius: 300, Mass: 8.0}},
			Walls: func() []WallDef {
				var walls []WallDef
				// Indestructible diamond shield
				walls = append(walls, genLine(640, 200, 800, 360, false)...)
				walls = append(walls, genLine(800, 360, 640, 520, false)...)
				walls = append(walls, genLine(640, 520, 480, 360, false)...)
				// The only gap is at the top left
				walls = append(walls, genLine(480, 360, 600, 240, false)...)
				return walls
			}(),
			Memory: MemoryNode{
				Position: core.Vector2{X: 640, Y: 360},
				Title:    "My Goddess",
				Descriptions: []string{
					"I see you as my 'Magnum Opus'—the greatest thing I've ever had the privilege to be part of. Whether you're just 'sitting kewt' or being your radiant self, you are my goddess. This is the peak of everything we've built.",
				},
				Color:  color.RGBA{255, 255, 100, 255},
				Photos: []string{"assets/magnumOpus.jpg", "assets/kewtcrunch.jpg", "assets/sitkewt.jpg"},
			},
			StartP1:  core.Vector2{X: 100, Y: 100},
			StartP2:  core.Vector2{X: 1180, Y: 100},
			Friction: 0.94,
		},
		// 8. Interlinked: Zero State Twist (Inevitable pull)
		{
			Name: "Interlinked",
			Wells: []GravityWell{
				{Position: core.Vector2{X: 640, Y: 360}, Radius: 50, Mass: 15.0}, // Absolute pull
			},
			Memory: MemoryNode{
				Position: core.Vector2{X: 640, Y: 360},
				Title:    "Zero State",
				Descriptions: []string{
					"No more noise. No more storms. Just 'Interlinked.' Like two souls that have finally found their frequency. We drift together in total peace. We were exactly what was missing in each other's life. Silence, at last, because words can't describe this anymore. We just are.",
				},
				Color:  color.RGBA{255, 255, 255, 255},
				Photos: []string{"assets/interlinked.jpg"},
			},
			StartP1:  core.Vector2{X: 200, Y: 360},
			StartP2:  core.Vector2{X: 1080, Y: 360},
			Friction: 0.96,
		},
	}
}




