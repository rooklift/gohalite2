package ai

import (
	"math/rand"
	"sort"

	hal "../gohalite2"
)

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

	// Choose target if needed... (i.e. we don't have a valid target already).

	all_enemy_ships := self.Game.EnemyShips()

	for _, pilot := range mobile_pilots {
		if CONFIG.Stateless || pilot.ValidateTarget() == false {
			pilot.ChooseTarget(all_enemy_ships)		// Also chooses from planets. But we cache ships for speed.
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

func (self *Overmind) UpdatePilots() {

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
		if pilot.Target == nil {
			panic("nil pilot.Target")
		}
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

func (self *Overmind) UpdateShipChases() {
	self.EnemyShipsChased = make(map[int][]int)
	for _, pilot := range self.Pilots {
		if pilot.Target.Type() == hal.SHIP {
			target := pilot.Target.(hal.Ship)
			self.EnemyShipsChased[target.Id] = append(self.EnemyShipsChased[target.Id], pilot.Id)
		}
	}
}
