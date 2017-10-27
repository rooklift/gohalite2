package gohalite2

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ---------------------------------------

type TokenParser struct {
	scanner		*bufio.Scanner
	count		int
}

func NewTokenParser() *TokenParser {
	ret := new(TokenParser)
	ret.scanner = bufio.NewScanner(os.Stdin)
	ret.scanner.Split(bufio.ScanWords)
	return ret
}

func (self *TokenParser) Int() int {
	self.scanner.Scan()
	ret, err := strconv.Atoi(self.scanner.Text())
	if err != nil {
		panic(fmt.Sprintf("TokenReader.Int(): Atoi failed at token %d: \"%s\"", self.count, self.scanner.Text()))
	}
	self.count++
	return ret
}

func (self *TokenParser) Float() float64 {
	self.scanner.Scan()
	ret, err := strconv.ParseFloat(self.scanner.Text(), 64)
	if err != nil {
		panic(fmt.Sprintf("TokenReader.Float(): ParseFloat failed at token %d: \"%s\"", self.count, self.scanner.Text()))
	}
	self.count++
	return ret
}

func (self *TokenParser) Bool() bool {
	self.scanner.Scan()
	s := self.scanner.Text()
	self.count++
	if s == "0" {
		return false
	} else if s == "1" {
		return true
	}

	panic(fmt.Sprintf("TokenReader.Bool(): Value wasn't 0 or 1 (was: \"%s\")", s))
}

// ---------------------------------------

func (self *Game) Parse() {

	self.orders = make(map[int]string)							// Clear all orders.

	if self.inited {
		self.turn++
	}

	// Set all objects to have 0 HP, on the assumption that they are only sent to us if they have
	// 1 or more HP. (Is this true?) Thus this is a default value if we receive no info (they are dead).

	for plid, planet := range self.planetMap {
		planet.HP = 0
		self.planetMap[plid] = planet
	}

	for sid, ship := range self.shipMap {
		ship.HP = 0
		self.shipMap[sid] = ship
	}

	// Clear the dock maps. We will recreate them during parsing.

	self.dockMap = make(map[int][]Ship)

	// Player parsing.............................................................................

	player_count := self.token_parser.Int()

	if self.initialPlayers == 0 {
		self.initialPlayers = player_count						// Only save this at init stage.
	}

	players_with_ships := 0

	for p := 0; p < player_count; p++ {

		pid := self.token_parser.Int()

		ship_count := self.token_parser.Int()

		if ship_count > 0 {
			players_with_ships++
		}

		for s := 0; s < ship_count; s++ {

			sid := self.token_parser.Int()
			ship, ok := self.shipMap[sid]

			if ok == false {
				ship.Id = sid
				ship.Owner = pid
				ship.Birth = Max(1, self.turn)					// If turn is 0 we are in init stage.
			}

			ship.X = self.token_parser.Float()
			ship.Y = self.token_parser.Float()
			ship.HP = self.token_parser.Int()
			self.token_parser.Float()							// Skip deprecated "speedx"
			self.token_parser.Float()							// Skip deprecated "speedy"
			ship.DockedStatus = self.token_parser.Int()
			ship.DockedPlanet = self.token_parser.Int()
			ship.DockingProgress = self.token_parser.Int()
			self.token_parser.Int()								// Skip deprecated "cooldown"

			self.shipMap[sid] = ship
		}
	}

	// Planet parsing.............................................................................

	planet_count := self.token_parser.Int()

	for p := 0; p < planet_count; p++ {

		plid := self.token_parser.Int()
		planet, ok := self.planetMap[plid]

		if ok == false {
			planet.Id = plid
		}

		planet.X = self.token_parser.Float()
		planet.Y = self.token_parser.Float()
		planet.HP = self.token_parser.Int()
		planet.Radius = self.token_parser.Float()
		planet.DockingSpots = self.token_parser.Int()
		planet.CurrentProduction = self.token_parser.Int()
		self.token_parser.Int()									// Skip deprecated "remaining production"
		planet.Owned = self.token_parser.Bool()
		planet.Owner = self.token_parser.Int()

		if planet.Owned == false {
			planet.Owner = -1
		}

		planet.DockedShips = self.token_parser.Int()

		// The dockMap is kept separately so that the Planet struct has no mutable fields.
		// i.e. the Planet struct itself does not get the following data:

		for s := 0; s < planet.DockedShips; s++ {

			// This relies on the fact that we've already been given info about the ships...

			sid := self.token_parser.Int()
			self.dockMap[plid] = append(self.dockMap[plid], self.GetShip(sid))
		}

		self.planetMap[plid] = planet
	}

	self.currentPlayers = players_with_ships
}

// ---------------------------------------

func (self *Game) Thrust(sid, speed, angle int) {
	self.orders[sid] = fmt.Sprintf("t %d %d %d", sid, speed, angle)
}

func (self *Game) Dock(sid, planet int) {
	self.orders[sid] = fmt.Sprintf("d %d %d", sid, planet)
}

func (self *Game) Undock(sid int) {
	self.orders[sid] = fmt.Sprintf("u %d", sid)
}

func (self *Game) ClearOrder(sid int) {
	delete(self.orders, sid)
}

func (self *Game) CurrentOrder(sid int) string {
	return self.orders[sid]
}

func (self *Game) Send() {
	var commands []string
	for _, s := range self.orders {
		commands = append(commands, s)
	}
	out := strings.Join(commands, " ")
	fmt.Printf(out)
	fmt.Printf("\n")
}
