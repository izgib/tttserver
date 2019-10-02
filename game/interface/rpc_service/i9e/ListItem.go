// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package i9e

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ListItem struct {
	_tab flatbuffers.Table
}

func GetRootAsListItem(buf []byte, offset flatbuffers.UOffsetT) *ListItem {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ListItem{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *ListItem) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ListItem) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *ListItem) ID() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *ListItem) MutateID(n int16) bool {
	return rcv._tab.MutateInt16Slot(4, n)
}

func (rcv *ListItem) Params(obj *GameParams) *GameParams {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(GameParams)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func ListItemStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func ListItemAddID(builder *flatbuffers.Builder, ID int16) {
	builder.PrependInt16Slot(0, ID, 0)
}
func ListItemAddParams(builder *flatbuffers.Builder, params flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(params), 0)
}
func ListItemEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}