package ai

import (
	hal "../../bot/gohalite2"
)

// Possible changes (FIXME?):
//
// -- Reset target on failure (possibly after failing for n turns)

const (
	TIME_STEPS = 21			// The true value for resolution. Higher is more exact. Can be safely changed.
	SPACE_RESOLUTION = 2	// Leave this alone.
	FILL_RADIUS = 1			// Leave this alone.
)

type XYT struct {
	X			int
	Y			int
	T			int
}

type AirTrafficControl struct {
	Game		*hal.Game					// Nice to have for debugging.
	Grid		map[XYT][]hal.Point
}

func NewATC(game *hal.Game) *AirTrafficControl {
	ret := new(AirTrafficControl)
	ret.Game = game
	ret.Grid = make(map[XYT][]hal.Point)
	return ret
}

func (self *AirTrafficControl) Clear() {
	self.Grid = make(map[XYT][]hal.Point)
}

func (self *AirTrafficControl) SetRestrict(ship hal.Ship, speed, degrees int, yes bool) {

	x2, y2 := hal.Projection(ship.X, ship.Y, float64(speed), degrees)

	stepx := (x2 - ship.X) / TIME_STEPS
	stepy := (y2 - ship.Y) / TIME_STEPS

	x := ship.X
	y := ship.Y

	for t := 0; t < TIME_STEPS; t++ {

		x += stepx
		y += stepy

		point := hal.Point{x, y}

		grid_x := int(x) * SPACE_RESOLUTION
		grid_y := int(y) * SPACE_RESOLUTION

		for index_x := grid_x - FILL_RADIUS; index_x <= grid_x + FILL_RADIUS; index_x++ {
			for index_y := grid_y - FILL_RADIUS; index_y <= grid_y + FILL_RADIUS; index_y++ {
				if yes {
					self.Grid[XYT{index_x, index_y, t}] = append(self.Grid[XYT{index_x, index_y, t}], point)
				} else {
					self.Grid[XYT{index_x, index_y, t}] = hal.RemovePointFromSlice(self.Grid[XYT{index_x, index_y, t}], point)
				}
			}
		}
	}
}

func (self *AirTrafficControl) Restrict(ship hal.Ship, speed, degrees int) {
	self.SetRestrict(ship, speed, degrees, true)
}

func (self *AirTrafficControl) Unrestrict(ship hal.Ship, speed, degrees int) {
	self.SetRestrict(ship, speed, degrees, false)
}

func (self *AirTrafficControl) PathIsFree(ship hal.Ship, speed, degrees int) bool {

	const (
		SAFETY_MARGIN = 0.001
	)

	x2, y2 := hal.Projection(ship.X, ship.Y, float64(speed), degrees)

	stepx := (x2 - ship.X) / TIME_STEPS
	stepy := (y2 - ship.Y) / TIME_STEPS

	x := ship.X
	y := ship.Y

	for t := 0; t < TIME_STEPS; t++ {

		x += stepx
		y += stepy

		point := hal.Point{x, y}

		grid_x := int(x) * SPACE_RESOLUTION
		grid_y := int(y) * SPACE_RESOLUTION

		for index_x := grid_x - FILL_RADIUS; index_x <= grid_x + FILL_RADIUS; index_x++ {
			for index_y := grid_y - FILL_RADIUS; index_y <= grid_y + FILL_RADIUS; index_y++ {
				for _, restriction := range self.Grid[XYT{index_x, index_y, t}] {
					if restriction.Dist(point) <= hal.SHIP_RADIUS * 2 + SAFETY_MARGIN {
						return false
					}
				}
			}
		}
	}

	return true
}
