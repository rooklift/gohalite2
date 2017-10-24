package gohalite2

const (
	SHIP_RADIUS = 0.5
	DOCKING_RADIUS = 4.0
)

const (						// Enum: Docking status
	UNDOCKED int = iota
	DOCKING
	DOCKED
	UNDOCKING
)

const (						// Enum: Ship target type
	NONE int = iota
	SHIP
	PLANET
)
