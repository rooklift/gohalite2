package core

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------

type TokenParser struct {
	scanner		*bufio.Scanner
	count		int
	all_tokens	[]string		// This is used for logging only. It is cleared each time it's asked-for.
}

func NewTokenParser() *TokenParser {
	ret := new(TokenParser)
	ret.scanner = bufio.NewScanner(os.Stdin)
	ret.scanner.Split(bufio.ScanWords)
	return ret
}

func (self *TokenParser) Int() int {
	bl := self.scanner.Scan()
	if bl == false {
		err := self.scanner.Err()
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		} else {
			panic(fmt.Sprintf("End of input."))
		}
	}
	self.all_tokens = append(self.all_tokens, self.scanner.Text())
	ret, err := strconv.Atoi(self.scanner.Text())
	if err != nil {
		panic(fmt.Sprintf("TokenReader.Int(): Atoi failed at token %d: \"%s\"", self.count, self.scanner.Text()))
	}
	self.count++
	return ret
}

func (self *TokenParser) DockedStatus() DockedStatus {
	return DockedStatus(self.Int())
}

func (self *TokenParser) Float() float64 {
	bl := self.scanner.Scan()
	if bl == false {
		err := self.scanner.Err()
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		} else {
			panic(fmt.Sprintf("End of input."))
		}
	}
	self.all_tokens = append(self.all_tokens, self.scanner.Text())
	ret, err := strconv.ParseFloat(self.scanner.Text(), 64)
	if err != nil {
		panic(fmt.Sprintf("TokenReader.Float(): ParseFloat failed at token %d: \"%s\"", self.count, self.scanner.Text()))
	}
	self.count++
	return ret
}

func (self *TokenParser) Bool() bool {
	val := self.Int()
	if val != 0 && val != 1 {
		panic(fmt.Sprintf("TokenReader.Bool(): Value wasn't 0 or 1 (was: \"%d\")", val))
	}
	return val == 1
}

func (self *TokenParser) Tokens(sep string) string {
	ret := strings.Join(self.all_tokens, sep)
	self.all_tokens = nil
	return ret
}

func (self *TokenParser) ClearTokens() {
	self.all_tokens = nil
}

// ---------------------------------------

