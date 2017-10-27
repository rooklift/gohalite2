package ai

import (
	hal "../gohalite2"
)

// --------------------------------------------

type Overmind struct {
	game			*hal.Game
	pilots			[]*Pilot
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
		pilot := new(Pilot)
		pilot.overmind = self
		pilot.game = game
		pilot.id = sid
		self.pilots = append(self.pilots, pilot)
	}

	// Delete AIs with dead ships from the slice...

	for i := 0; i < len(self.pilots); i++ {
		pilot := self.pilots[i]
		if pilot.Ship().Alive() == false {
			self.pilots = append(self.pilots[:i], self.pilots[i+1:]...)
			i--
		}
	}
}

// --------------------------------------------

func (self *Overmind) Step() {

	self.UpdatePilots()		// Fix the AI slices by adding / deleting AIs...

	for _, pilot := range self.pilots {
		pilot.Act()
	}
}
