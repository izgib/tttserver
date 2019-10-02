// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package i9e

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type WinLine struct {
	_tab flatbuffers.Table
}

func GetRootAsWinLine(buf []byte, offset flatbuffers.UOffsetT) *WinLine {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &WinLine{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *WinLine) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *WinLine) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *WinLine) Mark() MarkType {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return MarkType(rcv._tab.GetInt8(o + rcv._tab.Pos))
	}
	return 0
}

func (rcv *WinLine) MutateMark(n MarkType) bool {
	return rcv._tab.MutateInt8Slot(4, int8(n))
}

func (rcv *WinLine) Start(obj *Move) *Move {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(Move)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *WinLine) End(obj *Move) *Move {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(Move)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func WinLineStart(builder *flatbuffers.Builder) {
	builder.StartObject(3)
}
func WinLineAddMark(builder *flatbuffers.Builder, mark MarkType) {
	builder.PrependInt8Slot(0, int8(mark), 0)
}
func WinLineAddStart(builder *flatbuffers.Builder, start flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(start), 0)
}
func WinLineAddEnd(builder *flatbuffers.Builder, end flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(end), 0)
}
func WinLineEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}