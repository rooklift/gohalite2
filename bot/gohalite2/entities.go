package gohalite2

import (
	"fmt"
	"math"
)

// ------------------------------------------------------

type Entity interface {
	Type()							EntityType
	GetId()							int
	GetX()							float64
	GetY()							float64
	GetRadius()						float64
	Dist(other Entity)				float64
	SurfaceDist(other Entity)		float64
	Collides(other Entity)			bool
	Alive()							bool
	String()						string
}

func EntitiesDist(a, b Entity) float64 {
	dx := a.GetX() - b.GetX()
	dy := a.GetY() - b.GetY()
	return math.Sqrt(dx * dx + dy * dy)
}

func EntitiesSurfaceDist(a, b Entity) float64 {
	return EntitiesDist(a, b) - a.GetRadius() - b.GetRadius()
}

func EntitiesCollide(a, b Entity) bool {
	return EntitiesDist(a, b) <= a.GetRadius() + b.GetRadius()
}

// ------------------------------------------------------

type Planet struct {
	Id								int
	X								float64
	Y								float64
	HP								int
	Radius							float64
	DockingSpots					int
	CurrentProduction				int
	Owned							bool
	Owner							int			// Protocol will send 0 if not owned at all, but we "correct" this to -1
	DockedShips						int			// The ships themselves can be accessed via game.dockMap[]
}

func (p Planet) IsFull() bool {
	return p.DockedShips >= p.DockingSpots
}

// ------------------------------------------------------

type Ship struct {
	Id					int
	Owner				int
	X					float64
	Y					float64
	HP					int
	DockedStatus		DockedStatus
	DockedPlanet		int
	DockingProgress		int

	Birth				int			// Turn this ship was first seen
}

func (s Ship) CanDock(p Planet) bool {
	if s.Alive() && p.Alive() && p.IsFull() == false && (p.Owned == false || p.Owner == s.Owner) {
		return s.Dist(p) <= p.Radius + DOCKING_RADIUS
	}
	return false
}

func (s Ship) Projection(distance float64, degrees int) Ship {
	ret := s
	ret.X, ret.Y = Projection(s.X, s.Y, distance, degrees)
	return ret
}

// ------------------------------------------------------

type Point struct {
	X								float64
	Y								float64
}

// ------------------------------------------------------

// Interface satisfiers....

func (e Ship) Type() EntityType { return SHIP }
func (e Point) Type() EntityType { return POINT }
func (e Planet) Type() EntityType { return PLANET }

func (e Ship) GetId() int { return e.Id }
func (e Point) GetId() int { return -1 }
func (e Planet) GetId() int { return e.Id }

func (e Ship) GetX() float64 { return e.X }
func (e Point) GetX() float64 { return e.X }
func (e Planet) GetX() float64 { return e.X }

func (e Ship) GetY() float64 { return e.Y }
func (e Point) GetY() float64 { return e.Y }
func (e Planet) GetY() float64 { return e.Y }

func (e Ship) GetRadius() float64 { return SHIP_RADIUS }
func (e Point) GetRadius() float64 { return 0 }
func (e Planet) GetRadius() float64 { return e.Radius }

func (e Ship) Dist(other Entity) float64 { return EntitiesDist(e, other) }
func (e Point) Dist(other Entity) float64 { return EntitiesDist(e, other) }
func (e Planet) Dist(other Entity) float64 { return EntitiesDist(e, other) }

func (e Ship) SurfaceDist(other Entity) float64 { return EntitiesSurfaceDist(e, other) }
func (e Point) SurfaceDist(other Entity) float64 { return EntitiesSurfaceDist(e, other) }
func (e Planet) SurfaceDist(other Entity) float64 { return EntitiesSurfaceDist(e, other) }

func (e Ship) Collides(other Entity) bool { return EntitiesCollide(e, other) }
func (e Point) Collides(other Entity) bool { return EntitiesCollide(e, other) }
func (e Planet) Collides(other Entity) bool { return EntitiesCollide(e, other) }

func (e Ship) Alive() bool { return e.HP > 0 }
func (e Point) Alive() bool { return false }
func (e Planet) Alive() bool { return e.HP > 0 }

func (e Ship) String() string { return fmt.Sprintf("Ship %d [%d,%d]", e.Id, int(e.X), int(e.Y)) }
func (e Point) String() string { return fmt.Sprintf("Point [%d,%d]", int(e.X), int(e.Y)) }
func (e Planet) String() string { return fmt.Sprintf("Planet %d [%d,%d]", e.Id, int(e.X), int(e.Y)) }
