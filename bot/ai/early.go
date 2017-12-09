package ai

import (
	"sort"

	gen "../genetic"
	hal "../core"
)

func (self *Overmind) TurnZero() {
	assassin := self.ChooseInitialTargets()
	if assassin {
		self.TurnZeroCluster()
	} else {
		self.NormalStep()
	}
}

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

	best_dist := 999999.9

	for _, perm := range permutations {		// Find a non-crossing solution...

		target_0 := docks[perm[0]]
		target_1 := docks[perm[1]]
		target_2 := docks[perm[2]]

		if hal.Intersect(self.Pilots[0].Ship, target_0, self.Pilots[1].Ship, target_1) {
			continue
		}

		if hal.Intersect(self.Pilots[0].Ship, target_0, self.Pilots[2].Ship, target_2) {
			continue
		}

		if hal.Intersect(self.Pilots[1].Ship, target_1, self.Pilots[2].Ship, target_2) {
			continue
		}

		total_dist := self.Pilots[0].Dist(target_0) + self.Pilots[1].Dist(target_1) + self.Pilots[2].Dist(target_2)

		if total_dist < best_dist {
			best_dist = total_dist
			self.Pilots[0].SetTarget(target_0)
			self.Pilots[1].SetTarget(target_1)
			self.Pilots[2].SetTarget(target_2)
		}
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

func (self *Overmind) HandleRushFight() {
	if CONFIG.Conservative {
		self.NormalStep()
	} else {
		gen.FightRush(self.Game)
	}
}
