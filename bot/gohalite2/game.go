package gohalite2

type Game struct {
	inited				bool
	turn				int
	pid					int					// Our own ID
	width				int
	height				int

	initialPlayers		int					// Stored only once at startup. Never changes.
	currentPlayers		int

	planetMap			map[int]Planet		// Planet ID --> Planet			(can contain dead objects)
	shipMap				map[int]Ship		// Ship ID --> Ship				(can contain dead objects)
	dockMap				map[int][]Ship		// Planet ID --> Ship slice

	orders				map[int]string

	logfile             *Logfile
	token_parser		*TokenParser
}

var backend_dev_log = NewLog("backend_dev_log.txt")

func NewGame() *Game {
	game := new(Game)
	game.token_parser = NewTokenParser()
	game.pid = game.token_parser.Int()
	game.width = game.token_parser.Int()
	game.height = game.token_parser.Int()
	game.planetMap = make(map[int]Planet)
	game.shipMap = make(map[int]Ship)
	game.dockMap = make(map[int][]Ship)
	game.Parse()
	game.inited = true
	return game
}

func (self *Game) Turn() int { return self.turn }
func (self *Game) Pid() int { return self.pid }
func (self *Game) Width() int { return self.width }
func (self *Game) Height() int { return self.height }
func (self *Game) InitialPlayers() int { return self.initialPlayers }
func (self *Game) CurrentPlayers() int { return self.currentPlayers }
