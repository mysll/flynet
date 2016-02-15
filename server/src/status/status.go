package status

import (
	"libs/log"
	"server"
	"time"
)

var (
	App *StatusApp
)

type StatusApp struct {
	*server.Server
	rm       *RankManager
	pl       *PlayerList
	shutdown int
}

func (this *StatusApp) OnPrepare() bool {
	log.LogMessage(this.Id, " prepared")
	return true
}

func (this *StatusApp) OnMustAppReady() {
	App.rm.load()
}

func (this *StatusApp) OnShutdown() bool {
	this.shutdown = 1
	return false
}

func (this *StatusApp) OnLost(app string) {
	if server.GetAppByType("base") == nil && this.shutdown == 1 {
		this.Exit()
	}
}

func (this *StatusApp) Exit() {
	this.shutdown = 2
	App.rm.savetotal()
	App.AddHeartbeat("DelayQuit", time.Second, 1, this.DelayQuit, nil)
}

func (this *StatusApp) DelayQuit(t time.Duration, count int32, args interface{}) {
	App.Shutdown()
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {
	App = &StatusApp{
		rm: NewRankManager(),
		pl: NewPlayerList(),
	}
	server.RegisterRemote("RankManager", App.rm)
	server.RegisterRemote("PlayerList", App.pl)
}
