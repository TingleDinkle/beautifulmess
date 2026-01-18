package components

import (
	"image/color"

	"beautifulmess/pkg/core"

	"github.com/hajimehoshi/ebiten/v2"
	lua "github.com/yuin/gopher-lua"
)

type Transform struct {
	Position core.Vector2
	Rotation float64
}

type Physics struct {
	Velocity, Acceleration core.Vector2
	MaxSpeed, Friction, Mass float64
}

type Render struct {
	Sprite *ebiten.Image
	Color  color.RGBA
	Glow   bool
}

type AI struct {
	LState     *lua.LState
	ScriptPath string
	TargetID   int
}

type Tag struct {
	Name string
}
