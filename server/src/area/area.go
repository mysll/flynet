package area

import (
	"libs/log"
	"server"
)

var (
	App *AreaApp
)

type AreaApp struct {
	*server.Server
	players   *PlayerList
	baseProxy *BaseProxy
	cells     map[int]*cell
}

func (a *AreaApp) OnPrepare() bool {
	log.LogMessage(a.Id, " prepared")
	return true
}

func (a *AreaApp) OnBeatRun() {
	for _, cell := range a.cells {
		cell.Pump()
	}
	a.players.Pump()
}

func (a *AreaApp) OnBeginUpdate() {
	for _, cell := range a.cells {
		cell.OnBeginUpdate()
	}
}

func (a *AreaApp) OnUpdate() {
	for _, cell := range a.cells {
		cell.OnUpdate()
	}
}

func (a *AreaApp) OnLastUpdate() {
	for _, cell := range a.cells {
		cell.OnLastUpdate()
	}
}

func (a *AreaApp) OnFlush() {
	for _, cell := range a.cells {
		cell.OnFlush()
	}
}

func (a *AreaApp) OnFrame() {
	a.players.ClearDeleted()
}

func (a *AreaApp) GetCell(id int) *cell {
	if cell, ok := a.cells[id]; ok {
		return cell
	}
	cell := CreateCell(id, 1000, 1000)
	a.cells[id] = cell
	return cell
}

func (a *AreaApp) RemoveCell(id int) {
	if cell, ok := a.cells[id]; ok {
		cell.Delete()
		delete(a.cells, id)
	}
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {
	App = &AreaApp{
		players:   NewPlayerList(),
		baseProxy: NewBaseProxy(),
		cells:     make(map[int]*cell),
	}
	server.RegisterCallee("Player", &Player{})
	server.RegisterCallee("BaseScene", &Scene{})

	server.RegisterRemote("BaseProxy", App.baseProxy)
}
