package ai

import (
	"sort"

	hal "../core"
)

func (self *Overmind) DecideRush() {

	// Can leave things undecided, in which case it will be called again next iteration.

	if self.Config.Conservative || self.Game.InitialPlayers() > 2 || len(self.Game.AllShips()) > 6 || len(self.Game.MyShips()) < 3 {
		self.RushChoice = NOT_RUSHING
		return
	}

	centre_of_gravity := self.Game.AllShipsCentreOfGravity()

	my_ships := self.Game.MyShips()

	sort.Slice(my_ships, func(a, b int) bool {
		return my_ships[b].Dist(centre_of_gravity) < my_ships[b].Dist(centre_of_gravity)
	})

	if my_ships[0].Dist(centre_of_gravity) < 50 && my_ships[1].Dist(centre_of_gravity) < 53 && my_ships[1].Dist(centre_of_gravity) < 56 {
		self.RushChoice = RUSHING
		self.SetRushTargets()
	}
}

func (self *Overmind) SetRushTargets() {		// Called on Turn 0 only, iff Rush Flag is set.

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
		pilot.Target = enemies[index]
	}
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

	self.Pilots[0].PlanThrust(s0, d0)
	self.Pilots[1].PlanThrust(s1, d1)
	self.Pilots[2].PlanThrust(s2, d2)

	for _, pilot := range self.Pilots {
		pilot.ExecutePlan()
	}
}

func (self *Overmind) DetectRushFight() bool {

	// 2 players

	players := self.Game.SurvivingPlayerIDs()

	if len(players) != 2 {
		return false
	}

	// <= 3 ships each

	for _, pid := range players {
		ships := self.Game.ShipsOwnedBy(pid)
		if len(ships) > 3 {
			return false
		}
	}

	// My ships all undocked

	for _, ship := range self.Game.MyShips() {
		if ship.DockedStatus != hal.UNDOCKED {
			return false
		}
	}

	// All ships near centre of gravity

	centre_of_gravity := self.Game.AllShipsCentreOfGravity()

	for _, ship := range self.Game.AllShips() {
		if ship.Dist(centre_of_gravity) > 20 {
			return false
		}
	}

	// Now, return true if any of my ships is within critical distance of any enemy ship

	for _, my_ship := range self.Game.MyShips() {
		for _, enemy_ship := range self.Game.EnemyShips() {
			if my_ship.Dist(enemy_ship) <= 20 {
				return true
			}
		}
	}

	return false
}

func (self *Overmind) ChooseThreeDocks() {

	// Sort all planets by distance to our fleet...

	all_planets := self.Game.AllPlanets()

	sort.Slice(all_planets, func(a, b int) bool {
		return all_planets[a].ApproachDist(self.Pilots[0]) < all_planets[b].ApproachDist(self.Pilots[0])
	})

	closest_three := all_planets[:3]

	// Get docks...

	var docks []*hal.Port

	for _, planet := range closest_three {
		docks = append(docks, hal.OpeningDockHelper(planet, self.Pilots[0].Ship)...)
	}

	docks = docks[:3]

	best_acceptable_dist := 999999.9

	var permutations = [][]int{
		[]int{0,1,2},
		[]int{0,2,1},
		[]int{1,0,2},
		[]int{1,2,0},
		[]int{2,0,1},
		[]int{2,1,0},
	}

	for _, perm := range permutations {		// Find a non-crossing solution...

		dist := self.Pilots[0].Dist(docks[perm[0]]) + self.Pilots[1].Dist(docks[perm[1]]) + self.Pilots[2].Dist(docks[perm[2]])

		if dist > best_acceptable_dist {
			continue
		}

		self.Pilots[0].Target = docks[perm[0]]
		self.Pilots[1].Target = docks[perm[1]]
		self.Pilots[2].Target = docks[perm[2]]

		if hal.Intersect(self.Pilots[0].Ship, self.Pilots[0].Target, self.Pilots[1].Ship, self.Pilots[1].Target) {
			continue
		}

		if hal.Intersect(self.Pilots[0].Ship, self.Pilots[0].Target, self.Pilots[2].Ship, self.Pilots[2].Target) {
			continue
		}

		if hal.Intersect(self.Pilots[1].Ship, self.Pilots[1].Target, self.Pilots[2].Ship, self.Pilots[2].Target) {
			continue
		}

		best_acceptable_dist = dist
	}
}
