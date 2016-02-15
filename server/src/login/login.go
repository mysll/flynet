package login

import (
	"libs/log"
	"server"
)

var (
	App = &LoginApp{}
)

type LoginApp struct {
	*server.Server
}

func (l *LoginApp) OnPrepare() bool {
	log.LogMessage(l.Id, " prepared")
	return true
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {
	server.RegisterHandler("Account", &Account{})
}
