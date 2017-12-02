package pilot

import (
	hal "../core"
)

func (self *Pilot) Coward() {

	if self.DockedStatus == hal.DOCKED {
		self.PlanUndock()
		return
	}

	if self.DockedStatus != hal.UNDOCKED {
		return
	}

	// TO DO

}

