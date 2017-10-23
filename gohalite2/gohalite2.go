package gohalite2

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ---------------------------------------

var scanner = bufio.NewScanner(os.Stdin)

func get_tokens() []string {
	scanner.Scan()
	return strings.Fields(scanner.Text())
}

// ---------------------------------------

type TokenReader struct {
	tokens				[]string
	next_i				int
}

func NewTokenReader(tokens []string) *TokenReader {
	ret := new(TokenReader)
	ret.tokens = tokens
	ret.next_i = 0
	return ret
}

func (self *TokenReader) NextInt() int {
	ret, err := strconv.Atoi(self.tokens[self.next_i])
	if err != nil {
		panic("TokenReader.NextInt(): Atoi failed at token " + strconv.Itoa(self.next_i))
	}
	self.next_i++
	return ret
}

func (self *TokenReader) NextFloat() float64 {
	ret, err := strconv.ParseFloat(self.tokens[self.next_i], 64)
	if err != nil {
		panic("TokenReader.NextInt(): ParseFloat failed at token " + strconv.Itoa(self.next_i))
	}
	self.next_i++
	return ret
}

// ---------------------------------------

type Ship struct {
	Id					int
	X					float64
	Y					float64
	HP					int
	Speedx				float64
	Speedy				float64
	Docked				int			// Is this really a bool?
	DockedPlanet		int
	DockingProgress		int
	Cooldown			int
	Order				string
}

func (self *Ship) Thrust(speed, angle int) {
	self.Order = fmt.Sprintf("t %d %d %d", self.Id, speed, angle)
}

func (self *Ship) Dock(planet int) {
	self.Order = fmt.Sprintf("d %d %d", self.Id, planet)
}

func (self *Ship) Undock(planet int) {
	self.Order = fmt.Sprintf("u %d", self.Id)
}

type Player struct {
	Pid					int
	Ships				[]*Ship
}

type Planet struct {
	Id					int
	X					float64
	Y					float64
	HP					int
	Radius				float64
	DockingSpots		int
	CurrentProduction	int
	RemainingProduction	int
	Owned				int			// Is this really a bool?
	Owner				int			// Protocol will send 0 if not owned at all, but we "correct" this to -1
	DockedShips			[]int
}

type World struct {
	ShipMap				map[int]*Ship
	Players				[]*Player
	Planets				[]*Planet
	PlayerCount			int			// The initial value provided at startup (does this change if a player dies?)
	PlayerId			int			// Our own ID
	Width				int
	Height				int
}

func (self *World) GetShip(id int) (*Ship, bool) {
	ret, ok := self.ShipMap[id]
	return ret, ok
}

func (self *World) Init() {

	var tokens []string

	tokens = get_tokens()
	self.PlayerId, _ = strconv.Atoi(tokens[0])

	tokens = get_tokens()
	self.Width, _ = strconv.Atoi(tokens[0])
	self.Height, _ = strconv.Atoi(tokens[1])

	self.Parse()
}

func (self *World) Parse() {

	tokens := NewTokenReader(get_tokens())

	self.PlayerCount = tokens.NextInt()

	self.ShipMap = make(map[int]*Ship)
	self.Players = nil
	self.Planets = nil

	for p := 0; p < self.PlayerCount; p++ {

		player := new(Player)

		player.Pid = tokens.NextInt()

		ship_count := tokens.NextInt()

		for s := 0; s < ship_count; s++ {

			ship := new(Ship)

			ship.Id = tokens.NextInt()
			ship.X = tokens.NextFloat()
			ship.Y = tokens.NextFloat()
			ship.HP = tokens.NextInt()
			ship.Speedx = tokens.NextFloat()
			ship.Speedy = tokens.NextFloat()
			ship.Docked = tokens.NextInt()
			ship.DockedPlanet = tokens.NextInt()
			ship.DockingProgress = tokens.NextInt()
			ship.Cooldown = tokens.NextInt()

			player.Ships = append(player.Ships, ship)

			self.ShipMap[ship.Id] = ship
		}

		self.Players = append(self.Players, player)
	}

	planet_count := tokens.NextInt()

	for p := 0; p < planet_count; p++ {

		planet := new(Planet)

		planet.Id = tokens.NextInt()
		planet.X = tokens.NextFloat()
		planet.Y = tokens.NextFloat()
		planet.HP = tokens.NextInt()
		planet.Radius = tokens.NextFloat()
		planet.DockingSpots = tokens.NextInt()
		planet.CurrentProduction = tokens.NextInt()
		planet.RemainingProduction = tokens.NextInt()
		planet.Owned = tokens.NextInt()					// This should probably be converted to bool
		planet.Owner = tokens.NextInt()

		if planet.Owned == 0 {
			planet.Owner = -1
		}

		docked_ship_count := tokens.NextInt()

		// Docked ships are given to us as their IDs...

		for s := 0; s < docked_ship_count; s++ {
			planet.DockedShips = append(planet.DockedShips, tokens.NextInt())
		}

		self.Planets = append(self.Planets, planet)
	}
}
