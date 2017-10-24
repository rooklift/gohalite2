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
		panic("TokenReader.Int(): Atoi failed at token " + strconv.Itoa(self.count))
	}
	self.count++
	return ret
}

func (self *TokenParser) Float() float64 {
	self.scanner.Scan()
	ret, err := strconv.ParseFloat(self.scanner.Text(), 64)
	if err != nil {
		panic("TokenReader.Float(): ParseFloat failed at token " + strconv.Itoa(self.count))
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

	if self.Inited {
		self.Turn++
	}

	// Set all objects to have 0 HP, on the assumption that they are only sent to us if they have
	// 1 or more HP. Thus this is a default value if we receive no info.

	for _, planet := range(self.PlanetMap) {
		planet.HP = 0
	}

	for _, ship := range(self.ShipMap) {
		ship.HP = 0
		ship.ClearOrder()
	}

	// Player parsing.............................................................................

	self.PlayerCount = self.token_parser.Int()

	for p := 0; p < self.PlayerCount; p++ {

		// Get or create the player in memory...

		pid := self.token_parser.Int()
		player, ok := self.PlayerMap[pid]

		if ok == false {
			player = new(Player)
			player.Id = pid
			self.PlayerMap[pid] = player
		}

		ship_count := self.token_parser.Int()

		player.Ships = nil		// Clear the player's ship slice, it will be recreated now...

		for s := 0; s < ship_count; s++ {

			// Get or create the ship in memory...

			sid := self.token_parser.Int()
			ship, ok := self.ShipMap[sid]

			if ok == false {
				ship = new(Ship)
				ship.Id = sid
				ship.Birth = max(1, self.Turn)					// If turn is 0 we are in init stage.
				self.ShipMap[sid] = ship
			}

			ship.X = self.token_parser.Float()
			ship.Y = self.token_parser.Float()
			ship.HP = self.token_parser.Int()
			self.token_parser.Float()							// Skip deprecated "speedx"
			self.token_parser.Float()							// Skip deprecated "speedy"
			ship.Docked = self.token_parser.Int()
			ship.DockedPlanet = self.token_parser.Int()
			ship.DockingProgress = self.token_parser.Int()
			self.token_parser.Int()								// Skip deprecated "cooldown"

			player.Ships = append(player.Ships, ship)
		}
	}

	// Planet parsing.............................................................................

	planet_count := self.token_parser.Int()

	for p := 0; p < planet_count; p++ {

		plid := self.token_parser.Int()

		planet, ok := self.PlanetMap[plid]
		if ok == false {
			planet = new(Planet)
			planet.Id = plid
			self.PlanetMap[plid] = planet
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

		// Docked ships are given to us as their IDs...

		planet.DockedShips = nil	// Clear the planet's ship slice, it will be recreated now...

		docked_ship_count := self.token_parser.Int()

		for s := 0; s < docked_ship_count; s++ {
			sid := self.token_parser.Int()
			planet.DockedShips = append(planet.DockedShips, self.ShipMap[sid])
		}
	}
}

// ---------------------------------------

func (self *Game) Send() {
	me := self.GetMe()

	var commands []string

	for _, ship := range(me.Ships) {
		if ship.Order != "" {
			commands = append(commands, ship.Order)
		}
	}
	fmt.Printf(strings.Join(commands, " "))
	fmt.Printf("\n")
}
