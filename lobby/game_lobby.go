package lobby

import (
	"github.com/izgib/tttserver/game"
	"github.com/rs/zerolog"
)

type GameLobby struct {
	ID             uint32
	settings       game.GameSettings
	creatorMark    game.PlayerMark
	crReady        chan bool
	oppReady       chan bool
	startedChans   [2]chan bool
	gameController *GameController
	startedChan    int
	recorder       GameRecorder
}

func (g GameLobby) GetRecorder() GameRecorder {
	return g.recorder
}

func (g GameLobby) Settings() game.GameSettings {
	return g.settings
}

func (g GameLobby) CreatorMark() game.PlayerMark {
	return g.creatorMark
}

func (g GameLobby) OpponentMark() game.PlayerMark {
	return (g.creatorMark + 1) & 1
}

func (g GameLobby) GetGameController() GameController {
	return *g.gameController
}

func NewGameLobby(ID uint32, settings game.GameSettings, creatorMark game.PlayerMark, recorder GameRecorder) GameLobby {
	return GameLobby{ID: ID, settings: settings, creatorMark: creatorMark, crReady: make(chan bool), recorder: recorder,
		oppReady: make(chan bool), startedChan: 0, startedChans: [2]chan bool{make(chan bool), make(chan bool)},
	}
}

// Start block execution until game is canceled or game controller end execution
func (g *GameLobby) Start(logger *zerolog.Logger) (err error) {
	started := g.IsGameStarted()
	if started {
		gameController := NewGameController(g.settings, g.creatorMark, g.recorder)
		g.gameController = &gameController
		fanOutVal(started, g.startedChans)
		return gameController.Start(logger)
	} else {
		fanOutVal(started, g.startedChans)
	}
	return nil
}

func (g *GameLobby) IsGameStarted() bool {
	var crReady bool
	var oppReady bool
	for {
		select {
		case crReady = <-g.crReady:
			if crReady == false {
				return false
			}
		case oppReady = <-g.oppReady:
		}
		if crReady && oppReady {
			break
		}
	}
	return true
}

func (g *GameLobby) GameStartedChan() chan bool {
	defer func() { g.startedChan++ }()
	return g.startedChans[g.startedChan]
}

func (g *GameLobby) CreatorReadyChan() chan bool {
	return g.crReady
}

func (g *GameLobby) OpponentReadyChan() chan bool {
	return g.oppReady
}

func fanOutVal(val bool, slaves [2]chan bool) {
	go func() {
		for _, c := range slaves {
			c <- val
			// close(c)
		}
	}()
}
