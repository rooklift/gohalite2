package ai

import (
	"os"
	"sort"
	hal "../gohalite2"
)

// --------------------------------------------

type Overmind struct {
	Pilots					[]*Pilot
	Game					*hal.Game
	ATC						*AirTrafficControl
	EnemyMap				map[int][]hal.Ship		// Planet ID --> Enemy ships near the planet (not docked)
	FriendlyMap				map[int][]hal.Ship		// Planet ID --> Friendly ships near the planet (not docked)
	ShipsDockingMap			map[int]int				// Planet ID --> My ship count docking this turn
	EnemyShipsChased		map[int][]int			// Enemy Ship ID --> slice of my IDs chasing it
}

func NewOvermind(game *hal.Game) *Overmind {
	ret := new(Overmind)
	ret.Game = game
	ret.ATC = NewATC(game)
	return ret
}

func (self *Overmind) Step() {

	self.UpdatePilots()
	self.UpdateProximityMaps()
	self.UpdateShipChases()							// Must happen after self.Pilots is updated
	self.ShipsDockingMap = make(map[int]int)
	self.ATC.Clear()

	if self.Game.Turn() == 0 {
		assassin := self.ChooseInitialTargets()
		if assassin {
			self.TurnZeroCluster()
		} else {
			self.ExecuteMoves()
		}
		return
	}

	if self.DetectRushFight() {
		self.Game.LogOnce("Entering dangerous 3v3!")
		FightRush(self.Game)
	} else {
		self.ExecuteMoves()
	}
}

func (self *Overmind) ChooseInitialTargets() bool {		// Returns: are we assassinating?

	// As a good default...

	self.ChooseThreeDocks()

	if StringSliceContains(os.Args, "--conservative") {
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

/*

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
		pilot.SetTarget(closest_three[index])
	}
}

*/

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

		if Intersect(self.Pilots[0].Ship, self.Pilots[0].Target, self.Pilots[1].Ship, self.Pilots[1].Target) {
			continue
		}

		if Intersect(self.Pilots[0].Ship, self.Pilots[0].Target, self.Pilots[2].Ship, self.Pilots[2].Target) {
			continue
		}

		if Intersect(self.Pilots[1].Ship, self.Pilots[1].Target, self.Pilots[2].Ship, self.Pilots[2].Target) {
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
/*
	if centre_of_gravity.X > self.Pilots[0].X {
		self.Cluster(7, 10, 6, 0, 7, 350)
	} else if centre_of_gravity.X < self.Pilots[0].X {
		self.Cluster(7, 170, 6, 180, 7, 190)
	} else if centre_of_gravity.Y > self.Pilots[0].Y {
		self.Cluster(7, 90, 6, 100, 3, 60)
	} else if centre_of_gravity.Y < self.Pilots[0].Y {
		self.Cluster(3, 240, 6, 280, 7, 270)
	}
*/

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

/*

func (self *Overmind) DetectFailedRush() bool {

	if len(self.Pilots) != 3 {
		return false
	}

	for _, pilot := range self.Pilots {
		if pilot.Target.Type() != hal.SHIP {
			return false
		}
	}

	mid_point := hal.Point{
		X: float64(self.Game.Width()) / 2,
		Y: float64(self.Game.Height()) / 2,
	}

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].Dist(mid_point) < self.Pilots[b].Dist(mid_point)
	})

	if self.Pilots[0].Dist(mid_point) > 20 {
		return false
	}

	enemy_ships := self.Game.EnemyShips()

	sort.Slice(enemy_ships, func(a, b int) bool {
		return enemy_ships[a].Dist(mid_point) < enemy_ships[b].Dist(mid_point)
	})

	if enemy_ships[0].Dist(mid_point) < 50 {
		return false
	}

	return true
}

*/
