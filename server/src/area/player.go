package area

import (
	"bytes"
	"encoding/gob"
	"server"
	. "server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
	"time"
)

type Player struct {
	server.Callee
}

func (p *Player) OnEnterScene(self Entityer) int {
	mb := self.GetExtraData("base").(rpc.Mailbox)
	pl := App.players.GetPlayer(mb)
	if pl != nil {
		App.AddHeartbeat(self, "Store", time.Minute*5, -1, mb)
	}
	return 1
}

func StorePlayer(obj Entityer) (*EntityInfo, error) {
	item := &EntityInfo{}
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}
	item.Type = obj.ObjTypeName()
	item.Caps = obj.GetCapacity()
	item.DbId = obj.GetDbId()
	if obj.GetExtraData("linkObj") != nil {
		item.ObjId = obj.GetExtraData("linkObj").(ObjectID)
	}

	item.Index = obj.GetIndex()
	item.Data = buffer.Bytes()

	ls := obj.GetChilds()
	if len(ls) > 0 {
		item.Childs = make([]*EntityInfo, 0, len(ls))
	}
	for _, c := range ls {
		if c != nil {
			child, err := StorePlayer(c)
			if err != nil {
				return nil, err
			}
			if child != nil {
				item.Childs = append(item.Childs, child)
			}
		}
	}

	return item, nil
}

func (p *Player) OnStore(self Entityer, typ int) int {
	info, err := StorePlayer(self)
	if err != nil {
		log.LogError("save player failed")
		return 0
	}
	self.SetExtraData("saveData", info)
	return 0
}

func (p *Player) OnTimer(self Entityer, count int32, args interface{}) int {
	mb := args.(rpc.Mailbox)
	pl := App.players.GetPlayer(mb)
	if pl != nil {
		App.Save(pl.Entity, share.SAVETYPE_TIMER)
	}
	return 1
}

func (p *Player) OnLoad(self Entityer, typ int) int {
	if typ == share.LOAD_ARCHIVE {
	}

	return 1
}

func (p *Player) OnDestroy(self Entityer, sender Entityer) int {
	log.LogInfo("player destroy,", self.GetObjId())
	return 1
}
