package ai

import (
	hal "../gohalite2"
)

// Possible changes (FIXME?):
//
// -- Use a map[XYT]bool instead of a 3D slice? (Faster or slower??)
// -- Use a higher resolution?
// -- Reset target on failure (possibly after failing for n turns)

const (
	TIME_STEPS = 7
	RESOLUTION = 2		// i.e. double resolution
	FILL_RADIUS = 1		// filling from -1 to +1 inclusive
)

type XYT struct {
	X			int
	Y			int
	T			int
}

type AirTrafficControl struct {
	Grid		map[XYT]bool
	Width		int
	Height		int
}

func NewATC(world_width, world_height int) *AirTrafficControl {
	ret := new(AirTrafficControl)
	ret.Width = world_width * RESOLUTION
	ret.Height = world_height * RESOLUTION
	ret.Grid = make(map[XYT]bool)
	return ret
}

func (self *AirTrafficControl) Clear() {
	self.Grid = make(map[XYT]bool)
}

func (self *AirTrafficControl) SetRestrict(ship hal.Ship, speed, degrees int, val bool) {

	x2, y2 := hal.Projection(ship.X, ship.Y, float64(speed), degrees)

	stepx := (x2 - ship.X) / TIME_STEPS
	stepy := (y2 - ship.Y) / TIME_STEPS

	x := ship.X
	y := ship.Y

	for t := 0; t < TIME_STEPS; t++ {

		x += stepx
		y += stepy

		grid_x := int(x) * RESOLUTION
		grid_y := int(y) * RESOLUTION

		for index_x := grid_x - FILL_RADIUS; index_x <= grid_x + FILL_RADIUS; index_x++ {
			for index_y := grid_y - FILL_RADIUS; index_y <= grid_y + FILL_RADIUS; index_y++ {
				self.Grid[XYT{index_x, index_y, t}] = val
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

	x2, y2 := hal.Projection(ship.X, ship.Y, float64(speed), degrees)

	stepx := (x2 - ship.X) / TIME_STEPS
	stepy := (y2 - ship.Y) / TIME_STEPS

	x := ship.X
	y := ship.Y

	for t := 0; t < TIME_STEPS; t++ {

		x += stepx
		y += stepy

		grid_x := int(x) * RESOLUTION
		grid_y := int(y) * RESOLUTION

		for index_x := grid_x - FILL_RADIUS; index_x <= grid_x + FILL_RADIUS; index_x++ {
			for index_y := grid_y - FILL_RADIUS; index_y <= grid_y + FILL_RADIUS; index_y++ {
				if self.Grid[XYT{index_x, index_y, t}] {
					return false
				}
			}
		}
	}

	return true
}
