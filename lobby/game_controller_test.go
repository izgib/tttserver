package lobby

import (
	"github.com/izgib/tttserver/internal/logger"
	"sync"
	"testing"

	"github.com/rs/zerolog"

	"github.com/izgib/tttserver/game"
)

var wg sync.WaitGroup

func Test_GameControler(t *testing.T) {
	t.Run("success: X-win", func(t *testing.T) {
		settings := game.GameSettings{Rows: 3, Cols: 3, Win: 3}
		mark := game.CrossMark
		recorder := PlainGameRecorder(GameConfiguration{
			ID:       0,
			Settings: settings,
			Mark:     mark,
		})
		controller := NewGameController(settings, mark, recorder)
		moves := []game.Move{{1, 1}, {0, 2}, {2, 2}, {0, 1}, {0, 0}}

		go controller.Start(logger.CreateDebugLogger())

		crLogger := logger.CreateDebugLogger().With().Str("CrPlayer", game.EnumNamesMoveChoice[game.MoveChoice(controller.CreatorMark())]).Logger()
		crLogger.Debug().Msg("start")
		oppLogger := logger.CreateDebugLogger().With().Str("OppPlayer", game.EnumNamesMoveChoice[game.MoveChoice(controller.OpponentMark())]).Logger()
		oppLogger.Debug().Msg("start")
		wg.Add(2)
		go Player(controller.CreatorChannels(), crLogger, moves)
		go Player(controller.OpponentChannels(), oppLogger, moves)
		wg.Wait()
	})
}

func Player(chans *CommChannels, logger zerolog.Logger, moves []game.Move) {
	turn := 0
playerLoop:
	for {
		tLogger := logger.With().Int("turn", turn).Logger()
		action := <-chans.Actions
		switch a := action.(type) {
		case ReceiveMove:
			chans.Reactions <- ReceivedMove{&moves[turn], nil}
			tLogger.Debug().Dict("move", zerolog.Dict().
				Uint16("i", moves[turn].I).
				Uint16("j", moves[turn].J)).Msg("sent")
		case Step:
			chans.Reactions <- Status{Err: nil}
			if a.Pos != nil {
				move := a.Pos
				tLogger.Debug().Dict("move", zerolog.Dict().
					Uint16("i", move.I).
					Uint16("j", move.J)).
					Str("state", game.EnumNamesGameStateType[a.State.StateType]).
					Msg("received")
			}

			if a.State.StateType == game.Won {
				tLogger.Debug().Str("Winner", game.EnumNamesMoveChoice[game.MoveChoice(a.State.WinLine.Mark)]).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Uint16("i", a.State.WinLine.Start.I).
						Uint16("j", a.State.WinLine.Start.J)).
					Dict("end", zerolog.Dict().
						Uint16("i", a.State.WinLine.End.I).
						Uint16("j", a.State.WinLine.End.J)),
				).Msg("game ended")
			}
			turn++
			tLogger.Debug().Int("turn", turn)

			if a.State.StateType != game.Running {
				break playerLoop
			}
		case StepState:
			chans.Reactions <- Status{Err: nil}
			tLogger.Debug().Str("state", game.EnumNamesGameStateType[a.State.StateType]).Msg("got game state")
			if a.State.StateType == game.Won {
				tLogger.Debug().Str("Winner", game.EnumNamesMoveChoice[game.MoveChoice(a.State.WinLine.Mark)]).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Uint16("i", a.State.WinLine.Start.I).
						Uint16("j", a.State.WinLine.Start.J)).
					Dict("end", zerolog.Dict().
						Uint16("i", a.State.WinLine.End.I).
						Uint16("j", a.State.WinLine.End.J)),
				).Msg("game ended")
			}
			turn++

			tLogger.Debug().Int("turn", turn)

			if a.State.StateType != game.Running {
				break playerLoop
			}
		case Interruption:
			tLogger.Error().Str("event", EnumInterruption[a.Cause]).Msg("interruption")
			return
		}
	}
	wg.Done()
}
