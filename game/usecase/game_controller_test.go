package usecase

import (
	"github.com/izgib/tttserver/game/interface/recorder/text"
	"sync"
	"testing"

	"github.com/rs/zerolog"

	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/game/models"
	"github.com/izgib/tttserver/logger"
)

var wg sync.WaitGroup

func Test_GameControler(t *testing.T) {
	t.Run("success: X-win", func(t *testing.T) {
		settings := models.GameSettings{3, 3, 3}
		mark := models.CrossMark
		recorder := text.NewGameRecorder(game.GameConfiguration{
			ID:       0,
			Settings: settings,
			Mark:     mark,
		})
		controller := NewGameController(&settings, models.CrossMark, recorder)
		moves := []models.Move{{1, 1}, {0, 2}, {2, 2}, {0, 1}, {0, 0}}

		go controller.Start()

		crLogger := logger.CreateDebugLogger().With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(controller.GetCreatorMark())]).Logger()
		crLogger.Debug().Msg("start")
		oppLogger := logger.CreateDebugLogger().With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(controller.GetOpponentMark())]).Logger()
		oppLogger.Debug().Msg("start")
		wg.Add(2)
		go Player(controller, controller.GetCreatorComChannels(), crLogger, moves)
		go Player(controller, controller.GetOpponentComChannels(), oppLogger, moves)
		wg.Wait()
	})
}

func Player(controller game.GameController, chans game.PlayerComm, logger zerolog.Logger, moves []models.Move) {
	turn := 0
playerLoop:
	for {
		tLogger := logger.With().Int("turn", turn).Logger()
		switch <-chans.TypeChan {
		case game.MoveCh:
			if <-chans.MoveRequesterChan {
				chans.CommChan <- moves[turn]
				tLogger.Debug().Dict("move", zerolog.Dict().
					Int16("i", moves[turn].I).
					Int16("j", moves[turn].J)).Msg("sent")
			} else {
				move := <-chans.CommChan
				tLogger.Debug().Dict("move", zerolog.Dict().
					Int16("i", move.I).
					Int16("j", move.J)).Msg("received")
				chans.ErrChan <- nil
			}
		case game.StateCh:
			state := <-chans.StateChan
			tLogger.Debug().Str("state", models.EnumNamesGameStateType[state.StateType]).Msg("got game state")
			if state.StateType == models.Won {
				tLogger.Debug().Str("Winner", models.EnumNamesMoveChoice[models.MoveChoice(state.WinLine.Mark)]).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Int16("i", state.WinLine.Start.I).
						Int16("j", state.WinLine.Start.J)).
					Dict("end", zerolog.Dict().
						Int16("i", state.WinLine.End.I).
						Int16("j", state.WinLine.End.J)),
				).Msg("game ended")
			}
			turn++

			chans.ErrChan <- nil

			if state.StateType != models.Running {
				break playerLoop
			}
		case game.InterruptCh:
			tLogger.Error().Str("event", game.EnumInterruption[<-chans.InterruptChan]).Msg("interruption")
		}
	}
	wg.Done()
}
