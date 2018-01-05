package ai

import (
	"sort"

	hal "../core"
)

func (self *Overmind) DecideRush() {

	// Can leave things undecided, in which case it will be called again next iteration.

	if self.Config.ForceRush {
		self.RushChoice = RUSHING
		return
	}

	if self.Game.InitialPlayers() > 2 {
		self.RushChoice = NOT_RUSHING
		return
	}

	if self.Config.Conservative || len(self.Game.AllShips()) > 6 || len(self.Game.MyShips()) < 3 {
		self.RushChoice = NOT_RUSHING
		return
	}

	centre_of_gravity := self.Game.AllShipsCentreOfGravity()

	my_ships := self.Game.MyShips()

	for _, ship := range my_ships {
		if ship.DockedStatus != hal.UNDOCKED {
			self.RushChoice = NOT_RUSHING
			return
		}
	}

	sort.Slice(my_ships, func(a, b int) bool {
		return my_ships[b].Dist(centre_of_gravity) < my_ships[b].Dist(centre_of_gravity)
	})

	if my_ships[0].Dist(centre_of_gravity) < 45 && my_ships[1].Dist(centre_of_gravity) < 48 && my_ships[2].Dist(centre_of_gravity) < 51 {
		self.RushChoice = RUSHING
		return
	}
}

func (self *Overmind) SetRushTargets() {

	// FIXME: since this can now happen on turns other than 0, we need a better way.

	// Sort our pilots by Y...

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].Y < self.Pilots[b].Y
	})

	// Sort enemies by Y...

	enemies := self.Game.ShipsOwnedBy(self.RushEnemyID)
	sort.Slice(enemies, func(a, b int) bool {
		return enemies[a].Y < enemies[b].Y
	})

	// Pair pilots with enemies, and lock in (prevents getting reset, though they can still swap, right?)

	for index, pilot := range self.Pilots {
		if len(enemies) > index {
			pilot.Target = enemies[index]
			pilot.Locked = true
		}
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

	// Don't allow these targets to change.

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

func (self *Overmind) Check2v1() {

	// Called when DetectRushFight() has already returned true, i.e. we want to enter the genetic algorithm.
	// FIXME: was written assuming we only enter GA in 2p.

	// There is a special case that loses occasional games if we don't handle it... basically, if we are
	// 2v1 up but the opponent ever produced a ship, we can't just chase him forever.

	if self.Game.CountMyShips() < 2 || self.Game.CountEnemyShips() > 1 {
		return
	}

	// So, we have 2 or more ships, and the enemy has just 1.

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].HP > self.Pilots[b].HP			// Note reversed sort, highest HP first
	})

	enemy_ship := self.Game.EnemyShips()[0]

	enemy_player_id := enemy_ship.Owner

	if self.Game.GetCumulativeShipCount(enemy_player_id) <= self.Game.GetCumulativeShipCount(self.Game.Pid()) {
		return
	}

	// So we are indeed in the losing situation... our response is, if possible, to assign 1 ship to chase
	// the enemy while our other ship(s) dock. Set Conservative so we never try to enter GA again.

	if enemy_ship.ShotsToKill() <= self.Pilots[0].ShotsToKill() && enemy_ship.DockedStatus == hal.UNDOCKED {

		self.Game.Log("Losing 2v1 situation detected, setting Config.Conservative and choosing targets.")

		self.Config.Conservative = true

		self.Pilots[0].Target = enemy_ship
		self.Pilots[0].Locked = true

		for i := 1; i < len(self.Pilots); i++ {
			self.Pilots[i].Target = self.Game.FarthestPlanet(self.Pilots[i].Ship)
			self.Pilots[i].Locked = true
		}
	}
}

func (self *Overmind) FindRushEnemy() {

	self.RushEnemyID = -1

	switch self.Game.InitialPlayers() {

	case 2:

		if self.Game.Pid() == 0 {
			self.RushEnemyID = 1
		} else {
			self.RushEnemyID = 2
		}

	case 4:

		if self.Game.Pid() == 0 {
			self.RushEnemyID = 2
		} else if self.Game.Pid() == 1 {
			self.RushEnemyID = 3
		} else if self.Game.Pid() == 2 {
			self.RushEnemyID = 0
		} else if self.Game.Pid() == 3 {
			self.RushEnemyID = 1
		}
	}
}

func (self *Overmind) SetTargetsAfterGenetic() {

	// Lets say that entering GA commits us to wiping the enemy out.
	// Make sure all our ships are targeting the enemy, for when we
	// drop out of GA, if we do.

	// Note that we're working with a state of the world that hasn't
	// been updated yet since we haven't sent our GA moves yet.

	is_targeted := make(map[int]bool)

	relevant_enemies := self.Game.ShipsOwnedBy(self.RushEnemyID)

	if len(relevant_enemies) == 0 {				// Should be impossible, the GA wouldn't have been called.
		return
	}

	for _, ship := range relevant_enemies {
		is_targeted[ship.Id] = false
	}

	for _, pilot := range self.Pilots {
		if pilot.Target.Type() == hal.SHIP {
			is_targeted[pilot.Target.GetId()] = true
		}
	}

	var new_targets []*hal.Ship

	for sid, _ := range is_targeted {
		if is_targeted[sid] == false {
			ship, _ := self.Game.GetShip(sid)
			new_targets = append(new_targets, ship)
		}
	}

	// If there are no untargeted ships, just send our targetless guys at whatever ship...

	if len(new_targets) == 0 {
		new_targets = relevant_enemies
	}

	for _, pilot := range self.Pilots {

		// The only way we can have a target is if it was locked in some turns ago.
		// If it wasn't locked, we would have cleared it at Step() start.

		if pilot.Target.Type() != hal.SHIP {

			sort.Slice(new_targets, func(a, b int) bool {
				return pilot.Dist(new_targets[a]) < pilot.Dist(new_targets[b])
			})

			pilot.Target = new_targets[0]
			pilot.Locked = true
			pilot.Fearless = true

			pilot.Log("Because of GA, gained target: %v", pilot.Target)

			// Better not send all our guys after the same enemy...

			if len(new_targets) > 1 {
				new_targets = new_targets[1:]
			}
		}
	}
}
