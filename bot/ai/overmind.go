package ai

import (
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

	for _, pilot := range self.Pilots {
		pilot.MakePlan()
		pilot.PreliminaryRestrict(self.ATC)
		pilot.ExecutePlanIfStationary(self.ATC)
	}

	// Make multiple attempts to execute our orders. Should allow chains of ships to move OK.

	for n := 0; n < 5; n++ {
		for _, pilot := range self.Pilots {
			if pilot.HasOrdered == false {
				pilot.ExecutePlanIfSafe(self.ATC)
			}
		}
	}

	// Null thrust everyone who didn't move...

	for _, pilot := range self.Pilots {
		if pilot.HasOrdered == false && pilot.DockedStatus == hal.UNDOCKED {
			pilot.PlanThrust(0, 0)
			pilot.ExecutePlan()
		}
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
		pilot.Id = sid								// This has to be set so pilot.Update() can work.
		self.Pilots = append(self.Pilots, pilot)
	}

	// Update the Ships embedded in each Pilot... (yeah that makes sense)

	for _, pilot := range self.Pilots {
		pilot.Update()
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
