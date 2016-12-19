package server

import (
	"server/data/datatype"
	"server/libs/rpc"
)

var (
	codec ProtoCodec
)

type ProtoCodec interface {
	GetCodecInfo() string
	CreateObjectMessage(obj datatype.Entity, self bool, mailbox rpc.Mailbox) interface{}
	ErrorMsg(errno int32) interface{}
	CreateRpcMessage(svr, method string, args interface{}) (data []byte, err error)
	DecodeRpcMessage(msg *rpc.Message) (node, Servicemethod string, data []byte, err error)
	DecodeMessage(msg *rpc.Message, out interface{}) error
}

func RegisterProtoCodec(p ProtoCodec) {
	codec = p
}
