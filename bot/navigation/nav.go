package navigation

import (
	"fmt"

	hal "../core"
)

// In all calls, the NavStacker is just used for dumping info into its NavStack for debugging...

type NavStacker interface {
	AddToNavStack(format_string string, args ...interface{})
}

type NullNavStacker struct {}

func (self *NullNavStacker) AddToNavStack() {}

// ------------------------------------------------------------------------------------------------------------------------------------------

func CheckEntityCollision(ship *hal.Ship, distance float64, degrees int, other hal.Entity) bool {		// Would we hit some specific entity?

	const SAFETY_MARGIN = 0.00001

	endx, endy := hal.Projection(ship.X, ship.Y, distance, degrees)
	return hal.IntersectSegmentCircle(ship.X, ship.Y, endx, endy, other.GetX(), other.GetY(), other.GetRadius() + hal.SHIP_RADIUS + SAFETY_MARGIN)
}

func FirstCollision(ship *hal.Ship, distance float64, degrees int, possibles []hal.Entity) (hal.Entity, bool) {

	var collisions []hal.Entity

	for _, other := range possibles {
		if CheckEntityCollision(ship, distance, degrees, other) {
			collisions = append(collisions, other)
		}
	}

	if len(collisions) == 0 {
		return nil, false
	}

	var closest_ent hal.Entity
	var closest_distance float64 = 999999.9

	for _, c := range collisions {
		if ship.Dist(c) < closest_distance {
			closest_ent = c
			closest_distance = ship.Dist(c)
		}
	}

	return closest_ent, true
}

func GetCourse(ship *hal.Ship, target hal.Entity, avoid_list []hal.Entity, side Side, ns NavStacker) (int, int, error) {
	return GetCourseRecursive(ship, target, avoid_list, 10, side, ns)
}

func GetCourseRecursive(ship *hal.Ship, target hal.Entity, avoid_list []hal.Entity, depth int, side Side, ns NavStacker) (int, int, error) {

	// Try to navigate to (collide with) the target, but avoiding the list of entites,
	// which could include the target. Returns: speed, angle, error

	const (
		DODGE_MARGIN = 1.5
	)

	distance := ship.Dist(target)

	if distance < 0.5 {
		ns.AddToNavStack("GetCourseRecursive(): returning null move")
		return 0, 0, nil
	}

	// If we're close, only allow whole number distances so we don't hit things due to rounding later.

	if distance < hal.MAX_SPEED + 1 {
		distance = hal.RoundToFloat(distance)
	}

	degrees := hal.Angle(ship.X, ship.Y, target.GetX(), target.GetY())

	c, ok := FirstCollision(ship, distance, degrees, avoid_list)

	if ok == false {		// There is no collision
		speed := hal.Min(hal.Round(distance), hal.MAX_SPEED)
		ns.AddToNavStack("GetCourseRecursive(): succeeded with %v %v", speed, degrees)
		return speed, degrees, nil
	}

	if depth < 1 {
		ns.AddToNavStack("GetCourseRecursive(): exceeded max depth")
		return 0, 0, fmt.Errorf("GetCourseRecursive(): exceeded max depth")
	}

	// Reset our nav side iff the colliding object is a planet...

	if c.Type() == hal.PLANET {
		side = DecideSide(ship, target, c, ns)
		ns.AddToNavStack("GetCourseRecursive(): called DecideSide(): side is now %v", side)
	}

	var waypoint_angle int

	if side == RIGHT {
		waypoint_angle = degrees + 90
	} else {
		waypoint_angle = degrees - 90
	}

	waypointx, waypointy := hal.Projection(c.GetX(), c.GetY(), c.GetRadius() + DODGE_MARGIN, waypoint_angle)
	p := &hal.Point{waypointx, waypointy}

	ns.AddToNavStack("GetCourseRecursive(): collision: %v; recursing with %v", c, p)
	return GetCourseRecursive(ship, p, avoid_list, depth - 1, side, ns)
}

func GetApproach(ship *hal.Ship, target hal.Entity, margin float64, avoid_list []hal.Entity, side Side, ns NavStacker) (int, int, error) {

	// Navigate so that the ship's centre is definitely within <margin> of the target's edge.

	if ship.ApproachDist(target) < margin {
		return 0, 0, nil
	}

	// We add 0.51 in the calculation below to compensate for approximate navigation.
	// i.e. the GetCourseRecursive() call will put us within 0.5 of our requested point.

	travel_distance := ship.ApproachDist(target) + 0.51 - margin
	target_point_x, target_point_y := hal.Projection(ship.X, ship.Y, travel_distance, ship.Angle(target))

	p := &hal.Point{target_point_x, target_point_y}

	ns.AddToNavStack("GetApproach(): starting; side is %v, true target is %v, target is %v", side, target, p)
	return GetCourse(ship, p, avoid_list, side, ns)
}
