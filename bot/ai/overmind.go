package ai

import (
	atc "../atc"
	hal "../core"
	pil "../pilot"
)

// --------------------------------------------

type Overmind struct {
	Pilots					[]*pil.Pilot
	Game					*hal.Game
	ATC						*atc.AirTrafficControl
	EnemyMap				map[int][]hal.Ship		// Planet ID --> Enemy ships near the planet (not docked)
	FriendlyMap				map[int][]hal.Ship		// Planet ID --> Friendly ships near the planet (not docked)
	ShipsDockingCount		map[int]int				// Planet ID --> My ship count docking this turn
	EnemyShipChasers		map[int][]int			// Enemy Ship ID --> slice of my IDs chasing it
}

func NewOvermind(game *hal.Game) *Overmind {
	ret := new(Overmind)
	ret.Game = game
	ret.ATC = atc.NewATC(game)
	return ret
}

// pilot.go requires overminds to have the following interface satisfiers...

func (self *Overmind) NotifyTargetChange(pilot *pil.Pilot, old_target, new_target hal.Entity) {

	if old_target.Type() == hal.SHIP {
		self.EnemyShipChasers[old_target.(hal.Ship).Id] = hal.IntSliceWithout(self.EnemyShipChasers[old_target.(hal.Ship).Id], pilot.Id)
	}

	if new_target.Type() == hal.SHIP {
		self.EnemyShipChasers[new_target.(hal.Ship).Id] = append(self.EnemyShipChasers[new_target.(hal.Ship).Id], pilot.Id)
	}
}

func (self *Overmind) NotifyDock(planet hal.Planet) {
	self.ShipsDockingCount[planet.Id]++
}

func (self *Overmind) ShipsDockingAt(planet hal.Planet) int {
	return self.ShipsDockingCount[planet.Id]
}

func (self *Overmind) ShipsChasing(ship hal.Ship) int {
	return len(self.EnemyShipChasers[ship.Id])
}

func (self *Overmind) EnemiesNearPlanet(planet hal.Planet) []hal.Ship {
	ret := make([]hal.Ship, len(self.EnemyMap[planet.Id]))
	copy(ret, self.EnemyMap[planet.Id])
	return ret
}
