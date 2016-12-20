// Code generated by data parser.
// DO NOT EDIT!
package entity

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math"
	. "server/data/datatype"
	"server/libs/log"
	"server/util"
)

type Item_Save_Property struct {
	Capacity int32  `json:"C"` //容量
	ConfigId string `json:"I"`
	ID       string `json:"1"` //编号
	Time     int32  `json:"2"` //时效
	Amount   int16  `json:"3"` //物品叠加数量
}

//保存到DB的数据
type Item_Save struct {
	Item_Save_Property
}

func (s *Item_Save) Base() string {

	return ""

}

func (s *Item_Save) Marshal() (map[string]interface{}, error) {
	data, err := json.Marshal(*s)
	if err != nil {
		return nil, err
	}

	var ret map[string]interface{}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (s *Item_Save) Unmarshal(data map[string]interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(d, s)

	return err
}

func (s *Item_Save) InsertOrUpdate(eq ExecQueryer, insert bool, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) error {
	var sql string
	var args []interface{}
	if insert {
		sql = "INSERT INTO `tbl_item`(`id`,`capacity`,`configid`,`p_id`,`p_time`,`p_amount`%s ) VALUES(?,?,?,?,?,?%s) "
		args = []interface{}{dbId, s.Capacity, s.ConfigId, s.ID, s.Time, s.Amount}
		sql = fmt.Sprintf(sql, extfields, extplacehold)
		if extobjs != nil {
			args = append(args, extobjs...)
		}
	} else {
		sql = "UPDATE `tbl_item` SET %s`capacity`=?, `configid`=?,`p_id`=?,`p_time`=?,`p_amount`=? WHERE `id` = ?"
		if extobjs != nil {
			args = append(args, extobjs...)
		}
		args = append(args, []interface{}{s.Capacity, s.ConfigId, s.ID, s.Time, s.Amount, dbId}...)
		sql = fmt.Sprintf(sql, extfields)

	}

	if _, err := eq.Exec(sql, args...); err != nil {
		log.LogError("InsertOrUpdate error:", sql, args)
		return err
	}

	return nil
}

func (s *Item_Save) Update(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, false, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	return
}

func (s *Item_Save) Insert(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, true, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	return
}

func (s *Item_Save) Query(dbId uint64) (sql string, args []interface{}) {
	sql = "SELECT `id`,`capacity`,`configid`,`p_id`,`p_time`,`p_amount` %s FROM `tbl_item` WHERE `id`=? LIMIT 1"
	args = []interface{}{dbId}
	return
}

func (s *Item_Save) Load(eq ExecQueryer, dbId uint64, extfield string, extobjs ...interface{}) error {
	sql, a := s.Query(dbId)
	sql = fmt.Sprintf(sql, extfield)
	r, err := eq.Query(sql, a...)
	if err != nil {
		log.LogError("load error:", err)
		return err
	}
	defer r.Close()
	if !r.Next() {
		log.LogError("load error:", sql, a)
		return ErrSqlRowError
	}
	args := []interface{}{&dbId, &s.Capacity, &s.ConfigId, &s.ID, &s.Time, &s.Amount}
	if extobjs != nil {
		args = append(args, extobjs...)
	}
	if err = r.Scan(args...); err != nil {
		log.LogError("load error:", err)
		return err
	}

	return nil
}

type Item_Propertys struct {
	//属性定义
	Name string `type:"property",name:"Name",datatype:"string"` //名称

}

type Item_t struct {
	dirty           bool
	ObjectType      int
	DbId            uint64
	parent          Entity
	ObjId           ObjectID
	Deleted         bool
	NameHash_       int32
	ConfigIdHash    int32
	ContainerInited bool
	Index           int //在容器中的位置
	Childs          []Entity
	ChildNum        int

	Save bool //是否保存
}

type Item struct {
	uid          uint64 //全局ID
	InScene      bool   //是否在场景中
	InBase       bool   //是否在base中
	Mdirty       map[string]interface{}
	Mmodify      map[string]interface{}
	ExtraData    map[string]interface{}
	loading      bool
	quiting      bool
	propupdate   PropUpdater
	prophooker   PropChanger
	propcritical []uint64
	propflag     []uint64

	Item_t
	Item_Save
	Item_Propertys
}

func (obj *Item) SetUID(v uint64) {
	obj.uid = v
}

func (obj *Item) UID() uint64 {
	return obj.uid
}

func (obj *Item) SetInBase(v bool) {
	obj.InBase = v
}

func (obj *Item) IsInBase() bool {
	return obj.InBase
}

func (obj *Item) SetInScene(v bool) {
	obj.InScene = v
}

func (obj *Item) IsInScene() bool {
	return obj.InScene
}

func (obj *Item) SetLoading(loading bool) {
	obj.loading = loading
}

func (obj *Item) IsLoading() bool {
	return obj.loading
}

func (obj *Item) SetQuiting() {
	obj.quiting = true
}

func (obj *Item) IsQuiting() bool {
	return obj.quiting
}

func (obj *Item) Config() string {
	return obj.ConfigId
}

func (obj *Item) SetConfig(config string) {
	obj.ConfigId = config
	obj.ConfigIdHash = Hash(config)
}

func (obj *Item) SetSaveFlag() {
	root := obj.Root()
	if root != nil {
		root.SetSaveFlag()
	} else {
		obj.dirty = true
	}
}

func (obj *Item) ClearSaveFlag() {
	obj.dirty = false

}

func (obj *Item) NeedSave() bool {
	if obj.dirty {
		return true
	}
	return false
}

func (obj *Item) ChangeCapacity(capacity int32) error {

	if !obj.ContainerInited {
		return ErrContainerNotInit
	}

	if capacity == -1 || capacity == obj.Capacity {
		return ErrContainerCapacity
	}

	if capacity < int32(obj.ChildCount()) {
		return ErrContainerCapacity
	}

	newchilds := make([]Entity, capacity)
	if capacity < obj.Capacity { //缩容，位置将进行重排
		idx := 0
		for _, c := range obj.Childs {
			if c != nil {
				c.SetIndex(idx)
				newchilds[idx] = c
				idx++
			}
		}
	} else { //扩容，直接进行复制
		for k, c := range obj.Childs {
			newchilds[k] = c
		}
	}

	obj.ContainerInited = true
	obj.Capacity = capacity
	obj.Childs = newchilds
	return nil
}

//设置子对象的存储容量，-1为无限，无限时，需要提供初始容量。
func (obj *Item) SetCapacity(capacity int32, initcap int32) {

	if obj.ContainerInited {
		return
	}

	if capacity == -1 {
		obj.Childs = make([]Entity, 0, initcap)
	} else if capacity > 0 {
		obj.Childs = make([]Entity, capacity)
	} else {
		obj.Childs = nil
	}
	obj.Capacity = capacity
	obj.ContainerInited = true

}

func (obj *Item) Caps() int32 {
	return obj.Capacity
}

//获取实际的容量
func (obj *Item) RealCaps() int32 {
	if !obj.ContainerInited {
		return 0
	}
	return int32(len(obj.Childs))
}

func (obj *Item) Root() Entity {
	var ent Entity
	if obj.Parent() == nil {
		return nil
	}
	ent = obj
	for {
		if ent.Parent() == nil {
			break
		}
		if ent.Parent().ObjType() == SCENE {
			break
		}
		ent = ent.Parent()
	}
	return ent
}

//获取数据库id
func (obj *Item) DBId() uint64 {
	return obj.DbId
}

func (obj *Item) SetDBId(id uint64) {
	obj.DbId = id
}

func (obj *Item) SetParent(p Entity) {
	obj.parent = p
}

func (obj *Item) Parent() Entity {
	return obj.parent
}

func (obj *Item) SetDeleted(d bool) {
	obj.Deleted = d
}

func (obj *Item) IsDeleted() bool {
	return obj.Deleted
}

func (obj *Item) SetObjId(id ObjectID) {
	obj.ObjId = id
}

func (obj *Item) ObjectId() ObjectID {
	return obj.ObjId
}

//设置名字Hash
func (obj *Item) SetNameHash(v int32) {
	obj.NameHash_ = v
}

//获取名字Hash
func (obj *Item) NameHash() int32 {
	return obj.NameHash_
}

//名字比较
func (obj *Item) NameEqual(name string) bool {
	return obj.Name == name
}

//获取ConfigIdHash
func (obj *Item) ConfigHash() int32 {
	return obj.ConfigIdHash
}

//ID比较
func (obj *Item) ConfigIdEqual(id string) bool {
	return obj.ConfigId == id
}

func (obj *Item) ChildCount() int {
	return obj.ChildNum
}

//移除对象
func (obj *Item) RemoveChild(de Entity) error {
	idx := de.ChildIndex()
	e := obj.GetChild(idx)

	if e != nil && e.ObjectId().Equal(de.ObjectId()) {
		obj.Childs[idx] = nil
		de.SetParent(nil)
		obj.ChildNum--
		return nil
	}

	return ErrChildObjectNotFound
}

//获取子对象
func (obj *Item) AllChilds() []Entity {
	return obj.Childs
}

//获取容器中索引
func (obj *Item) ChildIndex() int {
	return obj.Index
}

//设置索引，逻辑层不要调用
func (obj *Item) SetIndex(idx int) {
	obj.Index = idx
}

//删除所有的子对象
func (obj *Item) ClearChilds() {
	for _, c := range obj.Childs {
		if c != nil {
			obj.RemoveChild(c)
		}
	}

	if obj.Capacity == -1 {
		obj.Childs = obj.Childs[:0]
	}
	obj.ChildNum = 0
}

//增加子对象
func (obj *Item) AddChild(idx int, e Entity) (index int, err error) {
	if !obj.ContainerInited {
		err = ErrContainerNotInit
		return
	}
	if obj.Capacity == -1 {
		for i, v := range obj.Childs {
			if v == nil {
				obj.Childs[i] = e
				e.SetIndex(i)
				e.SetParent(obj)
				obj.ChildNum++
				index = e.ChildIndex()
				return
			}
		}
		obj.Childs = append(obj.Childs, e)
		e.SetIndex(len(obj.Childs) - 1)
		e.SetParent(obj)
		index = e.ChildIndex()
		obj.ChildNum++
		return
	}

	if idx == -1 {
		for i, v := range obj.Childs {
			if v == nil {
				idx = i
				break
			}
		}
		if idx == -1 {
			err = ErrContainerFull
			return
		}
	}

	if idx >= len(obj.Childs) {
		log.LogError("out of range, ", idx, ",", len(obj.Childs))
		err = ErrContainerIndexOutOfRange
		return
	}

	if obj.Childs[idx] != nil {
		err = ErrContainerIndexHasChild
		return
	}

	obj.Childs[idx] = e
	e.SetIndex(idx)
	e.SetParent(obj)
	obj.ChildNum++
	index = e.ChildIndex()
	return

}

//获取子对象
func (obj *Item) GetChild(idx int) Entity {
	if !obj.ContainerInited {
		return nil
	}
	if idx < 0 || idx >= len(obj.Childs) {
		return nil
	}
	return obj.Childs[idx]
}

//通过ID获取子对象
func (obj *Item) FindChildByConfigId(id string) Entity {
	if !obj.ContainerInited {
		return nil
	}
	h := Hash(id)
	for _, v := range obj.Childs {
		if (v != nil) && (v.ConfigHash() == h) && v.ConfigIdEqual(id) {
			return v
		}
	}
	return nil
}
func (obj *Item) FindFirstChildByConfigId(id string) (int, Entity) {
	if !obj.ContainerInited {
		return -1, nil
	}
	h := Hash(id)
	for k, v := range obj.Childs {
		if (v != nil) && (v.ConfigHash() == h) && v.ConfigIdEqual(id) {
			return k + 1, v
		}
	}
	return -1, nil
}
func (obj *Item) NextChildByConfigId(start int, id string) (int, Entity) {

	if !obj.ContainerInited || start == -1 || start >= len(obj.Childs) {
		return -1, nil
	}
	h := Hash(id)
	for k, v := range obj.Childs[start:] {
		if (v != nil) && (v.ConfigHash() == h) && v.ConfigIdEqual(id) {
			return start + k + 1, v
		}
	}
	return -1, nil
}

//通过名称获取子对象
func (obj *Item) FindChildByName(name string) Entity {
	if !obj.ContainerInited {
		return nil
	}
	h := Hash(name)
	for _, v := range obj.Childs {
		if (v != nil) && (v.NameHash() == h) && v.NameEqual(name) {
			return v
		}
	}
	return nil
}
func (obj *Item) FindFirstChildByName(name string) (int, Entity) {
	if !obj.ContainerInited {
		return -1, nil
	}
	h := Hash(name)
	for k, v := range obj.Childs {
		if (v != nil) && (v.NameHash() == h) && v.NameEqual(name) {
			return k + 1, v
		}
	}
	return -1, nil
}
func (obj *Item) NextChildByName(start int, name string) (int, Entity) {

	if !obj.ContainerInited || start == -1 || start >= len(obj.Childs) {
		return -1, nil
	}
	h := Hash(name)
	for k, v := range obj.Childs[start:] {
		if (v != nil) && (v.NameHash() == h) && v.NameEqual(name) {
			return start + k + 1, v
		}
	}
	return -1, nil
}

//交换子对象的位置
func (obj *Item) SwapChild(src int, dest int) error {
	if !obj.ContainerInited {
		return ErrContainerNotInit
	}
	if src < 0 || src >= len(obj.Childs) || dest < 0 || dest >= len(obj.Childs) {
		return ErrContainerIndexOutOfRange
	}

	obj.Childs[src], obj.Childs[dest] = obj.Childs[dest], obj.Childs[src]
	if obj.Childs[src] != nil {
		obj.Childs[src].SetIndex(src)
	}
	if obj.Childs[dest] != nil {
		obj.Childs[dest].SetIndex(dest)
	}
	return nil
}

//获取基类
func (obj *Item) Base() Entity {
	return nil
}

//获取对象类型
func (obj *Item) ObjType() int {
	return obj.ObjectType
}

//额外的数据
func (obj *Item) SetExtraData(key string, value interface{}) {
	obj.ExtraData[key] = value
}

func (obj *Item) FindExtraData(key string) interface{} {
	if v, ok := obj.ExtraData[key]; ok {
		return v
	}
	return nil
}

func (obj *Item) ExtraDatas() map[string]interface{} {
	return obj.ExtraData
}

func (obj *Item) RemoveExtraData(key string) {
	if _, ok := obj.ExtraData[key]; ok {
		delete(obj.ExtraData, key)
	}
}

func (obj *Item) ClearExtraData() {
	for k := range obj.ExtraData {
		delete(obj.ExtraData, k)
	}
}

//获取对象是否保存
func (obj *Item) IsSave() bool {
	return obj.Save
}

//设置对象是否保存
func (obj *Item) SetSave(s bool) {
	obj.Save = s
}

//获取对象类型名
func (obj *Item) ObjTypeName() string {
	return "Item"
}

func (obj *Item) SetPropUpdate(sync PropUpdater) {
	obj.propupdate = sync
}

func (obj *Item) PropUpdate() PropUpdater {
	return obj.propupdate
}

//属性回调接口
func (obj *Item) SetPropHook(hooker PropChanger) {
	obj.prophooker = hooker
}

func (obj *Item) PropFlag(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propflag[index]&(uint64(1)<<bit) != 0
}

func (obj *Item) SetPropFlag(idx int, flag bool) {
	index := idx / 64
	bit := uint(idx) % 64
	if flag {
		obj.propflag[index] = obj.propflag[index] | (uint64(1) << bit)
		return
	}
	obj.propflag[index] = obj.propflag[index] & ^(uint64(1) << bit)
}

func (obj *Item) IsCritical(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propcritical[index]&(uint64(1)<<bit) != 0
}

func (obj *Item) SetCritical(prop string) {
	idx, err := obj.PropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] | (uint64(1) << bit)
}

