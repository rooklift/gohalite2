package gohalite2

import (
	"fmt"
)

func (self *Game) AngleCollisionID(ship Ship, distance float64, degrees int) int {

	// Returns ID of the first planet we would collide with, or -1 if no hits

	var collision_planets []Planet
	all_planets := self.AllPlanets()

	endx, endy := Projection(ship.X, ship.Y, distance, degrees)

	for _, planet := range all_planets {

		collides := IntersectSegmentCircle(ship.X, ship.Y, endx, endy, planet.X, planet.Y, planet.Radius + SHIP_RADIUS)

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
		if ship.Dist(c) < closest_distance {
			closest_planet = c
			closest_distance = ship.Dist(c)
		}
	}

	return closest_planet.Id
}

func (self *Game) GetCourse(ship Ship, target Entity) (int, int, error) {
	return self.GetCourseRecursive(ship, target, 10)
}

func (self *Game) GetCourseRecursive(ship Ship, target Entity, depth int) (int, int, error) {		// speed, angle, error

	// Navigate around planets (only).
	// If the target is in fact a planet, we don't navigate round it, but are happy to collide.

	const (
		SAFETY_MARGIN = 2		// For side waypoints only.
	)

	distance := ship.Dist(target)

	if distance < 0.5 {
		return 0, 0, nil
	}

	x1, y1, x2, y2 := ship.X, ship.Y, target.GetX(), target.GetY()
	degrees := Angle(x1, y1, x2, y2)

	colliding_planet_id := self.AngleCollisionID(ship, distance, degrees)

	if colliding_planet_id == -1 {
		speed = Min(Round(distance), MAX_SPEED)
		return speed, degrees, nil
	}

	if target_as_planet, ok := target.(Planet); ok {
		if target_as_planet.Id == colliding_planet_id {
			speed = Min(Round(distance), MAX_SPEED)
			return speed, degrees, nil
		}
	}

	if depth > 0 {

		planet := self.GetPlanet(colliding_planet_id)
		waypointx, waypointy := Projection(planet.X, planet.Y, planet.Radius + SAFETY_MARGIN, degrees + 90)
		return self.GetCourseRecursive(ship, Point{waypointx, waypointy}, depth - 1)

	}

	return 0, 0, fmt.Errorf("NavigateRecursive(): exceeded max depth")
}

func (self *Game) GetApproach(ship Ship, target Entity, margin float64) (int, int, error) {

	// Navigate so that the ship's centre comes near the target's edge. Target
	// can be a Planet or a Ship (or a Point).

	current_dist := ship.Dist(target)

	if current_dist < target.GetRadius() + margin {
		return 0, 0, nil
	}

	direct_angle := Angle(ship.X, ship.Y, target.GetX(), target.GetY())

	needed_distance := current_dist - target.GetRadius() - margin
	target_point_x, target_point_y := Projection(ship.X, ship.Y, needed_distance, direct_angle)

	return self.GetCourse(ship, Point{target_point_x, target_point_y})
}
