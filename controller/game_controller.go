package controller

import (
	"github.com/izgib/tttserver/base"
	"github.com/izgib/tttserver/corgroup"
	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/internal"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type playerError struct {
	error
	player int16
}

type gameController struct {
	settings    *game.GameSettings
	creatorMark game.PlayerMark
	game        *game.Game
	recorder    base.GameRecorder
	comChans    [2]base.CommChannels
}

//NewGameController create GameController
func NewGameController(settings *game.GameSettings, creatorMark game.PlayerMark, recorder base.GameRecorder) base.GameController {
	return &gameController{
		settings:    settings,
		creatorMark: creatorMark,
		game:        game.NewGame(*settings),
		recorder:    recorder,
		comChans: [2]base.CommChannels{
			{Actions: make(chan base.Action), Reactions: make(chan base.Reaction), Done: make(chan struct{}), Err: nil},
			{Actions: make(chan base.Action), Reactions: make(chan base.Reaction), Done: make(chan struct{}), Err: nil},
		},
	}
}

//GetCreatorMark return creator's mark
func (g *gameController) GetCreatorMark() game.PlayerMark {
	return g.creatorMark
}

//GetOpponentMark return opponent's mark
func (g *gameController) GetOpponentMark() game.PlayerMark {
	return (g.creatorMark + 1) & 1
}

func (g *gameController) GetOpponentChannels() *base.CommChannels {
	ind := g.GetOpponentMark()
	return &g.comChans[ind]
}

func (g *gameController) GetCreatorChannels() *base.CommChannels {
	ind := g.creatorMark
	return &g.comChans[ind]
}

func (g *gameController) receiveMove(player int16) error {
	select {
	case g.comChans[player].Actions <- base.ReceiveStep{}:
		return nil
	case <-g.comChans[player].Done:
		return g.comChans[player].Err
	}
}

func (g *gameController) sendMove(player int16, move game.Move) error {
	select {
	case g.comChans[player].Actions <- base.Step{move}:
		return nil
	case <-g.comChans[player].Done:
		return g.comChans[player].Err
	}
}

func (g *gameController) sendState(player int16, state game.GameState) error {
	select {
	case g.comChans[player].Actions <- base.State{state}:
		return nil
	case <-g.comChans[player].Done:
		return g.comChans[player].Err
	}
}

func (g *gameController) sendInterruption(player int16, cause base.InterruptionCause) error {
	select {
	case g.comChans[player].Actions <- base.Interruption{cause}:
		return nil
	case <-g.comChans[player].Done:
		return g.comChans[player].Err
	}
}

func (g *gameController) getStatus(player int16) error {
	return (<-g.comChans[player].Reactions).(base.Status).Err
}

