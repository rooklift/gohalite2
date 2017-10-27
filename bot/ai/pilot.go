package ai

import (
	"math/rand"
	hal "../gohalite2"
)

type Pilot struct {
	overmind		*Overmind
	game			*hal.Game
	id				int					// Ship ID
	target_type		hal.EntityType
	target_id		int
}

func (self *Pilot) Ship() hal.Ship {
	return self.game.GetShip(self.id)
}

func (self *Pilot) ValidateTarget() {

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

func (self *Pilot) ClosestPlanet() hal.Planet {
	return self.game.ClosestPlanet(self.Ship().X, self.Ship().Y)
}

func (self *Pilot) Thrust(speed, angle int) {
	self.game.Thrust(self.id, speed, angle)
}

func (self *Pilot) Dock(planet int) {
	self.game.Dock(self.id, planet)
}

func (self *Pilot) Undock() {
	self.game.Undock(self.id)
}

func (self *Pilot) ClearOrder() {
	self.game.ClearOrder(self.id)
}

func (self *Pilot) CurrentOrder() string {
	return self.game.CurrentOrder(self.id)
}

func (self *Pilot) Act() {

	// Clear dead / totally conquered targets...

	self.ValidateTarget()

	// Helpers can lock in an order by actually setting it.

	if self.CurrentOrder() == "" {
		self.DockIfPossible()
	}

	if self.CurrentOrder() == "" {
		self.ChooseTarget()
	}

	if self.CurrentOrder() == "" {
		self.EngageTarget()
	}
}

func (self *Pilot) DockIfPossible() {
	if self.Ship().DockedStatus == hal.UNDOCKED {
		closest_planet := self.ClosestPlanet()
		if self.Ship().CanDock(closest_planet) {
			self.Dock(closest_planet.Id)
		}
	}
}

func (self *Pilot) ChooseTarget() {
	game := self.game

	if self.target_type != hal.NONE {		// We already have a target.
		return
	}

	all_planets := game.AllPlanets()

	for n := 0; n < 5; n++ {

		i := rand.Intn(len(all_planets))
		planet := all_planets[i]

		if planet.Owner != game.Pid() || planet.IsFull() == false {
			self.target_id = planet.Id
			self.target_type = hal.PLANET
			break
		}
	}
}

func (self *Pilot) EngageTarget() {
	game := self.game

	if self.target_type == hal.NONE || self.Ship().DockedStatus != hal.UNDOCKED {
		return
	}

	if self.target_type == hal.PLANET {

		planet := game.GetPlanet(self.target_id)

		speed, degrees, err := game.Approach(self.Ship(), planet, 6.0)

		if err != nil {
			self.target_type = hal.NONE
		} else {
			if speed < 4 {
				// We could go faster and reach docking range, or do something else.
			} else {
				self.Thrust(speed, degrees)
			}
		}
	}
}
