package ai

import (
	"math/rand"
	"sort"

	gen "../genetic"
	hal "../core"
	pil "../pilot"
)

// --------------------------------------------

type Overmind struct {
	Pilots					[]*pil.Pilot
	Game					*hal.Game
	ShipsDockingCount		map[int]int				// Planet ID --> My ship count docking this turn
	CowardFlag				bool
	RushFlag				bool
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

func (self *Overmind) Step() {

	self.ResetPilots()
	self.ShipsDockingCount = make(map[int]int)
	self.SetCowardFlag()

	if self.Game.Turn() == 0 {
		self.SetRushFlag()
		if self.RushFlag {
			self.SetRushTargets()
			self.TurnZeroCluster()
			return
		} else {
			self.ChooseThreeDocks()
		}
	}

	if self.CowardFlag {
		self.CowardStep()
		return
	}

	if CONFIG.Conservative == false && self.DetectRushFight() {
		gen.FightRush(self.Game)
		return
	}

	self.ChooseTargets()
	self.ExecuteMoves()
}

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

		// The stateless version -- usually -- has no long term targets...

		if self.RushFlag == false || pilot.Id > 5 {
			if pilot.Target.Type() != hal.POINT {
				pilot.Target = hal.Nothing{}
				pilot.TurnTarget = hal.Nothing{}
			}
		}
	}
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
