package gohalite2

import (
	"sort"
)

type DistSortStruct struct {
	slice	[]Planet
	e		Entity
}

type ByDist DistSortStruct

func (dss ByDist) Len() int {
	return len(dss.slice)
}
func (dss ByDist) Swap(i, j int) {
	dss.slice[i], dss.slice[j] = dss.slice[j], dss.slice[i]
}
func (dss ByDist) Less(i, j int) bool {
	return dss.e.Dist(dss.slice[i]) < dss.e.Dist(dss.slice[j])
}

func (self *Game) AllPlanetsByDistance(e Entity) []Planet {
	dss := DistSortStruct{self.AllPlanets(), e}
	sort.Sort(ByDist(dss))
	return dss.slice
}
