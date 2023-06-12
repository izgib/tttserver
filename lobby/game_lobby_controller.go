package lobby

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/rs/zerolog"

	"github.com/izgib/tttserver/game"
)

type GameLobbyController struct {
	mu            sync.Locker
	recordCreator func(config GameConfiguration) GameRecorder
	lobbies       map[uint32]*GameLobby
	logger        *zerolog.Logger
}

func NewGameLobbyController(recordCreator func(config GameConfiguration) GameRecorder, logger *zerolog.Logger) GameLobbyController {
	return GameLobbyController{recordCreator: recordCreator, logger: logger, lobbies: make(map[uint32]*GameLobby), mu: &sync.Mutex{}}
}

func (g *GameLobbyController) CreateLobby(config GameConfiguration) (*GameLobby, error) {
	g.logger.Debug().
		Uint16("Rows", config.Settings.Rows).
		Uint16("Cols", config.Settings.Cols).
		Uint16("Win", config.Settings.Win).
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
	lobby := NewGameLobby(gameRecorder.ID(), config.Settings, config.Mark, gameRecorder)
	g.mu.Lock()
	g.lobbies[lobby.ID] = &lobby
	g.mu.Unlock()

	go func() {
		logger := g.logger.With().Uint32("session", lobby.ID).Logger()
		logger.Debug().Msg("before start")
		if err := lobby.Start(&logger); err != nil {
			(&logger).Info().Int("player games", lobby.gameController.gameCount).Err(err).Msg("ended")
		}
		g.mu.Lock()
		delete(g.lobbies, lobby.ID)
		g.mu.Unlock()
	}()

	return &lobby, nil
}

func (g *GameLobbyController) ListLobbies(filter GameFilter) ([]GameConfiguration, error) {
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

	var games []GameConfiguration
	for _, v := range g.lobbies {
		if (rows.Start <= v.Settings().Rows && v.Settings().Rows <= rows.End) &&
			(cols.Start <= v.Settings().Cols && v.Settings().Cols <= cols.End) &&
			(win.Start <= v.Settings().Win && v.Settings().Win <= win.End) {
			if mark == game.Empty || mark == game.MoveChoice(v.CreatorMark()) {
				g.logger.Debug().Uint32("game lobby", v.ID)
				games = append(games,
					GameConfiguration{
						ID:       v.ID,
						Settings: v.Settings(),
						Mark:     v.CreatorMark(),
					})
			}
		}
	}

	return games, nil
}

func (g *GameLobbyController) JoinLobby(ID uint32) (*GameLobby, error) {
	g.logger.Debug().Uint32("game", ID).Msg("request to join lobby")
	if gameLobby, ok := g.lobbies[ID]; ok {
		return gameLobby, nil
	} else {
		errString := fmt.Sprintf("GameConfiguration %d was not found", ID)
		err := NewGameLobbyError(errString)
		g.logger.Debug().Err(err)
		return nil, err
	}
}

func (g *GameLobbyController) DeleteLobby(lobby GameLobby) {
	g.mu.Lock()
	delete(g.lobbies, lobby.ID)
	g.mu.Unlock()
}

type GameCreationErrorType uint

const (
	BelowRange GameCreationErrorType = iota
	AboveRange
)

type GameCreationError struct {
	parameter game.GameParameter
	Bound     uint16
	errorType GameCreationErrorType
	Value     uint16
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
	LowBound  uint16
	HighBound uint16
	StartVal  uint16
	EndVal    uint16
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
