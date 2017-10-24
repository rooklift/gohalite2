package gohalite2

func (self *Game) AllPlanets() []Planet {
	var ret []Planet
	for key, _ := range self.planetMap {
		planet := *self.planetMap[key]
		if planet.HP > 0 {
			ret = append(ret, planet)
		}
	}
	return ret
}

func (self *Game) PlanetsOwnedBy(pid int) []Planet {
	var ret []Planet
	for key, _ := range self.planetMap {
		planet := *self.planetMap[key]
		if planet.HP > 0 && planet.Owned && planet.Owner == pid {
			ret = append(ret, planet)
		}
	}
	return ret
}

func (self *Game) MyPlanets() []Planet {
	return self.PlanetsOwnedBy(self.pid)
}
