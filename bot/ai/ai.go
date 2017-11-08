package ai

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	hal "../gohalite2"
)

const (
	NAME = "Fohristiwhirl"
	VERSION = "25 dev"			// (winner hash) ?? -- log ??
)

func Run() {

	game := hal.NewGame()

	defer func() {
		if p := recover(); p != nil {
			game.Log("Quitting: %v", p)
		}
	}()

	game.StartLog(fmt.Sprintf("log%d.txt", game.Pid()))
	game.LogWithoutTurn("--------------------------------------------------------------------------------")
	game.LogWithoutTurn("%s %s starting up at %s", NAME, VERSION, time.Now().Format("2006-01-02T15:04:05Z"))

	if StringSliceContains(os.Args, "--timeseed") {
		seed := time.Now().UTC().UnixNano()
		rand.Seed(seed)
		game.LogWithoutTurn("Seeding own RNG: %v", seed)
	}

	if len(os.Args) < 2 {
		fmt.Printf("%s %s\n", NAME, VERSION)
	} else {
		fmt.Printf("%s %s %s\n", NAME, VERSION, strings.Join(os.Args[1:], " "))
	}

	overmind := NewOvermind(game)

	var longest_turn time.Duration

	for {
		start_time := time.Now()

		game.Parse()
		overmind.Step()
		game.Send()

		if time.Now().Sub(start_time) > longest_turn {
			longest_turn = time.Now().Sub(start_time)
		}

		if len(game.MyShips()) < len(game.AllShips()) / 10 {
			game.LogOnce("Defeat immanent!")
			PrintFinalInfo(game, longest_turn)
		}

		if len(game.MyShips()) > (len(game.AllShips()) * 9) / 10 {
			game.LogOnce("Victory immanent!")
			PrintFinalInfo(game, longest_turn)
		}
	}
}

func PrintFinalInfo(game * hal.Game, longest_turn time.Duration) {
	game.LogOnce("Current ships...... %3d, %3d, %3d, %3d",
		len(game.ShipsOwnedBy(0)),
		len(game.ShipsOwnedBy(1)),
		len(game.ShipsOwnedBy(2)),
		len(game.ShipsOwnedBy(3)),
	)
	game.LogOnce("Cumulative ships... %3d, %3d, %3d, %3d",
		game.GetCumulativeShipCount(0),
		game.GetCumulativeShipCount(1),
		game.GetCumulativeShipCount(2),
		game.GetCumulativeShipCount(3),
	)
	game.LogOnce("Longest turn took %v", longest_turn)
	game.LogOnce("Board SHA-1: %v", hal.HashFromString(game.RawWorld()))
}
