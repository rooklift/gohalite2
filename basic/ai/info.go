package ai

import (
	"sort"
)

// ----------------------------------------------

func (self *Game) GetShip(sid int) (*Ship, bool) {
	ret, ok := self.shipMap[sid]
	return ret, ok
}

func (self *Game) GetPlanet(plid int) (*Planet, bool) {
	ret, ok := self.planetMap[plid]
	return ret, ok
}

// ----------------------------------------------

func (self *Game) AllShips() []*Ship {
	ret := make([]*Ship, len(self.all_ships_cache))
	copy(ret, self.all_ships_cache)
	return ret
}

func (self *Game) AllPlanets() []*Planet {
	ret := make([]*Planet, len(self.all_planets_cache))
	copy(ret, self.all_planets_cache)
	return ret
}

func (self *Game) AllImmobile() []Entity {						// Returns all planets and all docked ships
	var ret []Entity

	for _, planet := range self.all_planets_cache {
		ret = append(ret, planet)
	}

	for _, ship := range self.all_ships_cache {
		if ship.DockedStatus != UNDOCKED {
			ret = append(ret, ship)
		}
	}

	return ret
}

// ----------------------------------------------

func (self *Game) ShipsOwnedBy(pid int) []*Ship {
	ret := make([]*Ship, len(self.playershipMap[pid]))
	copy(ret, self.playershipMap[pid])
	return ret
}

func (self *Game) MyShips() []*Ship {
	return self.ShipsOwnedBy(self.pid)
}

func (self *Game) EnemyShips() []*Ship {
	ret := make([]*Ship, len(self.enemy_ships_cache))
	copy(ret, self.enemy_ships_cache)
	return ret
}

// ----------------------------------------------

func (self *Game) MyNewShipIDs() []int {			// My ships born this turn.
	var ret []int
	for sid, _ := range self.shipMap {
		ship := self.shipMap[sid]
		if ship.Birth == self.turn && ship.Owner == self.pid {
			ret = append(ret, ship.Id)
		}
	}
	sort.Slice(ret, func(a, b int) bool {
		return ret[a] < ret[b]
	})
	return ret
}

func (self *Game) ShipsDockedAt(planet *Planet) []*Ship {
	ret := make ([]*Ship, len(self.dockMap[planet.Id]))
	copy(ret, self.dockMap[planet.Id])
	return ret
}

func (self *Game) RawWorld() string {
	return self.raw
}

func (self *Game) EnemiesNearPlanet(planet *Planet) []*Ship {

	// Counts mobile enemies and enemies docked at the planet,
	// but not docked enemies at a different nearby planet.

	ret := make([]*Ship, len(self.enemies_near_planet[planet.Id]))
	copy(ret, self.enemies_near_planet[planet.Id])
	return ret
}

func (self *Game) DesiredSpots(planet *Planet) int {

	// If we don't own it, we want all its spots...

	if planet.Owner != self.pid {
		return planet.DockingSpots
	}

	// If we do, we want to fill it up...

	return planet.OpenSpots()
}
