package game

import (
	"reflect"
	"testing"
)

type fields struct {
	settings  GameSettings
	turn      uint16
	gameField [][]MoveChoice
}

var winFields = []fields{
	{settings: GameSettings{Rows: 3, Cols: 3, Win: 3},
		gameField: [][]MoveChoice{
			{Cross, Nought, Empty},
			{Cross, Nought, Empty},
			{Cross, Empty, Empty},
		},
		turn: 4,
	},
	{settings: GameSettings{Rows: 3, Cols: 3, Win: 3},
		gameField: [][]MoveChoice{
			{Nought, Nought, Nought},
			{Cross, Nought, Cross},
			{Cross, Empty, Cross},
		}},
	{settings: GameSettings{Rows: 3, Cols: 3, Win: 3},
		gameField: [][]MoveChoice{
			{Cross, Nought, Nought},
			{Cross, Cross, Empty},
			{Nought, Empty, Cross},
		}},
	{settings: GameSettings{Rows: 3, Cols: 3, Win: 3},
		gameField: [][]MoveChoice{
			{Cross, Nought, Nought},
			{Cross, Nought, Cross},
			{Nought, Empty, Cross},
		}},
	{settings: GameSettings{Rows: 3, Cols: 3, Win: 3},
		gameField: [][]MoveChoice{
			{Cross, Empty, Nought},
			{Cross, Empty, Nought},
			{Cross, Empty, Empty},
		}},
	{settings: GameSettings{Rows: 3, Cols: 3, Win: 3},
		gameField: [][]MoveChoice{
			{Cross, Nought, Cross},
			{Cross, Cross, Nought},
			{Nought, Nought, Cross},
		}},
}

var otherFields = []fields{
	{settings: GameSettings{Rows: 3, Cols: 3, Win: 3},
		gameField: [][]MoveChoice{
			{Nought, Cross, Cross},
			{Cross, Cross, Nought},
			{Nought, Nought, Cross},
		},
		turn: 8,
	},
	{settings: GameSettings{Rows: 3, Cols: 3, Win: 3},
		gameField: [][]MoveChoice{
			{Nought, Cross, Cross},
			{Cross, Empty, Nought},
			{Nought, Nought, Cross},
		},
		turn: 7,
	},
}

func TestGame_isWon(t *testing.T) {
	type args struct {
		mark MoveChoice
		i    uint16
		j    uint16
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *WinLine
	}{
		{"IsEndedVertical1",
			winFields[0],
			args{Cross, 0, 0},
			&WinLine{CrossMark, &Move{0, 0}, &Move{2, 0}},
		},
		{"IsEndedVertical2",
			winFields[0],
			args{Cross, 1, 0},
			&WinLine{CrossMark, &Move{0, 0}, &Move{2, 0}},
		},
		{"IsEndedHorizontal1",
			winFields[1],
			args{Nought, 0, 0},
			&WinLine{NoughtMark, &Move{0, 0}, &Move{0, 2}},
		},
		{"IsEndedHorizontal2",
			winFields[1],
			args{Nought, 0, 1},
			&WinLine{NoughtMark, &Move{0, 0}, &Move{0, 2}},
		},
		{"IsEndedMainDiagonal1",
			winFields[2],
			args{Cross, 0, 0},
			&WinLine{CrossMark, &Move{0, 0}, &Move{2, 2}},
		},
		{"IsEndedMainDiagonal2",
			winFields[2],
			args{Cross, 1, 1},
			&WinLine{CrossMark, &Move{0, 0}, &Move{2, 2}},
		},
		{"IsEndedAntiDiagonal1",
			winFields[3],
			args{Nought, 2, 0},
			&WinLine{NoughtMark, &Move{2, 0}, &Move{0, 2}},
		},
		{"IsEndedAntiDiagonal2",
			winFields[3],
			args{Nought, 1, 1},
			&WinLine{NoughtMark, &Move{2, 0}, &Move{0, 2}},
		},
		{"IsEndedRow",
			winFields[4],
			args{Cross, 2, 0},
			&WinLine{CrossMark, &Move{0, 0}, &Move{2, 0}},
		},
		{"IsEndedFull",
			winFields[5],
			args{Cross, 2, 2},
			&WinLine{CrossMark, &Move{0, 0}, &Move{2, 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				settings:  tt.fields.settings,
				turn:      tt.fields.turn,
				gameField: tt.fields.gameField,
			}
			res := g.isWon(tt.args.mark, tt.args.i, tt.args.j)
			t.Log(res)
			if got := g.isWon(tt.args.mark, tt.args.i, tt.args.j); !reflect.DeepEqual(got, tt.want) {
				// t.Logf("%+v", got)
				t.Log(got)
				t.Errorf("Game.isWon() = %v, endState %v", got, tt.want)
			}
		})
	}
}

