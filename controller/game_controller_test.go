package controller_test

import (
	"github.com/izgib/tttserver/base"
	"github.com/izgib/tttserver/controller"
	"github.com/izgib/tttserver/recorder/text"

	//"github.com/izgib/tttserver/lobby/recorder/text"
	"sync"
	"testing"

	"github.com/rs/zerolog"

	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/internal"
)

var wg sync.WaitGroup

func Test_GameControler(t *testing.T) {
	t.Run("success: X-win", func(t *testing.T) {
		settings := game.GameSettings{3, 3, 3}
		mark := game.CrossMark
		recorder := text.NewGameRecorder(base.GameConfiguration{
			ID:       0,
			Settings: settings,
			Mark:     mark,
		})
		controller := controller.NewGameController(&settings, mark, recorder)
		moves := []game.Move{{1, 1}, {0, 2}, {2, 2}, {0, 1}, {0, 0}}

		go controller.Start()

		crLogger := internal.CreateDebugLogger().With().Str("CrPlayer", game.EnumNamesMoveChoice[game.MoveChoice(controller.GetCreatorMark())]).Logger()
		crLogger.Debug().Msg("start")
		oppLogger := internal.CreateDebugLogger().With().Str("OppPlayer", game.EnumNamesMoveChoice[game.MoveChoice(controller.GetOpponentMark())]).Logger()
		oppLogger.Debug().Msg("start")
		wg.Add(2)
		go Player(controller.GetCreatorChannels(), crLogger, moves)
		go Player(controller.GetOpponentChannels(), oppLogger, moves)
		wg.Wait()
	})
}

func Player(chans *base.CommChannels, logger zerolog.Logger, moves []game.Move) {
	turn := 0
playerLoop:
	for {
		tLogger := logger.With().Int("turn", turn).Logger()
		action := <-chans.Actions
		switch a := action.(type) {
		case base.ReceiveStep:
			chans.Reactions <- base.ReceivedStep{moves[turn], nil}
			tLogger.Debug().Dict("move", zerolog.Dict().
				Int16("i", moves[turn].I).
				Int16("j", moves[turn].J)).Msg("sent")
		case base.Step:
			move := a.Pos
			tLogger.Debug().Dict("move", zerolog.Dict().
				Int16("i", move.I).
				Int16("j", move.J)).Msg("received")
			chans.Reactions <- base.Status{nil}
		case base.State:
			tLogger.Debug().Str("state", game.EnumNamesGameStateType[a.State.StateType]).Msg("got game state")
			if a.State.StateType == game.Won {
				tLogger.Debug().Str("Winner", game.EnumNamesMoveChoice[game.MoveChoice(a.State.WinLine.Mark)]).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Int16("i", a.State.WinLine.Start.I).
						Int16("j", a.State.WinLine.Start.J)).
					Dict("end", zerolog.Dict().
						Int16("i", a.State.WinLine.End.I).
						Int16("j", a.State.WinLine.End.J)),
				).Msg("game ended")
			}
			turn++
			chans.Reactions <- base.Status{nil}

			if a.State.StateType != game.Running {
				break playerLoop
			}
		case base.Interruption:
			tLogger.Error().Str("event", base.EnumInterruption[a.Cause]).Msg("interruption")
			return
		}
	}
	wg.Done()
}
