package lobby

import (
	"fmt"
	"github.com/izgib/tttserver/base"
	"strconv"
	"sync"

	"github.com/rs/zerolog"

	"github.com/izgib/tttserver/game"
)

type gameLobbyController struct {
	mu            sync.Locker
	recordCreator func(config base.GameConfiguration) base.GameRecorder
	lobbies       map[int16]base.GameLobby
	logger        *zerolog.Logger
}

func NewGameLobbyController(recordCreator func(config base.GameConfiguration) base.GameRecorder, logger *zerolog.Logger) base.GameLobbyController {
	return &gameLobbyController{recordCreator: recordCreator, logger: logger, lobbies: make(map[int16]base.GameLobby), mu: &sync.Mutex{}}
}

func (g *gameLobbyController) CreateLobby(config base.GameConfiguration) (base.GameLobby, error) {
	g.logger.Debug().
		Int16("Rows", config.Settings.Rows).
		Int16("Cols", config.Settings.Cols).
		Int16("Win", config.Settings.Win).
		Str("Mark", game.EnumNamesMoveChoice[game.MoveChoice(config.Mark)]).
		Msg("request to create lobby")
	if config.Settings.Rows < game.GameRules[game.Rows].Start {
		err := &GameCreationError{
			parameter: game.Rows,
			Bound:     game.GameRules[game.Rows].Start,
			errorType: BelowRange,
			Value:     config.Settings.Rows,
		}
		return nil, err
	}
	if game.GameRules[game.Rows].End < config.Settings.Rows {
		err := &GameCreationError{
			parameter: game.Rows,
			Bound:     game.GameRules[game.Rows].End,
			errorType: AboveRange,
			Value:     config.Settings.Rows,
		}
		return nil, err
	}

	if config.Settings.Cols < game.GameRules[game.Cols].Start {
		err := &GameCreationError{
			parameter: game.Cols,
			Bound:     game.GameRules[game.Cols].Start,
			errorType: BelowRange,
			Value:     config.Settings.Cols,
		}
		return nil, err
	}
	if game.GameRules[game.Cols].End < config.Settings.Cols {
		err := &GameCreationError{
			parameter: game.Cols,
			Bound:     game.GameRules[game.Cols].End,
			errorType: AboveRange,
			Value:     config.Settings.Cols,
		}
		return nil, err
	}

	if config.Settings.Win < game.GameRules[game.Win].Start {
		err := &GameCreationError{
			parameter: game.Win,
			Bound:     game.GameRules[game.Win].Start,
			errorType: BelowRange,
			Value:     config.Settings.Win,
		}
		return nil, err
	}
	highBound := config.Settings.Rows
	if config.Settings.Rows > config.Settings.Cols {
		highBound = config.Settings.Cols
	}
	if highBound < config.Settings.Win {
		err := &GameCreationError{
			parameter: game.Win,
			Bound:     highBound,
			errorType: AboveRange,
			Value:     config.Settings.Win,
		}
		return nil, err
	}

	gameRecorder := g.recordCreator(config)
	lobby := NewGameLobby(gameRecorder.GetID(), config.Settings, config.Mark, gameRecorder)
	g.mu.Lock()
	g.lobbies[gameRecorder.GetID()] = lobby
	g.mu.Unlock()

	go func() {
		if err := lobby.Start(); err != nil {
			logger := g.logger.With().Int16("game", lobby.GetID()).Logger()
			(&logger).Info().Err(err).Msg("ended")
		}
		g.mu.Lock()
		delete(g.lobbies, lobby.GetID())
		g.mu.Unlock()
	}()

	return lobby, nil
}

