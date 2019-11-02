package usecase

import (
	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/game/models"
)

type gameLobby struct {
	ID             int16
	Settings       models.GameSettings
	CreatorMark    models.PlayerMark
	CrReady        chan bool
	OppReady       chan bool
	startedChans   [2]chan bool
	gameController *game.GameController
	startedChan    int
	recorder       game.GameRecorder
}

func (g *gameLobby) GetRecorder() game.GameRecorder {
	return g.recorder
}

func (g *gameLobby) GetID() int16 {
	return g.ID
}

func (g *gameLobby) GetSettings() models.GameSettings {
	return g.Settings
}

func (g *gameLobby) GetCreatorMark() models.PlayerMark {
	return g.CreatorMark
}

func (g *gameLobby) GetOpponentMark() models.PlayerMark {
	return (g.CreatorMark + 1) & 1
}

func (g *gameLobby) GetGameController() game.GameController {
	return *g.gameController
}

func NewGameLobby(ID int16, settings models.GameSettings, creatorMark models.PlayerMark, recorder game.GameRecorder) game.GameLobby {
	return &gameLobby{ID: ID, Settings: settings, CreatorMark: creatorMark, CrReady: make(chan bool), recorder: recorder,
		OppReady: make(chan bool), startedChan: 0, startedChans: [2]chan bool{make(chan bool), make(chan bool)},
	}
}

//Start block execution until game is canceled or game controller end execution
func (g *gameLobby) Start() {
	started := g.IsGameStarted()
	if started {
		controller := NewGameController(&g.Settings, g.CreatorMark, g.recorder)
		g.gameController = &controller
		fanOutVal(started, g.startedChans)
		controller.Start()
	} else {
		fanOutVal(started, g.startedChans)
	}
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
