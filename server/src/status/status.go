package status

import (
	_ "logicdata/entity"
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
	if server.GetAppByType("base") == nil {
		status.Exit()
		return false
	}
	status.shutdown = 1
	return false
}

func (status *StatusApp) OnLost(app string) {
	if server.GetAppByType("base") == nil && status.shutdown == 1 {
		status.Exit()
	}
}

func (status *StatusApp) OnGlobalDataLoaded() {
	log.LogError("create test globaldata")
	if status.FindGlobalData("Test1") == -1 {
		if err := status.AddGlobalData("Test1", "GlobalData"); err != nil {
			log.LogError(err)
			return
		}

		index := status.FindGlobalData("Test1")
		status.GlobalDataSet(index, "Test1", "ddddddddd")
		status.GlobalDataSet(index, "Test2", "hhhhhhh")

		status.GlobalDataAddRowValues(index, "TestRec", -1, "sll", int8(1))
		status.GlobalDataAddRowValues(index, "TestRec", -1, "sll2", int8(2))
		status.GlobalDataSetGrid(index, "TestRec", 0, 0, "test")
		status.GlobalDataDelRow(index, "TestRec", 1)
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
