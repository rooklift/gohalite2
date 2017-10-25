package main

import (
	"fmt"
	hal "./gohalite2"
)

func main() {

	game := hal.NewGame()

	fmt.Printf("SuicideBot\n")

	for {
		game.Parse()

		my_ships := game.MyShips()

		for _, ship := range my_ships {
			game.Thrust(ship.Id, 7, 7)
		}

		game.Send()
	}
}

