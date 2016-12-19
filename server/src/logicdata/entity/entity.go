// Code generated by data parser.
// DO NOT EDIT!
package entity

import (
	"encoding/gob"
	. "server/data/datatype"
)

//获取类型
func GetType(name string) int {
	switch name {
	case "BaseScene":
		return SCENE
	case "Container":
		return ITEM
	case "Player":
		return PLAYER
	case "Item":
		return ITEM
	case "GlobalData":
		return HELPER
	case "GlobalSet":
		return HELPER
	default:
		return NONE
	}
}

func CreateSaveLoader(typ string) DBSaveLoader {
	switch typ {
	case "BaseScene":
		return &BaseScene_Save{}
	case "Container":
		return &Container_Save{}
	case "Player":
		return &Player_Save{}
	case "Item":
		return &Item_Save{}
	case "GlobalData":
		return &GlobalData_Save{}
	case "GlobalSet":
		return &GlobalSet_Save{}
	default:
		return nil
	}
}

func Hash(str string) int32 {
	hash := 5381
	for _, c := range str {
		hash += (hash << 5) + int(c)
	}

	return int32(hash & 0x7FFFFFFF)
}

func IsBaseScene(ent Entity) bool {
	return ent.ObjTypeName() == "BaseScene"
}

func IsContainer(ent Entity) bool {
	return ent.ObjTypeName() == "Container"
}

func IsPlayer(ent Entity) bool {
	return ent.ObjTypeName() == "Player"
}

func IsItem(ent Entity) bool {
	return ent.ObjTypeName() == "Item"
}

func IsGlobalData(ent Entity) bool {
	return ent.ObjTypeName() == "GlobalData"
}

func IsGlobalSet(ent Entity) bool {
	return ent.ObjTypeName() == "GlobalSet"
}

//初始化函数
func init() {

	Register("BaseScene", func() Entity {
		return CreateBaseScene()
	})
	BaseSceneInit()

	Register("Container", func() Entity {
		return CreateContainer()
	})
	ContainerInit()

	Register("Player", func() Entity {
		return CreatePlayer()
	})
	PlayerInit()

	Register("Item", func() Entity {
		return CreateItem()
	})
	ItemInit()

	Register("GlobalData", func() Entity {
		return CreateGlobalData()
	})
	GlobalDataInit()

	Register("GlobalSet", func() Entity {
		return CreateGlobalSet()
	})
	GlobalSetInit()

	gob.Register(&EntityInfo{})
}
