package ai

import (
	"math"
)

func CheckEntityCollision(ship *Ship, distance float64, degrees int, other Entity) bool {
	endx, endy := Projection(ship.X, ship.Y, distance, degrees)
	return IntersectSegmentCircle(ship.X, ship.Y, endx, endy, other.GetX(), other.GetY(), other.GetRadius() + SHIP_RADIUS)
}

func ShipsWillCollide(ship_a *Ship, speed_a, angle_a int, ship_b *Ship, speed_b, angle_b int) bool {

	// Work this out by pretending ship B is standing still, while ship A is moving possibly faster than allowed.

	radians_a := DegToRad(float64(angle_a))
	speedx_a := float64(speed_a) * math.Cos(radians_a)
	speedy_a := float64(speed_a) * math.Sin(radians_a)

	radians_b := DegToRad(float64(angle_b))
	speedx_b := float64(speed_b) * math.Cos(radians_b)
	speedy_b := float64(speed_b) * math.Sin(radians_b)

	startx := ship_a.X
	starty := ship_a.Y

	endx_adjusted := startx + speedx_a - speedx_b
	endy_adjusted := starty + speedy_a - speedy_b

	return IntersectSegmentCircle(ship_a.X, ship_a.Y, endx_adjusted, endy_adjusted, ship_b.X, ship_b.Y, SHIP_RADIUS * 2)
}

func IntersectSegmentCircle(startx, starty, endx, endy, circlex, circley, radius float64) bool {

	// Based on the Python version, I have no idea how this works.
	// "Mathematics not Zathras skill"

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
