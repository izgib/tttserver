package base

import (
	"github.com/izgib/tttserver/game"
)

type GameLobbyID interface {
	GetID() int16
}

type GameLobby interface {
	GameLobbyID
	GetSettings() game.GameSettings
	GetCreatorMark() game.PlayerMark
	GetOpponentMark() game.PlayerMark
	GameStartedChan() chan bool
	CreatorReadyChan() chan bool
	OpponentReadyChan() chan bool
	GetGameController() GameController
	IsGameStarted() bool
	GetRecorder() GameRecorder
	//Start block execution until game is canceled or game controller end execution
	Start() error
}
