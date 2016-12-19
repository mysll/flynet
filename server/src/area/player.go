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

func (p *Player) OnEnterScene(self Entity) int {
	mb := self.FindExtraData("base").(rpc.Mailbox)
	pl := App.Players.FindPlayer(mb.Uid)
	if pl != nil {
		App.AddHeartbeat(self, "Store", time.Minute*5, -1, mb)
	}
	return 1
}

func StorePlayer(obj Entity) (*EntityInfo, error) {
	item := &EntityInfo{}
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}
	item.Type = obj.ObjTypeName()
	item.Caps = obj.Caps()
	item.DbId = obj.DBId()
	if obj.FindExtraData("linkObj") != nil {
		item.ObjId = obj.FindExtraData("linkObj").(ObjectID)
	}

	item.Index = obj.ChildIndex()
	item.Data = buffer.Bytes()

	ls := obj.AllChilds()
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

func (p *Player) OnStore(self Entity, typ int) int {
	info, err := StorePlayer(self)
	if err != nil {
		log.LogError("save player failed")
		return 0
	}
	self.SetExtraData("saveData", info)
	return 0
}

func (p *Player) OnTimer(self Entity, beat string, count int32, args interface{}) int {
	switch beat {
	case "Store":
		mb := args.(rpc.Mailbox)
		pl := App.Players.FindPlayer(mb.Uid)
		if pl != nil {
			App.Save(pl.GetEntity(), share.SAVETYPE_TIMER)
		}
	}

	return 1
}

func (p *Player) OnLoad(self Entity, typ int) int {
	if typ == share.LOAD_ARCHIVE {
	}

	return 1
}

func (p *Player) OnDestroy(self Entity, sender Entity) int {
	log.LogInfo("player destroy,", self.ObjectId())
	return 1
}
