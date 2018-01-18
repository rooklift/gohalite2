package ai

import (
	"fmt"
)

// ------------------------------------------------------

type Entity interface {
	Type()							EntityType
	GetId()							int
	GetX()							float64
	GetY()							float64
	GetRadius()						float64
	Angle(other Entity)				int
	Dist(other Entity)				float64
	Alive()							bool
	String()						string
}

func EntitiesDist(a, b Entity) float64 {
	if a.Type() == NOTHING || b.Type() == NOTHING {
		return 0
	}
	return Dist(a.GetX(), a.GetY(), b.GetX(), b.GetY())
}

func EntitiesApproachDist(a, b Entity) float64 {					// Centre to edge dist. Note: a.ApproachDist(b) != b.ApproachDist(a)
	return EntitiesDist(a, b) - b.GetRadius()
}

func EntitiesAngle(a, b Entity) int {
	if a.Type() == NOTHING || b.Type() == NOTHING {
		panic("EntitiesAngle() called with NOTHING entity")
	}
	return Angle(a.GetX(), a.GetY(), b.GetX(), b.GetY())
}

// ------------------------------------------------------

type Planet struct {
	Game							*Game
	Id								int
	X								float64
	Y								float64
	HP								int
	Radius							float64
	DockingSpots					int
	CurrentProduction				int
	Owned							bool
	Owner							int			// Protocol will send 0 if not owned at all, but we "correct" this to -1
	DockedShips						int			// The ships themselves can be accessed via game.GetShip()
}

func (p *Planet) OpenSpots() int {
	return p.DockingSpots - p.DockedShips
}

func (p *Planet) IsFull() bool {
	return p.DockedShips >= p.DockingSpots
}

// ------------------------------------------------------

type Ship struct {
	Game							*Game
	Id								int
	Owner							int
	X								float64
	Y								float64
	HP								int
	DockedStatus					DockedStatus
	DockedPlanet					int
	DockingProgress					int

	Birth							int			// Turn this ship was first seen

	// Fields used by the AI...

	Target							Entity
	Validated						bool
}

func (s *Ship) CanDock(p *Planet) bool {
	if s.Alive() && p.Alive() && p.IsFull() == false && (p.Owned == false || p.Owner == s.Owner) {
		return s.Dist(p) - p.Radius < DOCKING_RADIUS + SHIP_RADIUS
	}
	return false
}

func (s *Ship) CanMove() bool {
	return s.DockedStatus == UNDOCKED
}

func (s *Ship) Thrust(speed, angle int) {
	s.Game.Thrust(s, speed, angle)
}

func (s *Ship) Dock(planet *Planet) {
	s.Game.Dock(s, planet)
}

func (s *Ship) CurrentCourse() (speed, angle int) {
	order := s.Game.CurrentOrder(s)
	return CourseFromString(order)
}

func (s *Ship) CurrentOrder() string {
	return s.Game.CurrentOrder(s)
}

func (s *Ship) SlowDown() {
	speed, angle := s.CurrentCourse()
	if speed > 0 {
		s.Thrust(speed - 1, angle)
	}
}

// ------------------------------------------------------

type Point struct {
	X								float64
	Y								float64
}

type Nothing struct {}

// ------------------------------------------------------

// Interface satisfiers....

func (e *Ship) Type() EntityType { return SHIP }
func (e *Point) Type() EntityType { return POINT }
func (e *Planet) Type() EntityType { return PLANET }
func (e *Nothing) Type() EntityType { return NOTHING }

func (e *Ship) GetId() int { return e.Id }
func (e *Point) GetId() int { return -1 }
func (e *Planet) GetId() int { return e.Id }
func (e *Nothing) GetId() int { return -1 }

func (e *Ship) GetX() float64 { return e.X }
func (e *Point) GetX() float64 { return e.X }
func (e *Planet) GetX() float64 { return e.X }
func (e *Nothing) GetX() float64 { panic("GetX() called on NOTHING entity") }

func (e *Ship) GetY() float64 { return e.Y }
func (e *Point) GetY() float64 { return e.Y }
func (e *Planet) GetY() float64 { return e.Y }
func (e *Nothing) GetY() float64 { panic("GetY() called on NOTHING entity") }

func (e *Ship) GetRadius() float64 { return SHIP_RADIUS }
func (e *Point) GetRadius() float64 { return 0 }
func (e *Planet) GetRadius() float64 { return e.Radius }
func (e *Nothing) GetRadius() float64 { return 0 }

func (e *Ship) Angle(other Entity) int { return EntitiesAngle(e, other) }
func (e *Point) Angle(other Entity) int { return EntitiesAngle(e, other) }
func (e *Planet) Angle(other Entity) int { return EntitiesAngle(e, other) }
func (e *Nothing) Angle(other Entity) int { return EntitiesAngle(e, other) }						// Will panic

func (e *Ship) Dist(other Entity) float64 { return EntitiesDist(e, other) }
func (e *Point) Dist(other Entity) float64 { return EntitiesDist(e, other) }
func (e *Planet) Dist(other Entity) float64 { return EntitiesDist(e, other) }
func (e *Nothing) Dist(other Entity) float64 { return EntitiesDist(e, other) }						// Will panic

func (e *Ship) ApproachDist(other Entity) float64 { return EntitiesApproachDist(e, other) }
func (e *Point) ApproachDist(other Entity) float64 { return EntitiesApproachDist(e, other) }
func (e *Planet) ApproachDist(other Entity) float64 { return EntitiesApproachDist(e, other) }
func (e *Nothing) ApproachDist(other Entity) float64 { return EntitiesApproachDist(e, other) }

func (e *Ship) Alive() bool { return e.HP > 0 }
func (e *Point) Alive() bool { return true }
func (e *Planet) Alive() bool { return e.HP > 0 }
func (e *Nothing) Alive() bool { return false }

func (e *Ship) String() string { return fmt.Sprintf("Ship %d [%d,%d]", e.Id, int(e.X), int(e.Y)) }
func (e *Point) String() string { return fmt.Sprintf("Point [%d,%d]", int(e.X), int(e.Y)) }
func (e *Planet) String() string { return fmt.Sprintf("Planet %d [%d,%d]", e.Id, int(e.X), int(e.Y)) }
func (e *Nothing) String() string { return "null entity" }
