package base

import "github.com/izgib/tttserver/game"

type GameRecorder interface {
	GetID() int16
	RecordMove(move game.Move) error
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
