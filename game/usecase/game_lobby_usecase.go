package usecase

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/rs/zerolog"

	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/game/models"
)

type gameLobbyUsecase struct {
	mu            sync.Locker
	recordCreator func(config game.GameConfiguration) game.GameRecorder
	lobbies       map[int16]game.GameLobby
	logger        *zerolog.Logger
}

func NewGameLobbyUsecase(recordCreator func(config game.GameConfiguration) game.GameRecorder, logger *zerolog.Logger) game.GameLobbyUsecase {
	return &gameLobbyUsecase{recordCreator: recordCreator, logger: logger, lobbies: make(map[int16]game.GameLobby), mu: &sync.Mutex{}}
}

func (g *gameLobbyUsecase) CreateLobby(config game.GameConfiguration) (game.GameLobby, error) {
	g.logger.Debug().
		Int16("Rows", config.Settings.Rows).
		Int16("Cols", config.Settings.Cols).
		Int16("Win", config.Settings.Win).
		Str("Mark", models.EnumNamesMoveChoice[models.MoveChoice(config.Mark)]).
		Msg("request to create lobby")
	if config.Settings.Rows < models.GameRules[models.Rows].Start {
		err := &GameCreationError{
			parameter: models.Rows,
			Bound:     models.GameRules[models.Rows].Start,
			errorType: BelowRange,
			Value:     config.Settings.Rows,
		}
		return nil, err
	}
	if models.GameRules[models.Rows].End < config.Settings.Rows {
		err := &GameCreationError{
			parameter: models.Rows,
			Bound:     models.GameRules[models.Rows].End,
			errorType: AboveRange,
			Value:     config.Settings.Rows,
		}
		return nil, err
	}

	if config.Settings.Cols < models.GameRules[models.Cols].Start {
		err := &GameCreationError{
			parameter: models.Cols,
			Bound:     models.GameRules[models.Cols].Start,
			errorType: BelowRange,
			Value:     config.Settings.Cols,
		}
		return nil, err
	}
	if models.GameRules[models.Cols].End < config.Settings.Cols {
		err := &GameCreationError{
			parameter: models.Cols,
			Bound:     models.GameRules[models.Cols].End,
			errorType: AboveRange,
			Value:     config.Settings.Cols,
		}
		return nil, err
	}

	if config.Settings.Win < models.GameRules[models.Win].Start {
		err := &GameCreationError{
			parameter: models.Win,
			Bound:     models.GameRules[models.Win].Start,
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
			parameter: models.Win,
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

	return lobby, nil
}

func (g *gameLobbyUsecase) ListLobbies(filter game.GameFilter) ([]game.GameConfiguration, error) {
	var err error = nil
	rows := filter.Rows
	cols := filter.Cols
	win := filter.Win
	mark := filter.Mark
	wrongOrder := fmt.Errorf("start must be <= end")

	if rows.Start > rows.End {
		return nil, wrongOrder
	}
	if !(models.GameRules[models.Rows].Start <= rows.Start && rows.End <= models.GameRules[models.Rows].End) {
		err = &ListFilterError{
			parameter: models.Rows,
			LowBound:  models.GameRules[models.Rows].Start,
			HighBound: models.GameRules[models.Rows].End,
			StartVal:  rows.Start,
			EndVal:    rows.End,
		}
		return nil, err
	}

	if cols.Start > cols.End {
		return nil, wrongOrder
	}
	if !(models.GameRules[models.Cols].Start <= cols.Start && cols.End <= models.GameRules[models.Cols].End) {
		err = &ListFilterError{
			parameter: models.Cols,
			LowBound:  models.GameRules[models.Cols].Start,
			HighBound: models.GameRules[models.Cols].End,
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
	if !(models.GameRules[models.Win].Start <= win.Start && win.End <= highBound) {
		err = &ListFilterError{
			parameter: models.Win,
			LowBound:  models.GameRules[models.Cols].Start,
			HighBound: models.GameRules[models.Cols].End,
			StartVal:  cols.Start,
			EndVal:    cols.End,
		}
		return nil, err
	}

	games := []game.GameConfiguration{}
	for _, v := range g.lobbies {
		if (rows.Start <= v.GetSettings().Rows && v.GetSettings().Rows <= rows.End) &&
			(cols.Start <= v.GetSettings().Cols && v.GetSettings().Cols <= cols.End) &&
			(win.Start <= v.GetSettings().Win && v.GetSettings().Win <= win.End) {
			if mark == models.Empty || mark == models.MoveChoice(v.GetCreatorMark()) {
				g.logger.Debug().Int16("game lobby", v.GetID())
				games = append(games,
					game.GameConfiguration{
						ID:       v.GetID(),
						Settings: v.GetSettings(),
						Mark:     v.GetCreatorMark(),
					})
			}
		}
	}

	return games, nil
}

func (g *gameLobbyUsecase) JoinLobby(ID int16) (game.GameLobby, error) {
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

func (g *gameLobbyUsecase) DeleteLobby(lobby game.GameLobby) {
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
	parameter models.GameParameter
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
		if e.parameter == models.Win {
			bound = fmt.Sprintf("%d(min of row and col length)", e.Bound)
		} else {
			bound = strconv.FormatInt(int64(e.Bound), 10)
		}
	}
	return fmt.Sprintf("%s must be %s %s, but got", models.EnumNamesGameParameter[e.parameter], sign, bound)
}

type ListFilterError struct {
	parameter models.GameParameter
	LowBound  int16
	HighBound int16
	StartVal  int16
	EndVal    int16
}

func (e *ListFilterError) Error() string {
	var highBound string
	switch e.parameter {
	case models.Win:
		highBound = fmt.Sprintf("%d(min of row and col length)", e.HighBound)
	default:
		highBound = strconv.FormatInt(int64(e.HighBound), 10)
	}
	return fmt.Sprintf("%s boundaries must be between %d and %s, but got %d..%d", models.EnumNamesGameParameter[e.parameter], e.LowBound, highBound, e.StartVal, e.EndVal)
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
