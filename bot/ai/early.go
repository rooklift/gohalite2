package ai

import (
	"sort"

	hal "../core"
)

func (self *Overmind) ChooseInitialTargets() bool {		// Returns: are we assassinating?

	// As a good default...

	self.ChooseThreeDocks()

	if CONFIG.Conservative {
		return false
	}

	if self.Game.InitialPlayers() == 2 {
		if self.Game.MyShips()[0].Dist(self.Game.EnemyShips()[0]) < 135 {
			self.ChooseAssassinateTargets()
			return true
		}
	}

	return false
}

func (self *Overmind) ChooseAssassinateTargets() {

	// Sort our pilots by Y...

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].Y < self.Pilots[b].Y
	})

	// Sort enemies by Y...

	enemies := self.Game.EnemyShips()
	sort.Slice(enemies, func(a, b int) bool {
		return enemies[a].Y < enemies[b].Y
	})

	// Pair pilots with enemies...

	for index, pilot := range self.Pilots {
		pilot.SetTarget(enemies[index])
	}
}

func (self *Overmind) ChooseThreeDocks() {

	// Sort all planets by distance to our fleet...

	all_planets := self.Game.AllPlanets()

	sort.Slice(all_planets, func(a, b int) bool {
		return all_planets[a].ApproachDist(self.Pilots[0]) < all_planets[b].ApproachDist(self.Pilots[0])
	})

	closest_three := all_planets[:3]

	// Get docks...

	var docks []hal.Point

	for _, planet := range closest_three {
		docks = append(docks, planet.OpeningDockHelper(self.Pilots[0].Ship)...)
	}

	docks = docks[:3]

	var permutations = [][]int{
		[]int{0,1,2},
		[]int{0,2,1},
		[]int{1,0,2},
		[]int{1,2,0},
		[]int{2,0,1},
		[]int{2,1,0},
	}

	for _, perm := range permutations {		// Find a non-crossing solution...

		self.Pilots[0].SetTarget(docks[perm[0]])
		self.Pilots[1].SetTarget(docks[perm[1]])
		self.Pilots[2].SetTarget(docks[perm[2]])

		if hal.Intersect(self.Pilots[0].Ship, self.Pilots[0].Target, self.Pilots[1].Ship, self.Pilots[1].Target) {
			continue
		}

		if hal.Intersect(self.Pilots[0].Ship, self.Pilots[0].Target, self.Pilots[2].Ship, self.Pilots[2].Target) {
			continue
		}

		if hal.Intersect(self.Pilots[1].Ship, self.Pilots[1].Target, self.Pilots[2].Ship, self.Pilots[2].Target) {
			continue
		}

		break
	}
}

func (self *Overmind) DetectRushFight() bool {

	// <= 3 ships each
	// All ships near each other
	// My ships all undocked

	players := self.Game.SurvivingPlayerIDs()

	if len(players) != 2 {
		return false
	}

	var all_ships []hal.Ship

	for _, pid := range players {
		ships := self.Game.ShipsOwnedBy(pid)
		if len(ships) > 3 {
			return false
		}
		for _, ship := range ships {
			if ship.DockedStatus != hal.UNDOCKED && ship.Owner == self.Game.Pid() {
				return false
			}
		}
		all_ships = append(all_ships, ships...)
	}

	for _, ship := range all_ships {
		if ship.Dist(self.Game.AllShipsCentreOfGravity()) > 20 {
			return false
		}
	}

	return true
}

func (self *Overmind) TurnZeroCluster() {

	centre_of_gravity := self.Game.AllShipsCentreOfGravity()

	if centre_of_gravity.X > self.Pilots[0].X {
		self.Cluster(7, 15, 6, 0, 7, 345)
	} else if centre_of_gravity.X < self.Pilots[0].X {
		self.Cluster(7, 165, 6, 180, 7, 195)
	} else if centre_of_gravity.Y > self.Pilots[0].Y {
		self.Cluster(7, 90, 5, 100, 2, 60)
	} else if centre_of_gravity.Y < self.Pilots[0].Y {
		self.Cluster(2, 240, 5, 280, 7, 270)
	}

}

func (self *Overmind) Cluster(s0, d0, s1, d1, s2, d2 int) {

	// Sort our pilots by Y...

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].Y < self.Pilots[b].Y
	})

	self.Pilots[0].PlanThrust(s0, d0, -1)
	self.Pilots[1].PlanThrust(s1, d1, -1)
	self.Pilots[2].PlanThrust(s2, d2, -1)

	for _, pilot := range self.Pilots {
		pilot.ExecutePlan()
	}
}
