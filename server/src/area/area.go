package area

import (
	_ "logicdata/entity"
	_ "pb"
	"server"
	"server/data/datatype"
	"server/libs/log"
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
	log.LogMessage(a.AppId, " prepared")
	return true
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
	a.AddDispatchNoName(cell, server.DP_BEGINUPDATE|server.DP_UPDATE|server.DP_LASTUPDATE|server.DP_FLUSH)
	return cell
}

func (a *AreaApp) FindCell(id int) *cell {
	if cell, ok := a.cells[id]; ok {
		return cell
	}
	return nil
}

func (a *AreaApp) RemoveCell(id int) {
	if cell, ok := a.cells[id]; ok {
		a.RemoveDispatch(cell.GetDispatchID())
		cell.Delete()
		delete(a.cells, id)
	}
}

func (a *AreaApp) OnTeleportFromBase(args []interface{}, player datatype.Entityer) bool {
	log.LogMessage(args)
	return true
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
}
