// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.12
// source: network.proto

package transport

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ListItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     int64       `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Params *GameParams `protobuf:"bytes,2,opt,name=params,proto3" json:"params,omitempty"`
}

func (x *ListItem) Reset() {
	*x = ListItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListItem) ProtoMessage() {}

func (x *ListItem) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListItem.ProtoReflect.Descriptor instead.
func (*ListItem) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{0}
}

func (x *ListItem) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ListItem) GetParams() *GameParams {
	if x != nil {
		return x.Params
	}
	return nil
}

type Range struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start int32 `protobuf:"varint,1,opt,name=start,proto3" json:"start,omitempty"`
	End   int32 `protobuf:"varint,2,opt,name=end,proto3" json:"end,omitempty"`
}

func (x *Range) Reset() {
	*x = Range{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Range) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Range) ProtoMessage() {}

func (x *Range) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Range.ProtoReflect.Descriptor instead.
func (*Range) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{1}
}

func (x *Range) GetStart() int32 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *Range) GetEnd() int32 {
	if x != nil {
		return x.End
	}
	return 0
}

type GameFilter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Rows *Range   `protobuf:"bytes,1,opt,name=rows,proto3" json:"rows,omitempty"`
	Cols *Range   `protobuf:"bytes,2,opt,name=cols,proto3" json:"cols,omitempty"`
	Win  *Range   `protobuf:"bytes,3,opt,name=win,proto3" json:"win,omitempty"`
	Mark MarkType `protobuf:"varint,4,opt,name=mark,proto3,enum=base.MarkType" json:"mark,omitempty"`
}

func (x *GameFilter) Reset() {
	*x = GameFilter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameFilter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameFilter) ProtoMessage() {}

func (x *GameFilter) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameFilter.ProtoReflect.Descriptor instead.
func (*GameFilter) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{2}
}

func (x *GameFilter) GetRows() *Range {
	if x != nil {
		return x.Rows
	}
	return nil
}

func (x *GameFilter) GetCols() *Range {
	if x != nil {
		return x.Cols
	}
	return nil
}

func (x *GameFilter) GetWin() *Range {
	if x != nil {
		return x.Win
	}
	return nil
}

func (x *GameFilter) GetMark() MarkType {
	if x != nil {
		return x.Mark
	}
	return MarkType_MARK_TYPE_UNSPECIFIED
}

type CreateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Payload:
	//	*CreateRequest_Move
	//	*CreateRequest_Action
	//	*CreateRequest_Params
	Payload isCreateRequest_Payload `protobuf_oneof:"payload"`
}

func (x *CreateRequest) Reset() {
	*x = CreateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateRequest) ProtoMessage() {}

func (x *CreateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateRequest.ProtoReflect.Descriptor instead.
func (*CreateRequest) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{3}
}

func (m *CreateRequest) GetPayload() isCreateRequest_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *CreateRequest) GetMove() *Move {
	if x, ok := x.GetPayload().(*CreateRequest_Move); ok {
		return x.Move
	}
	return nil
}

func (x *CreateRequest) GetAction() ClientAction {
	if x, ok := x.GetPayload().(*CreateRequest_Action); ok {
		return x.Action
	}
	return ClientAction_CLIENT_ACTION_LEAVE
}

func (x *CreateRequest) GetParams() *GameParams {
	if x, ok := x.GetPayload().(*CreateRequest_Params); ok {
		return x.Params
	}
	return nil
}

type isCreateRequest_Payload interface {
	isCreateRequest_Payload()
}

type CreateRequest_Move struct {
	Move *Move `protobuf:"bytes,1,opt,name=move,proto3,oneof"`
}

type CreateRequest_Action struct {
	Action ClientAction `protobuf:"varint,2,opt,name=action,proto3,enum=base.ClientAction,oneof"`
}

type CreateRequest_Params struct {
	Params *GameParams `protobuf:"bytes,3,opt,name=params,proto3,oneof"`
}

func (*CreateRequest_Move) isCreateRequest_Payload() {}

func (*CreateRequest_Action) isCreateRequest_Payload() {}

func (*CreateRequest_Params) isCreateRequest_Payload() {}

type CreateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Payload:
	//	*CreateResponse_Status
	//	*CreateResponse_WinLine
	//	*CreateResponse_GameId
	Payload isCreateResponse_Payload `protobuf_oneof:"payload"`
	Move    *Move                    `protobuf:"bytes,5,opt,name=move,proto3,oneof" json:"move,omitempty"`
}

func (x *CreateResponse) Reset() {
	*x = CreateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateResponse) ProtoMessage() {}

func (x *CreateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateResponse.ProtoReflect.Descriptor instead.
func (*CreateResponse) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{4}
}

func (m *CreateResponse) GetPayload() isCreateResponse_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *CreateResponse) GetStatus() GameStatus {
	if x, ok := x.GetPayload().(*CreateResponse_Status); ok {
		return x.Status
	}
	return GameStatus_GAME_STATUS_UNSPECIFIED
}

func (x *CreateResponse) GetWinLine() *WinLine {
	if x, ok := x.GetPayload().(*CreateResponse_WinLine); ok {
		return x.WinLine
	}
	return nil
}

func (x *CreateResponse) GetGameId() int64 {
	if x, ok := x.GetPayload().(*CreateResponse_GameId); ok {
		return x.GameId
	}
	return 0
}

func (x *CreateResponse) GetMove() *Move {
	if x != nil {
		return x.Move
	}
	return nil
}

type isCreateResponse_Payload interface {
	isCreateResponse_Payload()
}

type CreateResponse_Status struct {
	Status GameStatus `protobuf:"varint,1,opt,name=status,proto3,enum=base.GameStatus,oneof"`
}

type CreateResponse_WinLine struct {
	WinLine *WinLine `protobuf:"bytes,2,opt,name=win_line,json=winLine,proto3,oneof"`
}

type CreateResponse_GameId struct {
	GameId int64 `protobuf:"varint,4,opt,name=game_id,json=gameId,proto3,oneof"`
}

func (*CreateResponse_Status) isCreateResponse_Payload() {}

func (*CreateResponse_WinLine) isCreateResponse_Payload() {}

func (*CreateResponse_GameId) isCreateResponse_Payload() {}

type JoinRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Payload:
	//	*JoinRequest_Move
	//	*JoinRequest_Action
	//	*JoinRequest_GameId
	Payload isJoinRequest_Payload `protobuf_oneof:"payload"`
}

func (x *JoinRequest) Reset() {
	*x = JoinRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinRequest) ProtoMessage() {}

func (x *JoinRequest) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinRequest.ProtoReflect.Descriptor instead.
func (*JoinRequest) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{5}
}

func (m *JoinRequest) GetPayload() isJoinRequest_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *JoinRequest) GetMove() *Move {
	if x, ok := x.GetPayload().(*JoinRequest_Move); ok {
		return x.Move
	}
	return nil
}

func (x *JoinRequest) GetAction() ClientAction {
	if x, ok := x.GetPayload().(*JoinRequest_Action); ok {
		return x.Action
	}
	return ClientAction_CLIENT_ACTION_LEAVE
}

func (x *JoinRequest) GetGameId() int64 {
	if x, ok := x.GetPayload().(*JoinRequest_GameId); ok {
		return x.GameId
	}
	return 0
}

type isJoinRequest_Payload interface {
	isJoinRequest_Payload()
}

type JoinRequest_Move struct {
	Move *Move `protobuf:"bytes,1,opt,name=move,proto3,oneof"`
}

type JoinRequest_Action struct {
	Action ClientAction `protobuf:"varint,2,opt,name=action,proto3,enum=base.ClientAction,oneof"`
}

type JoinRequest_GameId struct {
	GameId int64 `protobuf:"varint,3,opt,name=game_id,json=gameId,proto3,oneof"`
}

func (*JoinRequest_Move) isJoinRequest_Payload() {}

func (*JoinRequest_Action) isJoinRequest_Payload() {}

func (*JoinRequest_GameId) isJoinRequest_Payload() {}

type JoinResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Payload:
	//	*JoinResponse_Status
	//	*JoinResponse_WinLine
	Payload isJoinResponse_Payload `protobuf_oneof:"payload"`
	Move    *Move                  `protobuf:"bytes,4,opt,name=move,proto3,oneof" json:"move,omitempty"`
}

func (x *JoinResponse) Reset() {
	*x = JoinResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinResponse) ProtoMessage() {}

func (x *JoinResponse) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinResponse.ProtoReflect.Descriptor instead.
func (*JoinResponse) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{6}
}

func (m *JoinResponse) GetPayload() isJoinResponse_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *JoinResponse) GetStatus() GameStatus {
	if x, ok := x.GetPayload().(*JoinResponse_Status); ok {
		return x.Status
	}
	return GameStatus_GAME_STATUS_UNSPECIFIED
}

func (x *JoinResponse) GetWinLine() *WinLine {
	if x, ok := x.GetPayload().(*JoinResponse_WinLine); ok {
		return x.WinLine
	}
	return nil
}

func (x *JoinResponse) GetMove() *Move {
	if x != nil {
		return x.Move
	}
	return nil
}

type isJoinResponse_Payload interface {
	isJoinResponse_Payload()
}

type JoinResponse_Status struct {
	Status GameStatus `protobuf:"varint,1,opt,name=status,proto3,enum=base.GameStatus,oneof"`
}

type JoinResponse_WinLine struct {
	WinLine *WinLine `protobuf:"bytes,2,opt,name=win_line,json=winLine,proto3,oneof"`
}

func (*JoinResponse_Status) isJoinResponse_Payload() {}

func (*JoinResponse_WinLine) isJoinResponse_Payload() {}

type Interruption struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cause StopCause `protobuf:"varint,1,opt,name=cause,proto3,enum=base.StopCause" json:"cause,omitempty"`
}

func (x *Interruption) Reset() {
	*x = Interruption{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Interruption) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Interruption) ProtoMessage() {}

func (x *Interruption) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Interruption.ProtoReflect.Descriptor instead.
func (*Interruption) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{7}
}

func (x *Interruption) GetCause() StopCause {
	if x != nil {
		return x.Cause
	}
	return StopCause_STOP_CAUSE_UNSPECIFIED
}

var File_network_proto protoreflect.FileDescriptor

var file_network_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x09, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x1a, 0x0a, 0x62, 0x61, 0x73, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x44, 0x0a, 0x08, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x74,
	0x65, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x28, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x10, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x50, 0x61,
	0x72, 0x61, 0x6d, 0x73, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x22, 0x2f, 0x0a, 0x05,
	0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x65,
	0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x65, 0x6e, 0x64, 0x22, 0xa0, 0x01,
	0x0a, 0x0a, 0x47, 0x61, 0x6d, 0x65, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x24, 0x0a, 0x04,
	0x72, 0x6f, 0x77, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x04, 0x72, 0x6f,
	0x77, 0x73, 0x12, 0x24, 0x0a, 0x04, 0x63, 0x6f, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x52, 0x61, 0x6e,
	0x67, 0x65, 0x52, 0x04, 0x63, 0x6f, 0x6c, 0x73, 0x12, 0x22, 0x0a, 0x03, 0x77, 0x69, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72,
	0x74, 0x2e, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x03, 0x77, 0x69, 0x6e, 0x12, 0x22, 0x0a, 0x04,
	0x6d, 0x61, 0x72, 0x6b, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x62, 0x61, 0x73,
	0x65, 0x2e, 0x4d, 0x61, 0x72, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x6d, 0x61, 0x72, 0x6b,
	0x22, 0x96, 0x01, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x20, 0x0a, 0x04, 0x6d, 0x6f, 0x76, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0a, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x4d, 0x6f, 0x76, 0x65, 0x48, 0x00, 0x52, 0x04,
	0x6d, 0x6f, 0x76, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x43, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x2a, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x10, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x50, 0x61,
	0x72, 0x61, 0x6d, 0x73, 0x48, 0x00, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x42, 0x09,
	0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0xbc, 0x01, 0x0a, 0x0e, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x10, 0x2e, 0x62,
	0x61, 0x73, 0x65, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x48, 0x00,
	0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x2a, 0x0a, 0x08, 0x77, 0x69, 0x6e, 0x5f,
	0x6c, 0x69, 0x6e, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x62, 0x61, 0x73,
	0x65, 0x2e, 0x57, 0x69, 0x6e, 0x4c, 0x69, 0x6e, 0x65, 0x48, 0x00, 0x52, 0x07, 0x77, 0x69, 0x6e,
	0x4c, 0x69, 0x6e, 0x65, 0x12, 0x19, 0x0a, 0x07, 0x67, 0x61, 0x6d, 0x65, 0x5f, 0x69, 0x64, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x03, 0x48, 0x00, 0x52, 0x06, 0x67, 0x61, 0x6d, 0x65, 0x49, 0x64, 0x12,
	0x23, 0x0a, 0x04, 0x6d, 0x6f, 0x76, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e,
	0x62, 0x61, 0x73, 0x65, 0x2e, 0x4d, 0x6f, 0x76, 0x65, 0x48, 0x01, 0x52, 0x04, 0x6d, 0x6f, 0x76,
	0x65, 0x88, 0x01, 0x01, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x42,
	0x07, 0x0a, 0x05, 0x5f, 0x6d, 0x6f, 0x76, 0x65, 0x22, 0x83, 0x01, 0x0a, 0x0b, 0x4a, 0x6f, 0x69,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x04, 0x6d, 0x6f, 0x76, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x4d, 0x6f,
	0x76, 0x65, 0x48, 0x00, 0x52, 0x04, 0x6d, 0x6f, 0x76, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x62, 0x61, 0x73,
	0x65, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00,
	0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x19, 0x0a, 0x07, 0x67, 0x61, 0x6d, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x48, 0x00, 0x52, 0x06, 0x67, 0x61, 0x6d,
	0x65, 0x49, 0x64, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0x9f,
	0x01, 0x0a, 0x0c, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x2a, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x10, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x48, 0x00, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x2a, 0x0a, 0x08, 0x77,
	0x69, 0x6e, 0x5f, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e,
	0x62, 0x61, 0x73, 0x65, 0x2e, 0x57, 0x69, 0x6e, 0x4c, 0x69, 0x6e, 0x65, 0x48, 0x00, 0x52, 0x07,
	0x77, 0x69, 0x6e, 0x4c, 0x69, 0x6e, 0x65, 0x12, 0x23, 0x0a, 0x04, 0x6d, 0x6f, 0x76, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x4d, 0x6f, 0x76,
	0x65, 0x48, 0x01, 0x52, 0x04, 0x6d, 0x6f, 0x76, 0x65, 0x88, 0x01, 0x01, 0x42, 0x09, 0x0a, 0x07,
	0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x6d, 0x6f, 0x76, 0x65,
	0x22, 0x35, 0x0a, 0x0c, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x72, 0x75, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x25, 0x0a, 0x05, 0x63, 0x61, 0x75, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x0f, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x53, 0x74, 0x6f, 0x70, 0x43, 0x61, 0x75, 0x73, 0x65,
	0x52, 0x05, 0x63, 0x61, 0x75, 0x73, 0x65, 0x32, 0xda, 0x01, 0x0a, 0x10, 0x47, 0x61, 0x6d, 0x65,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x3e, 0x0a, 0x0e,
	0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x4f, 0x66, 0x47, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x15,
	0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x46,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x1a, 0x13, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72,
	0x74, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x30, 0x01, 0x12, 0x45, 0x0a, 0x0a,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74,
	0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28,
	0x01, 0x30, 0x01, 0x12, 0x3f, 0x0a, 0x08, 0x4a, 0x6f, 0x69, 0x6e, 0x47, 0x61, 0x6d, 0x65, 0x12,
	0x16, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x4a, 0x6f, 0x69, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x28, 0x01, 0x30, 0x01, 0x42, 0x25, 0x0a, 0x15, 0x63, 0x6f, 0x6d, 0x2e, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x50, 0x01, 0x5a,
	0x0a, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_network_proto_rawDescOnce sync.Once
	file_network_proto_rawDescData = file_network_proto_rawDesc
)

func file_network_proto_rawDescGZIP() []byte {
	file_network_proto_rawDescOnce.Do(func() {
		file_network_proto_rawDescData = protoimpl.X.CompressGZIP(file_network_proto_rawDescData)
	})
	return file_network_proto_rawDescData
}

var file_network_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_network_proto_goTypes = []interface{}{
	(*ListItem)(nil),       // 0: transport.ListItem
	(*Range)(nil),          // 1: transport.Range
	(*GameFilter)(nil),     // 2: transport.GameFilter
	(*CreateRequest)(nil),  // 3: transport.CreateRequest
	(*CreateResponse)(nil), // 4: transport.CreateResponse
	(*JoinRequest)(nil),    // 5: transport.JoinRequest
	(*JoinResponse)(nil),   // 6: transport.JoinResponse
	(*Interruption)(nil),   // 7: transport.Interruption
	(*GameParams)(nil),     // 8: base.GameParams
	(MarkType)(0),          // 9: base.MarkType
	(*Move)(nil),           // 10: base.Move
	(ClientAction)(0),      // 11: base.ClientAction
	(GameStatus)(0),        // 12: base.GameStatus
	(*WinLine)(nil),        // 13: base.WinLine
	(StopCause)(0),         // 14: base.StopCause
}
var file_network_proto_depIdxs = []int32{
	8,  // 0: transport.ListItem.params:type_name -> base.GameParams
	1,  // 1: transport.GameFilter.rows:type_name -> transport.Range
	1,  // 2: transport.GameFilter.cols:type_name -> transport.Range
	1,  // 3: transport.GameFilter.win:type_name -> transport.Range
	9,  // 4: transport.GameFilter.mark:type_name -> base.MarkType
	10, // 5: transport.CreateRequest.move:type_name -> base.Move
	11, // 6: transport.CreateRequest.action:type_name -> base.ClientAction
	8,  // 7: transport.CreateRequest.params:type_name -> base.GameParams
	12, // 8: transport.CreateResponse.status:type_name -> base.GameStatus
	13, // 9: transport.CreateResponse.win_line:type_name -> base.WinLine
	10, // 10: transport.CreateResponse.move:type_name -> base.Move
	10, // 11: transport.JoinRequest.move:type_name -> base.Move
	11, // 12: transport.JoinRequest.action:type_name -> base.ClientAction
	12, // 13: transport.JoinResponse.status:type_name -> base.GameStatus
	13, // 14: transport.JoinResponse.win_line:type_name -> base.WinLine
	10, // 15: transport.JoinResponse.move:type_name -> base.Move
	14, // 16: transport.Interruption.cause:type_name -> base.StopCause
	2,  // 17: transport.GameConfigurator.GetListOfGames:input_type -> transport.GameFilter
	3,  // 18: transport.GameConfigurator.CreateGame:input_type -> transport.CreateRequest
	5,  // 19: transport.GameConfigurator.JoinGame:input_type -> transport.JoinRequest
	0,  // 20: transport.GameConfigurator.GetListOfGames:output_type -> transport.ListItem
	4,  // 21: transport.GameConfigurator.CreateGame:output_type -> transport.CreateResponse
	6,  // 22: transport.GameConfigurator.JoinGame:output_type -> transport.JoinResponse
	20, // [20:23] is the sub-list for method output_type
	17, // [17:20] is the sub-list for method input_type
	17, // [17:17] is the sub-list for extension type_name
	17, // [17:17] is the sub-list for extension extendee
	0,  // [0:17] is the sub-list for field type_name
}

func init() { file_network_proto_init() }
func file_network_proto_init() {
	if File_network_proto != nil {
		return
	}
	file_base_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_network_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListItem); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Range); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameFilter); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Interruption); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_network_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*CreateRequest_Move)(nil),
		(*CreateRequest_Action)(nil),
		(*CreateRequest_Params)(nil),
	}
	file_network_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*CreateResponse_Status)(nil),
		(*CreateResponse_WinLine)(nil),
		(*CreateResponse_GameId)(nil),
	}
	file_network_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*JoinRequest_Move)(nil),
		(*JoinRequest_Action)(nil),
		(*JoinRequest_GameId)(nil),
	}
	file_network_proto_msgTypes[6].OneofWrappers = []interface{}{
		(*JoinResponse_Status)(nil),
		(*JoinResponse_WinLine)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_network_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_network_proto_goTypes,
		DependencyIndexes: file_network_proto_depIdxs,
		MessageInfos:      file_network_proto_msgTypes,
	}.Build()
	File_network_proto = out.File
	file_network_proto_rawDesc = nil
	file_network_proto_goTypes = nil
	file_network_proto_depIdxs = nil
}
