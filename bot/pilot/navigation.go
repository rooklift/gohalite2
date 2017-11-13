package pilot

import (
	"fmt"

	hal "../core"
)

func (self *Pilot) CheckEntityCollision(distance float64, degrees int, other hal.Entity) bool {		// Would we hit some specific entity?

	const SAFETY_MARGIN = 0.001		// Needed to avoid floating point errors: the engine gives us data to 4 d.p.

	endx, endy := hal.Projection(self.X, self.Y, distance, degrees)
	return hal.IntersectSegmentCircle(self.X, self.Y, endx, endy, other.GetX(), other.GetY(), other.GetRadius() + hal.SHIP_RADIUS + SAFETY_MARGIN)
}

func (self *Pilot) FirstCollision(distance float64, degrees int, possibles []hal.Entity) (hal.Entity, bool) {

	var collisions []hal.Entity

	for _, other := range possibles {
		if self.CheckEntityCollision(distance, degrees, other) {
			collisions = append(collisions, other)
		}
	}

	if len(collisions) == 0 {
		return nil, false
	}

	var closest_ent hal.Entity
	var closest_distance float64 = 999999.9

	for _, c := range collisions {
		if self.Dist(c) < closest_distance {
			closest_ent = c
			closest_distance = self.Dist(c)
		}
	}

	return closest_ent, true
}

func (self *Pilot) GetCourse(target hal.Entity, avoid_list []hal.Entity, nav_side Side) (int, int, error) {
	return self.GetCourseRecursive(target, avoid_list, 10, nav_side)
}

func (self *Pilot) GetCourseRecursive(target hal.Entity, avoid_list []hal.Entity, depth int, nav_side Side) (int, int, error) {		// speed, angle, error

	// Try to navigate to (collide with) the target, but avoiding the list of entites,
	// which could include the target.

	const (
		DODGE_MARGIN = 1.5
	)

	distance := self.Dist(target)

	if distance < 0.5 {
		return 0, 0, nil
	}

	// If we're close, only allow whole number distances so we don't hit things due to rounding later.

	if distance < hal.MAX_SPEED + 1 {
		distance = hal.RoundToFloat(distance)
	}

	degrees := hal.Angle(self.X, self.Y, target.GetX(), target.GetY())

	c, ok := self.FirstCollision(distance, degrees, avoid_list)

	if ok == false {		// There is no collision
		speed := hal.Min(hal.Round(distance), hal.MAX_SPEED)
		return speed, degrees, nil
	}

	if depth < 1 {
		return 0, 0, fmt.Errorf("GetCourseRecursive(): exceeded max depth")
	}

	// Reset our nav side iff the colliding object is a planet / docked ship...

	if c.Type() == hal.PLANET {
		nav_side = self.DecideSide(target, c.(hal.Planet))
	} else {
		if c.(hal.Ship).DockedPlanet != -1 {
			p, _ := self.Game.GetPlanet(c.(hal.Ship).DockedPlanet)
			nav_side = self.DecideSide(target, p)
		}
	}

	var waypoint_angle int

	if nav_side == RIGHT {
		waypoint_angle = degrees + 90
	} else {
		waypoint_angle = degrees - 90
	}
	waypointx, waypointy := hal.Projection(c.GetX(), c.GetY(), c.GetRadius() + DODGE_MARGIN, waypoint_angle)
	return self.GetCourseRecursive(hal.Point{waypointx, waypointy}, avoid_list, depth - 1, nav_side)
}

func (self *Pilot) GetApproach(target hal.Entity, margin float64, avoid_list []hal.Entity) (int, int, error) {

	// Navigate so that the ship's centre is definitely within <margin> of the target's edge.

	if self.ApproachDist(target) < margin {
		return 0, 0, nil
	}

	nav_side := self.DecideSideFromTarget(target)

	// We add 0.51 in the calculation below to compensate for approximate navigation.
	// i.e. the GetCourseRecursive() call will put us within 0.5 of our requested point.

	travel_distance := self.ApproachDist(target) + 0.51 - margin
	target_point_x, target_point_y := hal.Projection(self.X, self.Y, travel_distance, self.Angle(target))
	return self.GetCourse(hal.Point{target_point_x, target_point_y}, avoid_list, nav_side)
}

// ---------------------------------------------------------------------

type Side int

func (s Side) String() string {
	if s == LEFT { return "LEFT" } else if s == RIGHT { return "RIGHT" }
	return "???"
}

const (
	LEFT Side = iota
	RIGHT
)

// Given a ship and some target, and some planet to navigate around,
// which side should we go?

func (self *Pilot) DecideSide(target hal.Entity, planet hal.Planet) Side {

	to_planet := self.Angle(planet)
	to_target := self.Angle(target)

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

	return RIGHT
}

func (self *Pilot) DecideSideFromTarget(target hal.Entity) Side {

	// As a default, use this arbitrary choice...

	var side Side; if self.Id % 2 == 0 { side = RIGHT } else { side = LEFT }

	// If the first planet in our path isn't our target planet, we choose a side to navigate around.
	// By using AllImmobile() as the avoid_list, any collision will be with a planet or docked ship.

	collision_entity, ok := self.FirstCollision(1000, self.Angle(self.Target), self.Game.AllImmobile())

	if ok {

		var blocking_planet hal.Planet

		// We also consider docked ships to be "part of the planet" for these purposes -+- we must use self.Game.AllImmobile() above

		if collision_entity.Type() == hal.PLANET {
			blocking_planet = collision_entity.(hal.Planet)
		} else {
			s := collision_entity.(hal.Ship)
			blocking_planet, _ = self.Game.GetPlanet(s.DockedPlanet)
		}

		if self.Target.Type() != hal.PLANET || blocking_planet.Id != self.Target.GetId() {
			side = self.DecideSide(self.Target, blocking_planet)
		}
	}

	return side
}
