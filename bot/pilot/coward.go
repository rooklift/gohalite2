package pilot

import (
	"sort"

	hal "../core"
)

func (self *Pilot) PlanCowardice(all_enemies []hal.Ship, avoid_list []hal.Entity) {

	if self.DockedStatus != hal.UNDOCKED {
		return
	}

	sort.Slice(all_enemies, func(a, b int) bool {
		return self.Dist(all_enemies[a]) < self.Dist(all_enemies[b])
	})

	enemy_ship := all_enemies[0]

	angle := self.Angle(enemy_ship) + 180

	x2, y2 := hal.Projection(self.X, self.Y, 7, angle)
	flee_point := hal.Point{x2, y2}

	side := self.DecideSideFromTarget()

	speed, degrees, err := self.GetApproach(flee_point, 1, avoid_list, side)

	var msg int
	if err != nil {
		msg = MSG_RECURSION
	} else {
		msg = MSG_COWARD
	}

	self.PlanThrust(speed, degrees)
	self.Message = msg

}
