package ai

import (
	"fmt"
	"time"

	hal "../core"
)

const (
	NAME = "Test"
	VERSION = "1"
)

func Run() {

	game := hal.NewGame()

	game.StartLog(fmt.Sprintf("log%d.txt", game.Pid()))
	game.LogWithoutTurn("--------------------------------------------------------------------------------")
	game.LogWithoutTurn("%s %s starting up at %s", NAME, VERSION, time.Now().Format("2006-01-02T15:04:05Z"))

	fmt.Printf("%s %s\n", NAME, VERSION)

	for {
		game.Parse()
		fantastic_unbeatable_AI(game)
		game.Send()
	}
}

func fantastic_unbeatable_AI(game *hal.Game) {

	for _, ship := range game.MyShips() {

		if ship.Id == 4 {

			switch game.Turn() {

			case 0:

				game.Thrust(ship, 1, 271)

			case 1:

				game.Thrust(ship, 6, 271)

			case 2: fallthrough
			case 3: fallthrough
			case 4:

				game.Thrust(ship, 7, 271)

			case 5:

				planet, _ := game.GetPlanet(0)
				game.Dock(ship, planet)
			}
		}
	}
}
