package datatype

import (
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
)

/*
基本数据类型
*/
const (
	DT_NONE = iota
	DT_INT8
	DT_INT16
	DT_INT32
	DT_INT64
	DT_UINT8
	DT_UINT16
	DT_UINT32
	DT_UINT64
	DT_FLOAT32
	DT_FLOAT64
	DT_STRING
	DT_EXT //内部类型分隔
	DT_VECTOR2
	DT_VECTOR3
	DT_OBJECTID
	DT_MAX
)

/*
标志声明
*/
const (
	FLAGCHANGE   uint16 = 0x1  //脏标志
	FLAGPRIVATE  uint16 = 0x2  //私有
	FLAGPUBLIC   uint16 = 0x4  //公有
	FLAGREALTIME uint16 = 0x8  //即时刷新
	FLAGSAVE     uint16 = 0x10 //保存
)

//objecttype
const (
	NONE   int = 0
	SCENE  int = 1
	PLAYER int = 2
	NPC    int = 3
	ITEM   int = 4
	HELPER int = 5
)

//用户自定义类型
var (
	types = make(map[string]interface{})
)

type usertyper interface {
	//返回值{"字段名", "数据类型","数据长度(字符串类需要提供)"}
	UserType(name string) [][]string
}

func GetUserType(t string, name string) [][]string {
	if user, dup := types[t]; dup {
		if ut, ok := user.(usertyper); ok {
			return ut.UserType(name)
		}

	}

	return nil
}

type Vector2 struct {
	X, Y float32
}

func (vec Vector2) UserType(name string) [][]string {
	return [][]string{{fmt.Sprintf("pv_%s_x", name), "float32", "0"},
		{fmt.Sprintf("pv_%s_y", name), "float32", "0"}}
}

func (vec Vector2) UserVal() []interface{} {
	return []interface{}{vec.X, vec.Y}
}

func (vec *Vector2) FromStr(s string) bool {
	vals := strings.Split(s, ",")
	if len(vals) == 2 {
		X, err := strconv.ParseFloat(vals[0], 32)
		if err != nil {
			return false
		}
		Y, err := strconv.ParseFloat(vals[1], 32)
		if err != nil {
			return false
		}
		vec.X, vec.Y = float32(X), float32(Y)
		return true
	}
	return false
}

type Vector3 struct {
	X, Y, Z float32
}

func (vec Vector3) UserType(name string) [][]string {
	return [][]string{{fmt.Sprintf("pv_%s_x", name), "float32", "0"},
		{fmt.Sprintf("pv_%s_y", name), "float32", "0"},
		{fmt.Sprintf("pv_%s_z", name), "float32", "0"}}
}

func (vec Vector3) UserVal() []interface{} {
	return []interface{}{vec.X, vec.Y, vec.Z}
}

func (vec *Vector3) FromStr(s string) bool {
	vals := strings.Split(s, ",")
	if len(vals) == 3 {
		X, err := strconv.ParseFloat(vals[0], 32)
		if err != nil {
			return false
		}
		Y, err := strconv.ParseFloat(vals[1], 32)
		if err != nil {
			return false
		}
		Z, err := strconv.ParseFloat(vals[2], 32)
		if err != nil {
			return false
		}
		vec.X, vec.Y, vec.Z = float32(X), float32(Y), float32(Z)
		return true
	}

	return false
}

type ObjectID struct {
	Index, Serial int32
}

func (objid ObjectID) Equal(other ObjectID) bool {
	if objid.Index == other.Index && objid.Serial == other.Serial {
		return true
	}
	return false
}

func (objid *ObjectID) IsNil() bool {
	if objid.Index == 0 || objid.Serial == 0 {
		return true
	}

	return false
}

func (objid ObjectID) UserType(name string) [][]string {
	return [][]string{{fmt.Sprintf("po_%s_index", name), "int32", "0"},
		{fmt.Sprintf("po_%s_serial", name), "int32", "0"}}
}

func (objid ObjectID) UserVal() []interface{} {
	return []interface{}{objid.Index, objid.Serial}
}

func init() {
	gob.Register(Vector2{})
	gob.Register(Vector3{})
	gob.Register(ObjectID{})
	types["Vector2"] = &Vector2{}
	types["Vector3"] = &Vector3{}
	types["ObjectID"] = &ObjectID{}
}
