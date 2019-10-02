// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package i9e

import "strconv"

type CreatorReqMsg byte

const (
	CreatorReqMsgNONE       CreatorReqMsg = 0
	CreatorReqMsgMove       CreatorReqMsg = 1
	CreatorReqMsgGameParams CreatorReqMsg = 2
)

var EnumNamesCreatorReqMsg = map[CreatorReqMsg]string{
	CreatorReqMsgNONE:       "NONE",
	CreatorReqMsgMove:       "Move",
	CreatorReqMsgGameParams: "GameParams",
}

var EnumValuesCreatorReqMsg = map[string]CreatorReqMsg{
	"NONE":       CreatorReqMsgNONE,
	"Move":       CreatorReqMsgMove,
	"GameParams": CreatorReqMsgGameParams,
}

func (v CreatorReqMsg) String() string {
	if s, ok := EnumNamesCreatorReqMsg[v]; ok {
		return s
	}
	return "CreatorReqMsg(" + strconv.FormatInt(int64(v), 10) + ")"
}