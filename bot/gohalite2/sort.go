package gohalite2

import (
	"sort"
)

type PlanetsByDist struct {
	slice	[]Planet
	e		Entity
}

func (pbd PlanetsByDist) Len() int {
	return len(pbd.slice)
}
func (pbd PlanetsByDist) Swap(i, j int) {
	pbd.slice[i], pbd.slice[j] = pbd.slice[j], pbd.slice[i]
}
func (pbd PlanetsByDist) Less(i, j int) bool {
	return pbd.e.Dist(pbd.slice[i]) < pbd.e.Dist(pbd.slice[j])
}

func (self *Game) AllPlanetsByDistance(e Entity) []Planet {
	pbd := PlanetsByDist{
		self.AllPlanets(),
		e,
	}
	sort.Sort(pbd)
	return pbd.slice
}

type PlanetsByY []Planet

func (slice PlanetsByY) Len() int {
	return len(slice)
}
func (slice PlanetsByY) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
func (slice PlanetsByY) Less(i, j int) bool {
	return slice[i].Y < slice[j].Y
}
