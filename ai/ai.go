package ai

import (
	hal "../gohalite2"
)

// --------------------------------------------

type ShipAI struct {
	mainAI			*AI
	game			*hal.Game
	id				int
}

func (self *ShipAI) State() hal.Ship {
	return self.game.GetShip(self.id)
}

// --------------------------------------------

type PlanetAI struct {
	mainAI			*AI
	game			*hal.Game
	id				int
}

func (self *PlanetAI) State() hal.Planet {
	return self.game.GetPlanet(self.id)
}

// --------------------------------------------

type AI struct {
	game			*hal.Game
	shipAIs			[]*ShipAI
	planetAIs		[]*PlanetAI
}

func NewAI(game *hal.Game) *AI {
	ret := new(AI)
	ret.game = game
	return ret
}

func (self *AI) Step() {
	self.UpdateShipAIs()
	self.UpdatePlanetAIs()
}

func (self *AI) UpdateShipAIs() {

	game := self.game

	// Add new AIs for new ships...

	my_new_ships := game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		ship_ai := new(ShipAI)
		ship_ai.mainAI = self
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
}

func (self *AI) UpdatePlanetAIs() {

	game := self.game

	if game.Turn() == 1 {

		for _, planet := range game.AllPlanets() {
			planet_ai := new(PlanetAI)
			planet_ai.mainAI = self
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
