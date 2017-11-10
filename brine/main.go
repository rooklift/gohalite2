package main

import (
	"fmt"
	"time"

	hal "../bot/gohalite2"
)

const (
	NAME = "brine"
	VERSION = "1"
)

func main() {

	game := hal.NewGame()

	defer func() {
		if p := recover(); p != nil {
			game.Log("Quitting: %v", p)
		}
	}()

	game.StartLog(fmt.Sprintf("%s%d.txt", NAME, game.Pid()))
	game.LogWithoutTurn("--------------------------------------------------------------------------------")
	game.LogWithoutTurn("%s %s starting up at %s", NAME, VERSION, time.Now().Format("2006-01-02T15:04:05Z"))

	fmt.Printf("%s %s\n", NAME, VERSION)

	ai := NewAI(game)

	for {
		game.Parse()
		ai.Step()
		game.Send()
	}
}

type AI struct {
	Game			*hal.Game
}

func NewAI(game *hal.Game) *AI {
	ret := new(AI)
	ret.Game = game
	return ret
}

func (self *AI) Step() {}
