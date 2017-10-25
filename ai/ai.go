package ai

import (
	hal "../gohalite2"
)

// --------------------------------------------

type Overmind struct {
	game			*hal.Game
	Pilots			[]*Pilot
}

func NewOvermind(game *hal.Game) *Overmind {
	ret := new(Overmind)
	ret.game = game
	return ret
}

func (self *Overmind) UpdatePilots() {

	game := self.game

	// Add new AIs for new ships...

	my_new_ships := game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		ship_ai := new(Pilot)
		ship_ai.overmind = self
		ship_ai.game = game
		ship_ai.id = sid
		self.Pilots = append(self.Pilots, ship_ai)
	}

	// Delete AIs with dead ships from the slice...

	for i := 0; i < len(self.Pilots); i++ {
		ship_ai := self.Pilots[i]
		if ship_ai.Ship().Alive() == false {
			self.Pilots = append(self.Pilots[:i], self.Pilots[i+1:]...)
			i--
		}
	}
}

// --------------------------------------------

func (self *Overmind) Step() {

	self.UpdatePilots()		// Fix the AI slices by adding / deleting AIs...

	for _, pilot := range self.Pilots {
		pilot.Act()
	}
}
