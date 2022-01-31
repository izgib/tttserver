package text

import (
	"github.com/izgib/tttserver/base"
	"github.com/izgib/tttserver/game"
	logger "github.com/izgib/tttserver/internal"
	"github.com/rs/zerolog"
)

type gameRecorder struct {
	gameId int16
	logger zerolog.Logger
	turn   int16
}

func (r *gameRecorder) GetID() int16 {
	return r.gameId
}

func (r *gameRecorder) RecordStatus(status base.GameEndStatus) error {
	var cause string
	switch status {
	case base.Tie:
		cause = "Tie"
	case base.XWon:
		cause = "Player X Won Game"
	case base.OWon:
		cause = "Player O Won Game"
	case base.XDisconnected:
		cause = "Player X Was Disconnected"
	case base.ODisconnected:
		cause = "Player O Was Disconnected"
	case base.XCheated:
		cause = "Player X make invalid move"
	case base.OCheated:
		cause = "Player O make invalid move"
	}
	r.logger.Debug().Str("cause", cause).Msg("ended")
	return nil
}

func (r *gameRecorder) RecordMove(move game.Move) error {
	logger := r.logger.With().Int16("turn", r.turn).Logger()
	r.turn++
	logger.Debug().Dict("move", zerolog.Dict().
		Int16("i", move.I).
		Int16("j", move.J),
	).Msg("move")
	return nil
}

var idCounter int16 = 1

func NewGameRecorder(config base.GameConfiguration) base.GameRecorder {
	logger := logger.CreateDebugLogger().With().Int16("game", idCounter).Logger()
	recorder := &gameRecorder{
		gameId: idCounter,
		logger: logger,
	}
	idCounter++
	logger.Info().
		Int16("Rows", config.Settings.Rows).
		Int16("Cols", config.Settings.Cols).
		Int16("Win", config.Settings.Win).
		Str("Mark", game.EnumNamesMoveChoice[game.MoveChoice(config.Mark)]).
		Msg("created game record")
	return recorder
}
