package game

import (
	"github.com/izgib/tttserver/game/models"
)

type GameLobbyUsecase interface {
	ListLobbies(filter GameFilter) ([]GameConfiguration, error)
	CreateLobby(config GameConfiguration) (GameLobby, error)
	JoinLobby(ID int16) (GameLobby, error)
	DeleteLobby(lobby GameLobby)
}

type GameConfiguration struct {
	ID       int16
	Settings models.GameSettings
	Mark     models.PlayerMark
}

type GameFilter struct {
	Rows Filter
	Cols Filter
	Win  Filter
	Mark models.MoveChoice
}
