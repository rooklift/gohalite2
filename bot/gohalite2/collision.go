package gohalite2

import (
	"fmt"
)

func (self *Game) AngleCollisionID(startx, starty, distance float64, degrees int) int {

	// Returns ID of the first planet we would collide with, or -1 if no hits

	var collision_planets []Planet
	all_planets := self.AllPlanets()

	endx, endy := Projection(startx, starty, distance, degrees)

	for _, planet := range all_planets {

		collides := IntersectSegmentCircle(startx, starty, endx, endy, planet.X, planet.Y, planet.Radius + SHIP_RADIUS)

		if collides {
			collision_planets = append(collision_planets, planet)
		}
	}

	if len(collision_planets) == 0 {
		return -1
	}

	closest_planet := Planet{}
	closest_distance := 999999.9

	for _, c := range collision_planets {
		if Dist(startx, starty, c.X, c.Y) < closest_distance {
			closest_planet = c
			closest_distance = Dist(startx, starty, c.X, c.Y)
		}
	}

	return closest_planet.Id
}

func (self *Game) Navigate(x1, y1, x2, y2 float64) (int, int, error) {
	return self.NavigateRecursive(x1, y1, x2, y2, 10)
}

func (self *Game) NavigateRecursive(x1, y1, x2, y2 float64, depth int) (int, int, error) {		// speed, angle, error

	// Navigate around planets (only).

	const (
		SAFETY_MARGIN = 2
	)

	distance := Dist(x1, y1, x2, y2)

	if distance < 0.5 {
		return 0, 0, nil
	}

	degrees := Angle(x1, y1, x2, y2)

	colliding_planet_id := self.AngleCollisionID(x1, y1, distance, degrees)

	if colliding_planet_id == -1 {

		speed := Round(distance)
		if speed > MAX_SPEED {
			speed = MAX_SPEED
		}

		return speed, degrees, nil
	}

	if depth > 0 {

		planet := self.GetPlanet(colliding_planet_id)
		waypointx, waypointy := Projection(planet.X, planet.Y, planet.Radius + SAFETY_MARGIN, degrees + 90)
		return self.NavigateRecursive(x1, y1, waypointx, waypointy, depth - 1)

	}

	return 0, 0, fmt.Errorf("NavigateRecursive(): exceeded max depth")
}

func (self *Game) Approach(ship Ship, target Entity, margin float64) (int, int, error) {

	// Navigate so that the ship's centre comes near the target's edge. Target
	// can be a Planet or a Ship (or a Point).

	current_dist := Dist(ship.X, ship.Y, target.GetX(), target.GetY())

	if current_dist < target.GetRadius() + margin {
		return 0, 0, nil
	}

	direct_angle := Angle(ship.X, ship.Y, target.GetX(), target.GetY())

	needed_distance := current_dist - target.GetRadius() - margin
	target_point_x, target_point_y := Projection(ship.X, ship.Y, needed_distance, direct_angle)

	return self.Navigate(ship.X, ship.Y, target_point_x, target_point_y)
}