func (self *Game) Parse() {

	// Do our first read before clearing things, so that it panics on EOF and we haven't corrupted our state...

	player_count := self.token_parser.Int()

	// Now reset various things...

	self.orders = make(map[int]string)			// Clear all orders.

	if self.inited {
		self.turn++
	}

	// Clear some info maps. We will recreate them during parsing.

	old_shipmap := self.shipMap					// We need last turn's ship info for inferring movement / birth.

	self.shipMap = make(map[int]Ship)
	self.planetMap = make(map[int]Planet)
	self.dockMap = make(map[int][]Ship)
	self.lastmoveMap = make(map[int]MoveInfo)
	self.playershipMap = make(map[int][]Ship)

	// Player parsing.............................................................................

	self.parse_time = time.Now()				// MUST happen AFTER the first token parse. <------------------------------------- important

	if self.initialPlayers == 0 {
		self.initialPlayers = player_count		// Only save this at init stage.
	}

	players_with_ships := 0

	for p := 0; p < player_count; p++ {

		pid := self.token_parser.Int()

		ship_count := self.token_parser.Int()

		if ship_count > 0 {
			players_with_ships++
		}

		for s := 0; s < ship_count; s++ {

			var ship Ship

			sid := self.token_parser.Int()

			ship.Id = sid
			ship.Owner = pid

			ship.X = self.token_parser.Float()
			ship.Y = self.token_parser.Float()
			ship.HP = self.token_parser.Int()
			self.token_parser.Float()								// Skip deprecated "speedx"
			self.token_parser.Float()								// Skip deprecated "speedy"
			ship.DockedStatus = self.token_parser.DockedStatus()
			ship.DockedPlanet = self.token_parser.Int()

			if ship.DockedStatus == UNDOCKED {
				ship.DockedPlanet = -1
			}

			ship.DockingProgress = self.token_parser.Int()
			self.token_parser.Int()									// Skip deprecated "cooldown"

			last_ship, ok := old_shipmap[sid]

			if ok == false {
				ship.Birth = Max(0, self.turn)						// Turn can be -1 in init stage.
				self.cumulativeShips[pid]++
				self.lastmoveMap[sid] = MoveInfo{Spawned: true}		// All other fields zero.
			} else {
				dx := ship.X - last_ship.X
				dy := ship.Y - last_ship.Y
				self.lastmoveMap[sid] = MoveInfo{
					Dx: dx,
					Dy: dy,
					Speed: Round(math.Sqrt(dx * dx + dy * dy)),
					Degrees: Angle(last_ship.X, last_ship.Y, ship.X, ship.Y),
					DockedStatus: ship.DockedStatus,
					Spawned: false,
				}
			}

			self.shipMap[sid] = ship
			self.playershipMap[pid] = append(self.playershipMap[pid], ship)
		}

		sort.Slice(self.playershipMap[pid], func(a, b int) bool {
			return self.playershipMap[pid][a].Id < self.playershipMap[pid][b].Id
		})
	}

	// Planet parsing.............................................................................

	planet_count := self.token_parser.Int()

	for p := 0; p < planet_count; p++ {

		var planet Planet

		plid := self.token_parser.Int()
		planet.Id = plid

		planet.X = self.token_parser.Float()
		planet.Y = self.token_parser.Float()
		planet.HP = self.token_parser.Int()
		planet.Radius = self.token_parser.Float()
		planet.DockingSpots = self.token_parser.Int()
		planet.CurrentProduction = self.token_parser.Int()
		self.token_parser.Int()										// Skip deprecated "remaining production"
		planet.Owned = self.token_parser.Bool()
		planet.Owner = self.token_parser.Int()

		if planet.Owned == false {
			planet.Owner = -1
		} else {
			self.lastownerMap[planet.Id] = planet.Owner
		}

		planet.DockedShips = self.token_parser.Int()

		// The dockMap is kept separately so that the Planet struct has no mutable fields.
		// i.e. the Planet struct itself does not get the following data:

		for s := 0; s < planet.DockedShips; s++ {

			// This relies on the fact that we've already been given info about the ships...

			sid := self.token_parser.Int()
			ship, ok := self.GetShip(sid)
			if ok == false {
				panic("Parser choked on GetShip(sid)")
			}
			self.dockMap[plid] = append(self.dockMap[plid], ship)
		}
		sort.Slice(self.dockMap[plid], func(a, b int) bool {
			return self.dockMap[plid][a].Id < self.dockMap[plid][b].Id
		})

		self.planetMap[plid] = planet
	}

	// Query responses (see info.go)... while these could be done interleaved with the above, they are separated for clarity.

	self.all_ships_cache = nil
	for _, ship := range self.shipMap {
		self.all_ships_cache = append(self.all_ships_cache, ship)
	}
	sort.Slice(self.all_ships_cache, func(a, b int) bool {
		return self.all_ships_cache[a].Id < self.all_ships_cache[b].Id
	})

	self.enemy_ships_cache = nil
	for _, ship := range self.shipMap {
		if ship.Owner != self.pid {
			self.enemy_ships_cache = append(self.enemy_ships_cache, ship)
		}
	}
	sort.Slice(self.enemy_ships_cache, func(a, b int) bool {
		return self.enemy_ships_cache[a].Id < self.enemy_ships_cache[b].Id
	})

	self.all_planets_cache = nil
	for _, planet := range self.planetMap {
		self.all_planets_cache = append(self.all_planets_cache, planet)
	}
	sort.Slice(self.all_planets_cache, func(a, b int) bool {
		return self.all_planets_cache[a].Id < self.all_planets_cache[b].Id
	})

	self.all_immobile_cache = nil
	for _, planet := range self.planetMap {
		self.all_immobile_cache = append(self.all_immobile_cache, planet)
		for _, ship := range self.ShipsDockedAt(planet) {
			self.all_immobile_cache = append(self.all_immobile_cache, ship)
		}
	}
	sort.Slice(self.all_immobile_cache, func(a, b int) bool {
		if self.all_immobile_cache[a].Type() == PLANET && self.all_immobile_cache[b].Type() == SHIP {
			return true
		}
		if self.all_immobile_cache[a].Type() == SHIP && self.all_immobile_cache[b].Type() == PLANET {
			return false
		}
		return self.all_immobile_cache[a].GetId() < self.all_immobile_cache[b].GetId()
	})

	// Some meta info...

	self.currentPlayers = players_with_ships
	self.raw = self.token_parser.Tokens(" ")
	self.UpdateEnemyMaps()
	self.UpdateFriendMap()
}

// ---------------------------------------

func (self *Game) Thrust(ship Ship, speed, degrees int) {
	for degrees < 0 { degrees += 360 }; degrees %= 360
	self.orders[ship.Id] = fmt.Sprintf("t %d %d %d", ship.Id, speed, degrees)
}

func (self *Game) ThrustWithMessage(ship Ship, speed, degrees int, message int) {
	for degrees < 0 { degrees += 360 }; degrees %= 360
	if message >= 0 && message <= 180 { degrees += (int(message) + 1) * 360 }
	self.orders[ship.Id] = fmt.Sprintf("t %d %d %d", ship.Id, speed, degrees)
}

func (self *Game) Dock(ship Ship, planet Planet) {
	self.orders[ship.Id] = fmt.Sprintf("d %d %d", ship.Id, planet.Id)
}

func (self *Game) Undock(ship Ship) {
	self.orders[ship.Id] = fmt.Sprintf("u %d", ship.Id)
}

func (self *Game) ClearOrder(ship Ship) {
	delete(self.orders, ship.Id)
}

func (self *Game) CurrentOrder(ship Ship) string {
	return self.orders[ship.Id]
}

func (self *Game) RawOrder(sid int, s string) {
	self.orders[sid] = s
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
