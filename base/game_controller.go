package base

import (
	"github.com/izgib/tttserver/game"
)

type GameController interface {
	GetCreatorChannels() *CommChannels
	GetOpponentChannels() *CommChannels
	Start() error
	GetOpponentMark() game.PlayerMark
	GetCreatorMark() game.PlayerMark
}

type CommChannels struct {
	Actions   chan Action
	Reactions chan Reaction
	Done      chan struct{}
	Err       error
}

type Filter struct {
	Start int16
	End   int16
}

type InterruptionCause int32

const (
	Disconnect InterruptionCause = iota + 1000
	Leave
	OppInvalidMove
	InvalidMove
)

var EnumInterruption = map[InterruptionCause]string{
	Disconnect:     "Disconnected",
	Leave:          "Leave",
	OppInvalidMove: "OppInvalidMove",
	InvalidMove:    "InvalidMove",
}

type Action interface {
	isAction()
}

type Step struct {
	Pos game.Move
}

func (n Step) isAction() {}

type ReceiveStep struct {
}

func (r ReceiveStep) isAction() {}

type State struct {
	State game.GameState
}

func (s State) isAction() {}

type Interruption struct {
	Cause InterruptionCause
}

func (i Interruption) isAction() {}

type Reaction interface {
	isReaction()
}

type ReceivedStep struct {
	Pos game.Move
	Err error
}

func (r ReceivedStep) isReaction() {}

type Status struct {
	Err error
}

func (e Status) isReaction() {}
