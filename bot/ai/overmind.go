package ai

import (
	"math/rand"
	"sort"

	hal "../gohalite2"
)

// --------------------------------------------

type Overmind struct {
	Pilots			[]*Pilot
	Game			*hal.Game
	ATC				*AirTrafficControl
}

func NewOvermind(game *hal.Game) *Overmind {
	ret := new(Overmind)
	ret.Game = game
	ret.ATC = NewATC(game.Width(), game.Height())
	return ret
}

func (self *Overmind) Step() {

	self.UpdatePilots()
	self.ATC.Clear()

	if self.Game.Turn() == 0 {
		self.ChooseInitialTargets()
	}

	// Setup data structures...

	var mobile_pilots []*Pilot
	var frozen_pilots []*Pilot		// Note that this doesn't include docked / docking / undocking ships.

	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.UNDOCKED {
			mobile_pilots = append(mobile_pilots, pilot)
		}
	}

	// Plan a Dock if possible. If we do, remove this pilot from the mobile pilots list and make it frozen.

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		ok := pilot.PlanDockIfPossible()
		if ok {
			mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
			frozen_pilots = append(frozen_pilots, pilot)
			i--
		}
	}

	// Choose target if needed... (i.e. we don't have a valid target already).

	for _, pilot := range mobile_pilots {
		valid := pilot.ValidateTarget()
		if valid == false {
			pilot.ChooseTarget()
		}
	}

	// Perhaps this pilot doesn't need to move? If so, consider it frozen.

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		pilot.PlanChase(self.Game.AllImmobile())	// Planets plus already-docked ships.
		if pilot.HasStationaryPlan() {
			mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
			frozen_pilots = append(frozen_pilots, pilot)
			i--
		}
	}

	// Our PlanChase() above didn't avoid these frozen ships. Remake plans with the new info.
	// Possibly avoiding collisions we would have (since the above ships won't use ATC).

	avoid_list := self.Game.AllImmobile()
	for _, pilot := range frozen_pilots {
		avoid_list = append(avoid_list, pilot.Ship)
	}

	for _, pilot := range mobile_pilots {
		pilot.Reset()
		pilot.PlanChase(avoid_list)
	}

	// Now the only danger is 2 "mobile" ships colliding. We use the ATC for this possibility.
	// Note that it's possible that one of the colliding ships will not actually be moving.

	for _, pilot := range mobile_pilots {
		self.ATC.Restrict(pilot.Ship, 0, 0)			// All initially restrict a null move.
	}

	for n := 0; n < 5; n++ {						// Try a few times to allow chains of ships.
		for _, pilot := range mobile_pilots {
			if pilot.HasExecuted == false {
				pilot.ExecutePlanWithATC(self.ATC)
			}
		}
	}

	// Randomly give up for half the ships that still aren't moving, and
	// retry the pathfinding with the other half.

	// Ships moved into the frozen slice can have their ATC restriction
	// cleared since we will navigate around them precisely.

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		if pilot.HasExecuted == false && rand.Intn(2) == 0 {
			pilot.PlanThrust(0, 0)
			self.ATC.Unrestrict(pilot.Ship, 0, 0)
			mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
			frozen_pilots = append(frozen_pilots, pilot)
			i--
		}
	}

	avoid_list = self.Game.AllImmobile()
	for _, pilot := range frozen_pilots {
		avoid_list = append(avoid_list, pilot.Ship)
	}

	for _, pilot := range mobile_pilots {
		if pilot.HasExecuted == false {
			pilot.Reset()
			pilot.PlanChase(avoid_list)
		}
	}

	for _, pilot := range mobile_pilots {
		if pilot.HasExecuted == false {
			pilot.ExecutePlanWithATC(self.ATC)
		}
	}

	// Null thrust every "mobile" ship that didn't move. This causes target info to be put into
	// the replay via the Angle Message system.

	for _, pilot := range mobile_pilots {
		if pilot.HasExecuted == false {
			pilot.PlanThrust(0, 0)
			pilot.ExecutePlan()
		}
	}

	// Don't forget our frozen ships!

	for _, pilot := range frozen_pilots {
		pilot.ExecutePlan()
	}
}

func (self *Overmind) ChooseInitialTargets() {

	closest_three := self.Game.AllPlanetsByDistance(self.Pilots[0])[:3]

	sort.Sort(hal.PlanetsByY(closest_three))
	sort.Sort(PilotsByY(self.Pilots))

	for index, pilot := range self.Pilots {
		pilot.TargetType = hal.PLANET
		pilot.TargetId = closest_three[index].Id
	}
}

func (self *Overmind) UpdatePilots() {

	game := self.Game

	// Add new AIs for new ships...

	my_new_ships := game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		pilot := new(Pilot)
		pilot.Overmind = self
		pilot.Game = game
		pilot.Id = sid								// This has to be set so pilot.Reset() can work.
		self.Pilots = append(self.Pilots, pilot)
	}

	// Set various variables to initial state (Target info is untouched by this)...

	for _, pilot := range self.Pilots {
		pilot.Reset()
	}

	// Delete AIs with dead ships from the slice...

	for i := 0; i < len(self.Pilots); i++ {
		pilot := self.Pilots[i]
		if pilot.Alive() == false {
			self.Pilots = append(self.Pilots[:i], self.Pilots[i+1:]...)
			i--
		}
	}
}
