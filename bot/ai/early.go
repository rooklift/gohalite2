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

	if len(self.Game.EnemyShips()) < 3 {	// If enemy ships crash, just beat the enemy normally.
		self.RushChoice = NOT_RUSHING
		self.Game.Log("Not rushing because: len(self.Game.EnemyShips()) < 3")
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
		return my_ships[a].Dist(centre_of_gravity) < my_ships[b].Dist(centre_of_gravity)
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

	// The point of the cluster now is to have 2 ships close enough to obliterate a single
	// enemy ship before it can ram us. But we like to have a spread so we can turn without
	// interfering with each other.

	if self.Game.InitialPlayers() == 4 {

		switch self.Game.Pid() {
			case 0: fallthrough
			case 1:	self.Cluster(6, 90, 5, 110, 2, 50)
			case 2: fallthrough
			case 3: self.Cluster(2, 230, 5, 290, 6, 270)
		}

	} else {

		centre_of_gravity := self.Game.AllShipsCentreOfGravity()

		if centre_of_gravity.X > self.Pilots[0].X {
			self.Cluster(7, 15, 5, 0, 7, 345)
		} else if centre_of_gravity.X < self.Pilots[0].X {
			self.Cluster(7, 165, 5, 180, 7, 195)
		} else if centre_of_gravity.Y > self.Pilots[0].Y {
			self.Cluster(6, 90, 5, 110, 2, 50)
		} else if centre_of_gravity.Y < self.Pilots[0].Y {
			self.Cluster(2, 230, 5, 290, 6, 270)
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

	// I have more than 1 ship, or the enemy isn't docked at all.

	if len(my_ships) == 1 {
		for _, ship := range relevant_enemies {
			if ship.DockedStatus != hal.UNDOCKED {
				return false
			}
		}
	}

	// (v96) 3 enemy ships have been closed down at some point,
	// or we ever docked (i.e. we are BEING rushed).

	if len(self.RushEnemiesTouched) < 3 {

		for _, enemy := range relevant_enemies {
			for _, ship := range my_ships {
				if enemy.Dist(ship) < 30 {
					self.RushEnemiesTouched[enemy.Id] = true
					break
				}
			}
		}

		if len(self.RushEnemiesTouched) < 3 && self.EverDocked == false {
			return false
		}
	}

	return true
}

func (self *Overmind) ChooseThreeDocks() {

	if self.Game.InitialPlayers() == 2 {
		self.MakeDefaultDockChoice()
		self.ConsiderCentreDocks()
	}

	if self.Game.InitialPlayers() == 4 {
		self.MakeDefaultDockChoice()
	}

	for _, pilot := range self.Pilots {
		pilot.Locked = true
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

func (self *Overmind) MakeDefaultDockChoice() {

	my_cog := self.Game.MyShipsCentreOfGravity()

	// Sort all planets by distance to our fleet...

	all_planets := self.Game.AllPlanets()

	sort.Slice(all_planets, func(a, b int) bool {
		return my_cog.ApproachDist(all_planets[a]) < my_cog.ApproachDist(all_planets[b])
	})

	closest_three := all_planets[:3]

	// Get docks...

	var docks []*hal.Port

	for _, planet := range closest_three {
		if self.Config.Split {
			docks = append(docks, hal.OpeningDockHelper(2, planet, my_cog)...)
		} else {
			docks = append(docks, hal.OpeningDockHelper(3, planet, my_cog)...)
		}
	}

	if len(docks) < 3 {
		return
	}

	self.SetNonIntersectingDockPaths(docks[:3])
}

func (self *Overmind) ConsiderCentreDocks() {

	// We've already chosen our default docks. Should we, instead, send all our ships to the centre?

	yes := false

	ships_to_centre := 0

	for _, pilot := range self.Pilots {
		if pilot.Target.GetId() < 4 {			// Centre planets have ID < 4, unless the engine changes.
			ships_to_centre++
		}
	}

	if ships_to_centre == 3 {					// Just to make logging consistent.
		yes = true
	}

	if yes == false && self.Config.Centre {
		self.Game.Log("Sending all ships to centre because of --centre flag.")
		yes = true
	}

	if yes == false && ships_to_centre > 0 && ships_to_centre < 3 {
		self.Game.Log("Was going to send %d ships to centre; sending all instead.", ships_to_centre)
		yes = true
	}

	// Difficult case: maybe the centre is close enough...

	if yes == false {

		sort.Slice(self.Pilots, func(a, b int) bool {
			return self.Pilots[a].Y < self.Pilots[b].Y
		})

		d := self.Pilots[1].Dist(self.Pilots[1].Target)		// Note that this target is a port, not a planet.

		a, _ := self.Game.GetPlanet(0)
		b, _ := self.Game.GetPlanet(2)

		cd := hal.MinFloat(self.Pilots[1].Dist(a), self.Pilots[1].Dist(b))

		if cd - d < 21 {
			self.Game.Log("Centre planets are close enough (diff == %v), going there.", cd - d)
			yes = true
		}
	}

	if yes {
		self.ChooseCentreDocks()
	}
}

func (self *Overmind) ChooseCentreDocks() {

	my_cog := self.Game.MyShipsCentreOfGravity()

	var centre_planets []*hal.Planet

	for n := 0; n < 4; n++ {
		planet, ok := self.Game.GetPlanet(n)
		if ok {
			centre_planets = append(centre_planets, planet)
		}
	}

	sort.Slice(centre_planets, func(a, b int) bool {
		return my_cog.Dist(centre_planets[a]) < my_cog.Dist(centre_planets[b])
	})

	var docks []*hal.Port

	for _, planet := range centre_planets {
		docks = append(docks, hal.OpeningDockHelper(3, planet, my_cog)...)
	}

	if len(docks) < 3 {
		return
	}

	self.SetNonIntersectingDockPaths(docks[:3])
}

/*
func (self *Overmind) MakeSafeDockChoice() {

	my_cog := self.Game.MyShipsCentreOfGravity()
	enemy_cog := self.Game.PartialCentreOfGravity(self.RushEnemyID)

	// Find three planets closest to me...

	all_planets := self.Game.AllPlanets()

	sort.Slice(all_planets, func(a, b int) bool {
		return my_cog.ApproachDist(all_planets[a]) < my_cog.ApproachDist(all_planets[b])
	})

	closest_three := all_planets[:3]

	// Sort by range to enemy...

	sort.Slice(closest_three, func(a, b int) bool {
		return closest_three[a].Dist(enemy_cog) < closest_three[b].Dist(enemy_cog)		// ApproachDist not so relevant here
	})

	// Discard planet closest to enemy if we can't reach it in time...

	if closest_three[0].Dist(enemy_cog) - closest_three[0].Dist(my_cog) < 70 {
		self.Game.Log("Discarding planet %v", closest_three[0].Id)
		closest_three = closest_three[1:]		// Now it's closest two, but whatever.
	}

	// Re-sort the surviving planets by distance to me...

	sort.Slice(closest_three, func(a, b int) bool {
		return my_cog.ApproachDist(closest_three[a]) < my_cog.ApproachDist(closest_three[b])
	})

	// Get docks...

	var docks []*hal.Port

	for _, planet := range closest_three {
		if self.Config.Split {
			docks = append(docks, hal.OpeningDockHelper(2, planet, self.Pilots[0].Ship)...)
		} else {
			docks = append(docks, hal.OpeningDockHelper(3, planet, self.Pilots[0].Ship)...)
		}
	}

	if len(docks) < 3 {
		return
	}

	self.SetNonIntersectingDockPaths(docks[:3])
}

*/

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
	self.Pilots[0].Locked = true

	for i := 1; i < len(self.Pilots); i++ {
		self.Pilots[i].Target = self.Game.FarthestPlanet(all_enemies[0])
		self.Pilots[i].Locked = true
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

	// NOTE! Can be called by MyBot.go for debugging purposes, in which case self.Pilots won't be up to date.
	// If we need to use self.Pilots here, do something about that.

	play_perfect := (self.Config.Imperfect == false)

	if play_perfect {		// Sometimes we need to turn it off anyway

		relevant_enemies := self.Game.ShipsOwnedBy(self.RushEnemyID)
		my_ships := self.Game.MyShips()

		// We are well winning but it's a 4p game and we need to end this quickly...

		if len(my_ships) == 3 && len(relevant_enemies) == 1 && self.Game.InitialPlayers() > 2 {
			play_perfect = false
		}

		// We are 3v2 or 3v1 or 2v1 on ships but the enemy built a ship and we're running out of time.
		// Some case not handled explicitly by the Avoid2v1() case.

		if len(my_ships) >= 2 && len(relevant_enemies) < len(my_ships) {
			if self.Game.GetCumulativeShipCount(self.RushEnemyID) > self.Game.GetCumulativeShipCount(self.Game.Pid()) {
				if self.Game.Turn() > 150 {
					play_perfect = false
				}
			}
		}

		// The enemy has some docked ships, gotta be more aggro...

		for _, ship := range relevant_enemies {
			if ship.DockedStatus != hal.UNDOCKED && ship.DockedStatus != hal.UNDOCKING {
				play_perfect = false
			}
		}

		// Nothing's happened for a while...

		if self.Game.RunOfSames() > 10 && rand.Intn(5) == 0 {
			if play_perfect {
				self.Game.Log("Taking a stab in the dark")
			}
			play_perfect = false
		}
	}

	gen.EvolveRush(self.Game, self.RushEnemyID, play_perfect)
}
