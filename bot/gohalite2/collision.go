package gohalite2

import (
	"fmt"
)

func (self *Game) CheckEntityCollision(ship Ship, distance float64, degrees int, other Entity) bool {

	// Would we hit some specific entity?

	endx, endy := Projection(ship.X, ship.Y, distance, degrees)

	if IntersectSegmentCircle(ship.X, ship.Y, endx, endy, other.GetX(), other.GetY(), other.GetRadius() + SHIP_RADIUS) {
		return true
	}

	return false
}

func (self *Game) Collision(ship Ship, distance float64, degrees int, possibles []Entity) (Entity, bool) {

	var collisions []Entity

	for _, other := range possibles {
		if self.CheckEntityCollision(ship, distance, degrees, other) {
			collisions = append(collisions, other)
		}
	}

	if len(collisions) == 0 {
		return nil, false
	}

	var closest_ent Entity
	var closest_distance float64 = 999999.9

	for _, c := range collisions {
		if ship.Dist(c) < closest_distance {
			closest_ent = c
			closest_distance = ship.Dist(c)
		}
	}

	return closest_ent, true
}

func (self *Game) GetCourse(ship Ship, target Entity, avoid []Entity) (int, int, error) {
	return self.GetCourseRecursive(ship, target, avoid, 10)
}

func (self *Game) GetCourseRecursive(ship Ship, target Entity, avoid []Entity, depth int) (int, int, error) {		// speed, angle, error

	const (
		DODGE_MARGIN = 1.5
	)

	distance := ship.Dist(target)

	if distance < 0.5 {
		return 0, 0, nil
	}

	degrees := Angle(ship.X, ship.Y, target.GetX(), target.GetY())

	c, does_hit := self.Collision(ship, distance, degrees, avoid)

	if does_hit == false {
		speed := Min(Round(distance), MAX_SPEED)
		return speed, degrees, nil
	}

	if depth > 0 {
		waypointx, waypointy := Projection(c.GetX(), c.GetY(), c.GetRadius() + DODGE_MARGIN, degrees + 90)
		return self.GetCourseRecursive(ship, Point{waypointx, waypointy}, avoid, depth - 1)
	}

	return 0, 0, fmt.Errorf("GetCourseRecursive(): exceeded max depth")
}

func (self *Game) GetApproach(ship Ship, target Entity, margin float64, avoid []Entity) (int, int, error) {

	// Navigate so that the ship's centre comes near the target's edge. Target
	// can be a Planet or a Ship (or a Point).

	current_dist := ship.Dist(target)

	if current_dist < target.GetRadius() + margin {
		return 0, 0, nil
	}

	direct_angle := Angle(ship.X, ship.Y, target.GetX(), target.GetY())

	needed_distance := current_dist - target.GetRadius() - margin
	target_point_x, target_point_y := Projection(ship.X, ship.Y, needed_distance, direct_angle)

	return self.GetCourse(ship, Point{target_point_x, target_point_y}, avoid)
}
