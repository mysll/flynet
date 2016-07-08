package server

import (
	"errors"
	"libs/log"
	"libs/rpc"
	"pb/c2s"
	"pb/s2c"

	"github.com/golang/protobuf/proto"
)

var (
	ERR_PROTO_NOT_MATCH = errors.New("proto not match")
	ERR_ARGS_NOT_MATCH  = errors.New("args must be proto.Message")
)

type ClientProtoer interface {
	CreateRpcMessage(svr, method string, args interface{}) (data []byte, err error)
	DecodeRpcMessage(msg *rpc.Message) (node, Servicemethod string, data []byte, err error)
	DecodeMessage(msg *rpc.Message, out interface{}) error
}

type PBProto struct {
}

func (pb *PBProto) CreateRpcMessage(svr, method string, args interface{}) (data []byte, err error) {
	r := &s2c.Rpc{}
	r.Sender = proto.String(svr)
	r.Servicemethod = proto.String(method)
	if val, ok := args.(proto.Message); ok {
		if r.Data, err = proto.Marshal(val); err != nil {
			return
		}
	} else {
		err = ERR_ARGS_NOT_MATCH
		return
	}

	data, err = proto.Marshal(r)
	return
}

func (pb *PBProto) DecodeRpcMessage(msg *rpc.Message) (node, Servicemethod string, data []byte, err error) {
	request := &c2s.Rpc{}

	if err = proto.Unmarshal(msg.Body, request); err != nil {
		log.LogError(err)
		return
	}

	return request.GetNode(), request.GetServicemethod(), request.GetData(), nil
}

func (pb *PBProto) DecodeMessage(msg *rpc.Message, out interface{}) error {
	r := NewMessageReader(msg)
	data, err := r.ReadData()
	if err != nil {
		return err
	}

	if pb, ok := out.(proto.Message); ok {
		return proto.Unmarshal(data, pb)
	}

	return ERR_PROTO_NOT_MATCH

}
