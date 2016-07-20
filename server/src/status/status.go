package status

import (
	_ "pb"
	"server"
	"server/libs/log"
	"time"
)

var (
	App *StatusApp
)

type StatusApp struct {
	*server.Server
	pl       *PlayerList
	shutdown int
}

func (status *StatusApp) OnPrepare() bool {
	log.LogMessage(status.Name, " prepared")
	return true
}

func (status *StatusApp) OnMustAppReady() {
}

func (status *StatusApp) OnShutdown() bool {
	status.shutdown = 1
	return false
}

func (status *StatusApp) OnLost(app string) {
	if server.GetAppByType("base") == nil && status.shutdown == 1 {
		status.Exit()
	}
}

func (status *StatusApp) Exit() {
	status.shutdown = 2
	App.AddTimer(time.Second, 1, status.DelayQuit, nil)
}

func (status *StatusApp) DelayQuit(intervalid server.TimerID, count int32, args interface{}) {
	App.Shutdown()
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {
	App = &StatusApp{
		pl: NewPlayerList(),
	}
	server.RegisterRemote("PlayerList", App.pl)
}
