package game

import "github.com/izgib/tttserver/game/models"

type GameRecorder interface {
	GetID() int16
	DeleteGameRecord(ID int16) error
	RecordMove(move models.Move) error
	RecordStatus(status GameEndStatus) error
}

type GameEndStatus byte

const (
	Tie           GameEndStatus = 0
	XWon          GameEndStatus = 1
	OWon          GameEndStatus = 2
	XDisconnected GameEndStatus = 3
	ODisconnected GameEndStatus = 4
	XCheated      GameEndStatus = 5
	OCheated      GameEndStatus = 6
)
