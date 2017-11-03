package gohalite2

// ----------------------------------------------

func (self *Game) GetShip(sid int) Ship {
	return self.shipMap[sid]
}

func (self *Game) GetPlanet(plid int) Planet {
	return self.planetMap[plid]
}

// ----------------------------------------------

func (self *Game) AllShips() []Ship {
	var ret []Ship
	for sid, _ := range self.shipMap {
		ship := self.GetShip(sid)
		ret = append(ret, ship)
	}
	return ret
}

func (self *Game) AllPlanets() []Planet {
	ret := make([]Planet, len(self.all_planets_cache))
	copy(ret, self.all_planets_cache)
	return ret
}

func (self *Game) AllImmobile() []Entity {						// Returns all planets and all docked ships
	ret := make([]Entity, len(self.all_immobile_cache))
	copy(ret, self.all_immobile_cache)
	return ret
}

// ----------------------------------------------

func (self *Game) PlanetsOwnedBy(pid int) []Planet {
	var ret []Planet
	for plid, _ := range self.planetMap {
		planet := self.GetPlanet(plid)
		if planet.Owned && planet.Owner == pid {
			ret = append(ret, planet)
		}
	}
	return ret
}

func (self *Game) ShipsOwnedBy(pid int) []Ship {
	var ret []Ship
	for sid, _ := range self.shipMap {
		ship := self.GetShip(sid)
		if ship.Owner == pid {
			ret = append(ret, ship)
		}
	}
	return ret
}

func (self *Game) MyPlanets() []Planet {
	return self.PlanetsOwnedBy(self.pid)
}

func (self *Game) MyShips() []Ship {
	return self.ShipsOwnedBy(self.pid)
}

func (self *Game) EnemyShips() []Ship {
	var ret []Ship
	for sid, _ := range self.shipMap {
		ship := self.GetShip(sid)
		if ship.Owner != self.Pid() {
			ret = append(ret, ship)
		}
	}
	return ret
}

// ----------------------------------------------

func (self *Game) MyNewShipIDs() []int {			// My ships born this turn.
	var ret []int
	for sid, _ := range self.shipMap {
		ship := self.shipMap[sid]
		if ship.Birth == self.turn && ship.Owner == self.pid {
			ret = append(ret, ship.Id)
		}
	}
	return ret
}

func (self *Game) ShipsDockedAt(planet Planet) []Ship {

	var ret []Ship

	for _, Ship := range self.dockMap[planet.Id] {
		ret = append(ret, Ship)
	}

	return ret
}

func (self *Game) LastTurnMoveBy(ship Ship) MoveInfo {
	return self.lastmoveMap[ship.Id]
}

func (self *Game) ClosestPlanet(e Entity) Planet {

	var best_dist float64 = 9999999
	var ret Planet

	for _, planet := range self.AllPlanets() {
		dist := e.ApproachDist(planet)
		if dist < best_dist {
			best_dist = dist
			ret = planet
		}
	}

	return ret
}

func (self *Game) RawWorld() string {
	return self.raw
}

func (self *Game) GetCumulativeShipCount(pid int) int {
	return self.cumulativeShips[pid]
}
