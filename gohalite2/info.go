package gohalite2

func (self *Game) AllPlanets() []Planet {
	var ret []Planet
	for plid, _ := range self.planetMap {
		planet := self.GetPlanet(plid)
		if planet.Alive() {
			ret = append(ret, planet)
		}
	}
	return ret
}

func (self *Game) PlanetsOwnedBy(pid int) []Planet {
	var ret []Planet
	for plid, _ := range self.planetMap {
		planet := self.GetPlanet(plid)
		if planet.Alive() && planet.Owned && planet.Owner == pid {
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
	for sid, _ := range self.shipMap {
		ship := self.shipMap[sid]
		if ship.Birth == self.turn && ship.Owner == self.pid {
			ret = append(ret, ship.Id)
		}
	}
	return ret
}

func (self *Game) GetShip(sid int) Ship {
	ship := self.shipMap[sid]
	return ship
}

func (self *Game) GetPlanet(plid int) Planet {
	planet := self.planetMap[plid]
	return planet
}

func (self *Game) ClosestPlanet(x, y float64) Planet {

	point := Point{
		X: x,
		Y: y,
	}

	var best_dist float64 = 9999999
	var ret Planet

	for _, planet := range self.AllPlanets() {
		dist := point.SurfaceDistance(planet)
		if dist < best_dist {
			ret = planet
		}
	}

	return ret
}

func (self *Game) ShipsDockedAt(plid int) []Ship {

	var ret []Ship

	for _, Ship := range self.dockMap[plid] {
		ret = append(ret, Ship)
	}

	return ret
}
