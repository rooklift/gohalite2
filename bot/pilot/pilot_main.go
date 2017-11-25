package pilot

import (
	"sort"

	hal "../core"
)

func (self *Pilot) PlanChase(avoid_list []hal.Entity) {

	if self.DockedStatus != hal.UNDOCKED {
		return
	}

	if self.Target.Type() == hal.NOTHING {
		self.PlanThrust(0, 0, MSG_NO_TARGET)
		return
	}

	switch self.Target.Type() {

	case hal.PLANET:

		planet := self.Target.(hal.Planet)

		if self.ApproachDist(planet) <= 100 {
			self.EngagePlanet(avoid_list)
			return
		}

		side := self.DecideSideFromTarget()
		speed, degrees, err := self.GetApproach(planet, 4.45, avoid_list, side)

		if err != nil {
			self.SetTarget(hal.Nothing{})
		} else {
			self.PlanThrust(speed, degrees, MessageInt(planet.Id))
		}

	case hal.SHIP:

		other_ship := self.Target.(hal.Ship)
		self.EngageShip(other_ship, avoid_list)

	case hal.POINT:

		point := self.Target.(hal.Point)

		side := self.DecideSideFromTarget()
		speed, degrees, err := self.GetCourse(point, avoid_list, side)

		if err != nil {
			self.SetTarget(hal.Nothing{})
		} else {
			self.PlanThrust(speed, degrees, MSG_POINT_TARGET)
		}
	}
}

func (self *Pilot) EngagePlanet(avoid_list []hal.Entity) {

	// We are "close" to our target planet. Do something about this.
	// As of v32 or so, "close" is actually quite far.

	if self.Target.Type() != hal.PLANET {
		self.Log("EngagePlanet() called but target wasn't a planet.")
		return
	}

	planet := self.Target.(hal.Planet)

	// Are there enemy ships near the planet? Includes docked enemies.

	enemies := self.Game.EnemiesNearPlanet(planet)

	if len(enemies) > 0 {

		// Find closest...

		sort.Slice(enemies, func(a, b int) bool {
			return enemies[a].Dist(self.Ship) < enemies[b].Dist(self.Ship)
		})

		self.EngageShip(enemies[0], avoid_list)
		return
	}

	// Is it available for us to dock?

	if planet.Owned == false || (planet.Owner == self.Game.Pid() && planet.IsFull() == false) {
		self.PlanetApproachForDock(avoid_list)
		return
	}

	// This function shouldn't have been called at all.

	self.Log("EngagePlanet() called but there's nothing to do here.")
	return
}

func (self *Pilot) EngageShip(enemy_ship hal.Ship, avoid_list []hal.Entity) {

	side := self.DecideSideFromTarget()
	speed, degrees, err := self.GetApproach(enemy_ship, ENEMY_SHIP_APPROACH_DIST, avoid_list, side)

	var msg MessageInt

	if err != nil {										// Pathfinding failed...
		msg = MSG_RECURSION
	} else if self.Target.Type() == hal.PLANET {		// We're fighting a ship because it's near our target planet...
		msg = MSG_ORBIT_FIGHT
	} else {											// We're directly targeting a ship...
		msg = MSG_ASSASSINATE
	}

	self.PlanThrust(speed, degrees, msg)
}

func (self *Pilot) PlanetApproachForDock(avoid_list []hal.Entity) {

	if self.Target.Type() != hal.PLANET {
		self.Log("PlanetApproachForDock() called but target wasn't a planet.")
		return
	}

	planet := self.Target.(hal.Planet)

	if self.CanDock(planet) {
		self.PlanDock(planet)
		return
	}

	side := self.DecideSideFromTarget()
	speed, degrees, err := self.GetApproach(planet, hal.DOCKING_RADIUS + hal.SHIP_RADIUS - 0.001, avoid_list, side)

	if err != nil {
		self.PlanThrust(speed, degrees, MSG_RECURSION)
	} else {
		self.PlanThrust(speed, degrees, MSG_DOCK_APPROACH)
	}
}
