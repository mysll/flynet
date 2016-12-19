package share

import (
	"bytes"
	"encoding/gob"
	"server/data/datatype"
	"server/libs/log"
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
	Entity       *datatype.EntityInfo
}

func GetPlayerInfo(acc string, name string, scene string, x, y, z, dir float32, obj datatype.Entity) (*PlayerInfo, error) {
	info, err := GetItemInfo(obj, true)
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

func GetItemInfo(obj datatype.Entity, syncchild bool) (*datatype.EntityInfo, error) {
	item := &datatype.EntityInfo{}
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	err := enc.Encode(obj)
	if err != nil {
		log.LogError("encode ", obj.ObjTypeName(), "error,", err)
		return nil, err
	}
	item.Type = obj.ObjTypeName()
	item.Caps = obj.Caps()
	item.DbId = obj.DBId()
	item.ObjId = obj.ObjectId()
	item.Index = obj.ChildIndex()
	item.Data = buffer.Bytes()

	if syncchild {
		ls := obj.AllChilds()
		if len(ls) > 0 {
			item.Childs = make([]*datatype.EntityInfo, 0, len(ls))
		}
		for _, c := range ls {
			if c != nil {
				child, err := GetItemInfo(c, syncchild)
				if err != nil {
					return nil, err
				}
				item.Childs = append(item.Childs, child)
			}
		}
	}

	return item, nil
}
