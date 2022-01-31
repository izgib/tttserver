package lobby

import (
	"github.com/izgib/tttserver/base"
	"github.com/izgib/tttserver/controller"
	"github.com/izgib/tttserver/game"
)

type gameLobby struct {
	ID             int16
	Settings       game.GameSettings
	CreatorMark    game.PlayerMark
	CrReady        chan bool
	OppReady       chan bool
	startedChans   [2]chan bool
	gameController *base.GameController
	startedChan    int
	recorder       base.GameRecorder
}

func (g *gameLobby) GetRecorder() base.GameRecorder {
	return g.recorder
}

func (g *gameLobby) GetID() int16 {
	return g.ID
}

func (g *gameLobby) GetSettings() game.GameSettings {
	return g.Settings
}

func (g *gameLobby) GetCreatorMark() game.PlayerMark {
	return g.CreatorMark
}

func (g *gameLobby) GetOpponentMark() game.PlayerMark {
	return (g.CreatorMark + 1) & 1
}

func (g *gameLobby) GetGameController() base.GameController {
	return *g.gameController
}

func NewGameLobby(ID int16, settings game.GameSettings, creatorMark game.PlayerMark, recorder base.GameRecorder) base.GameLobby {
	return &gameLobby{ID: ID, Settings: settings, CreatorMark: creatorMark, CrReady: make(chan bool), recorder: recorder,
		OppReady: make(chan bool), startedChan: 0, startedChans: [2]chan bool{make(chan bool), make(chan bool)},
	}
}

//Start block execution until game is canceled or game controller end execution
func (g *gameLobby) Start() (err error) {
	started := g.IsGameStarted()
	if started {
		gameController := controller.NewGameController(&g.Settings, g.CreatorMark, g.recorder)
		g.gameController = &gameController
		fanOutVal(started, g.startedChans)
		return gameController.Start()
	} else {
		fanOutVal(started, g.startedChans)
	}
	return nil
}

func (g *gameLobby) IsGameStarted() bool {
	var crReady bool
	var oppReady bool
	for {
		select {
		case crReady = <-g.CrReady:
			if crReady == false {
				return false
			}
		case oppReady = <-g.OppReady:
		}
		if crReady && oppReady {
			break
		}
	}
	return true
}

func (g *gameLobby) GameStartedChan() chan bool {
	defer func() { g.startedChan++ }()
	return g.startedChans[g.startedChan]
}

func (g *gameLobby) CreatorReadyChan() chan bool {
	return g.CrReady
}

func (g *gameLobby) OpponentReadyChan() chan bool {
	return g.OppReady
}

func fanOutVal(val bool, slaves [2]chan bool) {
	go func() {
		for _, c := range slaves {
			c <- val
			// close(c)
		}
	}()
}
