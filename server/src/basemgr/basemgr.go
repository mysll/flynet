package basemgr

import (
	"libs/log"
	"server"
)

var (
	App = &BaseMgr{}
)

type BaseMgr struct {
	*server.Server
	quit chan int
}

func (b *BaseMgr) OnPrepare() bool {
	log.LogMessage(b.Name, " prepared")
	return true
}

func (b *BaseMgr) OnStart() {
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {
	server.RegisterRemote("Session", &Session{})
}
