package gohalite2

import (
	"math"
)

type Entity interface {
	GetX()							float64
	GetY()							float64
	GetRadius()						float64
	Alive()							bool
	CentreDistance(other Entity)	float64
	SurfaceDistance(other Entity)	float64
}

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

func (p Planet) GetX() float64 {
	return p.X
}

func (p Planet) GetY() float64 {
	return p.Y
}

func (p Planet) GetRadius() float64 {
	return p.Radius
}

func (p Planet) Alive() bool {
	return p.HP > 0
}

func (p Planet) CentreDistance(other Entity) float64 {
	dx := p.X - other.GetX()
	dy := p.Y - other.GetY()
	return math.Sqrt(dx * dx + dy * dy)
}

func (p Planet) SurfaceDistance(other Entity) float64 {
	centre_distance := p.CentreDistance(other)
	return (centre_distance - p.Radius) - other.GetRadius()
}

func (p Planet) IsFull() bool {
	return p.DockedShips >= p.DockingSpots
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

func (s Ship) GetX() float64 {
	return s.X
}

func (s Ship) GetY() float64 {
	return s.Y
}

func (s Ship) GetRadius() float64 {
	return SHIP_RADIUS
}

func (s Ship) Alive() bool {
	return s.HP > 0
}

func (s Ship) CentreDistance(other Entity) float64 {
	dx := s.X - other.GetX()
	dy := s.Y - other.GetY()
	return math.Sqrt(dx * dx + dy * dy)
}

func (s Ship) SurfaceDistance(other Entity) float64 {
	centre_distance := s.CentreDistance(other)
	return (centre_distance - SHIP_RADIUS) - other.GetRadius()
}

func (s Ship) CanDock(p Planet) bool {
	if s.Alive() && p.Alive() && p.IsFull() == false {
		return s.CentreDistance(p) <= p.Radius + DOCKING_RADIUS
	}
	return false
}

type Point struct {
	X								float64
	Y								float64
}

func (p Point) GetX() float64 {
	return p.X
}

func (p Point) GetY() float64 {
	return p.Y
}

func (p Point) GetRadius() float64 {
	return 0
}

func (p Point) Alive() bool {
	return true
}

func (p Point) CentreDistance(other Entity) float64 {
	dx := p.X - other.GetX()
	dy := p.Y - other.GetY()
	return math.Sqrt(dx * dx + dy * dy)
}

func (p Point) SurfaceDistance(other Entity) float64 {
	centre_distance := p.CentreDistance(other)
	return centre_distance - other.GetRadius()
}