func (obj *Item) ClearCritical(prop string) {
	idx, err := obj.PropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] & ^(uint64(1) << bit)
}

//获取所有属性
func (obj *Item) Propertys() []string {
	return []string{
		"ID",
		"Time",
		"Amount",
		"Name",
	}
}

//获取所有可视属性
func (obj *Item) VisiblePropertys(typ int) []string {
	if typ == 0 {
		return []string{
			"ID",
			"Time",
			"Amount",
		}
	} else {
		return []string{}
	}

}

//获取属性类型
func (obj *Item) PropertyType(p string) (int, string, error) {
	switch p {
	case "ID":
		return DT_STRING, "string", nil
	case "Time":
		return DT_INT32, "int32", nil
	case "Amount":
		return DT_INT16, "int16", nil
	case "Name":
		return DT_STRING, "string", nil
	default:
		return DT_NONE, "", ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *Item) PropertyIndex(p string) (int, error) {
	switch p {
	case "ID":
		return 0, nil
	case "Time":
		return 1, nil
	case "Amount":
		return 2, nil
	case "Name":
		return 3, nil
	default:
		return -1, ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *Item) Inc(p string, v interface{}) error {
	switch p {
	case "Time":
		var dst int32
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.IncTime(dst)
		}
		return err
	case "Amount":
		var dst int16
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.IncAmount(dst)
		}
		return err
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性名设置值
func (obj *Item) Set(p string, v interface{}) error {
	switch p {
	case "ID":
		val, ok := v.(string)
		if ok {
			obj.SetID(val)
		} else {
			return ErrTypeMismatch
		}
	case "Time":
		var dst int32
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.SetTime(dst)
		}
		return err
	case "Amount":
		var dst int16
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.SetAmount(dst)
		}
		return err
	case "Name":
		val, ok := v.(string)
		if ok {
			obj.SetName(val)
		} else {
			return ErrTypeMismatch
		}
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性索引设置值
func (obj *Item) SetByIndex(index int16, v interface{}) error {
	switch index {
	case 0:
		val, ok := v.(string)
		if ok {
			obj.SetID(val)
		} else {
			return ErrTypeMismatch
		}
	case 1:
		var dst int32
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.SetTime(dst)
		}
		return err
	case 2:
		var dst int16
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.SetAmount(dst)
		}
		return err
	case 3:
		val, ok := v.(string)
		if ok {
			obj.SetName(val)
		} else {
			return ErrTypeMismatch
		}
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性名获取值
func (obj *Item) MustGet(p string) interface{} {
	switch p {
	case "ID":
		return obj.ID
	case "Time":
		return obj.Time
	case "Amount":
		return obj.Amount
	case "Name":
		return obj.Name
	default:
		return nil
	}
}

//通过属性名获取值
func (obj *Item) Get(p string) (val interface{}, err error) {
	switch p {
	case "ID":
		return obj.ID, nil
	case "Time":
		return obj.Time, nil
	case "Amount":
		return obj.Amount, nil
	case "Name":
		return obj.Name, nil
	default:
		return nil, ErrPropertyNotFound
	}
}

//是否需要同步到其它客户端
func (obj *Item) PropertyIsPublic(p string) bool {
	switch p {
	case "ID":
		return false
	case "Time":
		return false
	case "Amount":
		return false
	case "Name":
		return false
	default:
		return false
	}
}

//是否需要同步到自己的客户端
func (obj *Item) PropertyIsPrivate(p string) bool {
	switch p {
	case "ID":
		return true
	case "Time":
		return true
	case "Amount":
		return true
	case "Name":
		return false
	default:
		return false
	}
}

//是否需要存档
func (obj *Item) PropertyIsSave(p string) bool {
	switch p {
	case "ID":
		return true
	case "Time":
		return true
	case "Amount":
		return true
	case "Name":
		return false
	default:
		return false
	}
}

//脏标志(数据保存用)
func (obj *Item) setDirty(p string, v interface{}) {
	//obj.Mdirty[p] = v
	obj.SetSaveFlag()
}

func (obj *Item) Dirtys() map[string]interface{} {

	return obj.Mdirty
}

func (obj *Item) ClearDirty() {
	for k := range obj.Mdirty {
		delete(obj.Mdirty, k)
	}
}

//修改标志(数据同步用)
func (obj *Item) setModify(p string, v interface{}) {
	obj.Mmodify[p] = v
}

func (obj *Item) Modifys() map[string]interface{} {
	return obj.Mmodify
}

func (obj *Item) ClearModify() {
	for k := range obj.Mmodify {
		delete(obj.Mmodify, k)
	}
}

//编号
func (obj *Item) SetID(v string) {
	if obj.ID == v {
		return
	}

	old := obj.ID

	if !obj.InBase { //只有base能够修改自身的数据
		log.LogError("can't change base data")
	}

	obj.ID = v
	if obj.prophooker != nil && obj.IsCritical(0) && !obj.PropFlag(0) {
		obj.SetPropFlag(0, true)
		obj.prophooker.OnPropChange(obj, "ID", old)
		obj.SetPropFlag(0, false)
	}
	obj.setModify("ID", v)

	obj.setDirty("ID", v)
}
func (obj *Item) GetID() string {
	return obj.ID
}

//时效
func (obj *Item) SetTime(v int32) {
	if obj.Time == v {
		return
	}

	old := obj.Time

	if !obj.InBase { //只有base能够修改自身的数据
		log.LogError("can't change base data")
	}

	obj.Time = v
	if obj.prophooker != nil && obj.IsCritical(1) && !obj.PropFlag(1) {
		obj.SetPropFlag(1, true)
		obj.prophooker.OnPropChange(obj, "Time", old)
		obj.SetPropFlag(1, false)
	}
	obj.setModify("Time", v)

	obj.setDirty("Time", v)
}
func (obj *Item) GetTime() int32 {
	return obj.Time
}
func (obj *Item) IncTime(v int32) {
	obj.SetTime(obj.Time + v)
}

//物品叠加数量
func (obj *Item) SetAmount(v int16) {
	if obj.Amount == v {
		return
	}

	old := obj.Amount

	if !obj.InBase { //只有base能够修改自身的数据
		log.LogError("can't change base data")
	}

	obj.Amount = v
	if obj.prophooker != nil && obj.IsCritical(2) && !obj.PropFlag(2) {
		obj.SetPropFlag(2, true)
		obj.prophooker.OnPropChange(obj, "Amount", old)
		obj.SetPropFlag(2, false)
	}
	obj.setModify("Amount", v)

	obj.setDirty("Amount", v)
}
func (obj *Item) GetAmount() int16 {
	return obj.Amount
}
func (obj *Item) IncAmount(v int16) {
	obj.SetAmount(obj.Amount + v)
}

//名称
func (obj *Item) SetName(v string) {
	if obj.Name == v {
		return
	}

	old := obj.Name

	if !obj.InBase { //只有base能够修改自身的数据
		log.LogError("can't change base data")
	}

	obj.Name = v
	if obj.prophooker != nil && obj.IsCritical(3) && !obj.PropFlag(3) {
		obj.SetPropFlag(3, true)
		obj.prophooker.OnPropChange(obj, "Name", old)
		obj.SetPropFlag(3, false)
	}
	obj.NameHash_ = Hash(v)

}
func (obj *Item) GetName() string {
	return obj.Name
}

//初始化所有的表格
func (obj *Item) initRec() {

}

//获取某个表格
func (obj *Item) FindRec(rec string) Record {
	switch rec {
	default:
		return nil
	}
}

//获取所有表格名称
func (obj *Item) RecordNames() []string {
	return []string{}
}

func (obj *Item) ItemInit() {
	obj.quiting = false
	obj.Save = true
	obj.ObjectType = ITEM
	obj.InBase = false
	obj.InScene = false
	obj.uid = 0
}

//重置
func (obj *Item) Reset() {

	//属性初始化
	obj.Item_t = Item_t{}
	obj.Item_Save.Item_Save_Property = Item_Save_Property{}
	obj.Item_Propertys = Item_Propertys{}
	obj.ItemInit()
	//表格初始化

	obj.ClearDirty()
	obj.ClearModify()
	obj.ClearExtraData()
	for k := range obj.propcritical {
		obj.propcritical[k] = 0
		obj.propflag[k] = 0
	}

}

//对象拷贝
func (obj *Item) Copy(other Entity) error {
	if t, ok := other.(*Item); ok {
		//属性复制
		obj.DbId = t.DbId
		obj.NameHash_ = t.NameHash_
		obj.ConfigIdHash = t.ConfigIdHash
		obj.uid = t.uid

		obj.Item_t = t.Item_t
		obj.Item_Save.Item_Save_Property = t.Item_Save_Property
		obj.Item_Propertys = t.Item_Propertys

		//表格复制

		return nil
	}

	return ErrCopyObjError
}

//DB相关
//同步到数据库
func (obj *Item) SyncToDb() {

}

//从data中取出configid
func (obj *Item) GetConfigFromDb(data interface{}) string {
	if v, ok := data.(*Item_Save); ok {
		return v.ConfigId
	}

	return ""
}

//从数据库恢复
func (obj *Item) SyncFromDb(data interface{}) bool {
	if v, ok := data.(*Item_Save); ok {
		obj.Item_Save.Item_Save_Property = v.Item_Save_Property

		obj.NameHash_ = Hash(obj.Name)
		obj.ConfigIdHash = Hash(obj.ConfigId)
		return true
	}

	return false
}

func (obj *Item) SaveLoader() DBSaveLoader {
	return &obj.Item_Save
}

func (obj *Item) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	var err error

	err = encoder.Encode(obj.uid)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.Save)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.DbId)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.NameHash_)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.ConfigIdHash)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.Index)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(obj.Item_Save)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.Item_Propertys)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (obj *Item) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	var err error

	err = decoder.Decode(&obj.uid)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.Save)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.DbId)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.NameHash_)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.ConfigIdHash)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.Index)
	if err != nil {
		return err
	}

	err = decoder.Decode(&obj.Item_Save)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.Item_Propertys)
	if err != nil {
		return err
	}

	return nil
}

