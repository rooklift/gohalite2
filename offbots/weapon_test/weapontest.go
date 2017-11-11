package main

import (
	"fmt"
	hal "../../bot/core"
)

func main() {
	game := hal.NewGame()
	fmt.Printf("WeaponTestBot\n")
	for _, ship := range game.MyShips() {
		if ship.Id % 3 == 1 {
			do_test(game, ship)
		}
	}
}

func do_test(game *hal.Game, ship hal.Ship) {

	target := game.EnemyShips()[0]

	for {
		game.Parse()
		ship, _ := game.GetShip(ship.Id)
		target, _ := game.GetShip(target.Id)
		speed, degrees, _ := game.GetApproach(ship, target, 5.5, game.AllImmobile(), hal.RIGHT)
		game.Thrust(ship, speed, degrees)
		game.Send()
	}
}
