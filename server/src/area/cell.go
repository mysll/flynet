package area

import (
	"errors"
	"game/libs/aoi"
	"game/libs/aoi/toweraoi"
	"logicdata/inter"
	"server"
	. "server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
)

type CellInfo struct {
	Id           int
	Name         string
	Path         string  //地图信息路径
	X            float32 //地图左上角X坐标
	Y            float32 //地图左上角Y坐标
	Width        float32 //地图宽度
	Height       float32 //地图高度
	TileW        float32 //格子宽
	TileH        float32 //格子高
	MaxViewRange float32 //最大视野
	MaxPlayers   int32   //最大玩家数
	MaxVisual    int32   //最大可视人数
}

type cell struct {
	server.Dispatch
	aoi       *aoi.AOI
	id        int
	Objects   map[int]map[int32]Entity
	scene     Entity
	playernum int
	livetime  int
	width     float32
	height    float32
}

func (this *cell) enterScene(obj Entity) {
	if watcher, ok := obj.(inter.Watcher); ok {
		if mover, ok2 := obj.(inter.Mover); ok2 {
			this.aoi.AddWatcher(obj.ObjectId(), obj.ObjType(), mover.GetPos(), watcher.GetRange())
			this.aoi.AddObject(mover.GetPos(), obj.ObjectId(), obj.ObjType())
			ids := this.aoi.GetIdsByPos(mover.GetPos(), watcher.GetRange())
			var ch *server.Channel
			if obj.FindExtraData("aoiCh") == nil {
				ch = server.NewChannel()
				obj.SetExtraData("aoiCh", ch)
			} else {
				ch = obj.FindExtraData("aoiCh").(*server.Channel)
			}

			for _, id := range ids {
				if id.Equal(obj.ObjectId()) {
					continue
				}
				if ent := App.Kernel().GetEntity(id); ent != nil {
					watcher.AddObject(id, ent.ObjType())
					if ent.ObjType() == PLAYER {
						ch.Add(rpc.NewMailBoxFromUid(ent.UID()))
					}
				}
			}
		}
	}
}

func (this *cell) levelScene(obj Entity) {
	if watcher, ok := obj.(inter.Watcher); ok {
		if mover, ok2 := obj.(inter.Mover); ok2 {
			this.aoi.RemoveWatcher(obj.ObjectId(), obj.ObjType(), mover.GetPos(), watcher.GetRange())
			this.aoi.RemoveObject(mover.GetPos(), obj.ObjectId(), obj.ObjType())
			obj.FindExtraData("aoiCh").(*server.Channel).Clear()
			watcher.ClearAll()

			//如果是玩家，需要立即进行aoi同步，否则可能删除不了channel
			if obj.ObjType() == PLAYER {
				this.aoiEvent()
			}
		}
	}
}

func (this *cell) AddObject(obj Entity) error {
	id := obj.ObjectId()
	if _, ok := this.Objects[obj.ObjType()]; !ok {
		this.Objects[obj.ObjType()] = make(map[int32]Entity, 256)
	}

	if _, dup := this.Objects[obj.ObjType()][id.Index]; dup {
		return errors.New("object already added")
	}

	this.Objects[obj.ObjType()][id.Index] = obj
	if obj.ObjType() == PLAYER {
		App.Kernel().EntryScene(obj)
		this.enterScene(obj)
		App.baseProxy.entryScene(this.scene, obj)
		App.Kernel().EnterScene(obj)
		log.LogMessage("add player:", obj.ObjectId())
		this.playernum++
		this.livetime = -1
	} else {
		this.enterScene(obj)
	}

	return nil
}

func (this *cell) RemoveObject(obj Entity) {

	id := obj.ObjectId()
	if _, ok := this.Objects[obj.ObjType()]; !ok {
		return
	}

	if obj.ObjType() == PLAYER {
		mb := rpc.NewMailBoxFromUid(obj.UID())
		if pl := App.Players.FindPlayer(mb.Uid); pl != nil {
			pl.(*AreaPlayer).LevelScene()
		}
		log.LogMessage("remove player:", obj.ObjectId())
		this.playernum--
		if this.playernum == 0 { //没有玩家后，删除本场景
			this.livetime = 60
		}
	}

	this.levelScene(obj)

	delete(this.Objects[obj.ObjType()], id.Index)
}

func (this *cell) OnBeginUpdate() {

}

func (this *cell) OnUpdate() {

}

func (this *cell) OnLastUpdate() {
	childs := this.scene.AllChilds()
	for _, child := range childs {
		if child == nil {
			continue
		}
		d := child.Modifys()
		if len(d) > 0 {

		}

	}
}

func (this *cell) Delete() {
	if ps, ok := this.Objects[PLAYER]; ok {
		if len(ps) > 0 {
			for _, p := range ps {
				this.RemoveObject(p)
			}
		}
	}
	App.Kernel().Destroy(this.scene.ObjectId())
	log.LogMessage("cell deleted,", this.id)
}

