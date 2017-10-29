package gohalite2

import (
	"math"
	"strconv"
	"strings"
)

func IntersectSegmentCircle(startx, starty, endx, endy, circlex, circley, radius float64) bool {

	// Based on the Python version, I have no idea how this works.

	dx := endx - startx
	dy := endy - starty

	a := dx * dx + dy * dy

	b := -2 * (startx * startx - startx * endx - startx * circlex + endx * circlex +
	           starty * starty - starty * endy - starty * circley + endy * circley)

	if a == 0.0 {
		return Dist(startx, starty, circlex, circley) <= radius
	}

	t := MinFloat(-b / (2 * a), 1.0)

	if t < 0 {
		return false
	}

	closest_x := startx + dx * t
	closest_y := starty + dy * t

	return Dist(closest_x, closest_y, circlex, circley) <= radius
}

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

func Dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx * dx + dy * dy)
}

func CourseFromString(s string) (int, int) {

	tokens := strings.Fields(s)

	if len(tokens) != 4 {			// t sid speed angle
		return 0, 0
	}

	if tokens[0] != "t" {
		return 0, 0
	}

	speed, err1 := strconv.Atoi(tokens[2])
	degrees, err2 := strconv.Atoi(tokens[3])

	if err1 != nil || err2 != nil {
		return 0, 0
	}

	for degrees < 0 {
		degrees += 360
	}

	degrees %= 360

	return speed, degrees
}

func GetOrderType(s string) OrderType {

	tokens := strings.Fields(s)

	if len(tokens) == 0 {
		return NO_ORDER
	}
	if tokens[0] == "t" {
		return THRUST
	}
	if tokens[0] == "d" {
		return DOCK
	}
	if tokens[0] == "u" {
		return UNDOCK
	}

	return NO_ORDER
}
