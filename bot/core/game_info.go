package core

import (
	"sort"
)

// ----------------------------------------------

func (self *Game) GetShip(sid int) (*Ship, bool) {
	ret, ok := self.shipMap[sid]
	return ret, ok
}

func (self *Game) GetPlanet(plid int) (*Planet, bool) {
	ret, ok := self.planetMap[plid]
	return ret, ok
}

// ----------------------------------------------

func (self *Game) AllShips() []*Ship {
	ret := make([]*Ship, len(self.all_ships_cache))
	copy(ret, self.all_ships_cache)
	return ret
}

func (self *Game) AllPlanets() []*Planet {
	ret := make([]*Planet, len(self.all_planets_cache))
	copy(ret, self.all_planets_cache)
	return ret
}

func (self *Game) AllImmobile() []Entity {						// Returns all planets and all docked ships
	ret := make([]Entity, len(self.all_immobile_cache))
	copy(ret, self.all_immobile_cache)
	return ret
}

// ----------------------------------------------

func (self *Game) ShipsOwnedBy(pid int) []*Ship {
	ret := make([]*Ship, len(self.playershipMap[pid]))
	copy(ret, self.playershipMap[pid])
	return ret
}

func (self *Game) MyShips() []*Ship {
	return self.ShipsOwnedBy(self.pid)
}

func (self *Game) EnemyShips() []*Ship {
	ret := make([]*Ship, len(self.enemy_ships_cache))
	copy(ret, self.enemy_ships_cache)
	return ret
}

func (self *Game) WeHaveDockedShips() bool {
	for _, ship := range self.playershipMap[self.pid] {
		if ship.DockedStatus != UNDOCKED {
			return true
		}
	}
	return false
}

// ----------------------------------------------

func (self *Game) CountMyShips() int {
	return len(self.playershipMap[self.pid])
}

func (self *Game) CountEnemyShips() int {
	return len(self.enemy_ships_cache)
}

func (self *Game) CountPlanets() int {
	return len(self.all_planets_cache)
}

func (self *Game) CountOwnedPlanets() int {
	ret := 0
	for _, planet := range self.all_planets_cache {
		if planet.Owned {
			ret++
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
	sort.Slice(ret, func(a, b int) bool {
		return ret[a] < ret[b]
	})
	return ret
}

func (self *Game) ShipsDockedAt(planet *Planet) []*Ship {
	ret := make ([]*Ship, len(self.dockMap[planet.Id]))
	copy(ret, self.dockMap[planet.Id])
	return ret
}

func (self *Game) ClosestPlanet(e Entity) *Planet {

	var best_dist float64 = 9999999
	var ret *Planet

	for _, planet := range self.AllPlanets() {
		dist := e.ApproachDist(planet)
		if dist < best_dist {
			best_dist = dist
			ret = planet
		}
	}

	return ret
}

func (self *Game) FarthestPlanet(e Entity) *Planet {

	var best_dist float64 = -9999999
	var ret *Planet

	for _, planet := range self.AllPlanets() {
		dist := e.ApproachDist(planet)
		if dist > best_dist {
			best_dist = dist
			ret = planet
		}
	}

	return ret
}

func (self *Game) GetCumulativeShipCount(pid int) int {
	return self.cumulativeShips[pid]
}

func (self *Game) SurvivingPlayerIDs() []int {

	var ret []int

	for pid, ships := range self.playershipMap {
		if len(ships) > 0 {
			ret = append(ret, pid)
		}
	}

	sort.Slice(ret, func(a, b int) bool {
		return ret[a] < ret[b]
	})

	return ret
}

func (self *Game) CentreOfGravity(ships []*Ship) *Point {
	if len(ships) == 0 {
		return &Point{float64(self.width) / 2, float64(self.height) / 2}
	}
	avg_x := 0.0
	avg_y := 0.0
	for _, ship := range ships {
		avg_x += ship.X
		avg_y += ship.Y
	}
	avg_x /= float64(len(ships))
	avg_y /= float64(len(ships))
	return &Point{avg_x, avg_y}
}

func (self *Game) AllShipsCentreOfGravity() *Point {
	return self.CentreOfGravity(self.AllShips())
}

func (self *Game) PartialCentreOfGravity(player_ids ...int) *Point {
	var ships []*Ship
	for _, n := range player_ids {
		ships = append(ships, self.ShipsOwnedBy(n)...)
	}
	return self.CentreOfGravity(ships)
}

func (self *Game) MyShipsCentreOfGravity() *Point {
	return self.PartialCentreOfGravity(self.pid)
}

func (self *Game) DesiredSpots(planet *Planet) int {

	// If we don't own it, we want all its spots...

	if planet.Owner != self.pid {
		return planet.DockingSpots
	}

	// If we do, we want to fill it up...

	return planet.OpenSpots()
}

func (self *Game) EnemiesNearPlanet(planet *Planet) []*Ship {

	// See game.go for exact behaviour. Might not do what you'd expect.

	ret := make([]*Ship, len(self.enemies_near_planet[planet.Id]))
	copy(ret, self.enemies_near_planet[planet.Id])
	return ret
}

func (self *Game) MobileEnemiesNearPlanet(planet *Planet) []*Ship {
	ret := make([]*Ship, len(self.mobile_enemies_near_planet[planet.Id]))
	copy(ret, self.mobile_enemies_near_planet[planet.Id])
	return ret
}

func (self *Game) FriendsNearPlanet(planet *Planet) []*Ship {
	ret := make([]*Ship, len(self.friends_near_planet[planet.Id]))
	copy(ret, self.friends_near_planet[planet.Id])
	return ret
}

func (self *Game) LastOwner(planet *Planet) int {
	val, ok := self.lastownerMap[planet.Id]
	if ok == false {
		return -1
	}
	return val
}

func (self *Game) InBounds(x, y float64) bool {
	if x <= 0 { return false }
	if y <= 0 { return false }
	if x >= float64(self.width) { return false }
	if y >= float64(self.height) { return false }
	return true
}

func (self *Game) CourseStaysInBounds(ship *Ship, speed int, degrees int) bool {
	x2, y2 := Projection(ship.X, ship.Y, float64(speed), degrees)
	return self.InBounds(x2, y2)
}

func (self *Game) NearestEdge(ship *Ship) (e Edge, dist float64, point *Point) {

	e = LEFT
	dist = ship.X
	point = &Point{0, ship.Y}

	if float64(self.width) - ship.X < dist {
		e = RIGHT
		dist = float64(self.width) - ship.X
		point = &Point{float64(self.width), ship.Y}
	}

	if ship.Y < dist {
		e = TOP
		dist = ship.Y
		point = &Point{ship.X, 0}
	}

	if float64(self.height) - ship.Y < dist {
		e = BOTTOM
		dist = float64(self.height) - ship.Y
		point = &Point{ship.X, float64(self.height)}
	}

	return e, dist, point
}
