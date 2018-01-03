package pilot

import (
	"sort"

	hal "../core"
)

func (self *Pilot) SetTurnTarget() {				// Set our short term tactical target.

	if self.DockedStatus != hal.UNDOCKED || self.Target.Type() != hal.PLANET || self.Locked {
		return
	}

	planet := self.Target.(*hal.Planet)

	// Is the planet far away?

	if self.ApproachDist(planet) > 100 {
		return
	}

	// If no enemies, just return, leaving the planet as target...

	enemies := self.Game.EnemiesNearPlanet(planet)

	if len(enemies) == 0 {
		return
	}

	// Find closest...

	sort.Slice(enemies, func(a, b int) bool {
		return enemies[a].Dist(self.Ship) < enemies[b].Dist(self.Ship)
	})

	// We could consider only targetting non-doomed enemies. Problem: if there are none,
	// we will target the planet, potentially causing us to dock. But in that case,
	// a so-called "doomed" enemy can survive, since we didn't shoot it after all!

	self.Target = enemies[0]
}

func (self *Pilot) PlanChase(avoid_list []hal.Entity) {

	// We have our target, but what are we doing about it?

	self.ResetPlan()

	if self.DockedStatus != hal.UNDOCKED {
		return
	}

	switch self.Target.Type() {

	case hal.NOTHING:

		self.PlanThrust(0, 0)

	case hal.PLANET:

		self.PlanetApproachForDock(self.Target.(*hal.Planet), avoid_list)

	case hal.SHIP:

		other_ship := self.Target.(*hal.Ship)
		self.EngageShip(other_ship, avoid_list)

	case hal.POINT:

		point := self.Target.(*hal.Point)

		side := self.DecideSideFor(point)
		speed, degrees, err := self.GetCourse(point, avoid_list, side)

		if err != nil {
			self.Target = hal.Nothing
		} else {
			self.PlanThrust(speed, degrees)
		}

	case hal.PORT:

		port := self.Target.(*hal.Port)

		planet, ok := self.Game.GetPlanet(port.PlanetID)

		if ok == false {
			self.Target = hal.Nothing
			return
		}

		if self.CanDock(planet) {
			self.PlanDock(planet)
			return
		}

		side := self.DecideSideFor(port)
		speed, degrees, err := self.GetCourse(port, avoid_list, side)

		if err != nil {
			self.Target = hal.Nothing
		} else {
			self.PlanThrust(speed, degrees)
		}
	}
}

func (self *Pilot) EngageShip(enemy_ship *hal.Ship, avoid_list []hal.Entity) {

	// Sometimes approach, sometimes flee.

	if self.Firing {
		self.EngageShipFlee(enemy_ship, avoid_list)
		return
	}

	if (self.Inhibition > 0 && self.Dist(enemy_ship) <= 20) {

		// Special case if the enemy ship is docked and there are no dangerous ships nearby...

		if enemy_ship.DockedStatus != hal.UNDOCKED && self.DangerShips == 0 {
			self.EngageShipApproach(enemy_ship, avoid_list)
			return
		}

		// But normally, flee...

		self.EngageShipFlee(enemy_ship, avoid_list)
		return
	}

	self.EngageShipApproach(enemy_ship, avoid_list)
	return
}

func (self *Pilot) EngageShipApproach(enemy_ship *hal.Ship, avoid_list []hal.Entity) {
	side := self.DecideSideFor(enemy_ship)
	speed, degrees, err := self.GetApproach(enemy_ship, self.EnemyApproachDist, avoid_list, side)
	if err != nil {
		self.Message = MSG_RECURSION
	} else {
		self.PlanThrust(speed, degrees)
	}
}

func (self *Pilot) EngageShipFlee(enemy_ship *hal.Ship, avoid_list []hal.Entity) {

	// We were already within range of an enemy, so we will definitely attack this turn.
	// We can therefore back off.

	angle := self.Angle(enemy_ship) + 180

	x2, y2 := hal.Projection(self.X, self.Y, 7, angle)
	flee_point := &hal.Point{x2, y2}

	side := self.DecideSideFor(enemy_ship)											// Wrong, but to preserve behaviour while changing things
	speed, degrees, err := self.GetApproach(flee_point, 1, avoid_list, side)
	if err != nil {
		self.Message = MSG_RECURSION
	} else {
		self.PlanThrust(speed, degrees)
	}
}

func (self *Pilot) PlanetApproachForDock(planet *hal.Planet, avoid_list []hal.Entity) {

	if self.CanDock(planet) {
		self.PlanDock(planet)
		return
	}

	side := self.DecideSideFor(planet)
	speed, degrees, err := self.GetApproach(planet, hal.DOCKING_RADIUS + hal.SHIP_RADIUS - 0.001, avoid_list, side)
	if err != nil {
		self.Message = MSG_RECURSION
	} else {
		self.PlanThrust(speed, degrees)
	}
}

func (self *Pilot) SetInhibition(all_ships []*hal.Ship) {

	self.Inhibition = 0
	self.DangerShips = 0

	for _, ship := range all_ships {

		if ship == self.Ship {
			continue
		}

		// Skip enemy docked ships (but not our own, it's important we don't flee when defending)...

		if ship.Owner != self.Owner && ship.DockedStatus != hal.UNDOCKED {
			continue
		}

		dist := self.Dist(ship)

		if dist < 20 {
			if ship.Owner != self.Owner && ship.DockedStatus == hal.UNDOCKED {
				self.DangerShips++
			}
		}

		strength := 10000 / (dist * dist)

		if ship.Owner == self.Owner {
			strength *= -1
		}

		self.Inhibition += strength
	}
}
