package ai

import (
	hal "../gohalite2"
)

// --------------------------------------------

type ShipAI struct {
	overmind		*Overmind
	game			*hal.Game
	id				int
	target_type		int					// NONE (zero) / SHIP / PLANET
	target_id		int
}

func (self *ShipAI) State() hal.Ship {
	return self.game.GetShip(self.id)
}

func (self *ShipAI) ValidateTarget() {

	game := self.game

	if self.target_type == hal.SHIP {
		target := game.GetShip(self.target_id)
		if target.HP <= 0 {
			self.target_type = hal.NONE
		}
	} else if self.target_type == hal.PLANET {
		target := game.GetPlanet(self.target_id)
		if target.HP <= 0 {
			self.target_type = hal.NONE
		}
	}
}

// --------------------------------------------

type PlanetAI struct {
	overmind		*Overmind
	game			*hal.Game
	id				int
}

func (self *PlanetAI) State() hal.Planet {
	return self.game.GetPlanet(self.id)
}

// --------------------------------------------

type Overmind struct {
	game			*hal.Game
	shipAIs			[]*ShipAI
	planetAIs		[]*PlanetAI
}

func NewOvermind(game *hal.Game) *Overmind {
	ret := new(Overmind)
	ret.game = game
	return ret
}

func (self *Overmind) UpdateShipAIs() {

	game := self.game

	// Add new AIs for new ships...

	my_new_ships := game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		ship_ai := new(ShipAI)
		ship_ai.overmind = self
		ship_ai.game = game
		ship_ai.id = sid
		self.shipAIs = append(self.shipAIs, ship_ai)
		game.Log("Turn %d: received ship %d", game.Turn(), sid)
	}

	// Delete AIs with dead ships from the slice...

	for i := 0; i < len(self.shipAIs); i++ {
		ship_ai := self.shipAIs[i]
		if ship_ai.State().HP <= 0 {
			self.shipAIs = append(self.shipAIs[:i], self.shipAIs[i+1:]...)
			i--
			game.Log("Turn %d: ship %d destroyed", game.Turn(), ship_ai.id)
		}
	}

	// Clear dead targets...

	for _, ship_ai := range self.shipAIs {
		ship_ai.ValidateTarget()
	}
}

func (self *Overmind) UpdatePlanetAIs() {

	game := self.game

	// Create AIs on turn 1...

	if game.Turn() == 1 {
		for _, planet := range game.AllPlanets() {
			planet_ai := new(PlanetAI)
			planet_ai.overmind = self
			planet_ai.game = game
			planet_ai.id = planet.Id
			self.planetAIs = append(self.planetAIs, planet_ai)
		}
	}

	// Delete AIs with dead planets from the slice...

	for i := 0; i < len(self.planetAIs); i++ {
		planet_ai := self.planetAIs[i]
		if planet_ai.State().HP <= 0 {
			self.planetAIs = append(self.planetAIs[:i], self.planetAIs[i+1:]...)
			i--
			game.Log("Turn %d: planet %d destroyed", game.Turn(), planet_ai.id)
		}
	}
}

// --------------------------------------------

func (self *Overmind) Step() {
	self.UpdateShipAIs()
	self.UpdatePlanetAIs()
}
