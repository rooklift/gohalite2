package main

import (
	"fmt"
	hal "../../bot/core"
)

func main() {

	game := hal.NewGame()

	fmt.Printf("SuicideBot\n")

	for {
		game.Parse()

		my_ships := game.MyShips()

		for _, ship := range my_ships {

			if ship.Id % 3 == 1 {
				game.Thrust(ship, 7, 90)
			}

			if ship.Id % 3 == 2 {
				game.Thrust(ship, 7, 270)
			}
		}

		game.Send()
	}
}