func (this *cell) aoiAdd(id ObjectID, typ int, watchers []ObjectID) {
	for _, wid := range watchers {
		if ent := App.Kernel().GetEntity(wid); ent != nil {
			if watcher, ok := ent.(inter.Watcher); ok {
				if !ent.ObjectId().Equal(id) {
					watcher.AddObject(id, typ)
					if obj := App.Kernel().GetEntity(id); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.FindExtraData("aoiCh").(*server.Channel).Add(rpc.NewMailBoxFromUid(obj.UID()))
						}
					}
				}
			}
		}
	}
}

func (this *cell) aoiRemove(id ObjectID, typ int, watchers []ObjectID) {
	for _, wid := range watchers {
		if ent := App.Kernel().GetEntity(wid); ent != nil {
			if watcher, ok := ent.(inter.Watcher); ok {
				if !ent.ObjectId().Equal(id) {
					watcher.RemoveObject(id, typ)
					if obj := App.Kernel().GetEntity(id); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.FindExtraData("aoiCh").(*server.Channel).Remove(rpc.NewMailBoxFromUid(obj.UID()))
						}
					}
				}
			}
		}
	}
}

func (this *cell) aoiUpdate(id ObjectID, typ int, oldWatchers []ObjectID, newWatchers []ObjectID) {
	for _, wid := range oldWatchers {
		if ent := App.Kernel().GetEntity(wid); ent != nil {
			if watcher, ok := ent.(inter.Watcher); ok {
				if !ent.ObjectId().Equal(id) {
					watcher.RemoveObject(id, typ)
					if obj := App.Kernel().GetEntity(id); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.FindExtraData("aoiCh").(*server.Channel).Remove(rpc.NewMailBoxFromUid(obj.UID()))
						}
					}
				}
			}
		}
	}
	for _, wid := range newWatchers {
		if ent := App.Kernel().GetEntity(wid); ent != nil {
			if watcher, ok := ent.(inter.Watcher); ok {
				if !ent.ObjectId().Equal(id) {
					watcher.AddObject(id, typ)
					if obj := App.Kernel().GetEntity(id); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.FindExtraData("aoiCh").(*server.Channel).Add(rpc.NewMailBoxFromUid(obj.UID()))
						}
					}
				}
			}
		}
	}
}

func (this *cell) aoiUpdateWatcher(id ObjectID, typ int, addobjs []ObjectID, removeobjs []ObjectID) {
	if ent := App.Kernel().GetEntity(id); ent != nil {
		if watcher, ok := ent.(inter.Watcher); ok {
			for _, o := range addobjs {
				if !id.Equal(o) {
					watcher.AddObject(o, typ)
					if obj := App.Kernel().GetEntity(o); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.FindExtraData("aoiChan").(*server.Channel).Add(rpc.NewMailBoxFromUid(obj.UID()))
						}
					}
				}
			}
			for _, o := range removeobjs {
				if !id.Equal(o) {
					watcher.RemoveObject(o, typ)
					if obj := App.Kernel().GetEntity(o); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.FindExtraData("aoiChan").(*server.Channel).Remove(rpc.NewMailBoxFromUid(obj.UID()))
						}
					}
				}
			}
		}
	}
}

func (this *cell) aoiEvent() {
	event := this.aoi.GetEvent()

	for {
		e := event.Pop()
		if e == nil {
			break
		}
		switch e.Typ {
		case "add":
			this.aoiAdd(e.Args["id"].(ObjectID), e.Args["type"].(int), e.Args["watchers"].([]ObjectID))
		case "remove":
			this.aoiRemove(e.Args["id"].(ObjectID), e.Args["type"].(int), e.Args["watchers"].([]ObjectID))
		case "update":
			this.aoiUpdate(e.Args["id"].(ObjectID), e.Args["type"].(int), e.Args["oldWatchers"].([]ObjectID), e.Args["newWatchers"].([]ObjectID))
		case "updateWatcher":
			this.aoiUpdateWatcher(e.Args["id"].(ObjectID), e.Args["type"].(int), e.Args["addObjs"].([]ObjectID), e.Args["removeObjs"].([]ObjectID))
		}
		event.FreeEvent(e)
	}

}

func (this *cell) OnFlush() {
	this.aoiEvent()
	if this.livetime == 0 {
		App.RemoveCell(this.id)
		return
	}
	if this.livetime > 0 {
		this.livetime--
	}
}

func CreateCell(id int, width float32, height float32) *cell {
	c := &cell{}
	scene, err := App.Kernel().Create("BaseScene")
	if err != nil {
		return nil
	}
	c.width = width
	c.height = height
	c.aoi = aoi.NewAOI(toweraoi.NewTowerAOI(c.width, c.height, 50, 50, 5))
	c.id = id
	c.scene = scene
	scene.SetExtraData("cell", c)
	c.Objects = make(map[int]map[int32]Entity, 16)
	c.livetime = 60
	return c
}
