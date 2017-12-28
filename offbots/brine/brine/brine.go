package brine

import (
	"fmt"
	"math/rand"
	"sort"

	hal "../../../bot/core"
	pil "../../../bot/pilot"
)

// --------------------------------------------

type Overmind struct {
	Pilots					[]*pil.Pilot
	Game					*hal.Game
	ShipsDockingCount		map[int]int				// Planet ID --> My ship count docking this turn
	CowardFlag				bool
}

func NewOvermind(game *hal.Game) *Overmind {
	ret := new(Overmind)
	ret.Game = game
	ret.Game.SetThreatRange(20)
	return ret
}

func (self *Overmind) NotifyTargetChange(pilot *pil.Pilot, old_target, new_target hal.Entity) {
	// pass
}

func (self *Overmind) NotifyDock(planet hal.Planet) {
	self.ShipsDockingCount[planet.Id]++
}

// --------------------------------------------

func (self *Overmind) ShipsAboutToDock(planet hal.Planet) int {
	return self.ShipsDockingCount[planet.Id]
}

// --------------------------------------------

func (self *Overmind) ResetPilots() {

	game := self.Game

	// Add new AIs for new ships...

	my_new_ships := game.MyNewShipIDs()

	for _, sid := range my_new_ships {
		pilot := pil.NewPilot(sid, game, self)
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

		if pilot.Target.Type() != hal.POINT {
			pilot.Target = hal.Nothing{}				// Brine has no long term targets, except points during the opening.
			pilot.TurnTarget = hal.Nothing{}
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
	self.ShipsDockingCount = make(map[int]int)
	self.SetCowardFlag()

	if self.Game.Turn() == 0 {
		self.ChooseThreeDocks()
	}

	if self.CowardFlag {
		self.CowardStep()
	} else {
		self.ChooseTargets()
		self.ExecuteMoves()
	}
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

		if pilot.DockedStatus != hal.UNDOCKED || pilot.Target.Type() == hal.POINT {
			continue
		}

		sort.Slice(all_problems, func(a, b int) bool {
			return hal.Dist(pilot.X, pilot.Y, all_problems[a].X, all_problems[a].Y) < hal.Dist(pilot.X, pilot.Y, all_problems[b].X, all_problems[b].Y)
		})

		pilot.SetTarget(all_problems[0].Entity)
		all_problems[0].Need--
		if all_problems[0].Need <= 0 {
			all_problems = all_problems[1:]
		}
	}

	// See if we can optimise a bit...

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
			Need: 1,
		}
		all_problems = append(all_problems, problem)
	}

	return all_problems
}

func (self *Overmind) PlanetProblem(planet hal.Planet) *Problem {

	game := self.Game

	if game.DesiredSpots(planet) > 0 || len(game.EnemiesNearPlanet(planet)) > 0 {

		fight_strength := len(game.EnemiesNearPlanet(planet)) * 2
		capture_strength := game.DesiredSpots(planet)

		return &Problem{
			Entity: planet,
			X: planet.X,
			Y: planet.Y,
			Need: hal.Max(fight_strength, capture_strength),
		}
	}

	return nil
}

func (self *Overmind) ExecuteMoves() {

	avoid_list := self.Game.AllImmobile()		// To start with. AllImmobile() is planets + docked ships.

	// Setup data structures...

	var mobile_pilots []*pil.Pilot
	var frozen_pilots []*pil.Pilot				// Note that this doesn't include docked / docking / undocking ships.

	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.UNDOCKED {
			mobile_pilots = append(mobile_pilots, pilot)
		}
	}

	// As a special case (relevant for 1v1 rushes) sort 3 ships by distance to centre...

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
			ok := self.DockIfWise(pilot)
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

	for _, pilot := range frozen_pilots {
		avoid_list = append(avoid_list, pilot.Ship)
	}

	for _, pilot := range mobile_pilots {
		pilot.ResetPlan()
		pilot.PlanChase(avoid_list)
	}

	// Now the only danger is 2 "mobile" ships colliding. We use the ATC for this possibility.
	// Note that it's possible that one of the colliding ships will not actually be moving.

	pil.ExecuteSafely(mobile_pilots)

	// Randomly give up for half the ships that still aren't moving, and
	// retry the pathfinding with the other half.

	// Ships moved into the frozen slice can have their ATC restriction
	// cleared since we will navigate around them precisely.

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		if pilot.HasExecuted == false && rand.Intn(2) == 0 {
			pilot.PlanThrust(0, 0)
			pilot.Message = pil.MSG_ATC_DEACTIVATED
			mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
			frozen_pilots = append(frozen_pilots, pilot)
			i--
		}
	}

	// Remake the avoid_list...

	avoid_list = self.Game.AllImmobile()
	for _, pilot := range frozen_pilots {
		avoid_list = append(avoid_list, pilot.Ship)
	}

	// Remake plans for our non-moving ships that we didn't freeze...

	for _, pilot := range mobile_pilots {
		if pilot.HasExecuted == false {
			pilot.PlanChase(avoid_list)
		}
	}

	// And execute. Note that pilots that have already executed won't be affected...

	pil.ExecuteSafely(mobile_pilots)

	// Null thrust every "mobile" ship that didn't move. This causes target info to be put into
	// the replay via the Angle Message system.

	for _, pilot := range mobile_pilots {
		if pilot.HasExecuted == false {
			if pilot.Plan != "" {
				pilot.PlanThrust(0, 0)
				pilot.Message = pil.MSG_ATC_RESTRICT
				pilot.ExecutePlan()
			}
		}
	}

	// Don't forget our non-mobile ships!

	for _, pilot := range frozen_pilots {
		pilot.ExecutePlan()
	}
}

