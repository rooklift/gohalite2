package main

import (
	"fmt"
	hal "./gohalite2"
)

const (
	VERSION = "Fohristiwhirl 0.0.1"
)

func main() {
	game := new(hal.Game)
	game.Init("log.txt", true)
	game.LogStartup(VERSION)
	game.LogState()

	fmt.Printf(VERSION + "\n")

	for {
		game.Parse()
		game.Send()
	}
}
