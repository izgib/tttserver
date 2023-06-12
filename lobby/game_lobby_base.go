package lobby

type GameLobbyID interface {
	ID() uint32
}

/*type GameLobby interface {
	GameLobbyID
	Settings() game.GameSettings
	CreatorMark() game.PlayerMark
	OpponentMark() game.PlayerMark
	GameStartedChan() chan bool
	CreatorReadyChan() chan bool
	OpponentReadyChan() chan bool
	GetGameController() GameController
	IsGameStarted() bool
	GetRecorder() GameRecorder
	//Start block execution until game is canceled or game controller end execution
	Start(logger *zerolog.Logger) error
}*/
