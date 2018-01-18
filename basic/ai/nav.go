package ai

import (
	"fmt"
)

var Ignore_Collision_Dist float64 = 999		// Default means no collision ignored when pathing. But reduced by bot (100 default in bot).

func FirstCollision(ship *Ship, distance float64, degrees int, possibles []Entity) (Entity, bool) {

	var collisions []Entity

	for _, other := range possibles {
		if CheckEntityCollision(ship, distance, degrees, other) {
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

func GetCourse(ship *Ship, target Entity, avoid_list []Entity, side Side) (int, int, error) {
	return GetCourseRecursive(ship, target, avoid_list, 10, side)
}

func GetCourseRecursive(ship *Ship, target Entity, avoid_list []Entity, depth int, side Side) (int, int, error) {

	// Try to navigate to (collide with) the target, but avoiding the list of entites,
	// which could include the target. Returns: speed, angle, error

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

	c, ok := FirstCollision(ship, distance, degrees, avoid_list)

	if ok == false || ship.ApproachDist(c) > Ignore_Collision_Dist {			// There is no collision... or it's miles away (fixes replay 7710319)
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
	p := &Point{waypointx, waypointy}

	return GetCourseRecursive(ship, p, avoid_list, depth - 1, side)
}

func GetApproach(ship *Ship, target Entity, margin float64, avoid_list []Entity, side Side) (int, int, error) {

	// Navigate so that the ship's centre is definitely within <margin> of the target's edge.

	if ship.ApproachDist(target) < margin {
		return 0, 0, nil
	}

	// We add 0.51 in the calculation below to compensate for approximate navigation.
	// i.e. the GetCourseRecursive() call will put us within 0.5 of our requested point.

	travel_distance := ship.ApproachDist(target) + 0.51 - margin
	target_point_x, target_point_y := Projection(ship.X, ship.Y, travel_distance, ship.Angle(target))

	p := &Point{target_point_x, target_point_y}

	return GetCourse(ship, p, avoid_list, side)
}
