package gohalite2

import (
	"fmt"
	"strings"
)

type Game struct {
	Inited				bool
	Turn				int
	Pid					int					// Our own ID
	Width				int
	Height				int

	PlayerCount			int					// Does this change if a player dies?

	// These 3 things below can contain references to dead objects.
	// But the Planet and Player structs themselves do not.

	player_map			map[int]*Player
	planet_map			map[int]*Planet
	ship_map			map[int]*Ship

	logfile             *Logfile
	token_parser		*TokenParser
}

func NewGame() *Game {
	game := new(Game)
	game.token_parser = NewTokenParser()
	game.Pid = game.token_parser.Int()
	game.Width = game.token_parser.Int()
	game.Height = game.token_parser.Int()
	game.player_map = make(map[int]*Player)
	game.planet_map = make(map[int]*Planet)
	game.ship_map = make(map[int]*Ship)
	game.Parse()
	game.Inited = true
	return game
}

func (self *Game) GetMe() *Player {
	return self.player_map[self.Pid]
}

func (self *Game) GetPlanets() []*Planet {
	var ret []*Planet
	for key, _ := range self.planet_map {
		planet := self.planet_map[key]
		if planet.HP > 0 {
			ret = append(ret, planet)
		}
	}
	return ret
}

func (self *Game) Send() {
	me := self.GetMe()

	var commands []string

	for _, ship := range(me.Ships) {
		if ship.Order != "" {
			commands = append(commands, ship.Order)
		}
	}
	fmt.Printf(strings.Join(commands, " "))
	fmt.Printf("\n")
}
