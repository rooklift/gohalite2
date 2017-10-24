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

	// Based on the Python version

	const (
		FUDGE = 0.5
	)

	dx := endx - startx
	dy := endy - starty

	a := dx * dx + dy * dy

	b := -2 * (startx * startx - startx * endx - startx * circlex + endx * circlex +
			  starty * starty - starty * endy - starty * circley + endy * circley)

	// c := (startx - circlex) * (startx - circlex) + (starty - circley) * (starty - circley)

	if a == 0.0 {
		return dist(startx, starty, circlex, circley) <= radius + FUDGE
	}

	// Time along segment when closest to the circle (vertex of the quadratic)
	t := min_float(-b / (2 * a), 1.0)
	if t < 0 {
		return false
	}

	closest_x := startx + dx * t
	closest_y := starty + dy * t

	return dist(closest_x, closest_y, circlex, circley) <= radius + FUDGE
}
