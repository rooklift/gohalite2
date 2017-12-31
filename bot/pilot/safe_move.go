package pilot

import (
	hal "../core"
)

func ExecuteSafely(mobile_pilots []*Pilot) {

	if len(mobile_pilots) == 0 {
		return
	}

	game := mobile_pilots[0].Game

	// Assumption: we have already taken steps to ensure that any ship not included in the mobile_pilots
	// is avoided, i.e. those ships were explicitly avoided in the earlier navigation search.

	for n := 0; n < 10; n++ {

		total_executes := 0

		Pilot1Loop:

		for _, pilot1 := range mobile_pilots {

			if pilot1.HasExecuted {
				total_executes++
				continue
			}

			if n >= 5 {
				pilot1.SlowPlanDown()
			}

			pilot1_desired_speed, pilot1_desired_angle := pilot1.CourseFromPlan()

			if game.CourseStaysInBounds(pilot1.Ship, pilot1_desired_speed, pilot1_desired_angle) == false {
				continue
			}

			for _, pilot2 := range mobile_pilots {

				if pilot1 == pilot2 {
					continue
				}

				if pilot1.Dist(pilot2) > 15 {
					continue
				}

				pilot2_speed := 0
				pilot2_angle := 0

				if pilot2.HasExecuted {
					pilot2_speed, pilot2_angle = hal.CourseFromString(game.CurrentOrder(pilot2.Ship))
				}

				if hal.ShipsWillCollide(pilot1.Ship, pilot1_desired_speed, pilot1_desired_angle, pilot2.Ship, pilot2_speed, pilot2_angle) {
					continue Pilot1Loop
				}
			}

			pilot1.ExecutePlan()
			total_executes++
		}

		if total_executes >= len(mobile_pilots) {
			return
		}
	}
}