func TestGame_GameState(t *testing.T) {
	type args struct {
		move Move
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   GameState
	}{
		{"IsWonVertical",
			winFields[0],
			args{Move{0, 0}},
			GameState{
				StateType: Won,
				WinLine:   &WinLine{CrossMark, &Move{0, 0}, &Move{2, 0}},
			},
		},
		{"IsWonFilledGameField",
			winFields[5],
			args{Move{2, 2}},
			GameState{
				StateType: Won,
				WinLine:   &WinLine{CrossMark, &Move{0, 0}, &Move{2, 2}},
			},
		},
		{"IsTie",
			otherFields[0],
			args{Move{2, 2}},
			GameState{
				StateType: Tie,
				WinLine:   nil,
			},
		},
		{"IsRunning",
			otherFields[1],
			args{Move{0, 0}},
			GameState{
				StateType: Running,
				WinLine:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				settings:  tt.fields.settings,
				turn:      tt.fields.turn,
				gameField: tt.fields.gameField,
			}
			if got := g.GameState(tt.args.move); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GameState() = %v, endState %v", got, tt.want)
			}
		})
	}
}

func TestGame_GameExample(t *testing.T) {
	tests := []struct {
		name     string
		game     *Game
		moves    []Move
		endState GameState
	}{
		{"IsCrossWonHorizontal",
			NewGame(GameSettings{3, 3, 3}),
			[]Move{{1, 1}, {1, 2}, {0, 2}, {2, 0}, {0, 0}, {2, 2}, {0, 1}},
			GameState{
				StateType: Won,
				WinLine:   &WinLine{CrossMark, &Move{0, 0}, &Move{0, 2}},
			},
		},
		{"IsNoughtWonVertical",
			NewGame(GameSettings{3, 3, 3}),
			[]Move{{1, 1}, {0, 2}, {0, 0}, {2, 2}, {2, 0}, {1, 2}},
			GameState{
				StateType: Won,
				WinLine:   &WinLine{NoughtMark, &Move{0, 2}, &Move{2, 2}},
			},
		},
		{"IsTied",
			NewGame(GameSettings{3, 3, 3}),
			[]Move{{1, 1}, {0, 2}, {0, 0}, {2, 2}, {1, 2}, {1, 0}, {2, 1}, {0, 1}, {2, 0}},
			GameState{
				StateType: Tie,
				WinLine:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var state GameState
			for _, move := range tt.moves {
				tt.game.MoveTo(move)
				state = tt.game.GameState(move)
				switch state.StateType {
				case Running:
					continue
				default:
					break
				}
			}
			if !reflect.DeepEqual(state, tt.endState) {
				t.Errorf("GameState() = %v, endState %v", state, tt.endState)
			}
		})
	}
}

func TestGame_RestartTest(t *testing.T) {
	t.Run("GameRestart after Tie", func(t *testing.T) {
		moves := []Move{{1, 1}, {0, 2}, {0, 0}, {2, 2}, {1, 2}, {1, 0}, {2, 1}, {0, 1}, {2, 0}}
		game := NewGame(GameSettings{Rows: 3, Cols: 3, Win: 3})
		endState := GameState{StateType: Tie, WinLine: nil}
		var state GameState
		for _, move := range moves {
			game.MoveTo(move)
			state = game.GameState(move)
			switch state.StateType {
			case Running:
				continue
			default:
				break
			}
		}
		if !reflect.DeepEqual(state, endState) {
			t.Errorf("GameState() = %v, endState %v", state, endState)
		}
		game.Restart()
		t.Logf("after game restart")
		for _, move := range moves {
			game.MoveTo(move)
			state = game.GameState(move)
			switch state.StateType {
			case Running:
				continue
			default:
				break
			}
		}
		if !reflect.DeepEqual(state, endState) {
			t.Errorf("GameState() = %v, endState %v", state, endState)
		}
	})
}
