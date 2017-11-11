package main

import (
	"fmt"
	hal "../../bot/core"
)

func main() {
	game := hal.NewGame()
	fmt.Printf("NothingBot\n")
	for {
		game.Parse()
		game.Send()
	}
}
