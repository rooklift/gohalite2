package ai

import (
	hal "../gohalite2"
)

type AI struct {
	game	*hal.Game
}

func NewAI(game *hal.Game) *AI {
	ret := new(AI)
	ret.game = game
	return ret
}

func (self *AI) Step() {}
