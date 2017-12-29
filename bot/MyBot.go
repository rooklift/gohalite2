package ai

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	ai "./ai"
	hal "./core"
)

const (
	NAME = "Fohristiwhirl"
	VERSION = "49 dev"
)

func main() {

	config := new(ai.Config)
	flag.BoolVar(&config.Conservative, "conservative", false, "no rushing")
	flag.BoolVar(&config.Timeseed, "timeseed", false, "seed RNG with time")
	flag.Parse()

	game := hal.NewGame()

	var longest_turn time.Duration

	defer func() {
		if p := recover(); p != nil {
			game.Log("Quitting: %v", p)
			game.Log("Last known hash: %s", hal.HashFromString(game.RawWorld()))
			game.LogOnce("Current ships...... %3d, %3d, %3d, %3d",
				len(game.ShipsOwnedBy(0)),
				len(game.ShipsOwnedBy(1)),
				len(game.ShipsOwnedBy(2)),
				len(game.ShipsOwnedBy(3)))
			game.Log("Cumulative ships... %3d, %3d, %3d, %3d",
				game.GetCumulativeShipCount(0),
				game.GetCumulativeShipCount(1),
				game.GetCumulativeShipCount(2),
				game.GetCumulativeShipCount(3))
			game.Log("Longest turn took %v", longest_turn)
		}
	}()

	game.StartLog(fmt.Sprintf("log%d.txt", game.Pid()))
	game.LogWithoutTurn("--------------------------------------------------------------------------------")
	game.LogWithoutTurn("%s %s starting up at %s", NAME, VERSION, time.Now().Format("2006-01-02T15:04:05Z"))

	if config.Timeseed {
		seed := time.Now().UTC().UnixNano()
		rand.Seed(seed)
		game.LogWithoutTurn("Seeding own RNG: %v", seed)
	}

	if len(os.Args) < 2 {
		fmt.Printf("%s %s\n", NAME, VERSION)
	} else {
		fmt.Printf("%s %s %s\n", NAME, VERSION, strings.Join(os.Args[1:], " "))
	}

	overmind := ai.NewOvermind(game, config)

	for {
		start_time := time.Now()

		game.Parse()
		overmind.Step()
		game.Send()

		if time.Now().Sub(start_time) > longest_turn {
			longest_turn = time.Now().Sub(start_time)
		}
	}
}
