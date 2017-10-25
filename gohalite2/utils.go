package gohalite2

import (
	"math"
)

func max(a, b int) int {
	if a > b { return a }
	return b
}

func min_float(a, b float64) float64 {
	if a < b { return a }
	return b
}

func dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx * dx + dy * dy)
}

func intersect_segment_circle(startx, starty, endx, endy, circlex, circley, radius float64) bool {

	// Based on the Python version, I have no idea how this works.

	const (
		FUDGE = SHIP_RADIUS		// We almost certainly want this.
	)

	dx := endx - startx
	dy := endy - starty

	a := dx * dx + dy * dy

	b := -2 * (startx * startx - startx * endx - startx * circlex + endx * circlex +
			  starty * starty - starty * endy - starty * circley + endy * circley)

	// This is in the Python code, but is unused:
	// c := (startx - circlex) * (startx - circlex) + (starty - circley) * (starty - circley)

	if a == 0.0 {
		return dist(startx, starty, circlex, circley) <= radius + FUDGE
	}

	t := min_float(-b / (2 * a), 1.0)

	if t < 0 {
		return false
	}

	closest_x := startx + dx * t
	closest_y := starty + dy * t

	return dist(closest_x, closest_y, circlex, circley) <= radius + FUDGE
}

func projection(x1, y1, distance float64, degrees int) (x2, y2 float64) {

	// Given a coordinate, a distance and an angle, find a new coordinate.

	if distance == 0 {
		return x1, y1
	}

	radians := deg_to_rad(float64(degrees))

	x2 = distance * math.Cos(radians) + x1
	y2 = distance * math.Sin(radians) + y1

	return x2, y2
}

func deg_to_rad(d float64) float64 {
	return d / 180 * math.Pi
}

func rad_to_deg(r float64) float64 {
	return r / math.Pi * 180
}