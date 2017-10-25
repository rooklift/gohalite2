package ai

import (
	"math/rand"
	hal "../gohalite2"
)

type Pilot struct {
	overmind		*Overmind
	game			*hal.Game
	id				int					// Ship ID
	target_type		int					// NONE (zero) / SHIP / PLANET
	target_id		int
	course			int
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

func (self *Pilot) DockIfPossible() {
	if self.Ship().DockedStatus == hal.UNDOCKED {
		closest_planet := self.ClosestPlanet()
		if self.Ship().CanDock(closest_planet) {
			self.Dock(closest_planet.Id)
		}
	}
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
	game := self.game

	self.ValidateTarget()		// Clear dead / totally conquered targets...

	// Dock whenever possible...

	if self.CurrentOrder() == "" {
		self.DockIfPossible()
	}

	// Choose target and course...

	if self.CurrentOrder() == "" && self.target_type == hal.NONE {

		for n := 0; n < 10; n++ {

			degrees := rand.Intn(360)

			plid := game.AngleCollisionID(self.Ship().X, self.Ship().Y, 9999, degrees)

			if plid == -1 {
				continue
			}

			planet := game.GetPlanet(plid)

			if planet.Owner != game.Pid() || planet.IsFull() == false {

				self.target_id = plid
				self.target_type = hal.PLANET
				self.course = degrees
				break
			}
		}
	}

	// Move...

	if self.CurrentOrder() == "" {

		if self.target_type == hal.NONE || self.Ship().DockedStatus != hal.UNDOCKED {
			return
		}

		if self.game.AngleCollisionID(self.Ship().X, self.Ship().Y, 7, self.course) != -1 {
			self.Thrust(3, self.course)
		} else {
			self.Thrust(7, self.course)
		}
	}
}
