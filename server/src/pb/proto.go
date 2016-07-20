package pb

import (
	"errors"
	"io"
	"pb/c2s"
	"pb/s2c"
	"server"
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/util"

	"github.com/golang/protobuf/proto"
)

var (
	ERR_PROTO_NOT_MATCH = errors.New("proto not match")
	ERR_ARGS_NOT_MATCH  = errors.New("args must be proto.Message")
)

type PBProtoCodec struct {
}

func (pb *PBProtoCodec) GetCodecInfo() string {
	return "use protobuf"
}

func (pb *PBProtoCodec) CreateObjectMessage(obj datatype.Entityer, self bool, mailbox rpc.Mailbox) interface{} {
	data, _ := obj.Serial()
	create := &s2c.CreateObject{}
	create.Entity = proto.String(obj.ObjTypeName())
	create.Self = proto.Bool(self)
	create.Index = proto.Int32(0)
	create.Serial = proto.Int32(0)
	create.Typ = proto.Int32(0)
	create.Propinfo = data
	create.Mailbox = proto.String(mailbox.String())
	return create
}

func (pb *PBProtoCodec) ErrorMsg(errno int32) interface{} {
	errmsg := &s2c.Error{}
	errmsg.ErrorNo = proto.Int32(errno)
	return errmsg
}

func (pb *PBProtoCodec) CreateRpcMessage(svr, method string, args interface{}) (data []byte, err error) {
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

func (pb *PBProtoCodec) DecodeRpcMessage(msg *rpc.Message) (node, Servicemethod string, data []byte, err error) {
	request := &c2s.Rpc{}

	if err = proto.Unmarshal(msg.Body, request); err != nil {
		log.LogError(err)
		return
	}

	return request.GetNode(), request.GetServicemethod(), request.GetData(), nil
}

func (pb *PBProtoCodec) DecodeMessage(msg *rpc.Message, out interface{}) error {
	r := util.NewLoadArchiver(msg.Body)
	data, err := r.ReadData()
	if err != nil {
		return err
	}

	if pb, ok := out.(proto.Message); ok {
		return proto.Unmarshal(data, pb)
	}

	return ERR_PROTO_NOT_MATCH

}

func Encode(msg proto.Message) (data []byte, err error) {
	return proto.Marshal(msg)
}

func Decode(buf []byte, out proto.Message) error {
	return proto.Unmarshal(buf, out)
}

func SendMsg(r io.Writer, id int, msg proto.Message) error {
	data, err := Encode(msg)
	if err != nil {
		return err
	}

	out, err1 := util.CreateMsg(nil, data, id)
	if err1 != nil {
		return err1
	}

	r.Write(out)

	return nil
}

func init() {
	server.RegisterProtoCodec(&PBProtoCodec{})
	server.RegisterPropCodec(&PBPropCodec{})
	server.RegisterTableCodec(&PBTableCodec{})
	server.RegisterViewportCodec(&PBViewportCodec{})
}