// Start GameController
func (g *gameController) Start() error {
	logger := internal.CreateDebugLogger()
	curPlayer, nextPlayer := int16(game.CrossMark), int16(game.NoughtMark)
	var group corgroup.Group
	var move game.Move
	var err error

	for {
		if err = g.receiveMove(curPlayer); err == nil {
			var consumed bool
			select {
			case resp := <-g.comChans[curPlayer].Reactions:
				consumed = true
				step := resp.(base.ReceivedStep)
				if step.Err != nil {
					err = playerError{step.Err, curPlayer}
					logger.Error().Err(err)
					break
				}
				move = step.Pos
			case <-g.comChans[nextPlayer].Done:
				err = playerError{g.comChans[nextPlayer].Err, nextPlayer}
			}

			if !consumed {
				<-g.comChans[curPlayer].Reactions
			}
		} else {
			err = playerError{err, curPlayer}
		}

		if err != nil {
			plError := err.(playerError)
			curPlErr := wrapConnErr(plError.error, plError.player)
			activePlayer := getActivePlayer(plError.player)
			cause := getInterruptionCause(plError.error)
			if err = g.sendInterruption(activePlayer, cause); err != nil {
				connErr := wrapConnErr(err, activePlayer)
				if err = g.recorder.RecordStatus(enumDisconnectToStatus[game.PlayerMark(plError.player)]); err != nil {
					return errors.Errorf("Pl%d: %v, Pl%d: %v, recorder: %v", plError.player, connErr, activePlayer, connErr, err)
				}
				return errors.Errorf("Pl%d: %v, Pl%d: %v", plError.player, curPlErr, activePlayer, connErr)
			}
			return curPlErr
		}

		if err = g.recorder.RecordMove(move); err != nil {
			recErr := errors.Wrapf(err, "recorder: can not write move")
			logger.Error().Err(err).Msg("recorder error")

			group.Go(func() error { return g.sendInterruption(curPlayer, base.Disconnect) })
			group.Go(func() error { return g.sendInterruption(nextPlayer, base.Disconnect) })

			if err = group.Wait(); err != nil {
				corError := err.(*corgroup.GroupError)
				if corError.ErrorsCount == 1 {
					var plErr error
					if corError.Errors[0] != nil {
						plErr = wrapConnErr(corError.Errors[0], curPlayer)
					} else {
						plErr = wrapConnErr(corError.Errors[1], nextPlayer)
					}
					return errors.Errorf("%v, %v", plErr, recErr)
				} else {
					return errors.Errorf("%v, %v, %v",
						wrapConnErr(corError.Errors[0], curPlayer),
						wrapConnErr(corError.Errors[1], nextPlayer),
						recErr)
				}
			}

			return recErr
		}
		group = corgroup.Group{}

		if err = g.game.MoveTo(move); err != nil {
			gameErr := errors.Wrap(err, "game: invalid move")

			group.Go(func() error { return g.sendInterruption(curPlayer, base.InvalidMove) })
			group.Go(func() error { return g.sendInterruption(nextPlayer, base.OppInvalidMove) })

			var connErr error
			if connErr = group.Wait(); connErr != nil {
				corError := connErr.(*corgroup.GroupError)
				if corError.ErrorsCount == 1 {
					if corError.Errors[0] != nil {
						connErr = wrapConnErr(corError.Errors[0], curPlayer)
					} else {
						connErr = wrapConnErr(corError.Errors[1], nextPlayer)
					}
				} else {
					connErr = errors.Errorf("%v, %v",
						wrapConnErr(corError.Errors[0], curPlayer),
						wrapConnErr(corError.Errors[1], nextPlayer),
					)
				}
			}

			if err = g.recorder.RecordStatus(enumCheatingToStatus[game.PlayerMark(curPlayer)]); err != nil {
				recErr := errors.Wrapf(err, "recorder: can not write status")
				logger.Error().Err(recErr).Msg("recorder error")
				if connErr != nil {
					logger.Error().Err(connErr).Msg("connection error")
					return errors.Errorf("game: %v, recorder: %v, conn: %v", gameErr, recErr, connErr)
				}

				return errors.Errorf("%v, %v", gameErr, recErr)
			}

			if connErr != nil {
				plError := err.(playerError)
				connErr = wrapConnErr(plError.error, plError.player)
				return errors.Errorf("%v, conn: %v", gameErr, connErr)
			}
			return gameErr
		}

		state := g.game.GameState(move)

		group = corgroup.Group{}
		group.Go(func() error {
			var plErr error
			if plErr = g.sendState(curPlayer, state); plErr != nil {
				return plErr
			}
			return g.getStatus(curPlayer)
		})
		group.Go(func() error {
			var plErr error
			if plErr = g.sendMove(nextPlayer, move); plErr != nil {
				return plErr
			}
			if plErr = g.getStatus(nextPlayer); plErr != nil {
				return plErr
			}
			if plErr = g.sendState(nextPlayer, state); plErr != nil {
				return plErr
			}
			return g.getStatus(nextPlayer)
		})

		if err = group.Wait(); err != nil {
			corError := err.(*corgroup.GroupError)
			var optErr error
			if corError.ErrorsCount == 1 {
				if corError.Errors[0] != nil {
					err = wrapConnErr(corError.Errors[0], curPlayer)
					optErr = g.sendInterruption(nextPlayer, getInterruptionCause(corError.Errors[0]))
				} else {
					err = wrapConnErr(corError.Errors[1], nextPlayer)
					optErr = g.sendInterruption(curPlayer, getInterruptionCause(corError.Errors[1]))
				}
			} else {
				err = wrapConnErr(corError.Errors[0], curPlayer)
				optErr = wrapConnErr(corError.Errors[1], nextPlayer)
			}

			if optErr != nil {
				err = errors.Errorf("%v, %v", err, optErr)
			}

			logger.Error().Err(err).Msg("connection error")
			return err
		}

		if state.StateType != game.Running {
			switch state.StateType {
			case game.Tie:
				err = g.recorder.RecordStatus(base.Tie)
			case game.Won:
				err = g.recorder.RecordStatus(enumMarkToStatus[state.WinLine.Mark])
			}
			if err != nil {
				return errors.WithMessage(err, "db: can not write status")
			}
			return nil
		}
		curPlayer, nextPlayer = nextPlayer, curPlayer
	}
}

func wrapConnErr(err error, player int16) error {
	return errors.Wrapf(err, "connection error from player: %v", player)
}

func getActivePlayer(player int16) int16 {
	if player == int16(game.CrossMark) {
		return int16(game.NoughtMark)
	}
	return int16(game.CrossMark)
}

func getInterruptionCause(err error) base.InterruptionCause {
	s := status.Convert(err)
	switch s.Code() {
	case codes.Canceled:
		return base.Leave
	default:
		return base.Disconnect
	}
}

var enumMarkToStatus = map[game.PlayerMark]base.GameEndStatus{
	game.CrossMark:  base.XWon,
	game.NoughtMark: base.OWon,
}

var enumCheatingToStatus = map[game.PlayerMark]base.GameEndStatus{
	game.CrossMark:  base.XCheated,
	game.NoughtMark: base.OCheated,
}

var enumDisconnectToStatus = map[game.PlayerMark]base.GameEndStatus{
	game.CrossMark:  base.XDisconnected,
	game.NoughtMark: base.ODisconnected,
}
