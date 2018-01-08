package pilot

import (
	"fmt"

	hal "../core"
	nav "../navigation"
)

const (
	DEFAULT_ENEMY_SHIP_APPROACH_DIST = 5.45			// GetApproach uses centre-to-edge distances, so 5.5ish.
)

type RallyPoints struct {
	Points				[]*hal.Point
}

func (self *RallyPoints) Clear() {
	if self == nil { return }
	self.Points = nil
}

func (self *RallyPoints) Add(x, y float64) {
	if self == nil { return }
	self.Points = append(self.Points, &hal.Point{x, y})
}

func (self *RallyPoints) ClosestTo(e hal.Entity) (*hal.Point, float64, bool) {

	if self == nil || len(self.Points) == 0 {
		return nil, 999, false
	}

	var ret *hal.Point
	var best_dist float64

	for _, point := range self.Points {
		d := point.Dist(e)
		if ret == nil || d < best_dist {
			ret = point
			best_dist = d
		}
	}

	return ret, best_dist, true
}

type Pilot struct {
	*hal.Ship
	Plan				string						// Our planned order, valid for 1 turn only.
	Message				int							// Message for this turn. -1 for no message.
	HasExecuted			bool						// Have we actually "sent" the order? (Placed it in the game.orders map.)
	Game				*hal.Game
	Target				hal.Entity					// Use the hal.Nothing struct for no target.
	EnemyApproachDist	float64
	NavStack			[]string
	Inhibition			float64
	Locked				bool						// Whether Target can change. Use super-sparingly.
	DangerShips			int							// Enemy ships that could potentially hit us this turn.
}

func NewPilot(sid int, game *hal.Game) *Pilot {
	ret := new(Pilot)
	ret.Game = game
	ship, ok := game.GetShip(sid)
	if ok == false {
		panic("NewPilot called with invalid sid")
	}
	ret.Ship = ship
	ret.Target = hal.Nothing
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
	self.HasExecuted = false
	self.Game.RawOrder(self.Id, "")
}

func (self *Pilot) ResetAndUpdate() bool {				// Return true if we still exist.

	_, ok := self.Game.GetShip(self.Id)

	if ok == false {
		return false
	}

	self.ResetPlan()

	self.NavStack = nil
	self.Message = -1
	self.EnemyApproachDist = DEFAULT_ENEMY_SHIP_APPROACH_DIST
	self.Inhibition = 0
	self.DangerShips = 0

	// Delete our target if appropriate...

	if self.Locked == false {

		self.Target = hal.Nothing

	} else {

		if self.DockedStatus != hal.UNDOCKED {
			self.Target = hal.Nothing
		}

		if self.Target.Type() == hal.PLANET || self.Target.Type() == hal.SHIP {
			if self.Target.Alive() == false {
				self.Target = hal.Nothing
			}
		}

		if self.Target.Type() == hal.PORT {
			_, ok := self.Game.GetPlanet(self.Target.GetId())
			if ok == false {
				self.Target = hal.Nothing
			}
		}
	}

	if self.Target == hal.Nothing {
		self.Locked = false
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

func (self *Pilot) ClosestPlanet() *hal.Planet {
	return self.Game.ClosestPlanet(self)
}

// -------------------------------------------------------------------

func (self *Pilot) PlanThrust(speed, degrees int) {
	for degrees < 0 { degrees += 360 }
	degrees %= 360
	self.Plan = fmt.Sprintf("t %d %d %d", self.Id, speed, degrees)
}

func (self *Pilot) PlanDock(planet *hal.Planet) {
	self.Plan = fmt.Sprintf("d %d %d", self.Id, planet.Id)
}

func (self *Pilot) PlanUndock() {
	self.Plan = fmt.Sprintf("u %d", self.Id)
}

// -------------------------------------------------------------------

func (self *Pilot) ExecutePlan() {
	if self.Plan == "" {
		self.PlanThrust(0, 0)
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

func (self *Pilot) DecideSideFor(target hal.Entity) nav.Side {
	return nav.DecideSideFromTarget(self.Ship, target, self.Game, self)
}
