package pilot

import (
	"fmt"
	"sort"

	atc "../atc"
	hal "../gohalite2"
)

type MessageInt int

const (
	MSG_ATTACK_DOCKED = 121
	MSG_FIGHT_IN_ORBIT = 122
	MSG_ASSASSINATE = 123
	MSG_ATC_DEACTIVATED = 150
	MSG_ATC_RESTRICT = 151
	MSG_ATC_SLOWED = 152
	MSG_RECURSION = 175
	MSG_EXECUTED_NO_PLAN = 176
	MSG_SECRET_SAUCE = 177
	MSG_POINT_TARGET = 178
	MSG_DOCK_APPROACH = 179
	MSG_NO_TARGET = 180
)

type Pilot struct {
	hal.Ship
	Plan			string				// Our planned order, valid for 1 turn only
	HasExecuted		bool				// Have we actually "sent" the order? (Placed it in the game.orders map.)
	Overmind		Overmind
	Game			*hal.Game
	Target			hal.Entity			// Use a hal.Nothing{} struct for no target.
}

type Overmind interface {
	NotifyTargetChange(pilot *Pilot, old_target, new_target hal.Entity)
	NotifyDock(planet hal.Planet)
	ShipsDockingAt(planet hal.Planet) 												int
	ShipsChasing(ship hal.Ship) 													int
	EnemiesNearPlanet(planet hal.Planet) 											[]hal.Ship
}

func NewPilot(sid int, game *hal.Game, overmind Overmind) *Pilot {
	ret := new(Pilot)
	ret.Overmind = overmind
	ret.Game = game
	ret.Id = sid								// This has to be set so pilot.Reset() can work.
	ret.Target = hal.Nothing{}
	return ret
}

func (self *Pilot) Log(format_string string, args ...interface{}) {
	format_string = fmt.Sprintf("%v: ", self) + format_string
	self.Game.Log(format_string, args...)
}

