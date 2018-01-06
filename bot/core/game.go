package core

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	INITIAL_THREAT_RANGE = 10
	INITIAL_FRIEND_RANGE = 30
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
	inited						bool
	turn						int
	pid							int					// Our own ID
	width						int
	height						int

	initialPlayers				int					// Stored only once at startup. Never changes.
	currentPlayers				int

	planetMap					map[int]*Planet		// Planet ID --> Planet
	dockMap						map[int][]*Ship		// Planet ID --> Ship slice
	shipMap						map[int]*Ship		// Ship ID --> Ship
	lastmoveMap					map[int]*MoveInfo	// Ship ID --> MoveInfo struct
	playershipMap				map[int][]*Ship		// Player ID --> Ship slice
	cumulativeShips				map[int]int			// Player ID --> Count
	lastownerMap				map[int]int			// Planet ID --> Last owner (check OK for never owned)

	orders						map[int]string
	messages					map[int]int			// For the Chlorine viewer

	logfile						*Logfile
	token_parser				*TokenParser
	raw							string
	run_of_sames				int

	parse_time					time.Time

	// These slices are kept as answers to common queries...

	all_ships_cache				[]*Ship
	enemy_ships_cache			[]*Ship
	all_planets_cache			[]*Planet
	all_immobile_cache			[]Entity			// Planets and docked ships

	// Some more stuff maybe used by the AI...

	enemies_near_planet			map[int][]*Ship
	mobile_enemies_near_planet	map[int][]*Ship
	friends_near_planet			map[int][]*Ship
	threat_range				float64
	friend_range				float64
}

func NewGame() *Game {
	game := new(Game)
	game.turn = -1
	game.token_parser = NewTokenParser()
	game.pid = game.token_parser.Int()
	game.width = game.token_parser.Int()
	game.height = game.token_parser.Int()
	game.planetMap = make(map[int]*Planet)
	game.shipMap = make(map[int]*Ship)
	game.dockMap = make(map[int][]*Ship)
	game.lastmoveMap = make(map[int]*MoveInfo)
	game.cumulativeShips = make(map[int]int)
	game.lastownerMap = make(map[int]int)
	game.threat_range = INITIAL_THREAT_RANGE
	game.friend_range = INITIAL_FRIEND_RANGE
	game.token_parser.ClearTokens()				// This is just clearing the token_parser's "log".
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
func (self *Game) RunOfSames() int { return self.run_of_sames }

func (self *Game) UpdateEnemyMaps() {

	self.enemies_near_planet = make(map[int][]*Ship)
	self.mobile_enemies_near_planet = make(map[int][]*Ship)

	all_ships := self.AllShips()
	all_planets := self.AllPlanets()

	for _, ship := range all_ships {
		if ship.Owner != self.Pid() {
			for _, planet := range all_planets {
				if ship.ApproachDist(planet) < self.threat_range {

					// enemies_near_planet includes all mobile enemies, plus enemies docked at the planet...

					if ship.CanMove() || ship.DockedPlanet == planet.Id {
						self.enemies_near_planet[planet.Id] = append(self.enemies_near_planet[planet.Id], ship)
					}

					// mobile_enemies_near_planet only includes mobile enemies...

					if ship.CanMove() {
						self.mobile_enemies_near_planet[planet.Id] = append(self.mobile_enemies_near_planet[planet.Id], ship)
					}
				}
			}
		}
	}
}

func (self *Game) UpdateFriendMap() {

	self.friends_near_planet = make(map[int][]*Ship)

	my_ships := self.MyShips()
	all_planets := self.AllPlanets()

	for _, ship := range my_ships {
		for _, planet := range all_planets {
			if ship.CanMove() {
				if ship.ApproachDist(planet) < self.friend_range {
					self.friends_near_planet[planet.Id] = append(self.friends_near_planet[planet.Id], ship)
				}
			}
		}
	}
}

func (self *Game) SetThreatRange(d float64) {
	if d != self.threat_range {
		self.threat_range = d
		self.UpdateEnemyMaps()
	}
}

func (self *Game) SetFriendRange(d float64) {
	if d != self.friend_range {
		self.friend_range = d
		self.UpdateFriendMap()
	}
}

func (self *Game) RawWorld() string {
	return self.raw
}

func (self *Game) RawOutput(sorted, no_messages bool) string {

	var commands []string

	for sid, s := range self.orders {

		if s != "" {

			message, ok := self.messages[sid]

			if no_messages == false && s[0] == 't' && ok {

				speed, degrees := CourseFromString(s)

				// We put some extra info into the angle, which we can see in the Chlorine replayer...

				if message >= 0 && message <= 180 {
					degrees += (int(message) + 1) * 360
				}

				commands = append(commands, fmt.Sprintf("t %d %d %d", sid, speed, degrees))

			} else {

				commands = append(commands, s)

			}
		}
	}

	if sorted {
		sort.Slice(commands, func(a, b int) bool {
			return commands[a] < commands[b]
		})
	}

	return strings.Join(commands, " ")
}

func (self *Game) PredictTimeZero() {

	// Weapons fire at Time 0 is almost entirely predictable (unless ships involved dock, which is generally unlikely).
	// Therefore, we set 2 flags on every ship indicating about those events.

	var all_shots = make(map[int][]int)		// Ship ID --> target IDs
	var incoming = make(map[int]int)		// Ship ID --> damage coming in

	all_ships := self.AllShips()

	for i := 0; i < len(all_ships); i++ {

		ship1 := all_ships[i]

		for k := i + 1; k < len(all_ships); k++ {

			ship2 := all_ships[k]

			if ship1.Owner == ship2.Owner {
				continue
			}

			if ship1.Dist(ship2) > WEAPON_RANGE + SHIP_RADIUS * 2 {
				continue
			}

			// They will fire on each other, unless they are docked or dock now...

			if ship1.DockedStatus == UNDOCKED {
				all_shots[ship1.Id] = append(all_shots[ship1.Id], ship2.Id)
			}
			if ship2.DockedStatus == UNDOCKED {
				all_shots[ship2.Id] = append(all_shots[ship2.Id], ship1.Id)
			}
		}
	}

	for _, ship := range all_ships {

		shots := all_shots[ship.Id]

		if len(shots) == 0 {
			continue
		} else {
			ship.Firing = true
		}

		damage := WEAPON_DAMAGE / len(shots)		// Right? A straight up integer truncation?

		for _, target_id := range shots {
			incoming[target_id] += damage
		}
	}

	for _, ship := range all_ships {

		if incoming[ship.Id] >= ship.HP {
			ship.Doomed = true
		}
	}
}
