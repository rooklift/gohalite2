package gohalite2

type Game struct {
	PlayerCount			int				// The initial value provided at startup (does this change if a player dies?)
	PlayerId			int				// Our own ID
	Width				int
	Height				int

	Players				[]*Player
	Planets				[]*Planet

	logfile             *Logfile
	token_parser		*TokenParser
	ship_map			map[int]*Ship	// ship id --> ship pointer
}

func (self *Game) GetShip(id int) (*Ship, bool) {
	ret, ok := self.ship_map[id]
	return ret, ok
}

func (self *Game) Init(logfilename string, log_enabled bool) {

	self.token_parser = NewTokenParser()

	self.PlayerId = self.token_parser.Int()
	self.Width = self.token_parser.Int()
	self.Height = self.token_parser.Int()

	self.logfile = NewLog(logfilename, log_enabled)

	self.Parse()
}

func (self *Game) Parse() {

	self.PlayerCount = self.token_parser.Int()

	self.ship_map = make(map[int]*Ship)
	self.Players = nil
	self.Planets = nil

	for p := 0; p < self.PlayerCount; p++ {

		player := new(Player)

		player.Pid = self.token_parser.Int()

		ship_count := self.token_parser.Int()

		for s := 0; s < ship_count; s++ {

			ship := new(Ship)

			ship.Id = self.token_parser.Int()
			ship.X = self.token_parser.Float()
			ship.Y = self.token_parser.Float()
			ship.HP = self.token_parser.Int()
			ship.Speedx = self.token_parser.Float()
			ship.Speedy = self.token_parser.Float()
			ship.Docked = self.token_parser.Int()
			ship.DockedPlanet = self.token_parser.Int()
			ship.DockingProgress = self.token_parser.Int()
			ship.Cooldown = self.token_parser.Int()

			player.Ships = append(player.Ships, ship)

			self.ship_map[ship.Id] = ship
		}

		self.Players = append(self.Players, player)
	}

	planet_count := self.token_parser.Int()

	for p := 0; p < planet_count; p++ {

		planet := new(Planet)

		planet.Id = self.token_parser.Int()
		planet.X = self.token_parser.Float()
		planet.Y = self.token_parser.Float()
		planet.HP = self.token_parser.Int()
		planet.Radius = self.token_parser.Float()
		planet.DockingSpots = self.token_parser.Int()
		planet.CurrentProduction = self.token_parser.Int()
		planet.RemainingProduction = self.token_parser.Int()
		planet.Owned = self.token_parser.Int()					// This should probably be converted to bool
		planet.Owner = self.token_parser.Int()

		if planet.Owned == 0 {
			planet.Owner = -1
		}

		docked_ship_count := self.token_parser.Int()

		// Docked ships are given to us as their IDs...

		for s := 0; s < docked_ship_count; s++ {
			planet.DockedShips = append(planet.DockedShips, self.token_parser.Int())
		}

		self.Planets = append(self.Planets, planet)
	}
}
