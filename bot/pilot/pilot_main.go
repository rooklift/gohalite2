package pilot

import (
	"sort"

	hal "../core"
)

func (self *Pilot) SetTurnTarget() {				// Set our short term tactical target.

	// Default...

	self.TurnTarget = self.Target

	// Keep the default in many cases...

	if self.DockedStatus != hal.UNDOCKED {
		return
	}

	if self.Target.Type() != hal.PLANET {
		return
	}

	// But in the case of a planet, we might set TurnTarget to be a ship...

	planet := self.Target.(hal.Planet)

	// Is the planet far away?

	if self.ApproachDist(planet) > 100 {
		return
	}

	// Are there enemy ships near the planet? Includes docked enemies.

	enemies := self.Game.EnemiesNearPlanet(planet)

	if len(enemies) > 0 {

		// Find closest...

		sort.Slice(enemies, func(a, b int) bool {
			return enemies[a].Dist(self.Ship) < enemies[b].Dist(self.Ship)
		})

		self.TurnTarget = enemies[0]
		return
	}

	// Otherwise, just return (leaving the TurnTarget as the planet).

	return
}

func (self *Pilot) PlanChase(avoid_list []hal.Entity) {

	if self.DockedStatus != hal.UNDOCKED {
		return
	}

	if self.TurnTarget.Type() == hal.NOTHING {
		self.PlanThrust(0, 0)
		self.Message = MSG_NO_TARGET
		return
	}

	switch self.TurnTarget.Type() {

	case hal.PLANET:

		planet := self.TurnTarget.(hal.Planet)

		if self.ApproachDist(planet) <= 100 {
			self.PlanetApproachForDock(avoid_list)
			return
		}

		// Why do we bother with this, instead of always calling PlanetApproachForDock() ? - I can't recall.

		side := self.DecideSideFromTurnTarget()
		speed, degrees, err := self.GetApproach(planet, 4.45, avoid_list, side)

		if err != nil {
			self.SetTarget(hal.Nothing{})
		} else {
			self.PlanThrust(speed, degrees)
			self.Message = planet.Id
		}

	case hal.SHIP:

		other_ship := self.TurnTarget.(hal.Ship)
		self.EngageShip(other_ship, avoid_list)

	case hal.POINT:

		point := self.TurnTarget.(hal.Point)

		side := self.DecideSideFromTurnTarget()
		speed, degrees, err := self.GetCourse(point, avoid_list, side)

		if err != nil {
			self.SetTarget(hal.Nothing{})
		} else {
			self.PlanThrust(speed, degrees)
			self.Message = MSG_POINT_TARGET
		}
	}
}

func (self *Pilot) EngageShip(enemy_ship hal.Ship, avoid_list []hal.Entity) {

	// Flee if we're already in weapons range...

	if self.Dist(enemy_ship) < hal.WEAPON_RANGE + hal.SHIP_RADIUS * 2 {
		self.EngageShipFlee(enemy_ship, avoid_list)
		return
	}

	// If we're quite far, just approach...

	if self.Dist(enemy_ship) >= KILLER_THRESHOLD {
		self.EngageShipApproach(enemy_ship, avoid_list)
		return
	}

	// If enemy ship is docked, approaching is fine...

	if enemy_ship.DockedStatus != hal.UNDOCKED {
		self.EngageShipApproach(enemy_ship, avoid_list)
		return
	}

	// Enemy is undocked. We would need to close in to attack. But should we?

	if self.AvoidFight {
		self.EngageShipFlee(enemy_ship, avoid_list)
		return
	}

	self.EngageShipApproach(enemy_ship, avoid_list)
}

func (self *Pilot) EngageShipMessage(err error) int {
	if err != nil { return MSG_RECURSION }
	if self.Target.Type() == hal.PLANET { return self.Target.(hal.Planet).Id }
	return MSG_ASSASSINATE
}

func (self *Pilot) EngageShipApproach(enemy_ship hal.Ship, avoid_list []hal.Entity) {
	side := self.DecideSideFromTurnTarget()
	speed, degrees, err := self.GetApproach(enemy_ship, self.EnemyApproachDist, avoid_list, side)
	msg := self.EngageShipMessage(err)
	self.PlanThrust(speed, degrees)
	self.Message = msg
}

func (self *Pilot) EngageShipFlee(enemy_ship hal.Ship, avoid_list []hal.Entity) {

	// We were already within range of our target ship, so we will definitely attack it this turn.
	// We can therefore back off.

	angle := self.Angle(enemy_ship) + 180

	x2, y2 := hal.Projection(self.X, self.Y, 7, angle)
	flee_point := hal.Point{x2, y2}

	side := self.DecideSideFromTurnTarget()

	speed, degrees, err := self.GetApproach(flee_point, 1, avoid_list, side)

	msg := self.EngageShipMessage(err)

	self.PlanThrust(speed, degrees)
	self.Message = msg
}

func (self *Pilot) PlanetApproachForDock(avoid_list []hal.Entity) {

	if self.TurnTarget.Type() != hal.PLANET {
		self.Log("PlanetApproachForDock() called but target wasn't a planet.")
		return
	}

	planet := self.TurnTarget.(hal.Planet)

	if self.CanDock(planet) {
		self.PlanDock(planet)
		return
	}

	side := self.DecideSideFromTurnTarget()
	speed, degrees, err := self.GetApproach(planet, hal.DOCKING_RADIUS + hal.SHIP_RADIUS - 0.001, avoid_list, side)

	self.PlanThrust(speed, degrees)

	if err != nil {
		self.Message = MSG_RECURSION
	} else {
		self.Message = planet.Id
	}
}
