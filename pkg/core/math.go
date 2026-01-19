package core

import "math"

func DistWrapped(a, b Vector2) float64 {
	// Standard distance is kept for narrative logic where absolute units are required
	return math.Sqrt(DistSqWrapped(a, b))
}

func DistSqWrapped(a, b Vector2) float64 {
	// Squared distance avoids the computationally expensive Square Root operation during hot-path proximity checks
	d := VecToWrapped(a, b)
	return d.X*d.X + d.Y*d.Y
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

func WrapPosition(p *Vector2) {
	// Toroidal wrapping ensures that the coordinate space remains finite but boundless
	if p.X < 0 { p.X += ScreenWidth }
	if p.X >= ScreenWidth { p.X -= ScreenWidth }
	if p.Y < 0 { p.Y += ScreenHeight }
	if p.Y >= ScreenHeight { p.Y -= ScreenHeight }
}


