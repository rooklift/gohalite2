package ai

import (
	"sort"
)

func (self *Game) BasicAI() {
	self.ChooseTargets()
	self.MakeMoves()
	self.AntiCollision()
	self.HaltUnsafeOrders()
}

// -------------------------------------------------------------------------------

type Problem struct {
	Entity		Entity
	Value		float64
	Need		int
	Message		int
}

func (self *Game) ChooseTargets() {

	all_problems := self.AllProblems()

	for _, ship := range self.MyShips() {

		if len(all_problems) == 0 {
			all_problems = self.AllProblems()
			if len(all_problems) == 0 {
				break
			}
		}

		if ship.DockedStatus != UNDOCKED {
			continue
		}

		sort.Slice(all_problems, func(a, b int) bool {
			return ship.Dist(all_problems[a].Entity) / all_problems[a].Value <
			       ship.Dist(all_problems[b].Entity) / all_problems[b].Value
		})

		ship.Target = all_problems[0].Entity

		all_problems[0].Need--
		if all_problems[0].Need <= 0 {
			all_problems = all_problems[1:]
		}
	}
}

// -------------------------------------------------------------------------------

func (self *Game) AllProblems() []*Problem {

	var all_problems []*Problem

	for _, planet := range self.AllPlanets() {
		problems := self.PlanetProblems(planet)
		all_problems = append(all_problems, problems...)
	}

	for _, ship := range self.EnemyShips() {

		problem := &Problem{
			Entity: ship,
			Value: 1.0,
			Need: 1,
		}
		all_problems = append(all_problems, problem)
	}

	return all_problems
}

// -------------------------------------------------------------------------------

func (self *Game) PlanetProblems(planet *Planet) []*Problem {

	var ret []*Problem

	enemies := self.EnemiesNearPlanet(planet)
	capture_strength := self.DesiredSpots(planet)

	switch len(enemies) {

	case 0:

		if capture_strength > 0 {

			value := 1.0 / 1.4; if self.InitialPlayers() > 2 { value = 1.0 }

			ret = append(ret, &Problem{
				Entity: planet,
				Value: value,
				Need: capture_strength,
				Message: planet.Id,
			})
		}

	default:

		for _, enemy := range enemies {

			ret = append(ret, &Problem{
				Entity: enemy,
				Value: 1.0,
				Need: 2,
				Message: planet.Id,
			})
		}
	}

	return ret
}

// -------------------------------------------------------------------------------

func (self *Game) MakeMoves() {

	avoid_list := self.AllImmobile()

	for _, ship := range self.MyShips() {
		ship.PlanChase(avoid_list)
	}
}

func (self *Ship) PlanChase(avoid_list []Entity) {

	self.Validated = false

	if self.Target == nil || self.Target.Type() == NOTHING || self.DockedStatus != UNDOCKED {
		return
	}

	switch self.Target.Type() {

	case NOTHING:

		self.Thrust(0, 0)

	case PLANET:

		self.PlanetApproachForDock(self.Target.(*Planet), avoid_list)

	case SHIP:

		other_ship := self.Target.(*Ship)
		self.EngageShip(other_ship, avoid_list)
	}
}

func (self *Ship) PlanetApproachForDock(planet *Planet, avoid_list []Entity) {

	if self.CanDock(planet) {
		self.Dock(planet)
		return
	}

	side := self.DecideSideFor(planet)
	speed, degrees, _ := self.GetApproach(planet, DOCKING_RADIUS + SHIP_RADIUS - 0.001, avoid_list, side)
	self.Thrust(speed, degrees)
}

func (self *Ship) EngageShip(enemy_ship *Ship, avoid_list []Entity) {

	side := self.DecideSideFor(enemy_ship)
	speed, degrees, _ := self.GetApproach(enemy_ship, 5.45, avoid_list, side)
	self.Thrust(speed, degrees)

}

func (self *Ship) GetCourse(target Entity, avoid_list []Entity, side Side) (int, int, error) {
	return GetCourse(self, target, avoid_list, side)
}

func (self *Ship) GetApproach(target Entity, margin float64, avoid_list []Entity, side Side) (int, int, error) {
	return GetApproach(self, target, margin, avoid_list, side)
}

func (self *Ship) DecideSideFor(target Entity) Side {
	return DecideSideFromTarget(self, target, self.Game)
}

func (self *Game) AntiCollision() {

	var mobile_ships []*Ship

	for _, ship := range self.MyShips() {
		if ship.DockedStatus == UNDOCKED {
			mobile_ships = append(mobile_ships, ship)			// This does include ships that will dock or stay still.
		}
	}

	if len(mobile_ships) == 0 {
		return
	}

	// Assumption: we have already taken steps to ensure that any ship not included in the mobile_ships
	// is avoided, i.e. those ships were explicitly avoided in the earlier navigation search.

	for n := 0; n < 11; n++ {

		total_executes := 0

		ship1Loop:

		for _, ship1 := range mobile_ships {

			if ship1.Validated {
				total_executes++
				continue
			}

			if n >= 5 {
				ship1.SlowDown()
			}

			ship1_desired_speed, ship1_desired_angle := ship1.CurrentCourse()

			for _, ship2 := range mobile_ships {

				if ship2 == ship1 {
					continue
				}

				if ship2.Dist(ship1) > 15 {
					continue
				}

				ship2_speed := 0
				ship2_angle := 0

				if ship2.Validated {
					ship2_speed, ship2_angle = ship2.CurrentCourse()
				}

				if ShipsWillCollide(ship1, ship1_desired_speed, ship1_desired_angle, ship2, ship2_speed, ship2_angle) {
					continue ship1Loop
				}
			}

			ship1.Validated = true
			total_executes++
		}

		if total_executes >= len(mobile_ships) {
			return
		}
	}
}

func (self *Game) HaltUnsafeOrders() {
	for _, ship := range self.MyShips() {
		if ship.DockedStatus == UNDOCKED {
			if ship.Validated == false {
				ship.Thrust(0, 0)
			}
		}
	}
}
