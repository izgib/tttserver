package game

import (
	"github.com/izgib/tttserver/game/models"
)

type GameController interface {
	GetCreatorComChannels() PlayerComm
	GetOpponentComChannels() PlayerComm
	Start()
	GetOpponentMark() models.PlayerMark
	GetCreatorMark() models.PlayerMark
}

type PlayerComm struct {
	CommChan          chan models.Move
	MoveRequesterChan chan bool
	StateChan         chan models.GameState
	InterruptChan     <-chan Interruption
	ErrChan           chan<- error
	TypeChan          chan ObjType
}

type Filter struct {
	Start int16
	End   int16
}

type Interruption int8

const (
	Disconnect Interruption = iota
	Leave
	OppInvalidMove
	InvalidMove
)

var EnumInterruption = map[Interruption]string{
	Disconnect:     "Disconnected",
	Leave:          "Leave",
	OppInvalidMove: "OppInvalidMove",
	InvalidMove:    "InvalidMove",
}

type ObjType int8

const (
	MoveCh ObjType = iota
	StateCh
	InterruptCh
)
