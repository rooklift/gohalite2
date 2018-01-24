package ai

import (
	"fmt"
	"sort"

	hal "../core"
	pil "../pilot"
)

type Problem struct {
	Entity		hal.Entity
	Value		float64
	Need		int
	Message		int
}

func (self *Problem) String() string {
	return fmt.Sprintf("%v (%d, %f)", self.Entity, self.Need, self.Value)
}

// -------------------------------------------------------------------------------

func (self *Overmind) ChooseTargets() {

	all_problems := self.AllProblems()

	// Initial assignment of problems to pilots...

	for _, pilot := range self.Pilots {

		if len(all_problems) == 0 {
			all_problems = self.AllProblems()
			if len(all_problems) == 0 {
				break
			}
		}

		if pilot.DockedStatus != hal.UNDOCKED {
			continue
		}

		if pilot.Target.Type() != hal.NOTHING {			// Because our target wasn't reset for some reason.
			pilot.MessageWhileLocked()
			continue
		}

		// While one might think of using ApproachDist here, in the real world it lost a mu or more...

		sort.Slice(all_problems, func(a, b int) bool {
			return pilot.Dist(all_problems[a].Entity) / all_problems[a].Value <
			       pilot.Dist(all_problems[b].Entity) / all_problems[b].Value
		})

		pilot.Target = all_problems[0].Entity
		pilot.Message = all_problems[0].Message

		all_problems[0].Need--							// We could consider only doing this for non-doomed pilots. Hmm. They still do damage though.
		if all_problems[0].Need <= 0 {
			all_problems = all_problems[1:]
		}
	}
}

// -------------------------------------------------------------------------------

func (self *Overmind) AllProblems() []*Problem {

	if self.RushChoice == RUSHING && self.AvoidingBad2v1 == false {
		return self.RushProblems()
	}

	var all_problems []*Problem

	for _, planet := range self.Game.AllPlanets() {
		problems := self.PlanetProblems(planet)
		all_problems = append(all_problems, problems...)
	}

	for _, ship := range self.Game.EnemyShips() {

		if ship.Doomed == false {		// Skip the ship (as an assassination target) if we expect it to die at time 0.
			problem := &Problem{		// Note that we may end up targetting it as a planet's secondary target.
				Entity: ship,
				Value: 1.0,
				Need: 1,
				Message: pil.MSG_ASSASSINATE,
			}
			all_problems = append(all_problems, problem)
		}
	}

	return all_problems
}

func (self *Overmind) PlanetProblems(planet *hal.Planet) []*Problem {

	var ret []*Problem

	enemies := self.Game.EnemiesNearPlanet(planet)
	capture_strength := self.Game.DesiredSpots(planet)

	switch len(enemies) {

	case 0:

		if capture_strength > 0 {

			value := 1.0 / 1.4; if self.Game.InitialPlayers() > 2 { value = 1.0 }

			ret = append(ret, &Problem{
				Entity: planet,
				Value: value,
				Need: capture_strength,
				Message: planet.Id,
			})
		}

	default:

		for _, enemy := range enemies {

			// We can't skip Doomed targets here because we need to actually doom them before we dock.

			ret = append(ret, &Problem{
				Entity: enemy,
				Value: 1.0,
				Need: 2,
				Message: planet.Id,
			})
		}
	}

	return ret
}

func (self *Overmind) RushProblems() []*Problem {

	var problems []*Problem

	var helpable_docked_ships []*hal.Ship

	for _, ship := range self.Game.MyShips() {
		if ship.DockedStatus != hal.UNDOCKED && ship.Doomed == false {
			helpable_docked_ships = append(helpable_docked_ships, ship)
		}
	}

	if len(helpable_docked_ships) > 0 {

		for _, ship := range helpable_docked_ships {

			problem := &Problem{
				Entity: ship,
				Value: 1.0,
				Need: 1,
				Message: ship.Id,
			}
			problems = append(problems, problem)
		}

	} else {

		relevant_enemies := self.Game.ShipsOwnedBy(self.RushEnemyID)

		some_are_docked := false

		for _, ship := range relevant_enemies {
			if ship.DockedStatus != hal.UNDOCKED {
				some_are_docked = true
				break
			}
		}

		for _, ship := range relevant_enemies {

			if ship.DockedStatus != hal.UNDOCKED || some_are_docked == false {

				if ship.Doomed == false {

					problem := &Problem{
						Entity: ship,
						Value: 1.0,
						Need: 1,
						Message: ship.Id,
					}

					problems = append(problems, problem)
				}
			}
		}
	}

	return problems
}

// -------------------------------------------------------------------------------

func (self *Overmind) OptimisePilots() {

	if self.AvoidingBad2v1 {
		return
	}

	for n := 0; n < 5; n++ {

		for i := 0; i < len(self.Pilots); i++ {

			pilot_a := self.Pilots[i]

			if pilot_a.DockedStatus != hal.UNDOCKED || pilot_a.Locked {
				continue
			}

			for j := i + 1; j < len(self.Pilots); j++ {

				pilot_b := self.Pilots[j]

				if pilot_b.DockedStatus != hal.UNDOCKED || pilot_b.Locked {
					continue
				}

				// Dist or ApproachDist won't matter here as long as it's consistent.
				// Either way the comparison will come out the same.

				total_dist := pilot_a.Dist(pilot_a.Target) + pilot_b.Dist(pilot_b.Target)
				swap_dist := pilot_a.Dist(pilot_b.Target) + pilot_b.Dist(pilot_a.Target)

				if swap_dist < total_dist {
					pilot_a.Target, pilot_b.Target = pilot_b.Target, pilot_a.Target
					pilot_a.Message, pilot_b.Message = pilot_b.Message, pilot_a.Message
				}
			}
		}
	}
}
