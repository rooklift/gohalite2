package ai

import (
	"fmt"
	"math/rand"
	"sort"

	hal "../../bot/gohalite2"
)

// --------------------------------------------

type Overmind struct {
	Game					*hal.Game
	ATC						*AirTrafficControl
	EnemyMap				map[int][]hal.Ship		// Planet ID --> Enemy ships near the planet (not docked)
	FriendlyMap				map[int][]hal.Ship		// Planet ID --> Friendly ships near the planet (not docked)
	ShipsDockingMap			map[int]int				// Planet ID --> My ship count docking this turn
	Pilots					[]*Pilot
}

func NewOvermind(game *hal.Game) *Overmind {
	ret := new(Overmind)
	ret.Game = game
	ret.ATC = NewATC(game)
	return ret
}

func (self *Overmind) ResetPilots() {

	game := self.Game

	// Add new AIs for new ships...

	my_new_ships := game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		pilot := new(Pilot)
		pilot.Overmind = self
		pilot.Game = game
		pilot.Id = sid								// This has to be set so pilot.Reset() can work.
		pilot.Target = hal.Nothing{}				// The null target. We don't ever use nil here.
		self.Pilots = append(self.Pilots, pilot)
	}

	// Set various variables to initial state, but keeping current target.
	// Also update target info from the Game. Also delete pilot if the ship is dead.

	for i := 0; i < len(self.Pilots); i++ {
		pilot := self.Pilots[i]
		alive := pilot.ResetAndUpdate()
		if alive == false {
			self.Pilots = append(self.Pilots[:i], self.Pilots[i+1:]...)
			i--
		}
		pilot.Target = hal.Nothing{}				// Brine has no long term targets.
	}
}

func (self *Overmind) UpdateProximityMaps() {

	// Currently only includes non-docked ships.

	const (
		THREAT_RANGE = 10
	)

	self.EnemyMap = make(map[int][]hal.Ship)
	self.FriendlyMap = make(map[int][]hal.Ship)

	all_ships := self.Game.AllShips()
	all_planets := self.Game.AllPlanets()

	for _, ship := range all_ships {
		if ship.CanMove() {
			for _, planet := range all_planets {
				if ship.ApproachDist(planet) < THREAT_RANGE {
					if ship.Owner != self.Game.Pid() {
						self.EnemyMap[planet.Id] = append(self.EnemyMap[planet.Id], ship)
					} else {
						self.FriendlyMap[planet.Id] = append(self.FriendlyMap[planet.Id], ship)
					}
				}
			}
		}
	}
}

type Problem struct {
	Entity		hal.Entity
	X			float64
	Y			float64
	Need		int
}

func (self *Problem) String() string {
	return fmt.Sprintf("%v (%d)", self.Entity, self.Need)
}

func (self *Overmind) Step() {

	self.ResetPilots()
	self.UpdateProximityMaps()
	self.ShipsDockingMap = make(map[int]int)
	self.ATC.Clear()

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

		sort.Slice(all_problems, func(a, b int) bool {
			return Dist(pilot.X, pilot.Y, all_problems[a].X, all_problems[a].Y) < Dist(pilot.X, pilot.Y, all_problems[b].X, all_problems[b].Y)
		})

		pilot.SetTarget(all_problems[0].Entity)
		all_problems[0].Need--
		if all_problems[0].Need <= 0 {
			all_problems = all_problems[1:]
		}
	}

	// See if we can optimise a bit...

	swaps := 0

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
				swaps++
			}
		}
	}

	self.ExecuteMoves()
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
			Need: 1,
		}
		all_problems = append(all_problems, problem)
	}

	return all_problems
}

func (self *Overmind) PlanetProblem(planet hal.Planet) *Problem {

	game := self.Game

	if game.DesiredSpots(planet) > 0 || len(self.EnemyMap[planet.Id]) > 0 {

		fight_strength := len(self.EnemyMap[planet.Id]) * 2
		capture_strength := game.DesiredSpots(planet)

		return &Problem{
			Entity: planet,
			X: planet.X,
			Y: planet.Y,
			Need: Max(fight_strength, capture_strength),
		}
	}

	return nil
}

func (self *Overmind) ExecuteMoves() {

	avoid_list := self.Game.AllImmobile()		// To start with. AllImmobile() is planets + docked ships.

	// Setup data structures...

	var mobile_pilots []*Pilot
	var frozen_pilots []*Pilot		// Note that this doesn't include docked / docking / undocking ships.

	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.UNDOCKED {
			mobile_pilots = append(mobile_pilots, pilot)
		}
	}

	// As a special case (relevant for 1v1 rushes) sort 3 ships by distance to centre...
	// This is helpful for the ATC slowdown below.

	if len(mobile_pilots) <= 3 {

		centre_of_gravity := self.Game.AllShipsCentreOfGravity()

		sort.Slice(mobile_pilots, func(a, b int) bool {
			return mobile_pilots[a].Dist(centre_of_gravity) < mobile_pilots[b].Dist(centre_of_gravity)
		})
	}

	// Plan a Dock if possible. (And we're not chasing a ship.)
	// If we do, remove this pilot from the mobile pilots list and make it frozen.

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		if pilot.HasTarget() == false || pilot.Target.Type() == hal.PLANET || pilot.Target.Type() == hal.POINT {
			_, ok := pilot.PlanDockIfWise()
			if ok {
				mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
				frozen_pilots = append(frozen_pilots, pilot)
				i--
			}
		}
	}

	// Perhaps this pilot doesn't need to move? If so, consider it frozen.

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		pilot.PlanChase(avoid_list)			// avoid_list is, at this point, planets plus already-docked ships.
		if pilot.HasStationaryPlan() {
			mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
			frozen_pilots = append(frozen_pilots, pilot)
			i--
		}
	}

	// Our PlanChase() above didn't avoid these frozen ships. Remake plans with the new info.
	// Possibly avoiding collisions we would have (since the above ships won't use ATC).

	for _, pilot := range frozen_pilots {
		avoid_list = append(avoid_list, pilot.Ship)
	}

	for _, pilot := range mobile_pilots {
		pilot.ResetAndUpdate()
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

	// As a special case, at game start, allow retry with lower velocity...

	if len(self.Pilots) <= 3 {
		for n := 0; n < 2; n++ {
			for _, pilot := range mobile_pilots {
				if pilot.HasExecuted == false {
					pilot.SlowPlanDown()
					pilot.ExecutePlanWithATC(self.ATC)
				}
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
			pilot.PlanThrust(0, 0, MSG_ATC_DEACTIVATED)
			self.ATC.Unrestrict(pilot.Ship, 0, 0)
			mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
			frozen_pilots = append(frozen_pilots, pilot)
			i--
		}
	}

	avoid_list = self.Game.AllImmobile()			// Remake the avoid_list...
	for _, pilot := range frozen_pilots {
		avoid_list = append(avoid_list, pilot.Ship)
	}

	for _, pilot := range mobile_pilots {
		if pilot.HasExecuted == false {
			pilot.PlanChase(avoid_list)
			pilot.ExecutePlanWithATC(self.ATC)
		}
	}

	// Null thrust every "mobile" ship that didn't move. This causes target info to be put into
	// the replay via the Angle Message system.

	for _, pilot := range mobile_pilots {
		if pilot.HasExecuted == false {
			if pilot.Plan != "" {
				pilot.PlanThrust(0, 0, MSG_ATC_RESTRICT)
				pilot.ExecutePlan()
			}
		}
	}

	// Don't forget our non-mobile ships!

	for _, pilot := range frozen_pilots {
		pilot.ExecutePlan()
	}
}
