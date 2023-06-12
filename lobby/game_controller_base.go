package lobby

import (
	"github.com/izgib/tttserver/game"
)

type CommChannels struct {
	Actions   chan Action
	Reactions chan Reaction
	Done      chan struct{}
	Err       error
}

type Filter struct {
	Start uint16
	End   uint16
}

type InterruptionCause int32

const (
	Disconnect InterruptionCause = iota + 1000
	Leave
	InvalidMove
	GiveUp
	Internal
)

var EnumInterruption = map[InterruptionCause]string{
	Disconnect:  "Disconnected",
	Leave:       "Leave",
	InvalidMove: "InvalidMove",
	Internal:    "Internal",
	GiveUp:      "GiveUp",
}

type Action interface {
	isAction()
}

type Step struct {
	Pos   *game.Move
	State game.GameState
}

func (n Step) isAction() {}

type ReceiveMove struct{}

func (r ReceiveMove) isAction() {}

type StepState struct {
	State game.GameState
}

func (s StepState) isAction() {}

type Interruption struct {
	Cause InterruptionCause
}

func (i Interruption) isAction() {}

func (i Interruption) isReaction() {}

type Reaction interface {
	isReaction()
}

type ReceivedMove struct {
	Pos *game.Move
	Err error
}

func (r ReceivedMove) isReaction() {}

type Status struct {
	Err error
}

func (e Status) isReaction() {}

type playerError struct {
	error
	player int16
}

/*type GameController interface {
	CreatorChannels() *CommChannels
	OpponentChannels() *CommChannels
	Start() error
	OpponentMark() game.PlayerMark
	creatorMark() game.PlayerMark
}*/
