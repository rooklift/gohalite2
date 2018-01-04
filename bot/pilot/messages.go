package pilot

import (
	hal "../core"
)

const (
	MSG_ATTACK_DOCKED = 121
	MSG_ORBIT_FIGHT = 122
	MSG_ASSASSINATE = 123

	MSG_ATC_DEACTIVATED = 150
	MSG_ATC_RESTRICT = 151
	MSG_ATC_SLOWED = 152

	MSG_COWARD = 160

	MSG_PLANET_LOCKED = 165
	MSG_SHIP_LOCKED = 166
	MSG_POINT_LOCKED = 167
	MSG_PORT_LOCKED = 168

	MSG_SHIP_LOCKED_FEARLESS = 170

	MSG_DOCK_TARGET = 174

	MSG_RECURSION = 175
	MSG_EXECUTED_NO_PLAN = 176

	MSG_SECRET_SAUCE = 177

	MSG_POINT_TARGET = 178
	MSG_DOCK_APPROACH = 179
	MSG_NO_TARGET = 180
)

func (self *Pilot) MessageWhileLocked() {

	switch self.Target.Type() {

		case hal.PLANET:

			self.Message = MSG_PLANET_LOCKED

		case hal.SHIP:

			if self.Fearless {
				self.Message = MSG_SHIP_LOCKED_FEARLESS
			} else {
				self.Message = MSG_SHIP_LOCKED
			}

		case hal.POINT:

			self.Message = MSG_POINT_LOCKED

		case hal.PORT:

			self.Message = MSG_PORT_LOCKED

		default:

			self.Message = MSG_NO_TARGET

	}
}
