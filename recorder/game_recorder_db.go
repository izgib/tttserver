package recorder

import (
	"fmt"
	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/internal/logger"
	"github.com/izgib/tttserver/lobby"

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
	XIllegalMove  gameStatus = "x_illegal_move"
	XLeft         gameStatus = "x_left"
	ODisconnected gameStatus = "o_disconnected"
	OIllegalMove  gameStatus = "o_illegal_move"
	OLeft         gameStatus = "o_left"
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

var EnumValuesGameStatus = map[lobby.GameEndStatus]gameStatus{
	lobby.Tie:           Tie,
	lobby.XWon:          XWon,
	lobby.OWon:          OWon,
	lobby.XDisconnected: XDisconnected,
	lobby.ODisconnected: ODisconnected,
	lobby.XIllegalMove:  XIllegalMove,
	lobby.OIllegalMove:  OIllegalMove,
	lobby.XLeft:         XLeft,
	lobby.OLeft:         OLeft,
}

type gameRecorder struct {
	db     *sqlx.DB
	gameId uint32
}

func (r *gameRecorder) ID() uint32 {
	return r.gameId
}

func (r *gameRecorder) RecordMove(move game.Move) error {
	mq := `UPDATE game_session SET moves = array_append(moves, CAST(ROW($1, $2) AS move)) WHERE game_id=$3`
	_, err := r.db.Exec(mq, move.I, move.J, r.gameId)
	return err
}

func (r *gameRecorder) RecordStatus(status lobby.GameEndStatus) error {
	sq := `UPDATE game_session SET entities = array_append(entities, CAST (ROW(array_length(moves, 1) , $1) AS game_entity)) WHERE game_id = $2`
	_, err := r.db.Exec(sq, EnumValuesGameStatus[status], r.gameId)
	return err
}

func NewGameRecorder(config lobby.GameConfiguration) lobby.GameRecorder {
	psqlInfo := fmt.Sprintf("host=%s database=%s user=%s password=%s sslmode=disable",
		host, dbname, user, password)
	db := sqlx.MustConnect("pgx", psqlInfo)
	gq := `INSERT INTO game_session (rows, cols, win, creator_mark) VALUES ($1, $2, $3, $4) RETURNING game_id`
	var gameID uint32
	err := db.QueryRowx(gq, config.Settings.Rows, config.Settings.Cols, config.Settings.Win, EnumValuesPlayerMark[config.Mark]).Scan(&gameID)
	if err != nil {
		logger.CreateDebugLogger().Err(err).Msg("can not create db entity")
		return nil
	}
	return &gameRecorder{
		gameId: gameID,
		db:     db,
	}
}
