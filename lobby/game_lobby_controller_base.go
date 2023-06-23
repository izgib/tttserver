package lobby

import (
	"github.com/izgib/tttserver/game"
)

type GameConfiguration struct {
	ID       int64
	Settings game.GameSettings
	Mark     game.PlayerMark
}

type GameFilter struct {
	Rows Filter
	Cols Filter
	Win  Filter
	Mark game.MoveChoice
}
