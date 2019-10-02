package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/izgib/tttserver/game/interface/rpc_service/i9e"
)

type Move struct {
	I int16
	J int16
}
type MoveChoice int8

const (
	// Cross представляет крестик
	Cross = MoveChoice(i9e.MarkTypeFilterCross)
	// Nought представляет нолик
	Nought = MoveChoice(i9e.MarkTypeFilterNought)
	// Empty представляет пустое поле
	Empty = MoveChoice(i9e.MarkTypeFilterAny)
)

var EnumNamesMoveChoice = map[MoveChoice]string{
	Cross:  "X",
	Nought: "O",
	Empty:  " ",
}

var plMark = [2]PlayerMark{CrossMark, NoughtMark}

type GameStateType int8

const (
	Running GameStateType = iota
	Tie
	Won
)

var EnumNamesGameStateType = map[GameStateType]string{
	Running: "Running",
	Tie:     "Tie",
	Won:     "Won",
}

type GameState struct {
	StateType GameStateType
	WinLine   *WinLine
}

type PlayerMark int8

const (
	CrossMark  = PlayerMark(i9e.MarkTypeCross)
	NoughtMark = PlayerMark(i9e.MarkTypeNought)
)

type WinLine struct {
	Mark  PlayerMark
	Start Move
	End   Move
}

type iterationPair struct {
	start  Move
	iStep  int16
	jStep  int16
	length int16
}

var GameRules = map[GameParameter]struct {
	Start int16
	End   int16
}{
	Rows: {3, 15},
	Cols: {3, 15},
	Win:  {3, 15},
}

type GameSettings struct {
	Rows int16
	Cols int16
	Win  int16
}

type Game struct {
	settings  GameSettings
	turn      int16
	gameField [][]MoveChoice
}

func NewGame(settings GameSettings) *Game {
	sl := make([][]MoveChoice, settings.Rows)
	for i := range sl {
		sl[i] = make([]MoveChoice, settings.Cols)
		for j := int16(0); j < settings.Cols; j++ {
			sl[i][j] = Empty
		}
	}
	return &Game{
		settings:  settings,
		turn:      0,
		gameField: sl,
	}
}

// GameState checks game state, if state is running, going to next turn
func (g *Game) GameState(move Move) GameState {
	winLine := g.isWon(MoveChoice(g.getPlayerMark(g.CurPlayer())), move.I, move.J)
	if winLine != nil {
		return GameState{StateType: Won, WinLine: winLine}
	}
	if g.settings.Cols*g.settings.Rows-1 <= g.turn {
		return GameState{StateType: Tie, WinLine: nil}
	}
	g.turn++
	return GameState{StateType: Running, WinLine: nil}

}

func (g *Game) CurPlayer() int16 {
	return g.turn & 1
}

func (g *Game) OtherPlayer() int16 {
	return (g.turn + 1) & 1
}

func min(a int16, b int16) int16 {
	if a > b {
		return b
	}
	return a
}

func (g *Game) lineIterator(start Move, iStep int16, jStep int16) func() Move {
	i := start.I
	j := start.J
	return func() Move {
		defer func() {
			i += iStep
			j += jStep
		}()
		return Move{i, j}
	}
}

