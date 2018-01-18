package ai

const (
	SHIP_RADIUS = 0.5
	DOCKING_RADIUS = 4.0
	MAX_SPEED = 7
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
