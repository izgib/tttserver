package text

import (
	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/game/models"
	logger2 "github.com/izgib/tttserver/logger"
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

func (r *gameRecorder) RecordStatus(status game.GameEndStatus) error {
	var cause string
	switch status {
	case game.Tie:
		cause = "Tie"
	case game.XWon:
		cause = "Player X Won Game"
	case game.OWon:
		cause = "Player O Won Game"
	case game.XDisconnected:
		cause = "Player X Was Disconnected"
	case game.ODisconnected:
		cause = "Player O Was Disconnected"
	case game.XCheated:
		cause = "Player X make invalid move"
	case game.OCheated:
		cause = "Player O make invalid move"
	}
	r.logger.Debug().Str("cause", cause).Msg("ended")
	return nil
}

func (r *gameRecorder) DeleteGameRecord(ID int16) error {
	r.logger.Debug().Int16("game", r.gameId).Msg("deleted")
	return nil
}

func (r *gameRecorder) RecordMove(move models.Move) error {
	logger := r.logger.With().Int16("turn", r.turn).Logger()
	r.turn++
	logger.Debug().Dict("move", zerolog.Dict().
		Int16("i", move.I).
		Int16("j", move.J),
	).Msg("move")
	return nil
}

var idCounter int16 = 1

func NewGameRecorder(config game.GameConfiguration) game.GameRecorder {
	logger := logger2.CreateDebugLogger().With().Int16("game", idCounter).Logger()
	recorder := &gameRecorder{
		gameId: idCounter,
		logger: logger,
	}
	idCounter++
	logger.Debug().
		Int16("Rows", config.Settings.Rows).
		Int16("Cols", config.Settings.Cols).
		Int16("Win", config.Settings.Win).
		Str("Mark", models.EnumNamesMoveChoice[models.MoveChoice(config.Mark)]).
		Msg("created game record")
	return recorder
}
