package pilot

import (
	hal "../core"
)

func (self *Pilot) PlanChase(avoid_list []hal.Entity, ignore_inhibition bool) {

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
		self.EngageShip(other_ship, avoid_list, ignore_inhibition)

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

func (self *Pilot) EngageShip(other_ship *hal.Ship, avoid_list []hal.Entity, ignore_inhibition bool) {

	// Protect it if it's friendly...

	if other_ship.Owner == self.Game.Pid() {
		self.ProtectShip(other_ship, avoid_list)
		return
	}

	// Otherwise: sometimes approach, sometimes flee...

	if ignore_inhibition {
		self.EngageShipApproach(other_ship, avoid_list)
		return
	}

	if self.Firing && self.DangerShips > 0 {
		self.EngageShipFlee(other_ship, avoid_list)
		return
	}

	if (self.Inhibition > 0 && self.Dist(other_ship) <= 20) {

		// We are close to our enemy ship; if we both approach each other we will fight.
		// But should we actually approach?

		// Special case if the enemy ship is docked and there are no dangerous ships nearby...

		if other_ship.DockedStatus != hal.UNDOCKED && self.DangerShips == 0 {
			self.EngageShipApproach(other_ship, avoid_list)
			self.Log("Safe to ignore Inhibition and approach docked ship.")
			return
		}

		// Special case if the enemy ship is alone and we can kill it safely...
		// FIXME: this isn't actually correct, an enemy docked ship could absorb some damage intended for our real target...

		if other_ship.DockedStatus == hal.UNDOCKED && self.DangerShips == 1 && other_ship.ShotsToKill() == 1 && self.ShotsToKill() > 1 {
			self.EngageShipApproach(other_ship, avoid_list)
			self.Log("Safe to ignore Inhibition and go for the kill.")
			return
		}

		// But normally, flee...

		self.EngageShipFlee(other_ship, avoid_list)
		return
	}

	self.EngageShipApproach(other_ship, avoid_list)
	return
}

func (self *Pilot) ProtectShip(other_ship *hal.Ship, avoid_list []hal.Entity) {
	side := self.DecideSideFor(other_ship)
	speed, degrees, err := self.GetApproach(other_ship, 0, avoid_list, side)
	if err != nil {
		self.Message = MSG_RECURSION
	} else {
		self.PlanThrust(speed, degrees)
	}
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

	// EXPERIMENT (v90): ignore our target and retreat from the closest enemy instead.

	angle := self.Angle(self.ClosestEnemy) + 180

	x2, y2 := hal.Projection(self.ClosestEnemy.X, self.ClosestEnemy.Y, 14, angle)	// 13 + 1 which is fudged below (IIRC)
	flee_point := &hal.Point{x2, y2}

	side := self.DecideSideFor(flee_point)
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

		dist := self.Dist(ship)		// Consider: dist := hal.MaxFloat(5, self.Dist(ship)) // Don't let really close ships affect us too strongly...

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
