package core

// Fohristiwhirl, November 2017.
// A sort of alternative starter kit.

import (
	"math"
)

func Projection(x1, y1, distance float64, degrees int) (x2, y2 float64) {

	// Given a coordinate, a distance and an angle, find a new coordinate.

	if distance == 0 {
		return x1, y1
	}

	radians := DegToRad(float64(degrees))

	x2 = distance * math.Cos(radians) + x1
	y2 = distance * math.Sin(radians) + y1

	return x2, y2
}

func Angle(x1, y1, x2, y2 float64) int {

	rad := math.Atan2(y2 - y1, x2 - x1)
	deg := RadToDeg(rad)

	deg_int := Round(deg)

	for deg_int < 0 {
		deg_int += 360
	}

	return deg_int % 360
}

func DegToRad(d float64) float64 {
	return d / 180 * math.Pi
}

func RadToDeg(r float64) float64 {
	return r / math.Pi * 180
}

func Max(a, b int) int {
	if a > b { return a }
	return b
}

func Min(a, b int) int {
	if a < b { return a }
	return b
}

func MaxFloat(a, b float64) float64 {
	if a > b { return a }
	return b
}

func MinFloat(a, b float64) float64 {
	if a < b { return a }
	return b
}

func Round(n float64) int {
	return int(math.Floor(n + 0.5))
}

func RoundToFloat(n float64) float64 {
	return math.Floor(n + 0.5)
}

func Dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx * dx + dy * dy)
}
