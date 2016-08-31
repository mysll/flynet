package login

import (
	"pb/c2s"
	"pb/s2c"
	"server"
	"server/libs/rpc"
	"server/share"

	proto "github.com/golang/protobuf/proto"
)

type Account struct {
}

func (t *Account) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("Login", t.Login)
}

func (a *Account) Login(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	logindata := &c2s.Loginuser{}
	if server.Check(server.ParseProto(msg, logindata)) {
		return 0, nil
	}
	if logindata.GetPassword() == "123" {
		apps := server.GetAppByName("basemgr")
		if apps != nil {
			server.Check(apps.Call(&mailbox, "Session.GetBaseAndId", logindata.GetUser()))
			return 0, nil
		}
	} else {
		e := &s2c.Error{}
		e.ErrorNo = proto.Int32(share.ERROR_LOGIN_FAILED)
		server.MailTo(nil, &mailbox, "Login.Error", e)
	}

	return 0, nil
}
