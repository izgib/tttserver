package lobby

import (
	"github.com/izgib/tttserver/game"
	logger2 "github.com/izgib/tttserver/internal/logger"
	"github.com/rs/zerolog"
)

type GameRecorder interface {
	ID() int64
	RecordMove(move game.Move) error
	RecordStatus(status GameEndStatus) error
}

type GameEndStatus byte

const (
	Tie GameEndStatus = iota
	XWon
	XDisconnected
	XIllegalMove
	XLeft
	OWon
	ODisconnected
	OIllegalMove
	OLeft
)

type gameRecorder struct {
	gameId int64
	logger zerolog.Logger
	turn   uint16
}

func (r *gameRecorder) ID() int64 {
	return r.gameId
}

func (r *gameRecorder) RecordStatus(status GameEndStatus) error {
	var cause string
	switch status {
	case Tie:
		cause = "Tie"
	case XWon:
		cause = "Player X Won Game"
	case OWon:
		cause = "Player O Won Game"
	case XDisconnected:
		cause = "Player X Was Disconnected"
	case ODisconnected:
		cause = "Player O Was Disconnected"
	case XIllegalMove:
		cause = "Player X make invalid move"
	case OIllegalMove:
		cause = "Player O make invalid move"
	case XLeft:
		cause = "Player O left session "
	case OLeft:
		cause = "Player O left session"
	}
	r.logger.Debug().Str("cause", cause).Msg("ended")
	return nil
}

func (r *gameRecorder) RecordMove(move game.Move) error {
	logger := r.logger.With().Uint16("turn", r.turn).Logger()
	r.turn++
	logger.Debug().Dict("move", zerolog.Dict().
		Uint16("i", move.I).
		Uint16("j", move.J),
	).Msg("move")
	return nil
}

var idCounter int64 = 1

func PlainGameRecorder(config GameConfiguration) GameRecorder {
	logger := logger2.CreateDebugLogger().With().Int64("game", idCounter).Logger()
	recorder := &gameRecorder{
		gameId: idCounter,
		logger: logger,
	}
	idCounter++
	logger.Info().
		Uint16("Rows", config.Settings.Rows).
		Uint16("Cols", config.Settings.Cols).
		Uint16("Win", config.Settings.Win).
		Str("Mark", game.EnumNamesMoveChoice[game.MoveChoice(config.Mark)]).
		Msg("created game record")
	return recorder
}
