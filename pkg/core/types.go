package core

// Constants
const (
	ScreenWidth   = 1280
	ScreenHeight  = 720
	EntitySize    = 8.0
	MemoryRadius  = 70.0 // The size of the "Trap" / "Goal"
	MistWidth     = 320  // Low-res buffer for that retro feel
	MistHeight    = 180
)

// Vector2
type Vector2 struct {
	X, Y float64
}

// Entity ID
type Entity int
