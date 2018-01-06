package ai

import (
	"math/rand"
	"sort"

	gen "../genetic"
	hal "../core"
)

func (self *Overmind) DecideRush() {

	// Can leave things undecided, in which case it will be called again next iteration.

	if self.Game.InitialPlayers() > 2 {
		self.RushChoice = NOT_RUSHING
		self.Game.Log("Not rushing because: self.Game.InitialPlayers() > 2")
		return
	}

	if len(self.Game.MyShips()) < 3 {
		self.RushChoice = NOT_RUSHING
		self.Game.Log("Not rushing because: len(self.Game.MyShips()) < 3")
		return
	}

	if self.Game.WeHaveDockedShips() {
		self.RushChoice = NOT_RUSHING
		self.Game.Log("Not rushing because: self.Game.WeHaveDockedShips()")
		return
	}

	my_ships := self.Game.MyShips()
	centre_of_gravity := self.Game.AllShipsCentreOfGravity()

	sort.Slice(my_ships, func(a, b int) bool {
		return my_ships[b].Dist(centre_of_gravity) < my_ships[b].Dist(centre_of_gravity)
	})

	if my_ships[0].Dist(centre_of_gravity) < 45 && my_ships[1].Dist(centre_of_gravity) < 48 && my_ships[2].Dist(centre_of_gravity) < 51 {
		self.RushChoice = RUSHING
		self.Game.Log("RUSHING!")
		return
	}
}

func (self *Overmind) MaybeEndRush() {
	if len(self.Game.ShipsOwnedBy(self.RushEnemyID)) == 0 {
		self.Game.Log("Ending rush!")
		self.RushChoice = NOT_RUSHING
	}
}

func (self *Overmind) TurnZeroCluster() {

	if self.Game.InitialPlayers() == 4 {

		switch self.Game.Pid() {
			case 0: fallthrough
			case 1:	self.Cluster(7, 90, 5, 100, 2, 60)
			case 2: fallthrough
			case 3: self.Cluster(2, 240, 5, 280, 7, 270)
		}

	} else {

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

	relevant_enemies := self.Game.ShipsOwnedBy(self.RushEnemyID)			// In 4p, this is only the ships of the closest player
	my_ships := self.Game.MyShips()

	// Enemy exists...

	if len(relevant_enemies) == 0 {
		return false
	}

	// <= 3 ships each

	if len(my_ships) > 3 || len(relevant_enemies) > 3 {
		return false
	}

	// My ships all undocked

	for _, ship := range my_ships {
		if ship.DockedStatus != hal.UNDOCKED {
			return false
		}
	}

	// All ships near centre of gravity

	centre_of_gravity := self.Game.PartialCentreOfGravity(self.Game.Pid(), self.RushEnemyID)

	for _, ship := range relevant_enemies {
		if ship.Dist(centre_of_gravity) > 20 {
			return false
		}
	}

	for _, ship := range my_ships {
		if ship.Dist(centre_of_gravity) > 20 {
			return false
		}
	}

	// Now, return true if any of my ships is within critical distance of any enemy ship

	for _, my_ship := range my_ships {
		for _, enemy_ship := range relevant_enemies {
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

	if len(docks) < 3 {
		return
	}

	self.SetNonIntersectingDockPaths(docks[:3])

	// As a special case, never send less than 3 ships to the central planets in 2 player games...

	if self.Game.InitialPlayers() == 2 {

		ships_to_centre := 0

		for _, pilot := range self.Pilots {
			if pilot.Target.GetId() < 4 {			// Centre planets have ID < 4, unless the engine changes.
				ships_to_centre++
			}
		}

		if ships_to_centre > 0 && ships_to_centre < 3 {
			self.Game.Log("Was going to send %d ships to centre; sending all instead.", ships_to_centre)
			self.ChooseCentreDocks()
		}
	}
}

func (self *Overmind) SetNonIntersectingDockPaths(docks []*hal.Port) {

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

		for n := 0; n < 3; n++ {
			self.Pilots[n].Target = docks[perm[n]]
		}

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

func (self *Overmind) ChooseCentreDocks() {

	var centre_planets []*hal.Planet

	for n := 0; n < 4; n++ {
		planet, ok := self.Game.GetPlanet(n)
		if ok {
			centre_planets = append(centre_planets, planet)
		}
	}

	sort.Slice(centre_planets, func(a, b int) bool {
		return self.Pilots[0].Dist(centre_planets[a]) < self.Pilots[0].Dist(centre_planets[b])
	})

	var docks []*hal.Port

	for _, planet := range centre_planets {
		docks = append(docks, hal.OpeningDockHelper(planet, self.Pilots[0].Ship)...)
	}

	if len(docks) < 3 {
		return
	}

	self.SetNonIntersectingDockPaths(docks[:3])
}

func (self *Overmind) CanAvoidBad2v1() bool {

	// Called when DetectRushFight() has already returned true, i.e. we want to enter the genetic algorithm.
	// FIXME? Was written assuming we only enter GA in 2p.

	// There is a special case that loses occasional games if we don't handle it... basically, if we are
	// 2v1 up but the opponent ever produced a ship, we can't just chase him forever.

	if self.Game.CountMyShips() < 2 || self.Game.CountEnemyShips() > 1 {
		return false
	}

	// So, we have 2 or more ships, and the enemy has just 1.

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].HP > self.Pilots[b].HP			// Note reversed sort, highest HP first - it will chase enemy.
	})

	enemy_ship := self.Game.EnemyShips()[0]

	enemy_player_id := enemy_ship.Owner

	if self.Game.GetCumulativeShipCount(enemy_player_id) <= self.Game.GetCumulativeShipCount(self.Game.Pid()) {
		return false
	}

	// So we are indeed in the losing situation...
	// But can we do anything about it??

	if enemy_ship.ShotsToKill() <= self.Pilots[0].ShotsToKill() && enemy_ship.DockedStatus == hal.UNDOCKED {
		return true
	}

	return false
}

func (self *Overmind) AvoidBad2v1() {

	self.Game.Log("Avoiding Bad 2v1: setting Overmind.RushChoice, Overmind.NeverGA; and choosing targets.")

	self.RushChoice = RUSHING									// Ensures the chaser continues to chase.
	self.NeverGA = true
	self.AvoidingBad2v1 = true

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].HP > self.Pilots[b].HP			// Note reversed sort, highest HP first - it will chase enemy.
	})

	all_enemies := self.Game.EnemyShips()

	sort.Slice(all_enemies, func(a, b int) bool {
		return self.Pilots[0].Dist(all_enemies[a]) < self.Pilots[0].Dist(all_enemies[b])
	})

	self.Pilots[0].Target = all_enemies[0]

	for i := 1; i < len(self.Pilots); i++ {
		self.Pilots[i].Target = self.Game.FarthestPlanet(self.Pilots[i].Ship)
	}
}

