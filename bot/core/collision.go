package core

import (
	"fmt"
)

func (self *Game) CheckEntityCollision(ship Ship, distance float64, degrees int, other Entity) bool {	// Would we hit some specific entity?

	const SAFETY_MARGIN = 0.001		// Needed to avoid floating point errors: the engine gives us data to 4 d.p.

	endx, endy := Projection(ship.X, ship.Y, distance, degrees)
	return IntersectSegmentCircle(ship.X, ship.Y, endx, endy, other.GetX(), other.GetY(), other.GetRadius() + SHIP_RADIUS + SAFETY_MARGIN)
}

func (self *Game) FirstCollision(ship Ship, distance float64, degrees int, possibles []Entity) (Entity, bool) {

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

func (self *Game) GetCourse(ship Ship, target Entity, avoid_list []Entity, side Side) (int, int, error) {
	return self.GetCourseRecursive(ship, target, avoid_list, 10, side)
}

func (self *Game) GetCourseRecursive(ship Ship, target Entity, avoid_list []Entity, depth int, side Side) (int, int, error) {		// speed, angle, error

	// Try to navigate to (collide with) the target, but avoiding the list of entites,
	// which could include the target.

	const (
		DODGE_MARGIN = 1.5
	)

	distance := ship.Dist(target)

	if distance < 0.5 {
		return 0, 0, nil
	}

	// If we're close, only allow whole number distances so we don't hit things due to rounding later.

	if distance < MAX_SPEED + 1 {
		distance = RoundToFloat(distance)
	}

	degrees := Angle(ship.X, ship.Y, target.GetX(), target.GetY())

	c, ok := self.FirstCollision(ship, distance, degrees, avoid_list)

	if ok == false {		// There is no collision
		speed := Min(Round(distance), MAX_SPEED)
		return speed, degrees, nil
	}

	if depth < 1 {
		return 0, 0, fmt.Errorf("GetCourseRecursive(): exceeded max depth")
	}

	// Reset our nav side iff the colliding object is a planet...

	if c.Type() == PLANET {
		side = DecideSide(ship, target, c)
	}

	var waypoint_angle int

	if side == RIGHT {
		waypoint_angle = degrees + 90
	} else {
		waypoint_angle = degrees - 90
	}
	waypointx, waypointy := Projection(c.GetX(), c.GetY(), c.GetRadius() + DODGE_MARGIN, waypoint_angle)
	return self.GetCourseRecursive(ship, Point{waypointx, waypointy}, avoid_list, depth - 1, side)
}

func (self *Game) GetApproach(ship Ship, target Entity, margin float64, avoid_list []Entity, side Side) (int, int, error) {

	// Navigate so that the ship's centre is definitely within <margin> of the target's edge.

	if ship.ApproachDist(target) < margin {
		return 0, 0, nil
	}

	// We add 0.51 in the calculation below to compensate for approximate navigation.
	// i.e. the GetCourseRecursive() call will put us within 0.5 of our requested point.

	travel_distance := ship.ApproachDist(target) + 0.51 - margin
	target_point_x, target_point_y := Projection(ship.X, ship.Y, travel_distance, ship.Angle(target))
	return self.GetCourse(ship, Point{target_point_x, target_point_y}, avoid_list, side)
}

// ---------------------------------------------------------------------

type Side int

const (
	LEFT Side = iota
	RIGHT
)

// Given a ship and some target, and some planet to navigate around,
// which side should we go?

func DecideSide(ship Ship, target Entity, planet Entity) Side {

	to_planet := ship.Angle(planet)
	to_target := ship.Angle(target)

	diff := to_planet - to_target

	if diff >= 0 && diff <= 180 {
		return LEFT
	}

	if diff >= 180 {
		return RIGHT
	}

	if diff <= -180 {
		return LEFT
	}

	if diff >= -180 && diff <= 0 {
		return RIGHT
	}

	BackendDevLog.Log("DecideSide() failed.")
	return RIGHT
}
