package core

import "math"

func DistWrapped(a, b Vector2) float64 {
	// Wrapped distance calculation supports the non-Euclidean toroidal geometry of the game world
	d := VecToWrapped(a, b)
	return math.Sqrt(d.X*d.X + d.Y*d.Y)
}

func VecToWrapped(from, to Vector2) Vector2 {
	// Calculating the shortest path across boundaries allows AI and physics to ignore the coordinate discontinuity
	dx := to.X - from.X
	dy := to.Y - from.Y

	if dx > ScreenWidth/2 { dx -= ScreenWidth }
	if dx < -ScreenWidth/2 { dx += ScreenWidth }
	if dy > ScreenHeight/2 { dy -= ScreenHeight }
	if dy < -ScreenHeight/2 { dy += ScreenHeight }
	return Vector2{dx, dy}
}

