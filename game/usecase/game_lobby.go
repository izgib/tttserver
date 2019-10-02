package usecase

import (
	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/game/models"
)

type GameLobby struct {
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

func (g *GameLobby) GetRecorder() game.GameRecorder {
	return g.recorder
}

func (g *GameLobby) GetID() int16 {
	return g.ID
}

func (g *GameLobby) GetSettings() models.GameSettings {
	return g.Settings
}

func (g *GameLobby) GetCreatorMark() models.PlayerMark {
	return g.CreatorMark
}

func (g *GameLobby) GetOpponentMark() models.PlayerMark {
	return (g.CreatorMark + 1) & 1
}

func (g *GameLobby) GetGameController() game.GameController {
	return *g.gameController
}

func NewGameLobby(ID int16, settings models.GameSettings, creatorMark models.PlayerMark, recorder game.GameRecorder) game.GameLobby {
	return &GameLobby{ID: ID, Settings: settings, CreatorMark: creatorMark, CrReady: make(chan bool), recorder: recorder,
		OppReady: make(chan bool), startedChan: 0, startedChans: [2]chan bool{make(chan bool), make(chan bool)},
	}
}

func (g *GameLobby) OnStart() {
	started := g.IsGameStarted()
	if started {
		controller := NewGameController(&g.Settings, g.CreatorMark, g.recorder)
		g.gameController = &controller
		go controller.Start()
	}

	fanOutVal(started, g.startedChans)
}

func (g *GameLobby) IsGameStarted() bool {
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

func (g *GameLobby) GameStartedChan() chan bool {
	defer func() { g.startedChan++ }()
	return g.startedChans[g.startedChan]
}

func (g *GameLobby) CreatorReadyChan() chan bool {
	return g.CrReady
}

func (g *GameLobby) OpponentReadyChan() chan bool {
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
