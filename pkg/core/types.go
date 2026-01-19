package core

// Constants
const (
	ScreenWidth   = 1280
	ScreenHeight  = 720
	TimeStep      = 1.0 / 60.0 // Fixed delta-time ensures deterministic physics across hardware
	EntitySize    = 8.0
	MemoryRadius  = 70.0
	MistWidth     = 320
	MistHeight    = 180
)

// Vector2
type Vector2 struct {
	X, Y float64
}

// Entity ID
type Entity int
