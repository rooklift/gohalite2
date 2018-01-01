package ai

import (
	"fmt"
	"sort"

	hal "../core"
)

type Problem struct {
	Entity		hal.Entity
	Value		float64
	Need		int
}

func (self *Problem) String() string {
	return fmt.Sprintf("%v (%d, %f)", self.Entity, self.Need, self.Value)
}

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

		if pilot.DockedStatus != hal.UNDOCKED || pilot.Target.Type() != hal.NOTHING {
			continue
		}

		// While one might think of using ApproachDist here, in the real world it lost a mu or more...

		sort.Slice(all_problems, func(a, b int) bool {
			return pilot.Dist(all_problems[a].Entity) / all_problems[a].Value <
			       pilot.Dist(all_problems[b].Entity) / all_problems[b].Value
		})

		pilot.Target = all_problems[0].Entity
		all_problems[0].Need--							// We could consider only doing this for non-doomed pilots. Hmm. They still do damage though.
		if all_problems[0].Need <= 0 {
			all_problems = all_problems[1:]
		}
	}

	// Optimise (swap targets for better overall distance). Best do this before choosing short term targets...

	self.OptimisePilots()

	// Choose what tactical target we have this turn; i.e. if our main target is a planet, we may target a ship near that planet...

	for _, pilot := range self.Pilots {
		pilot.SetMessageFromTarget()
		pilot.SetTurnTarget()
	}
}

func (self *Overmind) AllProblems() []*Problem {

	var all_problems []*Problem

	for _, planet := range self.Game.AllPlanets() {
		problem := self.PlanetProblem(planet)
		if problem != nil {
			all_problems = append(all_problems, problem)
		}
	}

	for _, ship := range self.Game.EnemyShips() {

		if ship.Doomed == false {		// Skip the ship if we expect it to die at time 0
			problem := &Problem{
				Entity: ship,
				Value: 1.0,
				Need: 1,
			}
			all_problems = append(all_problems, problem)
		}
	}

	return all_problems
}

func (self *Overmind) PlanetProblem(planet *hal.Planet) *Problem {

	game := self.Game

	fight_strength := len(game.EnemiesNearPlanet(planet)) * 2
	capture_strength := game.DesiredSpots(planet)

	// Start with low value, but increase it to 1.0 if there's fighting to be done at the planet (enemies near it),
	// or if it's a 4 player game.
	//
	// However, since we do this every frame, it's not like the old stateful bot where target choice was made at
	// the moment that ship spawned and then kept.

	value := 1.0 / 1.4

	if fight_strength > 0 || self.Game.InitialPlayers() > 2 {
		value = 1.0
	}

	if capture_strength > 0 || fight_strength > 0 {

		return &Problem{
			Entity: planet,
			Value: value,
			Need: hal.Max(fight_strength, capture_strength),
		}

	}

	return nil
}

func (self *Overmind) OptimisePilots() {

	for n := 0; n < 5; n++ {

		for i := 0; i < len(self.Pilots); i++ {

			pilot_a := self.Pilots[i]

			if pilot_a.DockedStatus != hal.UNDOCKED || pilot_a.Target.Type() == hal.PORT {
				continue
			}

			for j := i + 1; j < len(self.Pilots); j++ {

				pilot_b := self.Pilots[j]

				if pilot_b.DockedStatus != hal.UNDOCKED || pilot_b.Target.Type() == hal.PORT {
					continue
				}

				// Dist or ApproachDist won't matter here as long as it's consistent.
				// Either way the comparison will come out the same.

				total_dist := pilot_a.Dist(pilot_a.Target) + pilot_b.Dist(pilot_b.Target)
				swap_dist := pilot_a.Dist(pilot_b.Target) + pilot_b.Dist(pilot_a.Target)

				if swap_dist < total_dist {
					pilot_a.Target, pilot_b.Target = pilot_b.Target, pilot_a.Target
				}
			}
		}
	}
}
