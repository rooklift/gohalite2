package gohalite2

type Game struct {
	Inited				bool
	Turn				int
	Pid					int					// Our own ID
	Width				int
	Height				int

	PlayerCount			int					// Does this change if a player dies?

	// These 3 things below can contain references to dead objects.
	// But the Planet and Player structs themselves do not.

	PlayerMap			map[int]*Player
	PlanetMap			map[int]*Planet
	ShipMap			map[int]*Ship

	logfile             *Logfile
	token_parser		*TokenParser
}

func NewGame() *Game {
	game := new(Game)
	game.token_parser = NewTokenParser()
	game.Pid = game.token_parser.Int()
	game.Width = game.token_parser.Int()
	game.Height = game.token_parser.Int()
	game.PlayerMap = make(map[int]*Player)
	game.PlanetMap = make(map[int]*Planet)
	game.ShipMap = make(map[int]*Ship)
	game.Parse()
	game.Inited = true
	return game
}
