package base

import (
	"math/rand"
	"pb/s2c"
	"server"
	"server/libs/log"
	"server/libs/rpc"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	EXPIREDTIME = 30 //seconds
)

type cacheUser struct {
	id        int32
	logintime time.Time
}

type Login struct {
	Cached map[string]cacheUser
}

func (t *Login) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("AddClient", t.AddClient)
	s.RegisterCallback("SwitchPlayer", t.SwitchPlayer)
}

func (l *Login) AddClient(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	user, err := r.ReadString()
	if server.Check(err) {
		return nil
	}
	serial := rand.Int31()
	log.LogMessage("client serial:", serial)
	l.Cached[user] = cacheUser{serial, time.Now()}

	ret := &s2c.Loginsucceed{}
	ret.Host = proto.String(App.ClientHost)
	ret.Port = proto.Int32(int32(App.ClientPort))
	ret.Key = proto.Int32(serial)
	server.Check(server.MailTo(nil, &mailbox, "Login.LoginResult", ret))
	return nil
}

func (l *Login) SwitchPlayer(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	user, err := r.ReadString()
	if server.Check(err) {
		return nil
	}
	serial := rand.Int31()
	log.LogMessage("client serial:", serial)
	l.Cached[user] = cacheUser{serial, time.Now()}

	ret := &s2c.Loginsucceed{}
	ret.Host = proto.String(App.ClientHost)
	ret.Port = proto.Int32(int32(App.ClientPort))
	ret.Key = proto.Int32(serial)
	server.Check(server.MailTo(nil, &mailbox, "Login.SwitchBase", ret))
	return nil
}

func (l *Login) checkClient(user string, key int32) bool {
	if k, ok := l.Cached[user]; ok {
		ret := (k.id == key)
		delete(l.Cached, user)
		return ret
	}
	return false
}

func (l *Login) checkCached() {
	for k, v := range l.Cached {
		if time.Now().Sub(v.logintime).Seconds() >= EXPIREDTIME {
			delete(l.Cached, k)
		}
	}
}

func NewLogin() *Login {
	l := &Login{}
	l.Cached = make(map[string]cacheUser)
	return l
}
