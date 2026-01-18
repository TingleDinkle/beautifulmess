package components

import (
	"image/color"

	"beautifulmess/pkg/core"

	"github.com/hajimehoshi/ebiten/v2"
)

type Transform struct {
	Position core.Vector2
	Rotation float64
}

type Physics struct {
	Velocity, Acceleration   core.Vector2
	MaxSpeed, Friction, Mass float64
	GravityMultiplier        float64
}

type Render struct {
	Sprite *ebiten.Image
	Color  color.RGBA
	Glow   bool
}

type AI struct {
	ScriptName string
	TargetID   int
}

type Tag struct {
	Name string
}

type GravityWell struct {
	Radius float64
	Mass   float64
}

type InputControlled struct{}
