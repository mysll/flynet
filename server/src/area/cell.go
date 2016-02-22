package area

import (
	. "data/datatype"
	"data/entity"
	"data/inter"
	"errors"
	"libs/aoi"
	"libs/aoi/toweraoi"
	"libs/log"
	"libs/rpc"
	"server"
)

type cell struct {
	*server.Heartbeat
	aoi       *aoi.AOI
	id        int
	Objects   map[int]map[int32]entity.Entityer
	scene     entity.Entityer
	playernum int
	livetime  int
	width     float32
	height    float32
}

func (this *cell) enterScene(obj entity.Entityer) {
	if watcher, ok := obj.(inter.Watcher); ok {
		if mover, ok2 := obj.(inter.Mover); ok2 {
			this.aoi.AddWatcher(obj.GetObjId(), obj.ObjType(), mover.GetPos(), watcher.GetRange())
			this.aoi.AddObject(mover.GetPos(), obj.GetObjId(), obj.ObjType())
			ids := this.aoi.GetIdsByPos(mover.GetPos(), watcher.GetRange())
			var ch *server.Channel
			if obj.GetExtraData("aoiCh") == nil {
				ch = server.NewChannel()
				obj.SetExtraData("aoiCh", ch)
			} else {
				ch = obj.GetExtraData("aoiCh").(*server.Channel)
			}

			for _, id := range ids {
				if id.Equal(obj.GetObjId()) {
					continue
				}
				if ent := App.GetEntity(id); ent != nil {
					watcher.AddObject(id, ent.ObjType())
					if ent.ObjType() == PLAYER {
						ch.Add(ent.GetExtraData("mailbox").(rpc.Mailbox))
					}
				}
			}
		}
	}
}

func (this *cell) levelScene(obj entity.Entityer) {
	if watcher, ok := obj.(inter.Watcher); ok {
		if mover, ok2 := obj.(inter.Mover); ok2 {
			this.aoi.RemoveWatcher(obj.GetObjId(), obj.ObjType(), mover.GetPos(), watcher.GetRange())
			this.aoi.RemoveObject(mover.GetPos(), obj.GetObjId(), obj.ObjType())
			obj.GetExtraData("aoiCh").(*server.Channel).Clear()
			watcher.ClearAll()

			//如果是玩家，需要立即进行aoi同步，否则可能删除不了channel
			if obj.ObjType() == PLAYER {
				this.aoiEvent()
			}
		}
	}
}

func (this *cell) AddObject(obj entity.Entityer) error {
	id := obj.GetObjId()
	if _, ok := this.Objects[obj.ObjType()]; !ok {
		this.Objects[obj.ObjType()] = make(map[int32]entity.Entityer, 256)
	}

	if _, dup := this.Objects[obj.ObjType()][id.Index]; dup {
		return errors.New("object already added")
	}

	this.Objects[obj.ObjType()][id.Index] = obj
	if obj.ObjType() == PLAYER {
		App.EntryScene(obj)
		this.enterScene(obj)
		App.baseProxy.entryScene(this.scene, obj)
		App.EnterScene(obj)
		log.LogMessage("add player:", obj.GetObjId())
		this.playernum++
		this.livetime = -1
	} else {
		this.enterScene(obj)
	}

	return nil
}

func (this *cell) RemoveObject(obj entity.Entityer) {

	id := obj.GetObjId()
	if _, ok := this.Objects[obj.ObjType()]; !ok {
		return
	}

	if obj.ObjType() == PLAYER {
		mb := obj.GetExtraData("mailbox").(rpc.Mailbox)
		if pl := App.players.GetPlayer(mb); pl != nil {
			pl.LevelScene()
		}
		log.LogMessage("remove player:", obj.GetObjId())
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
	childs := this.scene.GetChilds()
	for _, child := range childs {
		if child == nil {
			continue
		}
		d := child.GetModify()
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
	App.Destroy(this.scene.GetObjId())
	log.LogMessage("cell deleted,", this.id)
}

func (this *cell) aoiAdd(id ObjectID, typ int, watchers []ObjectID) {
	for _, wid := range watchers {
		if ent := App.GetEntity(wid); ent != nil {
			if watcher, ok := ent.(inter.Watcher); ok {
				if !ent.GetObjId().Equal(id) {
					watcher.AddObject(id, typ)
					if obj := App.GetEntity(id); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.GetExtraData("aoiCh").(*server.Channel).Add(obj.GetExtraData("mailbox").(rpc.Mailbox))
						}
					}
				}
			}
		}
	}
}

func (this *cell) aoiRemove(id ObjectID, typ int, watchers []ObjectID) {
	for _, wid := range watchers {
		if ent := App.GetEntity(wid); ent != nil {
			if watcher, ok := ent.(inter.Watcher); ok {
				if !ent.GetObjId().Equal(id) {
					watcher.RemoveObject(id, typ)
					if obj := App.GetEntity(id); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.GetExtraData("aoiCh").(*server.Channel).Remove(obj.GetExtraData("mailbox").(rpc.Mailbox))
						}
					}
				}
			}
		}
	}
}

func (this *cell) aoiUpdate(id ObjectID, typ int, oldWatchers []ObjectID, newWatchers []ObjectID) {
	for _, wid := range oldWatchers {
		if ent := App.GetEntity(wid); ent != nil {
			if watcher, ok := ent.(inter.Watcher); ok {
				if !ent.GetObjId().Equal(id) {
					watcher.RemoveObject(id, typ)
					if obj := App.GetEntity(id); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.GetExtraData("aoiCh").(*server.Channel).Remove(obj.GetExtraData("mailbox").(rpc.Mailbox))
						}
					}
				}
			}
		}
	}
	for _, wid := range newWatchers {
		if ent := App.GetEntity(wid); ent != nil {
			if watcher, ok := ent.(inter.Watcher); ok {
				if !ent.GetObjId().Equal(id) {
					watcher.AddObject(id, typ)
					if obj := App.GetEntity(id); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.GetExtraData("aoiCh").(*server.Channel).Add(obj.GetExtraData("mailbox").(rpc.Mailbox))
						}
					}
				}
			}
		}
	}
}

func (this *cell) aoiUpdateWatcher(id ObjectID, typ int, addobjs []ObjectID, removeobjs []ObjectID) {
	if ent := App.GetEntity(id); ent != nil {
		if watcher, ok := ent.(inter.Watcher); ok {
			for _, o := range addobjs {
				if !id.Equal(o) {
					watcher.AddObject(o, typ)
					if obj := App.GetEntity(o); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.GetExtraData("aoiChan").(*server.Channel).Add(obj.GetExtraData("mailbox").(rpc.Mailbox))
						}
					}
				}
			}
			for _, o := range removeobjs {
				if !id.Equal(o) {
					watcher.RemoveObject(o, typ)
					if obj := App.GetEntity(o); obj != nil {
						if obj.ObjType() == PLAYER {
							ent.GetExtraData("aoiChan").(*server.Channel).Remove(obj.GetExtraData("mailbox").(rpc.Mailbox))
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
	scene, err := App.Create("BaseScene")
	if err != nil {
		return nil
	}
	c.width = width
	c.height = height
	c.aoi = aoi.NewAOI(toweraoi.NewTowerAOI(c.width, c.height, 50, 50, 5))
	c.id = id
	c.scene = scene
	scene.SetExtraData("cell", c)
	c.Heartbeat = server.NewHeartbeat()
	c.Objects = make(map[int]map[int32]entity.Entityer, 16)
	c.livetime = 60
	return c
}
