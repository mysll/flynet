package share

import (
	"bytes"
	"data/entity"
	"encoding/gob"
	"libs/log"
)

const (
	REMOVE_OFFLINE = iota
	REMOVE_SWITCH
)

type PlayerInfo struct {
	Account      string
	Name         string
	Scene        string
	X, Y, Z, Dir float32
	Entity       *entity.EntityInfo
}

func GetPlayerInfo(acc string, name string, scene string, x, y, z, dir float32, obj entity.Entityer) (*PlayerInfo, error) {
	info, err := GetItemInfo(obj)
	if err != nil {
		return nil, err
	}

	p := &PlayerInfo{}
	p.Account = acc
	p.Name = name
	p.Scene = scene
	p.X, p.Y, p.Z, p.Dir = x, y, z, dir
	p.Entity = info

	return p, nil
}

func GetItemInfo(obj entity.Entityer) (*entity.EntityInfo, error) {
	item := &entity.EntityInfo{}
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	err := enc.Encode(obj)
	if err != nil {
		log.LogError("encode ", obj.ObjTypeName(), "error,", err)
		return nil, err
	}
	item.Type = obj.ObjTypeName()
	item.Caps = obj.GetCapacity()
	item.DbId = obj.GetDbId()
	item.ObjId = obj.GetObjId()
	item.Index = obj.GetIndex()
	item.Data = buffer.Bytes()

	ls := obj.GetChilds()
	if len(ls) > 0 {
		item.Childs = make([]*entity.EntityInfo, 0, len(ls))
	}
	for _, c := range ls {
		if c != nil {
			child, err := GetItemInfo(c)
			if err != nil {
				return nil, err
			}
			item.Childs = append(item.Childs, child)
		}
	}

	return item, nil
}