//由子类调用的初始化函数
func (obj *Item) baseInit(dirty, modify, extra map[string]interface{}) {
	//初始化表格
	obj.initRec()
	obj.Mdirty = dirty
	obj.Mmodify = modify
	obj.ExtraData = extra
}

func (obj *Item) Serial() ([]byte, error) {
	ar := util.NewStoreArchiver(nil)
	ps := obj.VisiblePropertys(0)
	ar.Write(int16(len(ps)))

	ar.Write(int16(0))
	ar.Write(obj.ID)
	ar.Write(int16(1))
	ar.Write(obj.Time)
	ar.Write(int16(2))
	ar.Write(obj.Amount)
	return ar.Data(), nil
}

func (obj *Item) SerialModify() ([]byte, error) {
	if len(obj.Mmodify) == 0 {
		return nil, nil
	}
	ar := util.NewStoreArchiver(nil)
	ar.Write(int16(len(obj.Mmodify)))
	for k, v := range obj.Mmodify {
		if !obj.PropertyIsPrivate(k) {
			continue
		}
		idx, _ := obj.PropertyIndex(k)

		ar.Write(int16(idx))
		ar.Write(v)
	}

	return ar.Data(), nil
}

func (obj *Item) IsSceneData(prop string) bool {
	idx, err := obj.PropertyIndex(prop)
	if err != nil {
		return false
	}

	return IsItemSceneData(idx)
}