func (g *gameLobbyController) ListLobbies(filter base.GameFilter) ([]base.GameConfiguration, error) {
	var err error = nil
	rows := filter.Rows
	cols := filter.Cols
	win := filter.Win
	mark := filter.Mark
	wrongOrder := fmt.Errorf("start must be <= end")

	if rows.Start > rows.End {
		return nil, wrongOrder
	}
	if !(game.GameRules[game.Rows].Start <= rows.Start && rows.End <= game.GameRules[game.Rows].End) {
		err = &ListFilterError{
			parameter: game.Rows,
			LowBound:  game.GameRules[game.Rows].Start,
			HighBound: game.GameRules[game.Rows].End,
			StartVal:  rows.Start,
			EndVal:    rows.End,
		}
		return nil, err
	}

	if cols.Start > cols.End {
		return nil, wrongOrder
	}
	if !(game.GameRules[game.Cols].Start <= cols.Start && cols.End <= game.GameRules[game.Cols].End) {
		err = &ListFilterError{
			parameter: game.Cols,
			LowBound:  game.GameRules[game.Cols].Start,
			HighBound: game.GameRules[game.Cols].End,
			StartVal:  cols.Start,
			EndVal:    cols.End,
		}
		return nil, err
	}

	if win.Start > win.End {
		return nil, wrongOrder
	}

	highBound := rows.End
	if rows.End > cols.End {
		highBound = cols.End
	}
	if !(game.GameRules[game.Win].Start <= win.Start && win.End <= highBound) {
		err = &ListFilterError{
			parameter: game.Win,
			LowBound:  game.GameRules[game.Cols].Start,
			HighBound: game.GameRules[game.Cols].End,
			StartVal:  cols.Start,
			EndVal:    cols.End,
		}
		return nil, err
	}

	var games []base.GameConfiguration
	for _, v := range g.lobbies {
		if (rows.Start <= v.GetSettings().Rows && v.GetSettings().Rows <= rows.End) &&
			(cols.Start <= v.GetSettings().Cols && v.GetSettings().Cols <= cols.End) &&
			(win.Start <= v.GetSettings().Win && v.GetSettings().Win <= win.End) {
			if mark == game.Empty || mark == game.MoveChoice(v.GetCreatorMark()) {
				g.logger.Debug().Int16("game lobby", v.GetID())
				games = append(games,
					base.GameConfiguration{
						ID:       v.GetID(),
						Settings: v.GetSettings(),
						Mark:     v.GetCreatorMark(),
					})
			}
		}
	}

	return games, nil
}

func (g *gameLobbyController) JoinLobby(ID int16) (base.GameLobby, error) {
	g.logger.Debug().Int16("game", ID).Msg("request to join lobby")
	if game, ok := g.lobbies[ID]; ok {
		return game, nil
	} else {
		errString := fmt.Sprintf("GameConfiguration %d was not found", ID)
		err := NewGameLobbyError(errString)
		g.logger.Debug().Err(err)
		return nil, err
	}
}

func (g *gameLobbyController) DeleteLobby(lobby base.GameLobby) {
	g.mu.Lock()
	delete(g.lobbies, lobby.GetID())
	g.mu.Unlock()
}

type GameCreationErrorType uint

const (
	BelowRange GameCreationErrorType = iota
	AboveRange
)

type GameCreationError struct {
	parameter game.GameParameter
	Bound     int16
	errorType GameCreationErrorType
	Value     int16
}

func (e *GameCreationError) Error() string {
	var sign string
	var bound string
	switch e.errorType {
	case BelowRange:
		sign = ">="
		bound = strconv.FormatInt(int64(e.Bound), 10)
	case AboveRange:
		sign = "<="
		if e.parameter == game.Win {
			bound = fmt.Sprintf("%d(min of row and col length)", e.Bound)
		} else {
			bound = strconv.FormatInt(int64(e.Bound), 10)
		}
	}
	return fmt.Sprintf("%s must be %s %s, but got", game.EnumNamesGameParameter[e.parameter], sign, bound)
}

type ListFilterError struct {
	parameter game.GameParameter
	LowBound  int16
	HighBound int16
	StartVal  int16
	EndVal    int16
}

func (e *ListFilterError) Error() string {
	var highBound string
	switch e.parameter {
	case game.Win:
		highBound = fmt.Sprintf("%d(min of row and col length)", e.HighBound)
	default:
		highBound = strconv.FormatInt(int64(e.HighBound), 10)
	}
	return fmt.Sprintf("%s boundaries must be between %d and %s, but got %d..%d", game.EnumNamesGameParameter[e.parameter], e.LowBound, highBound, e.StartVal, e.EndVal)
}

type gameLobbyError struct {
	msg string
}

func NewGameLobbyError(message string) error {
	return &gameLobbyError{message}
}

func (e gameLobbyError) Error() string {
	return e.msg
}
