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
	MSG_DOCK_TARGET = 174
	MSG_RECURSION = 175
	MSG_EXECUTED_NO_PLAN = 176
	MSG_SECRET_SAUCE = 177
	MSG_POINT_TARGET = 178
	MSG_DOCK_APPROACH = 179
	MSG_NO_TARGET = 180
)

func (self *Pilot) SetMessageFromTarget() {

	switch self.Target.Type() {

	case hal.NOTHING:
		self.Message = MSG_NO_TARGET

	case hal.PLANET:
		self.Message = self.Target.(hal.Planet).Id

	case hal.PORT:
		self.Message = MSG_DOCK_TARGET

	case hal.POINT:
		self.Message = MSG_POINT_TARGET

	case hal.SHIP:
		self.Message = MSG_ASSASSINATE
	}
}
