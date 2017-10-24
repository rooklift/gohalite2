package gohalite2

const (
	UNDOCKED int = iota
	DOCKING
	DOCKED
	UNDOCKING
)

type Planet struct {
	Id					int
	X					float64
	Y					float64
	HP					int
	Radius				float64
	DockingSpots		int
	CurrentProduction	int
	Owned				bool
	Owner				int			// Protocol will send 0 if not owned at all, but we "correct" this to -1
	DockedShips			[]int
}

type Ship struct {
	Id					int
	Owner				int
	X					float64
	Y					float64
	HP					int
	DockedStatus		int			// Enum
	DockedPlanet		int
	DockingProgress		int

	Birth				int			// Turn this ship was first seen
}
