package core

import "math"

func DistWrapped(a, b Vector2) float64 {
	dx := math.Abs(a.X - b.X)
	dy := math.Abs(a.Y - b.Y)
	if dx > ScreenWidth/2 {
		dx = ScreenWidth - dx
	}
	if dy > ScreenHeight/2 {
		dy = ScreenHeight - dy
	}
	return math.Sqrt(dx*dx + dy*dy)
}
