package ai

import (
	"sort"
	hal "../gohalite2"
)

// --------------------------------------------

type Overmind struct {
	Pilots					[]*Pilot
	Game					*hal.Game
	ATC						*AirTrafficControl
	EnemyMap				map[int][]hal.Ship		// Planet ID --> Enemy ships near the planet
}

func NewOvermind(game *hal.Game) *Overmind {
	ret := new(Overmind)
	ret.Game = game
	ret.ATC = NewATC(game)
	return ret
}

func (self *Overmind) Step() {

	self.UpdatePilots()
	self.UpdateProximityMap()
	self.ATC.Clear()

	if self.Game.Turn() == 0 {
		self.ChooseInitialTargets()
	}

	if self.DetectRushFight() {
		self.Game.LogOnce("Entering dangerous 3v3!")
		self.ExecuteMoves()
	} else {
		self.ExecuteMoves()
	}
}

func (self *Overmind) ChooseInitialTargets() bool {		// Returns: are we assassinating?

	// As a good default...

	self.ChooseThreePlanets()

	if self.Game.InitialPlayers() == 2 {
		if self.Game.MyShips()[0].Dist(self.Game.EnemyShips()[0]) < 150 {
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
		pilot.Target = enemies[index]
	}
}

func (self *Overmind) ChooseThreePlanets() {

	// Sort our pilots by Y...

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].Y < self.Pilots[b].Y
	})

	// Sort all planets by distance to our fleet...

	all_planets := self.Game.AllPlanets()

	sort.Slice(all_planets, func(a, b int) bool {
		return all_planets[a].Dist(self.Pilots[0]) < all_planets[b].Dist(self.Pilots[0])
	})

	// Sort closest 3 planets by Y...

	closest_three := all_planets[:3]

	sort.Slice(closest_three, func(a, b int) bool {
		return closest_three[a].Y < closest_three[b].Y
	})

	// Pair pilots with planets...

	for index, pilot := range self.Pilots {
		pilot.Target = closest_three[index]
	}
}

/*

func (self *Overmind) ChooseThreeDocks() {				// Pretty bad in internal testing.

	// Sort our pilots by Y...

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].Y < self.Pilots[b].Y
	})

	// Sort all planets by distance to our fleet...

	all_planets := self.Game.AllPlanets()

	sort.Slice(all_planets, func(a, b int) bool {
		return all_planets[a].Dist(self.Pilots[0]) < all_planets[b].Dist(self.Pilots[0])
	})

	closest_three := all_planets[:3]

	// Get docks...

	var docks []hal.Planet

	for _, planet := range closest_three {
		for n := 0; n < planet.OpenSpots(); n++ {
			docks = append(docks, planet)
		}
	}

	// Sort closest 3 docks by Y...

	closest_three_docks := docks[:3]

	sort.Slice(closest_three_docks, func(a, b int) bool {
		return closest_three_docks[a].Y < closest_three_docks[b].Y
	})

	// Pair pilots with planets...

	for index, pilot := range self.Pilots {
		pilot.Target = closest_three_docks[index]
	}
}

*/

func (self *Overmind) DetectRushFight() bool {

	// <= 3 ships each
	// All ships near each other
	// No docked ships on map

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
			if ship.DockedStatus != hal.UNDOCKED {
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

/*

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

*/
