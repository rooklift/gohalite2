package pilot

import (
	hal "../core"
)

type RallyPoints struct {
	Points				[]*hal.Point
}

func (self *RallyPoints) Clear() {
	if self == nil { return }
	self.Points = nil
}

func (self *RallyPoints) Add(x, y float64) {
	if self == nil { return }
	self.Points = append(self.Points, &hal.Point{x, y})
}

func (self *RallyPoints) ClosestTo(e hal.Entity) (*hal.Point, float64, bool) {

	if self == nil || len(self.Points) == 0 {
		return nil, 999, false
	}

	var ret *hal.Point
	var best_dist float64

	for _, point := range self.Points {
		d := point.Dist(e)
		if ret == nil || d < best_dist {
			ret = point
			best_dist = d
		}
	}

	return ret, best_dist, true
}
