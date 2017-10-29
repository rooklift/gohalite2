package ai

import (
	"fmt"
	"math/rand"

	hal "../gohalite2"
)

type Pilot struct {
	hal.Ship
	Plan			string				// Our planned order, valid for 1 turn only
	HasOrdered		bool				// Have we actually sent the order?
	Overmind		*Overmind
	Game			*hal.Game
	TargetType		hal.EntityType		// Long term target info
	TargetId		int					// Long term target info
}

func (self *Pilot) Log(format_string string, args ...interface{}) {
	format_string = fmt.Sprintf("%v: ", self) + format_string
	self.Game.Log(format_string, args...)
}

func (self *Pilot) Update() {
	self.Ship = self.Game.GetShip(self.Id)
	self.ClearPlan()					// Also clears the HasOrdered bool
}

func (self *Pilot) ClosestPlanet() hal.Planet {
	return self.Game.ClosestPlanet(self)
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

func (self *Pilot) MakePlan() {

	// Clear dead / totally conquered targets...

	self.ValidateTarget()

	// Helpers can lock in an order by actually setting it.

	if self.Plan == "" {
		self.DockIfPossible()
	}

	if self.Plan == "" {
		self.ChooseTarget()
	}

	if self.Plan == "" {
		self.ChaseTarget()
	}
}

func (self *Pilot) DockIfPossible() {
	if self.DockedStatus == hal.UNDOCKED {
		closest_planet := self.ClosestPlanet()
		if self.CanDock(closest_planet) {
			self.PlanDock(closest_planet)
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

		if self.ApproachDist(planet) < 4 {
			self.EngagePlanet()
			return
		}

		speed, degrees, err := game.GetApproach(self.Ship, planet, 4, game.AllImmobile())

		if err != nil {
			self.Log("ChaseTarget(): %v", err)
			self.TargetType = hal.NONE
		} else {
			self.PlanThrust(speed, degrees)
		}

	} else if self.TargetType == hal.SHIP {

		other_ship := game.GetShip(self.TargetId)

		speed, degrees, err := game.GetApproach(self.Ship, other_ship, 4.5, game.AllImmobile())		// GetApproach uses centre-to-edge distances, so 4.5

		if err != nil {
			self.Log("ChaseTarget(): %v", err)
			self.TargetType = hal.NONE
		} else {
			self.PlanThrust(speed, degrees)
			if speed == 0 && self.Dist(other_ship) >= hal.WEAPON_RANGE {
				self.Log("ChaseTarget(): not moving but not in range!")
			}
		}
	}
}

func (self *Pilot) EngagePlanet() {
	game := self.Game

	// We are very close to our target planet. Do something about this.

	if self.TargetType != hal.PLANET {
		self.Log("EngagePlanet() called but target wasn't a planet.", self.Id)
		return
	}

	planet := game.GetPlanet(self.TargetId)

	// Is it full and friendly? (This shouldn't be so.)

	if planet.Owned && planet.Owner == game.Pid() && planet.IsFull() {
		self.Log("EngagePlanet() called but my planet was full.", self.Id)
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

	speed, degrees, err := game.GetApproach(self.Ship, enemy_ship, 4.5, game.AllImmobile())			// GetApproach uses centre-to-edge distances, so 4.5

	if err != nil {
		self.Log("EngagePlanet(): %v", err)
		return
	}

	self.PlanThrust(speed, degrees)
}

func (self *Pilot) FinalPlanetApproachForDock() {
	game := self.Game

	if self.TargetType != hal.PLANET {
		self.Log("FinalPlanetApproachForDock() called but target wasn't a planet.", self.Id)
		return
	}

	planet := game.GetPlanet(self.TargetId)

	if self.CanDock(planet) {
		self.PlanDock(planet)
		return
	}

	speed, degrees, err := game.GetApproach(self.Ship, planet, 4, game.AllImmobile())

	if err != nil {
		self.Log("FinalPlanetApproachForDock(): %v", self.Id, err)
	}

	self.PlanThrust(speed, degrees)
}

// -------------------------------------------------------------------

func (self *Pilot) PlanThrust(speed, angle int) {
	self.Plan = fmt.Sprintf("t %d %d %d", self.Id, speed, angle)
}

func (self *Pilot) PlanDock(planet hal.Planet) {
	self.Plan = fmt.Sprintf("d %d %d", self.Id, planet.Id)
}

func (self *Pilot) PlanUndock() {
	self.Plan = fmt.Sprintf("u %d", self.Id)
}

func (self *Pilot) ClearPlan() {
	self.Plan = ""
	self.Game.RawOrder(self.Id, "")		// Also clear the actual order if needed
	self.HasOrdered = false
}

func (self *Pilot) ExecutePlan() {
	self.Game.RawOrder(self.Id, self.Plan)
	self.HasOrdered = true
	if hal.GetOrderType(self.Plan) == hal.THRUST {
		if self.TargetType == hal.PLANET {
			self.Game.EncodeSecretInfo(self.Ship, self.TargetId)
		} else if self.TargetType == hal.SHIP {
			self.Game.EncodeSecretInfo(self.Ship, 121)
		}
	}
}

func (self *Pilot) PreliminaryRestrict(atc *AirTrafficControl) {			// In case we end up not moving, restrict airspace.
	if self.DockedStatus == hal.UNDOCKED {			// Docked ships don't restrict airspace. We navigate around them using other means.
		atc.Restrict(self.Ship, 0, 0)
	}
}

func (self *Pilot) ExecutePlanIfStationary(atc *AirTrafficControl) {

	if self.HasOrdered {
		self.Log("ExecutePlanIfStationary(): already ordered!")
		return
	}

	speed, _ := hal.CourseFromString(self.Plan)
	if speed == 0 {
		self.ExecutePlan()
	}
}

func (self *Pilot) ExecutePlanIfSafe(atc *AirTrafficControl) {

	if self.HasOrdered {
		self.Log("ExecutePlanIfSafe(): already ordered!")
		return
	}

	speed, degrees := hal.CourseFromString(self.Plan)
	atc.Unrestrict(self.Ship, 0, 0)							// Unrestruct our preliminary null course so it doesn't block us.
	if atc.PathIsFree(self.Ship, speed, degrees) {
		self.ExecutePlan()
		atc.Restrict(self.Ship, speed, degrees)
	} else {
		atc.Restrict(self.Ship, 0, 0)						// Restrict our null course again.
		// Make the format string contain the turn and ship number so this message only gets logged once, but others can be.
		self.Game.LogOnce(fmt.Sprintf("t %d: %v: Refusing unsafe thrust %%d / %%d", self.Game.Turn(), self.Ship), speed, degrees)
	}
}

// ----------------------------------------------

type PilotsByY []*Pilot

func (slice PilotsByY) Len() int {
	return len(slice)
}
func (slice PilotsByY) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
func (slice PilotsByY) Less(i, j int) bool {
	return slice[i].Y < slice[j].Y
}
