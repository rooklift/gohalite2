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
	DockedShips						[]int
}

func (self *Planet) GetX() float64 {
	return self.X
}

func (self *Planet) GetY() float64 {
	return self.Y
}

func (self *Planet) GetRadius() float64 {
	return self.Radius
}

func (self *Planet) Alive() bool {
	return self.HP > 0
}

func (self *Planet) CentreDistance(other Entity) float64 {
	dx := self.X - other.GetX()
	dy := self.Y - other.GetY()
	return math.Sqrt(dx * dx + dy * dy)
}

func (self *Planet) SurfaceDistance(other Entity) float64 {
	centre_distance := self.CentreDistance(other)
	return (centre_distance - self.Radius) - other.GetRadius()
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

func (self *Ship) GetX() float64 {
	return self.X
}

func (self *Ship) GetY() float64 {
	return self.Y
}

func (self *Ship) GetRadius() float64 {
	return SHIP_RADIUS
}

func (self *Ship) Alive() bool {
	return self.HP > 0
}

func (self *Ship) CentreDistance(other Entity) float64 {
	dx := self.X - other.GetX()
	dy := self.Y - other.GetY()
	return math.Sqrt(dx * dx + dy * dy)
}

func (self *Ship) SurfaceDistance(other Entity) float64 {
	centre_distance := self.CentreDistance(other)
	return (centre_distance - SHIP_RADIUS) - other.GetRadius()
}

func (self *Ship) CanDock(p *Planet) bool {
	if self.Alive() && p.Alive() && len(p.DockedShips) < p.DockingSpots {
		return self.CentreDistance(p) <= p.Radius + DOCKING_RADIUS
	}
	return false
}
