package gohalite2

func (self *Game) GetMe() *Player {
	return self.PlayerMap[self.Pid]
}

func (self *Game) GetPlanets() []*Planet {
	var ret []*Planet
	for key, _ := range self.PlanetMap {
		planet := self.PlanetMap[key]
		if planet.HP > 0 {
			ret = append(ret, planet)
		}
	}
	return ret
}