func (self *Pilot) ResetAndUpdate() bool {			// Doesn't clear Target. Return true if we still exist.

	var ok bool
	self.Ship, ok = self.Game.GetShip(self.Id)

	if ok == false {
		return false
	}

	self.Plan = ""
	self.HasExecuted = false
	self.Game.RawOrder(self.Id, "")

	// Update the info about our target.

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

func (self *Pilot) SetTarget(e hal.Entity) {		// So we can update Overmind's info.
	old_target := self.Target
	self.Target = e
	self.Overmind.NotifyTargetChange(self, old_target, e)
}

func (self *Pilot) ClosestPlanet() hal.Planet {
	return self.Game.ClosestPlanet(self)
}

func (self *Pilot) ValidateTarget() bool {

	game := self.Game

	switch self.Target.Type() {

	case hal.SHIP:

		if self.Target.Alive() == false {
			self.SetTarget(hal.Nothing{})
		}

	case hal.PLANET:

		target := self.Target.(hal.Planet)

		if target.Alive() == false {
			self.SetTarget(hal.Nothing{})
		} else if self.Overmind.ShipsDockingAt(target) >= game.DesiredSpots(target) {			// We've enough guys (maybe 0) trying to dock...
			if len(self.Overmind.EnemiesNearPlanet(target)) == 0 {								// ...and the planet is safe
				self.SetTarget(hal.Nothing{})
			}
		}
	}

	if self.Target == (hal.Nothing{}) {
		return false
	}

	return true
}

func (self *Pilot) PlanDockIfWise() (hal.Planet, bool) {

	closest_planet := self.ClosestPlanet()

	if self.DockedStatus != hal.UNDOCKED {
		return hal.Planet{}, false
	}

	if self.CanDock(closest_planet) == false {
		return hal.Planet{}, false
	}

	if len(self.Overmind.EnemiesNearPlanet(closest_planet)) > 0 {
		return hal.Planet{}, false
	}

	if self.Overmind.ShipsDockingAt(closest_planet) >= self.Game.DesiredSpots(closest_planet) {
		return hal.Planet{}, false
	}

	self.PlanDock(closest_planet)
	return closest_planet, true
}

func (self *Pilot) ChooseTarget(all_planets []hal.Planet, all_enemy_ships []hal.Ship) {

	// We pass all_planets and all_enemy_ships for speed. They do get sorted in place, caller beware.

	game := self.Game

	var target_planets []hal.Planet

	for _, planet := range all_planets {

		ok := false

		if game.DesiredSpots(planet) > 0 && self.Overmind.ShipsDockingAt(planet) < game.DesiredSpots(planet) {
			ok = true
		} else if len(self.Overmind.EnemiesNearPlanet(planet)) > 0 {
			ok = true
		}

		if ok {
			target_planets = append(target_planets, planet)
		}
	}

	sort.Slice(target_planets, func(a, b int) bool {
		return self.ApproachDist(target_planets[a]) < self.ApproachDist(target_planets[b])
	})

	sort.Slice(all_enemy_ships, func(a, b int) bool {
		return self.Dist(all_enemy_ships[a]) < self.Dist(all_enemy_ships[b])
	})

	if len(all_enemy_ships) > 0 && len(target_planets) > 0 {
		if self.Dist(all_enemy_ships[0]) < self.Dist(target_planets[0]) {
			if self.Overmind.ShipsChasing(all_enemy_ships[0]) == 0 {
				self.SetTarget(all_enemy_ships[0])
			} else {
				self.SetTarget(target_planets[0])
			}
		} else {
			self.SetTarget(target_planets[0])
		}
	} else if len(target_planets) > 0 {
		self.SetTarget(target_planets[0])
	} else if len(all_enemy_ships) > 0 {
		self.SetTarget(all_enemy_ships[0])
	}
}

func (self *Pilot) PlanChase(avoid_list []hal.Entity) {

	game := self.Game

	if self.DockedStatus != hal.UNDOCKED {
		return
	}

	if self.Target.Type() == hal.NOTHING {
		self.PlanThrust(0, 0, MSG_NO_TARGET)
		return
	}

	// Which side of objects to navigate around? As a default, use this arbitrary choice...

	var side hal.Side; if self.Id % 2 == 0 { side = hal.RIGHT } else { side = hal.LEFT }

	// If the first planet in our path isn't our target planet, we choose a side to navigate around.
	// By using AllImmobile() as the avoid_list, any collision will be with a planet or docked ship.

	collision_entity, ok := game.FirstCollision(self.Ship, 1000, self.Angle(self.Target), game.AllImmobile())

	if ok {

		var blocking_planet hal.Planet

		// We also consider docked ships to be "part of the planet" for these purposes -+- we must use game.AllImmobile() above

		if collision_entity.Type() == hal.PLANET {
			blocking_planet = collision_entity.(hal.Planet)
		} else {
			s := collision_entity.(hal.Ship)
			blocking_planet, _ = game.GetPlanet(s.DockedPlanet)
		}

		if self.Target.Type() != hal.PLANET || blocking_planet.Id != self.Target.GetId() {
			side = hal.DecideSide(self.Ship, self.Target, blocking_planet)
		}
	}

	switch self.Target.Type() {

	case hal.PLANET:

		planet := self.Target.(hal.Planet)

		if self.ApproachDist(planet) <= 10 {		// If this is too low, we may get outside the action zone when navigating round allies.
			self.EngagePlanet(avoid_list)
			return
		}

		speed, degrees, err := game.GetApproach(self.Ship, planet, 4.45, avoid_list, side)

		if err != nil {
			self.Log("PlanChase(): %v", err)
			self.SetTarget(hal.Nothing{})
		} else {
			self.PlanThrust(speed, degrees, MessageInt(planet.Id))
		}

	case hal.SHIP:

		other_ship := self.Target.(hal.Ship)

		speed, degrees, err := game.GetApproach(self.Ship, other_ship, 5.45, avoid_list, side)		// GetApproach uses centre-to-edge distances, so 5.5ish

		if err != nil {
			self.Log("PlanChase(): %v", err)
			self.SetTarget(hal.Nothing{})
		} else {
			self.PlanThrust(speed, degrees, MSG_ASSASSINATE)
			if speed == 0 && self.Dist(other_ship) >= hal.WEAPON_RANGE + hal.SHIP_RADIUS * 2 {
				self.Log("PlanChase(): not moving but not in range!")
			}
		}

	case hal.POINT:

		point := self.Target.(hal.Point)

		speed, degrees, err := game.GetCourse(self.Ship, point, avoid_list, side)
		if err != nil {
			self.Log("PlanChase(): %v", err)
			self.SetTarget(hal.Nothing{})
		} else {
			self.PlanThrust(speed, degrees, MSG_POINT_TARGET)
		}
	}
}

func (self *Pilot) EngagePlanet(avoid_list []hal.Entity) {
	game := self.Game
	overmind := self.Overmind

	// We are very close to our target planet. Do something about this.

	if self.Target.Type() != hal.PLANET {
		self.Log("EngagePlanet() called but target wasn't a planet.")
		return
	}

	planet := self.Target.(hal.Planet)

	// Are there enemy ships near the planet?

	if len(overmind.EnemiesNearPlanet(planet)) > 0 || (planet.Owner != game.Pid() && planet.DockedShips > 0) {

		// We directly plan our move without changing our stored (planet) target.

		var enemies []hal.Ship
		enemies = append(enemies, overmind.EnemiesNearPlanet(planet)...)

		// Our target can also be one of the docked ships...

		if planet.Owner != game.Pid() {
			enemies = append(enemies, game.ShipsDockedAt(planet)...)
		}

		// Find closest...

		sort.Slice(enemies, func(a, b int) bool {
			return enemies[a].Dist(self.Ship) < enemies[b].Dist(self.Ship)
		})

		enemy_ship := enemies[0]
		side := hal.DecideSide(self.Ship, enemy_ship, planet)

		speed, degrees, err := game.GetApproach(self.Ship, enemy_ship, 5.45, avoid_list, side)		// GetApproach uses centre-to-edge distances, so 5.5ish
		if err != nil {
			self.PlanThrust(speed, degrees, MSG_RECURSION)
			self.Log("EngagePlanet(), while trying to engage ship: %v", err)
		} else {
			self.PlanThrust(speed, degrees, MSG_FIGHT_IN_ORBIT)
			if speed == 0 && self.Ship.Dist(enemy_ship) >= hal.WEAPON_RANGE + hal.SHIP_RADIUS * 2 {
				self.Log("EngagePlanet(), while approaching ship: stopped short of target.")
			}
		}
		return
	}

	// Is it available for us to dock?

	if planet.Owned == false || (planet.Owner == game.Pid() && planet.IsFull() == false) {
		self.FinalPlanetApproachForDock(avoid_list)
		return
	}

	// This function shouldn't have been called at all.

	self.Log("EngagePlanet() called but there's nothing to do here.")
	return
}

func (self *Pilot) FinalPlanetApproachForDock(avoid_list []hal.Entity) {
	game := self.Game

	if self.Target.Type() != hal.PLANET {
		self.Log("FinalPlanetApproachForDock() called but target wasn't a planet.", self.Id)
		return
	}

	planet := self.Target.(hal.Planet)

	if self.CanDock(planet) {
		self.PlanDock(planet)
		return
	}

	// Which side of objects to navigate around. At long range, use this arbitary choice...
	var side hal.Side; if self.Id % 2 == 0 { side = hal.RIGHT } else { side = hal.LEFT }

	speed, degrees, err := game.GetApproach(self.Ship, planet, hal.DOCKING_RADIUS + hal.SHIP_RADIUS - 0.001, avoid_list, side)

	if err != nil {
		self.Log("FinalPlanetApproachForDock(): %v", self.Id, err)
	}

	self.PlanThrust(speed, degrees, MSG_DOCK_APPROACH)
}

// -------------------------------------------------------------------

func (self *Pilot) PlanThrust(speed, degrees int, message MessageInt) {		// Send -1 as message for no message.

	for degrees < 0 {
		degrees += 360
	}

	degrees %= 360

	// We put some extra info into the angle, which we can see in the Chlorine replayer...

	if message >= 0 && message <= 180 {
		degrees += (int(message) + 1) * 360
	}

	self.Plan = fmt.Sprintf("t %d %d %d", self.Id, speed, degrees)
}

func (self *Pilot) PlanDock(planet hal.Planet) {
	self.Plan = fmt.Sprintf("d %d %d", self.Id, planet.Id)
	self.Overmind.NotifyDock(planet)
}

func (self *Pilot) PlanUndock() {
	self.Plan = fmt.Sprintf("u %d", self.Id)
}

// ----------------------------------------------

func (self *Pilot) ExecutePlan() {
	if self.Plan == "" {
		self.PlanThrust(0, 0, MSG_EXECUTED_NO_PLAN)
	}
	self.Game.RawOrder(self.Id, self.Plan)
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

	if atc.PathIsFree(self.Ship, speed, degrees) {

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

	self.PlanThrust(speed, degrees, MSG_ATC_SLOWED)
}
