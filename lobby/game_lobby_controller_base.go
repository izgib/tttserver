package lobby

import (
	"github.com/izgib/tttserver/game"
)

type GameConfiguration struct {
	ID       uint32
	Settings game.GameSettings
	Mark     game.PlayerMark
}

type GameFilter struct {
	Rows Filter
	Cols Filter
	Win  Filter
	Mark game.MoveChoice
}
