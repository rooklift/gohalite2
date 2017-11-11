package core

import (
	"fmt"
	"time"
)

type MoveInfo struct {
	Dx						float64
	Dy						float64
	Speed					int
	Degrees					int
	DockedStatus			DockedStatus
	Spawned					bool
}

func (self MoveInfo) String() string {
	if self.DockedStatus == DOCKED { return "is docked" }
	if self.DockedStatus == DOCKING { return "is docking" }
	if self.DockedStatus == UNDOCKING { return "is undocking" }
	return fmt.Sprintf("dx: %.2f, dy: %.2f (%d / %d)", self.Dx, self.Dy, self.Speed, self.Degrees)
}

type Game struct {
	inited					bool
	turn					int
	pid						int					// Our own ID
	width					int
	height					int

	initialPlayers			int					// Stored only once at startup. Never changes.
	currentPlayers			int

	planetMap				map[int]Planet		// Planet ID --> Planet
	shipMap					map[int]Ship		// Ship ID --> Ship
	dockMap					map[int][]Ship		// Planet ID --> Ship slice
	lastmoveMap				map[int]MoveInfo	// Ship ID --> MoveInfo struct
	playershipMap			map[int][]Ship		// Player ID --> Ship slice
	cumulativeShips			map[int]int			// Player ID --> Count

	orders					map[int]string

	logfile					*Logfile
	token_parser			*TokenParser
	raw						string

	parse_time				time.Time

	// These slices are kept as answers to common queries...

	all_ships_cache			[]Ship
	enemy_ships_cache		[]Ship
	all_planets_cache		[]Planet
	all_immobile_cache		[]Entity

	enemies_near_planet		map[int][]Ship
}

func NewGame() *Game {
	game := new(Game)
	game.turn = -1
	game.token_parser = NewTokenParser()
	game.pid = game.token_parser.Int()
	game.width = game.token_parser.Int()
	game.height = game.token_parser.Int()
	game.planetMap = make(map[int]Planet)
	game.shipMap = make(map[int]Ship)
	game.dockMap = make(map[int][]Ship)
	game.lastmoveMap = make(map[int]MoveInfo)
	game.cumulativeShips = make(map[int]int)
	game.token_parser.ClearTokens()			// This is just clearing the token_parser's "log".
	game.Parse()
	game.inited = true		// Just means Parse() will increment the turn value before parsing.
	return game
}

func (self *Game) Turn() int { return self.turn }
func (self *Game) Pid() int { return self.pid }
func (self *Game) Width() int { return self.width }
func (self *Game) Height() int { return self.height }
func (self *Game) InitialPlayers() int { return self.initialPlayers }
func (self *Game) CurrentPlayers() int { return self.currentPlayers }
func (self *Game) ParseTime() time.Time { return self.parse_time }

func (self *Game) UpdateProximityMap() {
	const (
		THREAT_RANGE = 10
	)

	self.enemies_near_planet = make(map[int][]Ship)

	all_ships := self.AllShips()
	all_planets := self.AllPlanets()

	for _, ship := range all_ships {
		for _, planet := range all_planets {
			if ship.CanMove() || ship.DockedPlanet == planet.Id {
				if ship.ApproachDist(planet) < THREAT_RANGE {
					if ship.Owner != self.Pid() {
						self.enemies_near_planet[planet.Id] = append(self.enemies_near_planet[planet.Id], ship)
					}
				}
			}
		}
	}
}
