package gohalite2

import (
	"fmt"
)

type MoveInfo struct {
	Dx					float64
	Dy					float64
	Speed				int
	Degrees				int
	DockedStatus		DockedStatus
	Spawned				bool
}

func (self MoveInfo) String() string {
	if self.DockedStatus == DOCKED { return "is docked" }
	if self.DockedStatus == DOCKING { return "is docking" }
	if self.DockedStatus == UNDOCKING { return "is undocking" }
	return fmt.Sprintf("dx: %.2f, dy: %.2f (%d / %d)", self.Dx, self.Dy, self.Speed, self.Degrees)
}

type Game struct {
	inited				bool
	turn				int
	pid					int					// Our own ID
	width				int
	height				int

	initialPlayers		int					// Stored only once at startup. Never changes.
	currentPlayers		int

	planetMap			map[int]Planet		// Planet ID --> Planet
	shipMap				map[int]Ship		// Ship ID --> Ship
	dockMap				map[int][]Ship		// Planet ID --> Ship slice
	lastmoveMap			map[int]MoveInfo	// Ship ID --> MoveInfo struct

	orders				map[int]string

	logfile             *Logfile
	token_parser		*TokenParser
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
	game.Parse()
	game.inited = true		// Just means the parser will increment the turn value before parsing.
	return game
}

func (self *Game) Turn() int { return self.turn }
func (self *Game) Pid() int { return self.pid }
func (self *Game) Width() int { return self.width }
func (self *Game) Height() int { return self.height }
func (self *Game) InitialPlayers() int { return self.initialPlayers }
func (self *Game) CurrentPlayers() int { return self.currentPlayers }
