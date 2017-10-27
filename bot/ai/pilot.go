package ai

import (
	"math/rand"
	hal "../gohalite2"
)

type Pilot struct {
	hal.Ship
	Overmind		*Overmind
	Game			*hal.Game
	TargetType		hal.EntityType
	TargetId		int
}

func (self *Pilot) Log(format_string string, args ...interface{}) {
	self.Game.Log("Ship %d: " + format_string, args...)
}

func (self *Pilot) Update() {
	self.Ship = self.Game.GetShip(self.Id)
}

func (self *Pilot) ClosestPlanet() hal.Planet {
	return self.Game.ClosestPlanet(self.X, self.Y)
}

func (self *Pilot) CurrentOrder() string {
	return self.Game.CurrentOrder(self.Id)
}

func (self *Pilot) Thrust(speed, angle int) {
	self.Game.Thrust(self.Id, speed, angle)
}

func (self *Pilot) Dock(planet hal.Planet) {
	self.Game.Dock(self.Id, planet.Id)
}

func (self *Pilot) Undock() {
	self.Game.Undock(self.Id)
}

func (self *Pilot) ClearOrder() {
	self.Game.ClearOrder(self.Id)
}

func (self *Pilot) ValidateTarget() {

	game := self.Game

	if self.TargetType == hal.SHIP {
		target := game.GetShip(self.TargetId)
		if target.Alive() == false {
			self.TargetType = hal.NONE
			closest_planet := self.ClosestPlanet()
			if self.Dist(closest_planet) < 50 {
				if closest_planet.IsFull() == false || closest_planet.Owner != game.Pid() {
					self.TargetType = hal.PLANET
					self.TargetId = closest_planet.Id
				}
			}
		}
	} else if self.TargetType == hal.PLANET {
		target := game.GetPlanet(self.TargetId)
		if target.Alive() == false {
			self.TargetType = hal.NONE
		} else if target.Owner == game.Pid() && target.IsFull() {
			self.TargetType = hal.NONE
		}
	}
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
	if self.DockedStatus == hal.UNDOCKED {
		closest_planet := self.ClosestPlanet()
		if self.CanDock(closest_planet) {
			self.Dock(closest_planet)
		}
	}
}

func (self *Pilot) ChooseTarget() {
	game := self.Game

	if self.TargetType != hal.NONE {		// We already have a target.
		return
	}

	all_planets := game.AllPlanets()

	for n := 0; n < 5; n++ {

		i := rand.Intn(len(all_planets))
		planet := all_planets[i]

		if planet.Owner != game.Pid() || planet.IsFull() == false {
			self.TargetId = planet.Id
			self.TargetType = hal.PLANET
			break
		}
	}
}

func (self *Pilot) ChaseTarget() {
	game := self.Game

	if self.TargetType == hal.NONE || self.DockedStatus != hal.UNDOCKED {
		return
	}

	if self.TargetType == hal.PLANET {

		planet := game.GetPlanet(self.TargetId)

		speed, degrees, err := game.GetApproach(self.Ship, planet, 6.0)

		if err != nil {
			self.TargetType = hal.NONE
		} else {
			if speed < 4 {
				self.EngagePlanet()
			} else {
				self.Thrust(speed, degrees)
			}
		}

	} else if self.TargetType == hal.SHIP {

		other_ship := game.GetShip(self.TargetId)

		speed, degrees, err := game.GetApproach(self.Ship, other_ship, 3.0)

		if err != nil {
			self.TargetType = hal.NONE
		} else {
			self.Thrust(speed, degrees)
		}
	}
}

func (self *Pilot) EngagePlanet() {
	game := self.Game

	// We are very close to our target planet. Do something about this.

	if self.TargetType != hal.PLANET {
		game.Log("Pilot %d: EngagePlanet() called but target wasn't a planet.", self.Id)
		return
	}

	planet := game.GetPlanet(self.TargetId)

	// Is it full and friendly? (This shouldn't be so.)

	if planet.Owned && planet.Owner == game.Pid() && planet.IsFull() {
		game.Log("Pilot %d: EngagePlanet() called but my planet was full.", self.Id)
		return
	}

	// Is it available for us to dock?

	if planet.Owned == false || (planet.Owner == game.Pid() && planet.IsFull() == false) {
		self.FinalPlanetApproachForDock()
		return
	}

	// So it's hostile...

	docked_targets := game.ShipsDockedAt(planet)

	enemy_ship := docked_targets[rand.Intn(len(docked_targets))]
	self.TargetType = hal.SHIP
	self.TargetId = enemy_ship.Id

	speed, degrees, err := game.GetApproach(self.Ship, enemy_ship, 3.0)

	if err != nil {
		return
	}

	self.Thrust(speed, degrees)
}

func (self *Pilot) FinalPlanetApproachForDock() {
	game := self.Game

	if self.TargetType != hal.PLANET {
		game.Log("Pilot %d: FinalPlanetApproachForDock() called but target wasn't a planet.", self.Id)
		return
	}

	planet := game.GetPlanet(self.TargetId)

	if self.CanDock(planet) {
		self.Dock(planet)
		return
	}

	speed, degrees, err := game.GetApproach(self.Ship, planet, 3.5)

	if err != nil {
		game.Log("Pilot %d: FinalPlanetApproachForDock(): %v", self.Id, err)
	}

	self.Thrust(speed, degrees)
}
