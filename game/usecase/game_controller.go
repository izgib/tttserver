package usecase

import (
	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/game/models"
	"github.com/izgib/tttserver/internal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gameController struct {
	settings    *models.GameSettings
	creatorMark models.PlayerMark
	game        *models.Game
	recorder    game.GameRecorder
	// true is to receive move, false to send
	typeChan      [2]chan game.ObjType
	moveRequester [2]chan bool
	playerChan    [2]chan models.Move
	stateChan     [2]chan models.GameState
	interruptChan [2]chan game.Interruption
	errChan       [2]chan error
}

//NewGameController create GameController
func NewGameController(settings *models.GameSettings, creatorMark models.PlayerMark, recorder game.GameRecorder) game.GameController {
	return &gameController{
		settings:      settings,
		creatorMark:   creatorMark,
		game:          models.NewGame(*settings),
		typeChan:      [2]chan game.ObjType{make(chan game.ObjType), make(chan game.ObjType)},
		moveRequester: [2]chan bool{make(chan bool), make(chan bool)},
		playerChan:    [2]chan models.Move{make(chan models.Move), make(chan models.Move)},
		stateChan:     [2]chan models.GameState{make(chan models.GameState), make(chan models.GameState)},
		interruptChan: [2]chan game.Interruption{make(chan game.Interruption), make(chan game.Interruption)},
		errChan:       [2]chan error{make(chan error), make(chan error)},
		recorder:      recorder,
	}
}

//GetOpponentMark return opponent's mark
func (g *gameController) GetOpponentMark() models.PlayerMark {
	return (g.creatorMark + 1) & 1
}

//GetCreatorMark return creator's mark
func (g *gameController) GetCreatorMark() models.PlayerMark {
	return g.creatorMark
}

// GetCreatorComChannels return Creator's channels for communication with GameController
func (g *gameController) GetCreatorComChannels() game.PlayerComm {
	ind := g.creatorMark
	return game.PlayerComm{
		CommChan:          g.playerChan[ind],
		MoveRequesterChan: g.moveRequester[ind],
		StateChan:         g.stateChan[ind],
		InterruptChan:     g.interruptChan[ind],
		ErrChan:           g.errChan[ind],
		TypeChan:          g.typeChan[ind],
	}
}

// GetOpponentComChannels return Opponent's channels for communication with GameController
func (g *gameController) GetOpponentComChannels() game.PlayerComm {
	ind := g.GetOpponentMark()
	return game.PlayerComm{
		CommChan:          g.playerChan[ind],
		MoveRequesterChan: g.moveRequester[ind],
		StateChan:         g.stateChan[ind],
		InterruptChan:     g.interruptChan[ind],
		ErrChan:           g.errChan[ind],
		TypeChan:          g.typeChan[ind],
	}
}

func (g *gameController) requestMove(player int16) (move models.Move, err error) {
	g.typeChan[player] <- game.MoveCh
	g.moveRequester[player] <- true
	select {
	case move = <-g.playerChan[player]:
		return move, nil
	case err = <-g.errChan[player]:
		return models.Move{}, err
	}
}

func (g *gameController) sendMove(player int16, move models.Move) {
	g.typeChan[player] <- game.MoveCh
	g.moveRequester[player] <- false
	g.playerChan[player] <- move
}

func (g *gameController) sendState(player int16, state models.GameState) {
	g.typeChan[player] <- game.StateCh
	g.stateChan[player] <- state
}

func (g *gameController) sendInterruption(player int16, cause game.Interruption) {
	g.typeChan[player] <- game.InterruptCh
	g.interruptChan[player] <- cause
}

// Start GameController
func (g *gameController) Start() {
	logger := internal.CreateDebugLogger()
	curPlayer, nextPlayer := int16(models.Cross), int16(models.Nought)
	for {
		var move models.Move
		var err error

		if move, err = g.requestMove(curPlayer); err != nil {
			s := status.Convert(err)
			switch s.Code() {
			case codes.Canceled:
				g.sendInterruption(nextPlayer, game.Leave)
			default:
				g.sendInterruption(nextPlayer, game.Disconnect)
			}
			if err = g.recorder.RecordStatus(enumDisconnectToStatus[models.PlayerMark(curPlayer)]); err != nil {
				logger.Error().Err(err).Msg("db error: can not write status")
			}
			return
		}

		if err = g.recorder.RecordMove(move); err != nil {
			logger.Error().Err(err).Msg("db error: can not write move")
		}

		if err = g.game.MoveTo(move); err != nil {
			if err = g.recorder.RecordStatus(enumCheatingToStatus[models.PlayerMark(curPlayer)]); err != nil {
				logger.Error().Err(err).Msg("db error: can not write status")
			}
			g.sendInterruption(curPlayer, game.InvalidMove)
			g.sendInterruption(nextPlayer, game.OppInvalidMove)
			return
		}

		state := g.game.GameState(move)
		g.sendState(curPlayer, state)
		if err = <-g.errChan[curPlayer]; err != nil {
			s := status.Convert(err)
			switch s.Code() {
			case codes.Canceled:
				g.sendInterruption(nextPlayer, game.Leave)
			default:
				g.sendInterruption(nextPlayer, game.Disconnect)
			}
			if err = g.recorder.RecordStatus(enumDisconnectToStatus[models.PlayerMark(curPlayer)]); err != nil {
				logger.Error().Err(err).Msg("db error: can not write status")
			}
			return
		}

		g.sendMove(nextPlayer, move)
		if err = <-g.errChan[nextPlayer]; err != nil {
			s := status.Convert(err)
			switch s.Code() {
			case codes.Canceled:
				g.sendInterruption(curPlayer, game.Leave)
			default:
				g.sendInterruption(curPlayer, game.Disconnect)
			}
			if err = g.recorder.RecordStatus(enumDisconnectToStatus[models.PlayerMark(nextPlayer)]); err != nil {
				logger.Error().Err(err).Msg("db error: can not write status")
			}
			return
		}

		g.sendState(nextPlayer, state)
		if err = <-g.errChan[nextPlayer]; err != nil {
			s := status.Convert(err)
			switch s.Code() {
			case codes.Canceled:
				g.sendInterruption(curPlayer, game.Leave)
			default:
				g.sendInterruption(curPlayer, game.Disconnect)
			}
			if err = g.recorder.RecordStatus(enumDisconnectToStatus[models.PlayerMark(nextPlayer)]); err != nil {
				logger.Error().Err(err).Msg("db error: can not write status")
			}
			return
		}

		if state.StateType != models.Running {
			switch state.StateType {
			case models.Tie:
				err = g.recorder.RecordStatus(game.Tie)
			case models.Won:
				err = g.recorder.RecordStatus(enumMarkToStatus[state.WinLine.Mark])
			}
			if err != nil {
				logger.Error().Err(err).Msg("db error: can not write status")
			}
			return
		}
		curPlayer, nextPlayer = nextPlayer, curPlayer
	}
}

var enumMarkToStatus = map[models.PlayerMark]game.GameEndStatus{
	models.CrossMark:  game.XWon,
	models.NoughtMark: game.OWon,
}

var enumCheatingToStatus = map[models.PlayerMark]game.GameEndStatus{
	models.CrossMark:  game.XCheated,
	models.NoughtMark: game.OCheated,
}

var enumDisconnectToStatus = map[models.PlayerMark]game.GameEndStatus{
	models.CrossMark:  game.XDisconnected,
	models.NoughtMark: game.ODisconnected,
}
