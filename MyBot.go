package main

import (
	hal "./gohalite2"
)

func main() {
	game := new(hal.Game)
	game.Init("log.txt", true)
	game.LogState()
}
