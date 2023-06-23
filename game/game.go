package game

import (
	"fmt"
	"github.com/izgib/tttserver/internal/logger"
	"github.com/rs/zerolog"
)

type Move struct {
	I uint16
	J uint16
}
type MoveChoice int8

const (
	Cross  = MoveChoice(CrossMark)
	Nought = MoveChoice(NoughtMark)
	Empty  = MoveChoice(2)
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
	CrossMark  = PlayerMark(0)
	NoughtMark = PlayerMark(1)
)

type WinLine struct {
	Mark  PlayerMark
	Start *Move
	End   *Move
}

type iterationPair struct {
	start  Move
	iStep  int16
	jStep  int16
	length uint16
}

var GameRules = map[GameParameter]struct {
	Start uint16
	End   uint16
}{
	Rows: {3, 15},
	Cols: {3, 15},
	Win:  {3, 15},
}

type GameSettings struct {
	Rows uint16
	Cols uint16
	Win  uint16
}

type Game struct {
	settings  GameSettings
	turn      uint16
	gameField [][]MoveChoice
}

func NewGame(settings GameSettings) *Game {
	sl := make([][]MoveChoice, settings.Rows)
	for i := range sl {
		sl[i] = make([]MoveChoice, settings.Cols)
		for j := uint16(0); j < uint16(settings.Cols); j++ {
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
	winLine := g.isWon(MoveChoice(g.GetPlayerMark(g.CurPlayer())), move.I, move.J)
	if winLine != nil {
		return GameState{StateType: Won, WinLine: winLine}
	}
	if g.settings.Cols*g.settings.Rows-1 <= g.turn {
		return GameState{StateType: Tie, WinLine: nil}
	}
	g.turn++
	return GameState{StateType: Running, WinLine: nil}
}

func (g *Game) Settings() GameSettings {
	return g.settings
}

func (g *Game) Restart() {
	g.turn = 0
	for i := uint16(0); i < g.settings.Rows; i++ {
		for j := uint16(0); j < g.settings.Cols; j++ {
			g.gameField[i][j] = Empty
		}
	}
}

func (g *Game) CurPlayer() uint16 {
	return g.turn & 1
}

func (g *Game) OtherPlayer() uint16 {
	return (g.turn + 1) & 1
}

func min(a uint16, b uint16) uint16 {
	if a > b {
		return b
	}
	return a
}

func (g *Game) lineIterator(start Move, iStep int16, jStep int16) func() Move {
	i := int16(start.I)
	j := int16(start.J)
	return func() Move {
		defer func() {
			i += iStep
			j += jStep
		}()
		return Move{uint16(i), uint16(j)}
	}
}

func (g *Game) isWon(mark MoveChoice, i uint16, j uint16) *WinLine {
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
		length := uint16(0)
		if iterParams.length >= g.settings.Win {
			iterator := g.lineIterator(iterParams.start, iterParams.iStep, iterParams.jStep)
			for k := uint16(0); k < iterParams.length; k++ {
				coord := iterator()
				if g.gameField[coord.I][coord.J] == mark {
					if lineStart == nil {
						lineStart = &coord
					}
					length++
					if length >= g.settings.Win {
						return &WinLine{Mark: markType, Start: lineStart, End: &coord}
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

func (g *Game) GetPlayerMark(player uint16) PlayerMark {
	return plMark[player]
}

// MoveTo place mark on the field, if cell is not empty return error
func (g *Game) MoveTo(move Move) error {
	mark := g.GetPlayerMark(g.CurPlayer())

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

func CreateGameDebugLogger(ID int64) *zerolog.Logger {
	logger := logger.CreateDebugLogger().With().Int64("game", ID).Timestamp().Logger()
	return &logger
}

type TakenGameError struct {
	Move Move
}

func (e *TakenGameError) Error() string {
	return fmt.Sprintf("game cell(%d,%d) is not empty", e.Move.I, e.Move.J)
}

type OutsideGameError struct {
	Move Move
	rows uint16
	cols uint16
}

func NewOutOfFieldError(move Move, rows uint16, cols uint16) error {
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
