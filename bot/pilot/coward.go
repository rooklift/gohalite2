package pilot

import (
	"sort"

	hal "../core"
)

func (self *Pilot) PlanCowardice(all_enemies []hal.Ship, avoid_list []hal.Entity) {

	if self.DockedStatus != hal.UNDOCKED {
		return
	}

	edge, dist, flee_point := self.Game.NearestEdge(self.Ship)

	if dist < 3 {
		self.PlanEdgeCowardice(edge, all_enemies)
		return
	}

	self.Target = flee_point
	side := self.DecideSideFor(flee_point)

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

func (self *Pilot) PlanEdgeCowardice(edge hal.Edge, all_enemies []hal.Ship) {

	self.Message = MSG_COWARD

	sort.Slice(all_enemies, func(a, b int) bool {
		return self.Dist(all_enemies[a]) < self.Dist(all_enemies[b])
	})

	enemy_ship := all_enemies[0]

	switch edge {

	case hal.TOP: fallthrough
	case hal.BOTTOM:

		if enemy_ship.X > self.X {
			self.PlanThrust(7, 180)
		} else {
			self.PlanThrust(7, 0)
		}

	case hal.LEFT: fallthrough
	case hal.RIGHT:

		if enemy_ship.Y > self.Y {
			self.PlanThrust(7, 270)
		} else {
			self.PlanThrust(7, 90)
		}
	}
}
