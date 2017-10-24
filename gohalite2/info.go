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

func (self *Game) MyNewShipIDs() []int {
	var ret []int
	for key, _ := range self.shipMap {
		ship := *self.shipMap[key]
		if ship.Birth == self.turn && ship.Owner == self.pid {
			ret = append(ret, ship.Id)
		}
	}
	return ret
}

func (self *Game) GetShip(id int) Ship {
	ship := *self.shipMap[id]
	return ship
}
