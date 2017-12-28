package ai

import (
	"math/rand"
	"sort"

	hal "../core"
	pil "../pilot"
)

const (
	DEBUG_SHIP_ID = -1
	DEBUG_TURN = -1
)

func (self *Overmind) Step() {

	self.UpdatePilots()
	self.UpdateChasers()						// Must happen after self.Pilots is updated
	self.ShipsDockingCount = make(map[int]int)
	self.SetCowardFlag()

	// -----------------

	if self.Game.Turn() == 0 {
		self.TurnZero()
	} else if self.DetectRushFight() {
		self.HandleRushFight()
	} else if self.CowardFlag {
		self.CowardStep()
	} else {
		self.NormalStep()
	}
}

func (self *Overmind) NormalStep() {

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
	// This is helpful for the ATC slowdown below.

	if len(self.Pilots) <= 3 {

		centre_of_gravity := self.Game.AllShipsCentreOfGravity()

		sort.Slice(mobile_pilots, func(a, b int) bool {
			return mobile_pilots[a].Dist(centre_of_gravity) < mobile_pilots[b].Dist(centre_of_gravity)
		})
	}

	// Plan a Dock if possible. Not allowed if:
	//     - We have no target because our target just died. (Empirically, this seems a bad time to dock.)
	//     - We are chasing a ship.
	// If we do, remove this pilot from the mobile pilots list and make it frozen.

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		if (pilot.HasTarget() == false && pilot.Birth == self.Game.Turn()) || pilot.Target.Type() == hal.PLANET || pilot.Target.Type() == hal.POINT {
			if self.DockIfWise(pilot) {
				mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
				frozen_pilots = append(frozen_pilots, pilot)
				i--
			}
		}
	}

	// Choose target if needed... (i.e. we don't have a valid target already, or we are stateless).

	if CONFIG.Stateless && len(self.Game.MyShips()) > 3 {
		for _, pilot := range mobile_pilots {
			pilot.SetTarget(hal.Nothing{})										// Also causes the Chaser maps to be updated.
		}
	}

	all_enemy_ships := self.Game.EnemyShips()
	all_planets := self.Game.AllPlanets()

	for _, pilot := range mobile_pilots {
		if self.ValidateTarget(pilot) == false {
			self.ChooseTarget(pilot, all_planets, all_enemy_ships)
		}
	}

	// Swap targets if this results in decreased travel distance...
	// FIXME? Can do more passes of this...

	for a := 0; a < len(mobile_pilots); a++ {

		pa := mobile_pilots[a]

		if pa.Target.Type() == hal.NOTHING {
			continue
		}

		for b := a + 1; b < len(mobile_pilots); b++ {

			pb := mobile_pilots[b]

			if pb.Target.Type() == hal.NOTHING {
				continue
			}

			current_dist := pa.Dist(pa.Target) + pb.Dist(pb.Target)

			swap_dist := pa.Dist(pb.Target) + pb.Dist(pa.Target)

			if swap_dist < current_dist {
				// self.Game.Log("Swapping targets: %v (%v), %v (%v), gain %.2f", pa.Id, pa.Target, pb.Id, pb.Target, current_dist - swap_dist)
				pa.Target, pb.Target = pb.Target, pa.Target
			}
		}
	}

	// Set tactical (single turn) targets...

	for _, pilot := range mobile_pilots {
		pilot.SetTurnTarget()
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
		pilot.ResetPlan()
		pilot.PlanChase(avoid_list)
	}

	// Now the only danger is 2 "mobile" ships colliding. (Though one of them may not actually be moving.)

	pil.ExecuteSafely(mobile_pilots)

	// Randomly give up for half the ships that still aren't moving, and
	// retry the pathfinding with the other half.

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

	// Debug...

	if self.Game.Turn() == DEBUG_TURN {
		for _, pilot := range self.Pilots {
			if pilot.Id == DEBUG_SHIP_ID {
				pilot.LogNavStack()
				break
			}
		}
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
