package gohalite2

type Planet struct {
	Id					int
	X					float64
	Y					float64
	HP					int
	Radius				float64
	DockingSpots		int
	CurrentProduction	int
	RemainingProduction	int
	Owned				int			// Is this really a bool?
	Owner				int			// Protocol will send 0 if not owned at all, but we "correct" this to -1
	DockedShips			[]int
}
