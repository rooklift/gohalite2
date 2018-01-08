package core

// Fohristiwhirl, November 2017.
// A sort of alternative starter kit.

const (
	SHIP_RADIUS = 0.5
	DOCKING_RADIUS = 4.0
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