func (self *Overmind) FindRushEnemy() {

	self.RushEnemyID = -1

	switch self.Game.InitialPlayers() {

	case 2:

		if self.Game.Pid() == 0 {
			self.RushEnemyID = 1
		} else {
			self.RushEnemyID = 0
		}

		my_first_ship := self.Game.MyShips()[0]

		if my_first_ship.X < float64(self.Game.Width()) / 2 - 20 {
			self.MyRushSide = hal.LEFT
		} else if my_first_ship.X > float64(self.Game.Width()) / 2 + 20 {
			self.MyRushSide = hal.RIGHT
		} else if my_first_ship.Y < float64(self.Game.Height()) / 2 - 20 {
			self.MyRushSide = hal.TOP
		} else if my_first_ship.Y > float64(self.Game.Height()) / 2 + 20 {
			self.MyRushSide = hal.BOTTOM
		} else {
			panic("Couldn't determine MyRushSide")
		}

	case 4:

		if self.Game.Pid() == 0 {
			self.RushEnemyID = 2
			self.MyRushSide = hal.TOP
		} else if self.Game.Pid() == 1 {
			self.RushEnemyID = 3
			self.MyRushSide = hal.TOP
		} else if self.Game.Pid() == 2 {
			self.RushEnemyID = 0
			self.MyRushSide = hal.BOTTOM
		} else if self.Game.Pid() == 3 {
			self.RushEnemyID = 1
			self.MyRushSide = hal.BOTTOM
		}
	}
}

func (self *Overmind) EnterGeneticAlgorithm() {

	play_perfect := (self.Config.Imperfect == false)

	if play_perfect {		// Sometimes we need to turn it off anyway

		relevant_enemies := self.Game.ShipsOwnedBy(self.RushEnemyID)

		if len(self.Game.MyShips()) == 3 && len(relevant_enemies) == 1 && self.Game.InitialPlayers() > 2 {
			play_perfect = false
		}

		if len(self.Game.MyShips()) == 1 {
			play_perfect = false
		}

		for _, ship := range relevant_enemies {
			if ship.DockedStatus != hal.UNDOCKED {
				play_perfect = false
			}
		}

		if self.Game.RunOfSames() > 10 && rand.Intn(5) == 0 {
			play_perfect = false
		}
	}

	gen.FightRush(self.Game, self.RushEnemyID, play_perfect)
}
