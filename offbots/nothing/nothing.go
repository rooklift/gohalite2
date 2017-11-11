package main

import (
	"fmt"
	hal "../../bot/gohalite2"
)

func main() {
	game := hal.NewGame()
	fmt.Printf("NothingBot\n")
	for {
		game.Parse()
		game.Send()
	}
}
