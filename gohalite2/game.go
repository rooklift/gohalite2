package gohalite2

type Game struct {
	inited				bool
	turn				int
	pid					int					// Our own ID
	width				int
	height				int

	players				int					// Stored only once at startup. Never changes.

	// These 3 things below can contain references to dead objects.
	// But the Planet and Player structs themselves do not.

	planetMap			map[int]*Planet
	shipMap				map[int]*Ship

	orders				map[int]string

	logfile             *Logfile
	token_parser		*TokenParser
}

func NewGame() *Game {
	game := new(Game)
	game.token_parser = NewTokenParser()
	game.pid = game.token_parser.Int()
	game.width = game.token_parser.Int()
	game.height = game.token_parser.Int()
	game.planetMap = make(map[int]*Planet)
	game.shipMap = make(map[int]*Ship)
	game.Parse()
	game.inited = true
	return game
}

func (self *Game) Pid() int {
	return self.pid
}
