package ai

import (
	atc "../atc"
	hal "../core"
	pil "../pilot"
)

// --------------------------------------------

type Overmind struct {
	Pilots					[]*pil.Pilot			// Not kept in any particular order, can be sorted by whatever
	Game					*hal.Game
	ATC						*atc.AirTrafficControl
	ShipsDockingCount		map[int]int				// Planet ID --> My ship count docking this turn
	EnemyShipChasers		map[int][]int			// Enemy Ship ID --> slice of my IDs chasing it
	PlanetChasers			map[int][]int			// Planet ID --> slice of my IDs going there
}

func NewOvermind(game *hal.Game) *Overmind {
	ret := new(Overmind)
	ret.Game = game
	ret.ATC = atc.NewATC(game)

	if ret.Game.InitialPlayers() == 2 {
		ret.Game.SetThreatRange(20)					// This value seems to be surprisingly fine-tuned
	} else {
		ret.Game.SetThreatRange(10)
	}

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

	if old_target.Type() == hal.PLANET {
		self.PlanetChasers[old_target.(hal.Planet).Id] = hal.IntSliceWithout(self.PlanetChasers[old_target.(hal.Planet).Id], pilot.Id)
	}

	if new_target.Type() == hal.PLANET {
		self.PlanetChasers[new_target.(hal.Planet).Id] = append(self.PlanetChasers[new_target.(hal.Planet).Id], pilot.Id)
	}
}

func (self *Overmind) NotifyDock(planet hal.Planet) {
	self.ShipsDockingCount[planet.Id]++
}

// Other useful utilities...

func (self *Overmind) ShipsChasing(ship hal.Ship) int {
	return len(self.EnemyShipChasers[ship.Id])
}

func (self *Overmind) ShipsAboutToDock(planet hal.Planet) int {
	return self.ShipsDockingCount[planet.Id]
}

// Updaters called by Step()...

func (self *Overmind) UpdatePilots() {

	game := self.Game

	// Add new AIs for new ships...

	my_new_ships := game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		pilot := pil.NewPilot(sid, game, self)
		self.Pilots = append(self.Pilots, pilot)
	}

	// Set various variables to initial state, but keeping current target.
	// Also update target info from the Game. Also delete pilot if the ship is dead.

	for i := 0; i < len(self.Pilots); i++ {
		pilot := self.Pilots[i]
		alive := pilot.ResetAndUpdate(true)
		if alive == false {
			self.Pilots = append(self.Pilots[:i], self.Pilots[i+1:]...)
			i--
		}
		if pilot.Target == nil {
			panic("nil pilot.Target")
		}
	}
}

func (self *Overmind) UpdateChasers() {

	self.EnemyShipChasers = make(map[int][]int)
	self.PlanetChasers = make(map[int][]int)

	for _, pilot := range self.Pilots {

		if pilot.DockedStatus != hal.UNDOCKED {
			continue
		}

		if pilot.Target.Type() == hal.SHIP {

			target := pilot.Target.(hal.Ship)
			self.EnemyShipChasers[target.Id] = append(self.EnemyShipChasers[target.Id], pilot.Id)

		} else if pilot.Target.Type() == hal.PLANET {

			target := pilot.Target.(hal.Planet)
			self.PlanetChasers[target.Id] = append(self.PlanetChasers[target.Id], pilot.Id)

		}
	}
}
