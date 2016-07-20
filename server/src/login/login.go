package login

import (
	_ "pb"
	"server"
	"server/libs/log"
)

var (
	App = &LoginApp{}
)

type LoginApp struct {
	*server.Server
}

func (l *LoginApp) OnPrepare() bool {
	log.LogMessage(l.Name, " prepared")
	return true
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {
	server.RegisterHandler("Account", &Account{})
}
