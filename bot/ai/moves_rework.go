package ai

import (
	"sort"

	hal "../core"
	nav "../navigation"
	pil "../pilot"
)

func (self *Overmind) ExecuteMoves2() {

	sort.Slice(self.Pilots, func(a, b int) bool {
		return self.Pilots[a].Dist(self.Pilots[a].Target) < self.Pilots[b].Dist(self.Pilots[b].Target)
	})

	ignore_inhibition := (self.RushChoice == RUSHING)

	raw_avoid_list := self.Game.AllImmobile()
	var avoid_list []hal.Entity

	for _, entity := range raw_avoid_list {
		switch entity.Type() {
		case hal.SHIP:
			if entity.(*hal.Ship).Doomed == false {
				avoid_list = append(avoid_list, entity)
			}
		default:
			avoid_list = append(avoid_list, entity)
		}
	}

	// Setup data structures...

	var mobile_pilots []*pil.Pilot
	var frozen_pilots []*pil.Pilot				// Note that this doesn't include (already) docking / docked / undocking ships.

	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.UNDOCKED && pilot.Doomed == false {
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

	// Plan moves, add non-moving ships to the avoid list, then scrap other moves and plan them again...

	for _, pilot := range mobile_pilots {
		pilot.PlanChase(avoid_list, ignore_inhibition)
	}

	for i := 0; i < len(mobile_pilots); i++ {
		pilot := mobile_pilots[i]
		if pilot.HasStationaryPlan() {
			mobile_pilots = append(mobile_pilots[:i], mobile_pilots[i+1:]...)
			frozen_pilots = append(frozen_pilots, pilot)
			avoid_list = append(avoid_list, pilot.Ship)
			i--
		}
	}

	for _, pilot := range mobile_pilots {
		pilot.PlanChase(avoid_list, ignore_inhibition)
	}

	// Execute...

	ExecuteSafely2(mobile_pilots, avoid_list)

	for _, pilot := range frozen_pilots {
		pilot.ExecutePlan()
	}
}


func ExecuteSafely2(mobile_pilots []*pil.Pilot, avoid_list []hal.Entity) {

	// So avoid_list is planets, docked ships, and our own ships that are committed to not moving.

	if len(mobile_pilots) == 0 {
		return
	}

	game := mobile_pilots[0].Game

	for _, pilot := range mobile_pilots {
		if pilot.NavTarget == nil {
			pilot.NavTarget = pilot.Target
		}
	}

	sort.Slice(mobile_pilots, func(a, b int) bool {
		return mobile_pilots[a].Dist(mobile_pilots[a].NavTarget) < mobile_pilots[b].Dist(mobile_pilots[b].NavTarget)
	})


	// FIXME: add wall detection.


	for n := 0; n < 67; n++ {

		adjustment := (n + 1) / 2			// 0, 1, 1, 2, 2, 3, 3...
		if n > 60 {
			adjustment = 0
		}

		total_executes := 0

		Pilot1Loop:

		for _, pilot1 := range mobile_pilots {

			if pilot1.HasExecuted {
				total_executes++
				continue
			}

			if (pilot1.Id + n) % 2 == 0 {
				adjustment = -adjustment
			}

			if n > 60 {
				pilot1.SlowPlanDown()
			}

			pilot1_desired_speed, pilot1_desired_angle := pilot1.CourseFromPlan()
			pilot1_desired_angle += adjustment

			if pilot1_desired_angle < 0 { pilot1_desired_angle += 360 }
			if pilot1_desired_angle > 359 { pilot1_desired_angle -= 360 }

			msg1 := pilot1.Message

			for _, e := range avoid_list {
				if nav.CheckEntityCollision(pilot1.Ship, float64(pilot1_desired_speed), pilot1_desired_angle, e) {
					continue Pilot1Loop
				}
			}

			for _, pilot2 := range mobile_pilots {

				if pilot2 == pilot1 {
					continue
				}

				if pilot2.Dist(pilot1) > 15 {
					continue
				}

				if pilot2.Doomed {
					continue
				}

				pilot2_speed := 0
				pilot2_angle := 0

				if pilot2.HasExecuted {
					pilot2_speed, pilot2_angle = hal.CourseFromString(game.CurrentOrder(pilot2.Ship))
				}

				msg2 := pilot2.Message

				if hal.ShipsWillCollide(pilot1.Ship, pilot1_desired_speed, pilot1_desired_angle, msg1, pilot2.Ship, pilot2_speed, pilot2_angle, msg2) {
					continue Pilot1Loop
				}
			}

			pilot1.PlanThrust(pilot1_desired_speed, pilot1_desired_angle)
			pilot1.ExecutePlan()
			total_executes++
		}

		if total_executes >= len(mobile_pilots) {
			return
		}
	}
}
