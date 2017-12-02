package core

const (
	SHIP_RADIUS = 0.5
	DOCKING_RADIUS = 4.0
	MAX_SPEED = 7
	WEAPON_RANGE = 5.0
)

type DockedStatus int

const (
	UNDOCKED DockedStatus = iota
	DOCKING
	DOCKED
	UNDOCKING
)

type EntityType int

const (
	UNSET EntityType = iota
	SHIP
	PLANET
	POINT
	NOTHING
)

type OrderType int

const (
	NO_ORDER OrderType = iota
	THRUST
	DOCK
	UNDOCK
)

type Edge int

const (
	TOP Edge = iota
	BOTTOM
	LEFT
	RIGHT
)

var BackendDevLog = NewLog("backend_dev_log.txt")
