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
	Speedx				float64
	Speedy				float64
	Docked				int			// Is this really a bool?
	DockedPlanet		int
	DockingProgress		int
	Cooldown			int
	Order				string
}

func (self *Ship) Thrust(speed, angle int) {
	self.Order = fmt.Sprintf("t %d %d %d", self.Id, speed, angle)
}

func (self *Ship) Dock(planet int) {
	self.Order = fmt.Sprintf("d %d %d", self.Id, planet)
}

func (self *Ship) Undock(planet int) {
	self.Order = fmt.Sprintf("u %d", self.Id)
}
