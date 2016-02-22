package base

import (
	"bytes"
	. "data/datatype"
	"data/entity"
	"encoding/gob"
	"errors"
	"libs/log"
	"libs/rpc"
	"server"
)

type Sync struct {
	childs map[ObjectID]entity.Entityer
	newobj map[ObjectID]ObjectID
}

func (s *Sync) getAllChilds(obj entity.Entityer) {
	s.childs[obj.GetObjId()] = obj
	cs := obj.GetChilds()
	for _, c := range cs {
		if c != nil {
			s.getAllChilds(c)
		}
	}
}
func (s *Sync) sync(info *entity.EntityInfo) (obj entity.Entityer, err error) {
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
		var cobj entity.Entityer
		cobj, err = s.sync(c)
		if err != nil {
			return
		}
		obj.AddChild(info.Index, cobj)
	}

	return
}

func (s *Sync) SyncPlayer(src rpc.Mailbox, infos map[string]interface{}) error {
	mb := infos["mailbox"].(rpc.Mailbox)
	info := infos["data"].(*entity.EntityInfo)
	playerid := info.ObjId
	player := App.GetEntity(playerid)
	if player == nil || player.ObjType() != PLAYER {
		log.LogError("sync player player not found", playerid)
		return nil
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

	err := server.MailTo(&mb, &src, "BaseProxy.SyncPlayerBak", s.newobj)
	return err
}

func NewSync() *Sync {
	s := &Sync{}
	s.childs = make(map[ObjectID]entity.Entityer, 1024)
	s.newobj = make(map[ObjectID]ObjectID, 1024)
	return s
}
