package base

import (
	"bytes"
	"encoding/gob"
	"errors"
	"server"
	. "server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
)

type Sync struct {
	childs map[ObjectID]Entity
	newobj map[ObjectID]ObjectID
}

func (t *Sync) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("SyncPlayer", t.SyncPlayer)
}

func (s *Sync) getAllChilds(obj Entity) {
	s.childs[obj.GetObjId()] = obj
	cs := obj.GetChilds()
	for _, c := range cs {
		if c != nil {
			s.getAllChilds(c)
		}
	}
}
func (s *Sync) sync(info *EntityInfo) (obj Entity, err error) {
	if info.ObjId.IsNil() {
		obj, err = App.CreateContainer(info.Type, int(info.Caps))
		if err == nil {
			s.newobj[info.ObjId] = obj.GetObjId()
		}
	} else {
		obj = App.GetEntity(info.ObjId)
		if obj != nil {
			delete(s.childs, info.ObjId)
		}
	}
	if obj == nil {
		err = errors.New("get entity failed")
		return
	}

	buf := bytes.NewBuffer(info.Data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(obj)
	if err != nil {
		return
	}

	obj.ClearChilds()
	for _, c := range info.Childs {
		var cobj Entity
		cobj, err = s.sync(c)
		if err != nil {
			return
		}
		obj.AddChild(info.Index, cobj)
	}

	return
}

func (s *Sync) SyncPlayer(src rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	r := server.NewMessageReader(msg)
	infos := make(map[string]interface{})
	if server.Check(r.ReadObject(&infos)) {
		return 0, nil
	}
	mb := infos["mailbox"].(rpc.Mailbox)
	info := infos["data"].(*EntityInfo)
	playerid := info.ObjId
	player := App.GetEntity(playerid)
	if player == nil || player.ObjType() != PLAYER {
		log.LogError("sync player player not found", playerid)
		return 0, nil
	}
	for k := range s.childs {
		delete(s.childs, k)
	}
	for k := range s.newobj {
		delete(s.newobj, k)
	}
	s.getAllChilds(player)
	s.sync(info)
	for k := range s.childs { //剩余的表示需要删除
		App.Destroy(k)
	}

	server.Check(server.MailTo(&mb, &src, "BaseProxy.SyncPlayerBak", s.newobj))
	return 0, nil
}

func NewSync() *Sync {
	s := &Sync{}
	s.childs = make(map[ObjectID]Entity, 1024)
	s.newobj = make(map[ObjectID]ObjectID, 1024)
	return s
}
