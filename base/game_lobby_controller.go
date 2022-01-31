package base

import (
	"github.com/izgib/tttserver/game"
)

type GameLobbyController interface {
	ListLobbies(filter GameFilter) ([]GameConfiguration, error)
	CreateLobby(config GameConfiguration) (GameLobby, error)
	JoinLobby(ID int16) (GameLobby, error)
	DeleteLobby(lobby GameLobby)
}

type GameConfiguration struct {
	ID       int16
	Settings game.GameSettings
	Mark     game.PlayerMark
}

type GameFilter struct {
	Rows Filter
	Cols Filter
	Win  Filter
	Mark game.MoveChoice
}
