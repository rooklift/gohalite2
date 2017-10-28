package ai

import (
	"fmt"
	"time"

	hal "../gohalite2"
)

const (
	NAME = "Fohristiwhirl"
	VERSION = "9 dev"
)

func Run() {

	game := hal.NewGame()

	game.StartLog(fmt.Sprintf("log%d.txt", game.Pid()))
	game.LogWithoutTurn("--------------------------------------------------------------------------------")
	game.LogWithoutTurn("%s %s starting up at %s", NAME, VERSION, time.Now().Format("2006-01-02T15:04:05Z"))

	fmt.Printf("%s %s\n", NAME, VERSION)

	overmind := NewOvermind(game)

	game.Parse()
	overmind.FirstTurn()
	game.Send()

	for {
		game.Parse()
		overmind.Step()
		game.Send()
	}
}
