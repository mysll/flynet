package server

import (
	"libs/log"
	"libs/rpc"
	"util"

	"github.com/mysll/log4go"
)

type BodyWriter struct {
	*util.StoreArchive
	msg *rpc.Message
}

func (w *BodyWriter) GetMessage() *rpc.Message {
	return w.msg
}

func (w *BodyWriter) Flush() {
	w.msg.Body = w.msg.Body[:w.Len()]
}

func (w *BodyWriter) Free() {
	w.msg.Free()
}

type HeadWriter struct {
	*util.StoreArchive
	msg *rpc.Message
}

func (w *HeadWriter) Flush() {
	w.msg.Header = w.msg.Header[:w.Len()]
}

func NewMessageReader(msg *rpc.Message) *util.LoadArchive {
	return util.NewLoadArchiver(msg.Body)
}

func NewMessageWriter(msg *rpc.Message) *BodyWriter {
	msg.Body = msg.Body[:0]
	w := &BodyWriter{util.NewStoreArchiver(msg.Body), msg}
	return w
}

func NewMessage() *BodyWriter {
	msg := rpc.NewMessage(rpc.MAX_BUF_LEN)
	w := &BodyWriter{util.NewStoreArchiver(msg.Body), msg}
	return w
}

func NewHeadReader(msg *rpc.Message) *util.LoadArchive {
	return util.NewLoadArchiver(msg.Header)
}

func NewHeadWriter(msg *rpc.Message) *HeadWriter {
	msg.Header = msg.Header[:0]
	w := &HeadWriter{util.NewStoreArchiver(msg.Header), msg}
	return w
}

func ProtoParse(msg *rpc.Message, obj interface{}) error {
	return core.rpcProto.DecodeMessage(msg, obj)
}

func ParseArgs(msg *rpc.Message, args ...interface{}) error {
	if len(args) == 0 || msg == nil {
		return nil
	}

	ar := NewMessageReader(msg)
	for i := 0; i < len(args); i++ {
		err := ar.Read(args[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func Check(err error) bool {
	if err != nil {
		log4go.LogCallerDepth = 4
		log.LogError(err)
		log4go.LogCallerDepth = 3
		return true
	}

	return false
}

func Check2(_ interface{}, err error) bool {
	if err != nil {
		log4go.LogCallerDepth = 4
		log.LogError(err)
		log4go.LogCallerDepth = 3
		return true
	}

	return false
}

func CreateMessage(args ...interface{}) (*rpc.Message, error) {
	if len(args) > 0 {
		msg := NewMessage()
		for i := 0; i < len(args); i++ {
			err := msg.Write(args[i])
			if err != nil {
				msg.Free()
				return nil, err
			}
		}
		msg.Flush()
		return msg.GetMessage(), nil
	}
	return nil, nil
}