func (g *Game) isWon(mark MoveChoice, i int16, j int16) *WinLine {
	var markType PlayerMark
	if mark == Cross {
		markType = CrossMark
	} else if mark == Nought {
		markType = NoughtMark
	}
	leftStep := min(j, g.settings.Win-1)
	rightStep := min((g.settings.Cols-1)-j, g.settings.Win-1)
	topStep := min(i, g.settings.Win-1)
	bottomStep := min((g.settings.Rows-1)-i, g.settings.Win-1)

	lines := [4]iterationPair{
		// horizontal
		{Move{i, j - leftStep}, 0, 1, rightStep + 1 + leftStep},
		// vertical
		{Move{i - topStep, j}, 1, 0, topStep + 1 + bottomStep},
		// main diagonal
		{Move{i - min(leftStep, topStep), j - min(leftStep, topStep)},
			1,
			1,
			min(leftStep, topStep) + 1 + min(rightStep, bottomStep)},
		// anti diagonal
		{Move{i + min(leftStep, bottomStep), j - min(leftStep, bottomStep)},
			-1,
			1,
			min(leftStep, bottomStep) + 1 + min(rightStep, topStep)},
	}

	for _, iterParams := range lines {
		var lineStart *Move = nil
		length := int16(0)
		if iterParams.length >= g.settings.Win {
			iterator := g.lineIterator(iterParams.start, iterParams.iStep, iterParams.jStep)
			for k := int16(0); k < iterParams.length; k++ {
				coord := iterator()
				if g.gameField[coord.I][coord.J] == mark {
					if lineStart == nil {
						lineStart = &coord
					}
					length++
					if length >= g.settings.Win {
						return &WinLine{Mark: markType, Start: *lineStart, End: coord}
					}
				} else {
					lineStart = nil
					length = 0
				}
			}
		}
	}

	return nil
}

func (g *Game) getPlayerMark(player int16) PlayerMark {
	return plMark[player]
}

// MoveTo place mark on the field, if cell is not empty return error
func (g *Game) MoveTo(move Move) error {
	mark := g.getPlayerMark(g.CurPlayer())

	var err error = nil
	if (0 <= move.I && move.I < g.settings.Rows) && (0 <= move.J && move.J < g.settings.Cols) {
		if g.gameField[move.I][move.J] == Empty {
			g.gameField[move.I][move.J] = MoveChoice(mark)
		} else {
			err = &TakenGameError{move}
		}
	} else {
		err = &OutsideGameError{move, g.settings.Rows, g.settings.Cols}
	}
	return err
}

func CreateGameDebugLogger(ID int16) *zerolog.Logger {
	writer := zerolog.NewConsoleWriter()
	writer.TimeFormat = time.RFC3339
	writer.FormatCaller = func(i interface{}) string {
		if i == nil {
			return ""
		}
		value := fmt.Sprintf("%v", i)
		// далее использутся только репозитории из github и gitlab, отщивыаем всё вплоть до этих слов
		pos := strings.Index(value, "github.com")
		if pos < 0 {
			pos = strings.Index(value, "gitlab.")
			if pos >= 0 {
				value = value[pos:]
			}
		} else {
			value = value[pos:]
		}
		return "(" + value + ")"
	}
	writer.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("\033[1m%v\033[0m", i)
	}
	writer.FormatTimestamp = func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("\033[33;1m%v\033[0m", i)
	}
	writer.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("\033[35m%s\033[0m", i)
	}
	writer.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("[%s]", i)
	}
	writer.FormatErrFieldName = func(i interface{}) string {
		return fmt.Sprintf("\033[31m%s\033[0m", i)
	}
	writer.FormatErrFieldValue = func(i interface{}) string {
		return fmt.Sprintf("\033[31m[%v]\033[0m", i)
	}
	logger := zerolog.New(writer).Level(zerolog.DebugLevel).With().Int16("game", ID).Timestamp().Logger()
	return &logger
}

var isTakerFmt = "game cell(%d,%d) is not empty"

type TakenGameError struct {
	Move Move
}

func (e *TakenGameError) Error() string {
	return fmt.Sprintf("game cell(%d,%d) is not empty", e.Move.I, e.Move.J)
}

var outOfFieldFmt = "gameField is (%d,%d), got move to (%d,%d)"

type OutsideGameError struct {
	Move Move
	rows int16
	cols int16
}

func NewOutOfFieldError(move Move, rows int16, cols int16) error {
	return &OutsideGameError{move, rows, cols}
}

func (e *OutsideGameError) Error() string {
	return fmt.Sprintf("gameField is (%d,%d), got move to (%d,%d)", e.rows, e.cols, e.Move.I, e.Move.J)
}

type GameParameter uint

const (
	Rows GameParameter = iota
	Cols
	Win
)

var EnumNamesGameParameter = map[GameParameter]string{
	Rows: "rows",
	Cols: "cols",
	Win:  "win",
}
