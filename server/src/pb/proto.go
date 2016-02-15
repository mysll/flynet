package pb

import (
	"github.com/golang/protobuf/proto"
	"io"
	"util"
)

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
