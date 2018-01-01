package pilot

import (
	"sort"

	hal "../core"
)

func (self *Pilot) SetTurnTarget() {				// Set our short term tactical target.

	if self.DockedStatus != hal.UNDOCKED || self.Target.Type() != hal.PLANET {
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

	// Flee if we're firing at time 0...

	if self.Firing {
		self.EngageShipFlee(enemy_ship, avoid_list)
	} else {
		self.EngageShipApproach(enemy_ship, avoid_list)
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

	// We were already within range of our target ship, so we will definitely attack it this turn.
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
