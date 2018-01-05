package ai

import (
	hal "../core"
	pil "../pilot"
)

func (self *Overmind) CowardStep() {

	var mobile_pilots []*pil.Pilot

	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.UNDOCKED {
			mobile_pilots = append(mobile_pilots, pilot)
		}
	}

	all_enemies := self.Game.EnemyShips()
	avoid_list := self.Game.AllImmobile()

	for _, pilot := range mobile_pilots {
		pilot.PlanCowardice(all_enemies, avoid_list)
	}

	pil.ExecuteSafely(mobile_pilots)

	// Also undock any docked ships...

	for _, pilot := range self.Pilots {
		if pilot.DockedStatus == hal.DOCKED {
			pilot.PlanUndock()
			pilot.ExecutePlan()
		}
	}
}

func (self *Overmind) SetCowardFlag() {

	if self.Game.CurrentPlayers() <= 2 {
		self.CowardFlag = false
		return
	}

	if self.CowardFlag {
		return				// i.e. leave it true
	}

	// So currently CowardFlag is false; should we make it true?

	if self.Game.Turn() < 100 {
		return
	}

	if self.Game.CountMyShips() < self.Game.CountEnemyShips() / 10 {
		self.CowardFlag = true
	}
}
