package game

import (
	"github.com/izgib/tttserver/game/models"
)

type GameLobby interface {
	GetID() int16
	GetSettings() models.GameSettings
	GetCreatorMark() models.PlayerMark
	GetOpponentMark() models.PlayerMark
	GameStartedChan() chan bool
	CreatorReadyChan() chan bool
	OpponentReadyChan() chan bool
	GetGameController() GameController
	IsGameStarted() bool
	GetRecorder() GameRecorder
	OnStart()
}
