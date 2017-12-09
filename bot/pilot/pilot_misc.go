package pilot

import (
	"fmt"

	atc "../atc"
	hal "../core"
	nav "../navigation"
)

// The point of making Pilot its own module is that the logic of dealing with targets is
// mostly independent of grand strategy. Still, there are a few things the Overmind may
// need back from us, hence the Overmind interface below which allows us to update it.

type Overmind interface {
	NotifyTargetChange(pilot *Pilot, old_target, new_target hal.Entity)
	NotifyDock(planet hal.Planet)
}

const DEFAULT_ENEMY_SHIP_APPROACH_DIST = 5.45		// GetApproach uses centre-to-edge distances, so 5.5ish.

type Pilot struct {
	hal.Ship
	Plan				string						// Our planned order, valid for 1 turn only.
	Message				int							// Message for this turn. -1 for no message.
	HasExecuted			bool						// Have we actually "sent" the order? (Placed it in the game.orders map.)
	Overmind			Overmind
	Game				*hal.Game
	Target				hal.Entity					// Use a hal.Nothing{} struct for no target.
	TurnTarget			hal.Entity
	EnemyApproachDist	float64
	NavStack			[]string
}

func NewPilot(sid int, game *hal.Game, overmind Overmind) *Pilot {
	ret := new(Pilot)
	ret.Overmind = overmind
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

func (self *Pilot) ResetAndUpdate(clear_stack, reset_ead, reset_tt bool) bool {		// Doesn't clear Target. Return true if we still exist.

	if clear_stack {
		self.NavStack = nil
	}

	if reset_ead {
		self.EnemyApproachDist = DEFAULT_ENEMY_SHIP_APPROACH_DIST
	}

	if reset_tt {
		self.TurnTarget = self.Target
	}

	current_ship, alive := self.Game.GetShip(self.Id)

	if alive == false {
		self.SetTarget(hal.Nothing{})							// Means the overmind will be notified about our lack of target.
		return false
	}

	self.Ship = current_ship									// Don't do this until after the (possible) self.SetTarget() above.
	self.Plan = ""
	self.Message = -1
	self.HasExecuted = false
	self.Game.RawOrder(self.Id, "")

	// Update the info about our target.

	if self.DockedStatus != hal.UNDOCKED {
		self.SetTarget(hal.Nothing{})
	}

	switch self.Target.Type() {

	case hal.SHIP:

		if self.Target.Alive() == false {
			self.SetTarget(hal.Nothing{})
		} else {
			var ok bool
			self.Target, ok = self.Game.GetShip(self.Target.GetId())
			if ok == false {
				self.SetTarget(hal.Nothing{})
			}
		}

	case hal.PLANET:

		if self.Target.Alive() == false {
			self.SetTarget(hal.Nothing{})
		} else {
			var ok bool
			self.Target, ok = self.Game.GetPlanet(self.Target.GetId())
			if ok == false {
				self.SetTarget(hal.Nothing{})
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

func (self *Pilot) SetTarget(e hal.Entity) {		// So we can update Overmind's info.
	old_target := self.Target
	self.Target = e
	self.Overmind.NotifyTargetChange(self, old_target, e)
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
	self.Overmind.NotifyDock(planet)
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

func (self *Pilot) ExecutePlanWithATC(atc *atc.AirTrafficControl) bool {

	speed, degrees := hal.CourseFromString(self.Plan)
	atc.Unrestrict(self.Ship, 0, 0)							// Unrestruct our preliminary null course so it doesn't block us.

	if atc.PathIsFree(self.Ship, speed, degrees) && self.Game.CourseStaysInBounds(self.Ship, speed, degrees) {

		self.ExecutePlan()
		atc.Restrict(self.Ship, speed, degrees)
		return true

	} else {

		atc.Restrict(self.Ship, 0, 0)						// Restrict our null course again.
		return false

	}
}

func (self *Pilot) SlowPlanDown() {

	speed, degrees := hal.CourseFromString(self.Plan)

	if speed < 1 {
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

func (self *Pilot) DecideSideFromTarget() nav.Side {
	return nav.DecideSideFromTarget(self.Ship, self.Target, self.Game, self)
}
