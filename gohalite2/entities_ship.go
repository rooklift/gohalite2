package gohalite2

import (
	"fmt"
)

// ---------------------------------------

type Ship struct {
	Id					int
	X					float64
	Y					float64
	HP					int
	Docked				int			// Is this really a bool?
	DockedPlanet		int
	DockingProgress		int
	Cooldown			int
	Order				string
}

func (self *Ship) Thrust(speed, angle int) {
	self.Order = fmt.Sprintf("t %d %d %d\n", self.Id, speed, angle)
}

func (self *Ship) Dock(planet int) {
	self.Order = fmt.Sprintf("d %d %d\n", self.Id, planet)
}

func (self *Ship) Undock(planet int) {
	self.Order = fmt.Sprintf("u %d\n", self.Id)
}

func (self *Ship) Noop() {
	self.Order = ""
}
