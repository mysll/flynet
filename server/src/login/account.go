package login

import (
	"libs/rpc"
	"pb/c2s"
	"pb/s2c"
	"server"
	"share"

	proto "github.com/golang/protobuf/proto"
)

type Account struct {
}

func (a *Account) Login(mailbox rpc.Mailbox, logindata c2s.Loginuser) error {
	if logindata.GetPassword() == "123" {
		apps := server.GetApp("basemgr")
		if apps != nil {
			return apps.Call(&mailbox, "Session.GetBaseAndId", logindata.GetUser())
		}
	} else {
		e := &s2c.Error{}
		e.ErrorNo = proto.Int32(share.ERROR_LOGIN_FAILED)
		server.MailTo(nil, &mailbox, "Login.Error", e)
	}

	return nil
}
