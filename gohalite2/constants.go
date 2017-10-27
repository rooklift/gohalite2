package gohalite2

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
	NONE EntityType = iota
	SHIP
	PLANET
	POINT
)
