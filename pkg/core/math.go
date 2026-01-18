package core

import "math"

// DistWrapped computes the Euclidean distance between two points
// while accounting for the toroidal (wrapped) screen topology.
func DistWrapped(a, b Vector2) float64 {
	d := VecToWrapped(a, b)
	return math.Sqrt(d.X*d.X + d.Y*d.Y)
}

// VecToWrapped calculates the shortest vector from one point to another
// across the screen boundaries, essential for physics and AI to perceive
// the world as a continuous torus.
func VecToWrapped(from, to Vector2) Vector2 {
	dx := to.X - from.X
	dy := to.Y - from.Y

	if dx > ScreenWidth/2 {
		dx -= ScreenWidth
	}
	if dx < -ScreenWidth/2 {
		dx += ScreenWidth
	}
	if dy > ScreenHeight/2 {
		dy -= ScreenHeight
	}
	if dy < -ScreenHeight/2 {
		dy += ScreenHeight
	}
	return Vector2{dx, dy}
}
