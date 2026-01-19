package components

import (
	"image/color"

	"beautifulmess/pkg/core"

	"github.com/hajimehoshi/ebiten/v2"
)

type Transform struct {
	Position core.Vector2
	Rotation float64 // Radians are used to maintain compatibility with standard trigonometric functions
}

type Physics struct {
	Velocity, Acceleration   core.Vector2
	MaxSpeed, Friction, Mass float64
	GravityMultiplier        float64 // Allows entities to react differently to the curvature of space
}

type Render struct {
	Sprite *ebiten.Image
	Color  color.RGBA
	Glow   bool
	Scale  float64 // Non-zero scale values enable resolution-independent sprite sizing
}

type AI struct {
	ScriptName string
	TargetID   int
}

type Tag struct {
	Name string // String tags facilitate data-driven logic without hardcoded type-checking
}

type GravityWell struct {
	Radius float64
	Mass   float64
}

type InputControlled struct{} // Marker component delegates entity control to the input system

type Wall struct {
	Size         float64
	Destructible bool
	IsDestroyed  bool
}

type ProjectileEmitter struct {
	Interval float64 // Fixed-step intervals ensure deterministic fire rates across different hardware
	LastTime float64
}

type Lifetime struct {
	TimeRemaining float64
}