func (self *Overmind) DockIfWise(pilot *pil.Pilot) bool {

	if pilot.DockedStatus != hal.UNDOCKED {
		return false
	}

	closest_planet := pilot.ClosestPlanet()

	if pilot.CanDock(closest_planet) == false {
		return false
	}

	// Pilots with point targets should always succeed in docking...

	if pilot.Target.Type() == hal.POINT {
		pilot.SetTarget(closest_planet)			// It would be sad to stay with a Point target forever...
		pilot.PlanDock(closest_planet)
		return true
	}

	// Otherwise we check some things...

	if len(self.Game.EnemiesNearPlanet(closest_planet)) > 0 {
		return false
	}

	if self.ShipsAboutToDock(closest_planet) >= self.Game.DesiredSpots(closest_planet) {
		return false
	}

	pilot.PlanDock(closest_planet)
	return true
}

func (self *Overmind) ChooseThreeDocks() {

	// Sort all planets by distance to our fleet...

	all_planets := self.Game.AllPlanets()

	sort.Slice(all_planets, func(a, b int) bool {
		return all_planets[a].ApproachDist(self.Pilots[0]) < all_planets[b].ApproachDist(self.Pilots[0])
	})

	closest_three := all_planets[:3]

	// Get docks...

	var docks []hal.Point

	for _, planet := range closest_three {
		docks = append(docks, planet.OpeningDockHelper(self.Pilots[0].Ship)...)
	}

	docks = docks[:3]

	var permutations = [][]int{
		[]int{0,1,2},
		[]int{0,2,1},
		[]int{1,0,2},
		[]int{1,2,0},
		[]int{2,0,1},
		[]int{2,1,0},
	}

	for _, perm := range permutations {		// Find a non-crossing solution...

		self.Pilots[0].SetTarget(docks[perm[0]])
		self.Pilots[1].SetTarget(docks[perm[1]])
		self.Pilots[2].SetTarget(docks[perm[2]])

		if hal.Intersect(self.Pilots[0].Ship, self.Pilots[0].Target, self.Pilots[1].Ship, self.Pilots[1].Target) {
			continue
		}

		if hal.Intersect(self.Pilots[0].Ship, self.Pilots[0].Target, self.Pilots[2].Ship, self.Pilots[2].Target) {
			continue
		}

		if hal.Intersect(self.Pilots[1].Ship, self.Pilots[1].Target, self.Pilots[2].Ship, self.Pilots[2].Target) {
			continue
		}

		break
	}
}

func (self *Overmind) CowardStep() {

	var mobile_pilots []*pil.Pilot

	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.UNDOCKED {
			mobile_pilots = append(mobile_pilots, pilot)
		}
	}

	all_enemies := self.Game.EnemyShips()
	avoid_list := self.Game.AllImmobile()

	for _, pilot := range mobile_pilots {
		pilot.PlanCowardice(all_enemies, avoid_list)
	}

	pil.ExecuteSafely(mobile_pilots)

	// Also undock any docked ships...

	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.DOCKED {
			pilot.PlanUndock()
			pilot.ExecutePlan()
		}
	}
}

func (self *Overmind) SetCowardFlag() {

	if self.Game.CurrentPlayers() <= 2 {
		self.CowardFlag = false
		return
	}

	if self.CowardFlag {
		return				// i.e. leave it true
	}

	if self.Game.CountMyShips() < self.Game.CountEnemyShips() / 10 {
		self.CowardFlag = true
	}
}
