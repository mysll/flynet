package server

import (
	"server/libs/log"
	"server/libs/rpc"
	"server/util"

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

//获取rpc消息的错误代码，返回0没有错误
func GetReplyError(msg *rpc.Message) int32 {
	ar := util.NewLoadArchiver(msg.Header)
	if len(msg.Header) <= 8 {
		return 0
	}
	ar.Seek(8, 0)

	haserror, err := ar.ReadInt8()
	if err != nil {
		return 0
	}

	if haserror != 1 {
		return 0
	}

	errcode, err := ar.ReadInt32()
	if err != nil {
		return 0
	}

	return errcode
}

func ParseProto(msg *rpc.Message, obj interface{}) error {
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

func CreateMessage(args ...interface{}) *rpc.Message {
	if len(args) > 0 {
		msg := NewMessage()
		for i := 0; i < len(args); i++ {
			err := msg.Write(args[i])
			if err != nil {
				msg.Free()
				panic("write args failed")
			}
		}
		msg.Flush()
		return msg.GetMessage()
	}
	return nil
}

func AppendArgs(msg *rpc.Message, args ...interface{}) {
	w := &BodyWriter{util.NewStoreArchiver(msg.Body), msg}
	if len(args) > 0 {
		for i := 0; i < len(args); i++ {
			err := w.Write(args[i])
			if err != nil {
				msg.Free()
				panic("write args failed")
			}
		}
		w.Flush()
	}
}

//错误消息
func ReplyErrorMessage(errcode int32) *rpc.Message {
	msg := rpc.NewMessage(1)
	if errcode == 0 {
		return msg
	}
	sr := util.NewStoreArchiver(msg.Header)
	sr.Write(int8(1))
	sr.Write(errcode)
	msg.Header = msg.Header[:sr.Len()]
	return msg
}

func ReplyErrorAndArgs(errcode int32, args ...interface{}) *rpc.Message {
	msg := rpc.NewMessage(rpc.MAX_BUF_LEN)

	if errcode > 0 {
		sr := util.NewStoreArchiver(msg.Header)
		sr.Write(int8(1))
		sr.Write(errcode)
		msg.Header = msg.Header[:sr.Len()]
	}

	if len(args) > 0 {
		mw := NewMessageWriter(msg)
		for i := 0; i < len(args); i++ {
			err := mw.Write(args[i])
			if err != nil {
				msg.Free()
				panic("write args failed")
				return nil
			}
		}
		mw.Flush()
	}

	return msg
}
