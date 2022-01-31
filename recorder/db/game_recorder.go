package db

import (
	"fmt"
	"github.com/izgib/tttserver/base"
	"github.com/izgib/tttserver/game"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	host     = "db"
	user     = "postgres"
	password = "postgres"
	dbname   = "tic_tac_toe"
)

type gameStatus string

const (
	XWon          gameStatus = "x_won"
	OWon          gameStatus = "o_won"
	Tie           gameStatus = "tie"
	XDisconnected gameStatus = "x_disconnected"
	ODisconnected gameStatus = "o_disconnected"
	XCheated      gameStatus = "x_cheated"
	OCheated      gameStatus = "o_cheated"
)

type creatorMark string

const (
	Cross  creatorMark = "cross"
	Nought creatorMark = "nought"
)

var EnumValuesCreatorMark = map[creatorMark]game.PlayerMark{
	Cross:  game.CrossMark,
	Nought: game.NoughtMark,
}

var EnumValuesPlayerMark = map[game.PlayerMark]creatorMark{
	game.CrossMark:  Cross,
	game.NoughtMark: Nought,
}

var EnumValuesGameStatus = map[base.GameEndStatus]gameStatus{
	base.Tie:           Tie,
	base.XWon:          XWon,
	base.OWon:          OWon,
	base.XDisconnected: XDisconnected,
	base.ODisconnected: ODisconnected,
	base.XCheated:      XCheated,
	base.OCheated:      OCheated,
}

type gameRecorder struct {
	db     *sqlx.DB
	gameId int16
}

func (r *gameRecorder) GetID() int16 {
	return r.gameId
}

func (r *gameRecorder) RecordMove(move game.Move) error {
	mq := `UPDATE game SET moves = array_append(moves, CAST(ROW($1, $2) AS move)) WHERE game_id=$3`
	_, err := r.db.Exec(mq, move.I, move.J, r.gameId)
	return err
}

func (r *gameRecorder) RecordStatus(status base.GameEndStatus) error {
	sq := `UPDATE game SET status=$1 WHERE game_id = $2`
	_, err := r.db.Exec(sq, EnumValuesGameStatus[status], r.gameId)
	return err
}

func NewGameRecorder(config base.GameConfiguration) base.GameRecorder {
	psqlInfo := fmt.Sprintf("host=%s database=%s user=%s password=%s sslmode=disable",
		host, dbname, user, password)
	db := sqlx.MustConnect("pgx", psqlInfo)
	gq := `INSERT INTO game (rows, cols, win, creator_mark) VALUES ($1, $2, $3, $4) RETURNING game_id`
	var gameID int16
	err := db.QueryRowx(gq, config.Settings.Rows, config.Settings.Cols, config.Settings.Win, EnumValuesPlayerMark[config.Mark]).Scan(&gameID)
	if err != nil {
		panic(err)
	}
	return &gameRecorder{
		gameId: gameID,
		db:     db,
	}
}
