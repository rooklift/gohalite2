package ai

import (
	"fmt"
	"time"

	hal "../gohalite2"
)

const (
	NAME = "Fohristiwhirl"
	VERSION = "18 final"
)

func Run() {

	game := hal.NewGame()

	game.StartLog(fmt.Sprintf("log%d.txt", game.Pid()))
	game.LogWithoutTurn("--------------------------------------------------------------------------------")
	game.LogWithoutTurn("%s %s starting up at %s", NAME, VERSION, time.Now().Format("2006-01-02T15:04:05Z"))

	fmt.Printf("%s %s\n", NAME, VERSION)

	overmind := NewOvermind(game)

	var longest_turn time.Duration

	for {
		start_time := time.Now()

		game.Parse()
		overmind.Step()
		game.Send()

		if time.Now().Sub(start_time) > longest_turn {
			longest_turn = time.Now().Sub(start_time)
		}

		if len(game.MyShips()) < len(game.AllShips()) / 10 {
			game.LogOnce("Defeat immanent. Longest turn took %v", longest_turn)
		}

		if len(game.MyShips()) > (len(game.AllShips()) * 9) / 10 {
			game.LogOnce("Victory immanent. Longest turn took %v", longest_turn)
		}
	}
}
