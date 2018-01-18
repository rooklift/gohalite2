package ai

import (
	"time"
)

type Game struct {
	inited					bool
	turn					int
	pid						int					// Our own ID
	width					int
	height					int

	initialPlayers			int					// Stored only once at startup. Never changes.
	currentPlayers			int

	planetMap				map[int]*Planet		// Planet ID --> Planet
	shipMap					map[int]*Ship		// Ship ID --> Ship
	dockMap					map[int][]*Ship		// Planet ID --> Ship slice
	playershipMap			map[int][]*Ship		// Player ID --> Ship slice

	orders					map[int]string

	logfile					*Logfile
	token_parser			*TokenParser
	raw						string				// The raw input line sent by halite.exe

	parse_time				time.Time

	// These slices are kept as answers to common queries...

	all_ships_cache			[]*Ship
	enemy_ships_cache		[]*Ship
	all_planets_cache		[]*Planet

	// This is needed by the AI...

	enemies_near_planet		map[int][]*Ship
}

func NewGame() *Game {
	game := new(Game)
	game.turn = -1
	game.token_parser = NewTokenParser()
	game.pid = game.token_parser.Int()
	game.width = game.token_parser.Int()
	game.height = game.token_parser.Int()
	game.token_parser.ClearTokens()				// This is just clearing the token_parser's "log".
	game.Parse()
	game.inited = true							// Just means Parse() will increment the turn value before parsing.
	return game
}

func (self *Game) Turn() int { return self.turn }
func (self *Game) Pid() int { return self.pid }
func (self *Game) Width() int { return self.width }
func (self *Game) Height() int { return self.height }
func (self *Game) InitialPlayers() int { return self.initialPlayers }
func (self *Game) CurrentPlayers() int { return self.currentPlayers }
func (self *Game) ParseTime() time.Time { return self.parse_time }

func (self *Game) UpdateEnemyMaps() {

	self.enemies_near_planet = make(map[int][]*Ship)

	all_ships := self.AllShips()
	all_planets := self.AllPlanets()

	for _, ship := range all_ships {

		if ship.Owner != self.Pid() {

			for _, planet := range all_planets {

				if ship.ApproachDist(planet) < 20 {

					// enemies_near_planet includes all mobile enemies, plus enemies docked at the planet...

					if ship.CanMove() || ship.DockedPlanet == planet.Id {
						self.enemies_near_planet[planet.Id] = append(self.enemies_near_planet[planet.Id], ship)
					}
				}
			}
		}
	}
}
