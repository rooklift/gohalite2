package main

import (
	"fmt"
	"time"

	hal "./gohalite2"
)

const (
	NAME = "Fohristiwhirl"
	VERSION = 1
)

func main() {

	game := new(hal.Game)
	game.Init("log.txt", true)

	game.Log("--------------------------------------------------------------------------------")
	game.Log("%s %d starting up at %s", NAME, VERSION, time.Now().Format("2006-01-02T15:04:05Z"))

	game.LogState()

	fmt.Printf("%s %d\n", NAME, VERSION)

	for {
		game.Parse()
		game.Send()
	}
}
