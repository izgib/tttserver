package lobby

import (
	"github.com/izgib/tttserver/corgroup"
	"github.com/izgib/tttserver/game"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GameController struct {
	creatorMark game.PlayerMark
	game        *game.Game
	recorder    GameRecorder
	comChans    [2]CommChannels
	gameCount   int
}

// NewGameController create GameController
func NewGameController(settings game.GameSettings, creatorMark game.PlayerMark, recorder GameRecorder) GameController {
	return GameController{
		creatorMark: creatorMark,
		game:        game.NewGame(settings),
		recorder:    recorder,
		comChans: [2]CommChannels{
			{Actions: make(chan Action), Reactions: make(chan Reaction), Done: make(chan struct{}), Err: nil},
			{Actions: make(chan Action), Reactions: make(chan Reaction), Done: make(chan struct{}), Err: nil},
		},
		gameCount: 0,
	}
}

func (g GameController) GameSettings() game.GameSettings {
	return g.game.Settings()
}

// CreatorMark return creator's mark
func (g GameController) CreatorMark() game.PlayerMark {
	return g.creatorMark
}

// OpponentMark return opponent's mark
func (g GameController) OpponentMark() game.PlayerMark {
	return (g.creatorMark + 1) & 1
}

// OpponentChannels returns communication channels with Opponent
func (g GameController) OpponentChannels() *CommChannels {
	ind := g.OpponentMark()
	return &g.comChans[ind]
}

// CreatorChannels returns communication channels with Creator
func (g GameController) CreatorChannels() *CommChannels {
	ind := g.creatorMark
	return &g.comChans[ind]
}

func (g GameController) receiveMove(player int16) error {
	select {
	case g.comChans[player].Actions <- ReceiveMove{}:
		return nil
	case <-g.comChans[player].Done:
		return g.comChans[player].Err
	}
}

func (g *GameController) restartGame() {
	g.game.Restart()
	g.comChans[0], g.comChans[1] = g.comChans[1], g.comChans[0]
}

func (g GameController) sendStep(player int16, move game.Move, state game.GameState) error {
	select {
	case g.comChans[player].Actions <- Step{Pos: &move, State: state}:
		return (<-g.comChans[player].Reactions).(Status).Err
	case <-g.comChans[player].Done:
		return g.comChans[player].Err
	}
}

func (g GameController) sendState(player int16, state game.GameState) error {
	select {
	case g.comChans[player].Actions <- Step{State: state}:
		return (<-g.comChans[player].Reactions).(Status).Err
	case <-g.comChans[player].Done:
		return g.comChans[player].Err
	}
}

func (g GameController) sendInterruption(player int16, cause InterruptionCause) error {
	select {
	case g.comChans[player].Actions <- Interruption{Cause: cause}:
		return nil
	case <-g.comChans[player].Done:
		return g.comChans[player].Err
	}
}

// Start GameController
func (g GameController) Start(logger *zerolog.Logger) error {
	gLogger := logger
	curPlayer, nextPlayer := int16(game.CrossMark), int16(game.NoughtMark)
	var group corgroup.Group
	var move game.Move
	var err error
gameSession:
	for {
		for {
			if err = g.receiveMove(curPlayer); err == nil {
				var consumed bool
				select {
				case resp := <-g.comChans[curPlayer].Reactions:
					consumed = true
					switch reaction := resp.(type) {
					case ReceivedMove:
						if reaction.Err != nil {
							err = playerError{reaction.Err, curPlayer}
							gLogger.Error().Err(err)
							break
						}
						move = *reaction.Pos
					case Interruption:
						switch reaction.Cause {
						case GiveUp:
							winState := game.GameState{StateType: game.Won, WinLine: &game.WinLine{
								Mark:  g.game.GetPlayerMark(g.game.OtherPlayer()),
								Start: nil,
								End:   nil,
							}}
							group.Go(func() error { return g.sendState(nextPlayer, winState) })
							group.Go(func() error { return g.sendState(curPlayer, winState) })
							if err = group.Wait(); err != nil {
								corError := err.(*corgroup.GroupError)
								if corError.ErrorsCount == 1 {
									var plErr error
									if corError.Errors[0] != nil {
										plErr = wrapConnErr(corError.Errors[0], curPlayer)
									} else {
										plErr = wrapConnErr(corError.Errors[1], nextPlayer)
									}
									return errors.Errorf("%v", plErr)
								} else {
									return errors.Errorf("%v, %v",
										wrapConnErr(corError.Errors[0], curPlayer),
										wrapConnErr(corError.Errors[1], nextPlayer),
									)
								}
							}
						case Leave:
							g.sendInterruption(nextPlayer, reaction.Cause)
						default:
							err = g.sendInterruption(nextPlayer, reaction.Cause)
							gLogger.Error().Err(err)
						}

						return err
					}
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
				gLogger.Error().Err(err).Msg("recorder error")

				group.Go(func() error { return g.sendInterruption(curPlayer, Disconnect) })
				group.Go(func() error { return g.sendInterruption(nextPlayer, Disconnect) })

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

				group.Go(func() error { return g.sendInterruption(curPlayer, InvalidMove) })
				group.Go(func() error { return g.sendInterruption(nextPlayer, InvalidMove) })

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
					gLogger.Error().Err(recErr).Msg("recorder error")
					if connErr != nil {
						gLogger.Error().Err(connErr).Msg("connection error")
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
				return g.sendState(curPlayer, state)
			})
			group.Go(func() error {
				return g.sendStep(nextPlayer, move, state)
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

				gLogger.Error().Err(err).Msg("connection error")
				return err
			}

			if state.StateType != game.Running {
				switch state.StateType {
				case game.Tie:
					err = g.recorder.RecordStatus(Tie)
				case game.Won:
					err = g.recorder.RecordStatus(enumMarkToStatus[state.WinLine.Mark])
				}
				if err != nil {
					return errors.WithMessage(err, "db: can not write status")
				}
				g.restartGame()
				g.gameCount++
				continue gameSession
			}
			curPlayer, nextPlayer = nextPlayer, curPlayer
		}
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

func getInterruptionCause(err error) InterruptionCause {
	s := status.Convert(err)
	switch s.Code() {
	case codes.Canceled:
		return Leave
	default:
		return Disconnect
	}
}

var enumMarkToStatus = map[game.PlayerMark]GameEndStatus{
	game.CrossMark:  XWon,
	game.NoughtMark: OWon,
}

var enumCheatingToStatus = map[game.PlayerMark]GameEndStatus{
	game.CrossMark:  XIllegalMove,
	game.NoughtMark: OIllegalMove,
}

var enumDisconnectToStatus = map[game.PlayerMark]GameEndStatus{
	game.CrossMark:  XDisconnected,
	game.NoughtMark: ODisconnected,
}
