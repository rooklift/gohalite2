package pilot

import (
	"fmt"

	hal "../core"
	nav "../navigation"
)

const (
	DEFAULT_ENEMY_SHIP_APPROACH_DIST = 5.45			// GetApproach uses centre-to-edge distances, so 5.5ish.
)

type Pilot struct {
	hal.Ship
	Plan				string						// Our planned order, valid for 1 turn only.
	Message				int							// Message for this turn. -1 for no message.
	HasExecuted			bool						// Have we actually "sent" the order? (Placed it in the game.orders map.)
	Game				*hal.Game
	Target				hal.Entity					// Use a hal.Nothing{} struct for no target.
	TurnTarget			hal.Entity
	EnemyApproachDist	float64
	NavStack			[]string
}

func NewPilot(sid int, game *hal.Game) *Pilot {
	ret := new(Pilot)
	ret.Game = game
	ret.Id = sid
	ret.Target = hal.Nothing{}
	return ret
}

func (self *Pilot) AddToNavStack(format_string string, args ...interface{}) {
	s := fmt.Sprintf(format_string, args...)
	self.NavStack = append(self.NavStack, s)
}

func (self *Pilot) LogNavStack() {
	self.Game.Log("%v Nav Stack:", self)
	for _, s := range self.NavStack {
		self.Game.Log("        %v", s)
	}
}

func (self *Pilot) Log(format_string string, args ...interface{}) {
	format_string = fmt.Sprintf("%v: ", self) + format_string
	self.Game.Log(format_string, args...)
}

func (self *Pilot) ResetPlan() {
	self.Plan = ""
	self.Message = -1
	self.HasExecuted = false
	self.Game.RawOrder(self.Id, "")
}

func (self *Pilot) ResetAndUpdate() bool {						// Doesn't clear Target. Return true if we still exist.

	self.NavStack = nil
	self.EnemyApproachDist = DEFAULT_ENEMY_SHIP_APPROACH_DIST
	self.TurnTarget = self.Target

	current_ship, alive := self.Game.GetShip(self.Id)

	if alive == false {
		return false
	}

	self.Ship = current_ship

	self.ResetPlan()

	// Update the info about our target.

	if self.DockedStatus != hal.UNDOCKED {
		self.Target = hal.Nothing{}
	}

	switch self.Target.Type() {

	case hal.SHIP:

		if self.Target.Alive() == false {
			self.Target = hal.Nothing{}
		} else {
			var ok bool
			self.Target, ok = self.Game.GetShip(self.Target.GetId())
			if ok == false {
				self.Target = hal.Nothing{}
			}
		}

	case hal.PLANET:

		if self.Target.Alive() == false {
			self.Target = hal.Nothing{}
		} else {
			var ok bool
			self.Target, ok = self.Game.GetPlanet(self.Target.GetId())
			if ok == false {
				self.Target = hal.Nothing{}
			}
		}
	}

	return true
}

func (self *Pilot) HasStationaryPlan() bool {		// true iff we DO have a plan, which doesn't move us.
	if self.Plan == "" {
		return false
	}
	speed, _ := self.CourseFromPlan()
	return speed == 0
}

func (self *Pilot) CourseFromPlan() (int, int) {
	return hal.CourseFromString(self.Plan)
}

func (self *Pilot) HasTarget() bool {				// We don't use nil ever, so we can e.g. call hal.Type()
	return self.Target.Type() != hal.NOTHING
}

func (self *Pilot) ClosestPlanet() hal.Planet {
	return self.Game.ClosestPlanet(self)
}

// -------------------------------------------------------------------

func (self *Pilot) PlanThrust(speed, degrees int) {
	for degrees < 0 { degrees += 360 }
	degrees %= 360
	self.Plan = fmt.Sprintf("t %d %d %d", self.Id, speed, degrees)
}

func (self *Pilot) PlanDock(planet hal.Planet) {
	self.Plan = fmt.Sprintf("d %d %d", self.Id, planet.Id)
}

func (self *Pilot) PlanUndock() {
	self.Plan = fmt.Sprintf("u %d", self.Id)
}

// -------------------------------------------------------------------

func (self *Pilot) ExecutePlan() {
	if self.Plan == "" {
		self.PlanThrust(0, 0)
		self.Message = MSG_EXECUTED_NO_PLAN
	}
	self.Game.RawOrder(self.Id, self.Plan)
	self.Game.SetMessage(self.Ship, self.Message)			// Fails silently if message < 0 or > 180
	self.HasExecuted = true
}

func (self *Pilot) ExecutePlanIfStationary() {
	speed, _ := hal.CourseFromString(self.Plan)
	if speed == 0 {
		self.ExecutePlan()
	}
}

func (self *Pilot) SlowPlanDown() {

	speed, degrees := hal.CourseFromString(self.Plan)

	if speed <= 1 {											// Don't slow our plan to zero, which is like having no plan.
		return
	}

	speed--

	self.PlanThrust(speed, degrees)
	self.Message = MSG_ATC_SLOWED
}

// -------------------------------------------------------------------

func (self *Pilot) GetCourse(target hal.Entity, avoid_list []hal.Entity, side nav.Side) (int, int, error) {
	return nav.GetCourse(self.Ship, target, avoid_list, side, self)
}

func (self *Pilot) GetApproach(target hal.Entity, margin float64, avoid_list []hal.Entity, side nav.Side) (int, int, error) {
	return nav.GetApproach(self.Ship, target, margin, avoid_list, side, self)
}

func (self *Pilot) DecideSideFromTurnTarget() nav.Side {
	return nav.DecideSideFromTarget(self.Ship, self.TurnTarget, self.Game, self)
}
