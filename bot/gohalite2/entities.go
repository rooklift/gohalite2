package gohalite2

import (
	"math"
)

// ------------------------------------------------------

type Entity interface {
	GetX()							float64
	GetY()							float64
	GetRadius()						float64
	Alive()							bool
	Dist(other Entity)				float64
	SurfaceDist(other Entity)		float64
	Type()							EntityType
}

func EntitiesDist(a, b Entity) float64 {
	dx := a.GetX() - b.GetX()
	dy := a.GetY() - b.GetY()
	return math.Sqrt(dx * dx + dy * dy)
}

func EntitiesSurfaceDist(a, b Entity) float64 {
	return EntitiesDist(a, b) - a.GetRadius() - b.GetRadius()
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

// ------------------------------------------------------

type Point struct {
	X								float64
	Y								float64
}

// Interface satisfiers....

func (s Ship) GetX() float64 { return s.X }
func (p Point) GetX() float64 { return p.X }
func (p Planet) GetX() float64 { return p.X }

func (s Ship) GetY() float64 { return s.Y }
func (p Point) GetY() float64 { return p.Y }
func (p Planet) GetY() float64 { return p.Y }

func (s Ship) GetRadius() float64 { return SHIP_RADIUS }
func (p Point) GetRadius() float64 { return 0 }
func (p Planet) GetRadius() float64 { return p.Radius }

func (s Ship) Alive() bool { return s.HP > 0 }
func (p Point) Alive() bool { return true }
func (p Planet) Alive() bool { return p.HP > 0 }

func (s Ship) Dist(other Entity) float64 { return EntitiesDist(s, other) }
func (p Point) Dist(other Entity) float64 { return EntitiesDist(p, other) }
func (p Planet) Dist(other Entity) float64 { return EntitiesDist(p, other) }

func (s Ship) SurfaceDist(other Entity) float64 { return EntitiesSurfaceDist(s, other) }
func (p Point) SurfaceDist(other Entity) float64 { return EntitiesSurfaceDist(p, other) }
func (p Planet) SurfaceDist(other Entity) float64 { return EntitiesSurfaceDist(p, other) }

func (s Ship) Type() EntityType { return SHIP }
func (p Point) Type() EntityType { return POINT }
func (p Planet) Type() EntityType { return PLANET }
