package ai

import (
	"math/rand"

	hal "../gohalite2"
)

// --------------------------------------------

type ShipAI struct {
	overmind		*Overmind
	game			*hal.Game
	id				int
	target_type		int					// NONE (zero) / SHIP / PLANET
	target_id		int
	course			int
}

func (self *ShipAI) Ship() hal.Ship {
	return self.game.GetShip(self.id)
}

func (self *ShipAI) ValidateTarget() {

	game := self.game

	if self.target_type == hal.SHIP {
		target := game.GetShip(self.target_id)
		if target.Alive() == false {
			self.target_type = hal.NONE
		}
	} else if self.target_type == hal.PLANET {
		target := game.GetPlanet(self.target_id)
		if target.Alive() == false {
			self.target_type = hal.NONE
		} else if target.Owner == game.Pid() && target.IsFull() {
			self.target_type = hal.NONE
		}
	}
}

func (self *ShipAI) ClosestPlanet() hal.Planet {
	return self.game.ClosestPlanet(self.Ship().X, self.Ship().Y)
}

func (self *ShipAI) DockIfPossible() {
	if self.Ship().DockedStatus == hal.UNDOCKED {
		closest_planet := self.ClosestPlanet()
		if self.Ship().CanDock(closest_planet) {
			self.game.Log("Docking ship %d", self.id)
			self.Dock(closest_planet.Id)
		}
	}
}

func (self *ShipAI) Thrust(speed, angle int) {
	self.game.Thrust(self.id, speed, angle)
}

func (self *ShipAI) Dock(planet int) {
	self.game.Dock(self.id, planet)
}

func (self *ShipAI) Undock() {
	self.game.Undock(self.id)
}

func (self *ShipAI) ClearOrder() {
	self.game.ClearOrder(self.id)
}

func (self *ShipAI) CurrentOrder() string {
	return self.game.CurrentOrder(self.id)
}

func (self *ShipAI) Act() {

	if self.target_type == hal.NONE || self.Ship().DockedStatus != hal.UNDOCKED {
		return
	}

	if self.game.AngleCollisionID(self.Ship().X, self.Ship().Y, 7, self.course) != -1 {
		self.Thrust(3, self.course)
	} else {
		self.Thrust(7, self.course)
	}
}

// --------------------------------------------

type PlanetAI struct {
	overmind		*Overmind
	game			*hal.Game
	id				int
}

func (self *PlanetAI) Planet() hal.Planet {
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
		if ship_ai.Ship().Alive() == false {
			self.shipAIs = append(self.shipAIs[:i], self.shipAIs[i+1:]...)
			i--
			game.Log("Turn %d: ship %d destroyed", game.Turn(), ship_ai.id)
		}
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
		if planet_ai.Planet().Alive() == false {
			self.planetAIs = append(self.planetAIs[:i], self.planetAIs[i+1:]...)
			i--
			game.Log("Turn %d: planet %d destroyed", game.Turn(), planet_ai.id)
		}
	}
}

// --------------------------------------------

func (self *Overmind) Step() {

	game := self.game

	// Fix the AI slices by adding / deleting AIs...

	self.UpdateShipAIs()
	self.UpdatePlanetAIs()

	// Clear dead / totally conquered targets...

	for _, ship_ai := range self.shipAIs {
		ship_ai.ValidateTarget()
	}

	// Dock whenever possible...

	for _, ship_ai := range self.shipAIs {
		if ship_ai.CurrentOrder() == "" {
			ship_ai.DockIfPossible()
		}
	}

	for _, ship_ai := range self.shipAIs {

		if ship_ai.CurrentOrder() == "" && ship_ai.target_type == hal.NONE {

			for n := 0; n < 10; n++ {

				degrees := rand.Intn(360)

				plid := game.AngleCollisionID(ship_ai.Ship().X, ship_ai.Ship().Y, 9999, degrees)

				if plid == -1 {
					continue
				}

				planet := game.GetPlanet(plid)

				if planet.Owner != game.Pid() || planet.IsFull() == false {

					ship_ai.target_id = plid
					ship_ai.target_type = hal.PLANET
					ship_ai.course = degrees
					break
				}
			}
		}
	}

	// Each ship now acts based on its assigned target...

	for _, ship_ai := range self.shipAIs {
		if ship_ai.CurrentOrder() == "" {
			ship_ai.Act()
		}
	}
}
