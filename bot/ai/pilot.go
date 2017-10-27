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

func (self *Pilot) Dock(planet hal.Planet) {
	self.game.Dock(self.id, planet.Id)
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

func (self *Pilot) CanDock(p hal.Planet) bool {
	return self.Ship().CanDock(p)
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
		self.ChaseTarget()
	}
}

func (self *Pilot) DockIfPossible() {
	if self.Ship().DockedStatus == hal.UNDOCKED {
		closest_planet := self.ClosestPlanet()
		if self.Ship().CanDock(closest_planet) {
			self.Dock(closest_planet)
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

func (self *Pilot) ChaseTarget() {
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
				self.EngagePlanet()
			} else {
				self.Thrust(speed, degrees)
			}
		}
	}
}

func (self *Pilot) EngagePlanet() {
	game := self.game

	// We are very close to our target planet. Do something about this.

	if self.target_type != hal.PLANET {
		game.Log("Pilot %d: EngagePlanet() called but target wasn't a planet.", self.id)
		return
	}

	planet := game.GetPlanet(self.target_id)

	// Is it full and friendly? (This shouldn't be so.)

	if planet.Owned && planet.Owner == game.Pid() && planet.IsFull() {
		game.Log("Pilot %d: EngagePlanet() called but my planet was full.", self.id)
		return
	}

	// Is it available for us to dock?

	if planet.Owned == false || (planet.Owner == game.Pid() && planet.IsFull() == false) {
		self.FinalPlanetApproachForDock()
		return
	}

	// So it's hostile...

	speed, degrees, err := game.Navigate(self.Ship().X, self.Ship().Y, planet.X, planet.Y)

	if err != nil {
		return
	}

	self.Thrust(speed, degrees)
}

func (self *Pilot) FinalPlanetApproachForDock() {
	game := self.game

	if self.target_type != hal.PLANET {
		game.Log("Pilot %d: FinalPlanetApproachForDock() called but target wasn't a planet.", self.id)
		return
	}

	planet := game.GetPlanet(self.target_id)

	if self.CanDock(planet) {
		self.Dock(planet)
		return
	}

	speed, degrees, err := game.Approach(self.Ship(), planet, 3.5)

	if err != nil {
		game.Log("Pilot %d: FinalPlanetApproachForDock(): %v", self.id, err)
	}

	self.Thrust(speed, degrees)
}