//通过scenedata同步
func (obj *Item) SyncFromSceneData(val interface{}) error {
	var sd *ItemSceneData
	var ok bool
	if sd, ok = val.(*ItemSceneData); !ok {
		return fmt.Errorf("type not ItemSceneData", sd)
	}

	return nil
}

func (obj *Item) SceneData() interface{} {
	sd := &ItemSceneData{}

	//属性

	//表格
	return sd
}

//创建函数
func CreateItem() *Item {
	obj := &Item{}

	obj.ObjectType = ITEM
	obj.initRec()

	obj.propcritical = make([]uint64, int(math.Ceil(float64(4)/64)))
	obj.propflag = make([]uint64, int(math.Ceil(float64(4)/64)))
	obj.ItemInit()

	obj.Mdirty = make(map[string]interface{}, 32)
	obj.Mmodify = make(map[string]interface{}, 32)
	obj.ExtraData = make(map[string]interface{}, 16)

	return obj
}

type ItemSceneData struct {
}

func IsItemSceneData(idx int) bool {
	switch idx {
	case 0: //编号
		return false
	case 1: //时效
		return false
	case 2: //物品叠加数量
		return false
	case 3: //名称
		return false
	}
	return false
}

func ItemInit() {
	gob.Register(&Item_Save{})
	gob.Register(&Item{})
	gob.Register(&ItemSceneData{})
}
