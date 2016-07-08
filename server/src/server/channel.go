package server

import (
	"libs/log"
	"libs/rpc"

	"github.com/golang/protobuf/proto"
)

//频道
type Channel struct {
	receivers  map[int32]map[uint64]rpc.Mailbox
	sessionbuf []int64
}

//频道内增加用户
func (this *Channel) Add(mailbox rpc.Mailbox) {
	if recvs, ok := this.receivers[mailbox.App]; !ok {
		recvs = make(map[uint64]rpc.Mailbox, 256)
		recvs[mailbox.Uid] = mailbox
		this.receivers[mailbox.App] = recvs
	} else {
		recvs[mailbox.Uid] = mailbox
	}
}

//频道内移除用户
func (this *Channel) Remove(mailbox rpc.Mailbox) {
	if recvs, ok := this.receivers[mailbox.App]; ok {
		delete(recvs, mailbox.Uid)
	}
}

//清除所有用户
func (this *Channel) Clear() {
	for k := range this.receivers {
		for k1 := range this.receivers[k] {
			delete(this.receivers[k], k1)
		}
	}
}

//向频道内广播消息
func (this *Channel) SendMsg(src rpc.Mailbox, method string, args proto.Message) {

	for appid, recvs := range this.receivers {
		app := GetAppById(appid)
		if app == nil {
			log.LogError(ErrAppNotFound.Error())
			continue
		}
		this.sessionbuf = this.sessionbuf[:0]
		for _, m := range recvs {
			this.sessionbuf = append(this.sessionbuf, m.Id)
		}
		err := app.ClientBroadcast(&src, this.sessionbuf, method, args)
		if err != nil {
			log.LogError(err)
		}
	}
}

//创建一个新的频道
func NewChannel() *Channel {
	channel := &Channel{}
	channel.receivers = make(map[int32]map[uint64]rpc.Mailbox, 32)
	channel.sessionbuf = make([]int64, 0, 256)
	return channel
}
