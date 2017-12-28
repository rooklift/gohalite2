package ai

import (
	"fmt"
	"sort"

	hal "../core"
)

type Problem struct {
	Entity		hal.Entity
	X			float64
	Y			float64
	Value		float64
	Need		int
}

func (self *Problem) String() string {
	return fmt.Sprintf("%v (%d)", self.Entity, self.Need)
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

		sort.Slice(all_problems, func(a, b int) bool {
			return hal.Dist(pilot.X, pilot.Y, all_problems[a].X, all_problems[a].Y) / all_problems[a].Value <
			       hal.Dist(pilot.X, pilot.Y, all_problems[b].X, all_problems[b].Y) / all_problems[b].Value
		})

		pilot.SetTarget(all_problems[0].Entity)
		all_problems[0].Need--
		if all_problems[0].Need <= 0 {
			all_problems = all_problems[1:]
		}
	}

	// See if we can optimise a bit...

	for n := 0; n < 5; n++ {

		for i := 0; i < len(self.Pilots); i++ {

			pilot_a := self.Pilots[i]

			if pilot_a.DockedStatus != hal.UNDOCKED {
				continue
			}

			for j := i + 1; j < len(self.Pilots); j++ {

				pilot_b := self.Pilots[j]

				if pilot_b.DockedStatus != hal.UNDOCKED {
					continue
				}

				total_dist := pilot_a.Dist(pilot_a.Target) + pilot_b.Dist(pilot_b.Target)
				swap_dist := pilot_a.Dist(pilot_b.Target) + pilot_b.Dist(pilot_a.Target)

				if swap_dist < total_dist {
					pilot_a.Target, pilot_b.Target = pilot_b.Target, pilot_a.Target
				}
			}
		}
	}

	for _, pilot := range self.Pilots {
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
		problem := &Problem{
			Entity: ship,
			X: ship.X,
			Y: ship.Y,
			Value: 1.0,
			Need: 1,
		}
		all_problems = append(all_problems, problem)
	}

	return all_problems
}

func (self *Overmind) PlanetProblem(planet hal.Planet) *Problem {

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
			X: planet.X,
			Y: planet.Y,
			Value: value,
			Need: hal.Max(fight_strength, capture_strength),
		}

	}

	return nil
}
