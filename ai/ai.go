package ai

import (
	hal "../gohalite2"
)

type ShipAI struct {
	game			*hal.Game
	sid				int
}

func (self *ShipAI) State() hal.Ship {
	return self.game.GetShip(self.sid)
}

type AI struct {
	game			*hal.Game
	shipAIs			[]*ShipAI
}

func NewAI(game *hal.Game) *AI {
	ret := new(AI)
	ret.game = game
	return ret
}

func (self *AI) Step() {
	self.UpdateShipAIs()
}

func (self *AI) UpdateShipAIs() {

	game := self.game

	// Add new AIs for new ships...

	my_new_ships := game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		ship_ai := new(ShipAI)
		ship_ai.game = game
		ship_ai.sid = sid
		self.shipAIs = append(self.shipAIs, ship_ai)
		game.Log("Turn %d: received ship %d", game.Turn(), sid)
	}

	// Delete AIs with dead ships from the slice...

	for i := 0; i < len(self.shipAIs); i++ {
		ship_ai := self.shipAIs[i]
		if ship_ai.State().HP <= 0 {
			self.shipAIs = append(self.shipAIs[:i], self.shipAIs[i+1:]...)
			game.Log("Turn %d: ship %d destroyed", game.Turn(), ship_ai.sid)
		}
	}
}
