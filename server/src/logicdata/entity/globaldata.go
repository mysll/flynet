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

//Test任务表行定义
type GlobalDataTestRecRow struct {
	ID   string `json:"1"` //Test1
	Flag int8   `json:"2"` //Test2
}

//Test任务表
type GlobalDataTestRec struct {
	MaxRows int `json:"-"`
	Cols    int `json:"-"`
	Rows    []GlobalDataTestRecRow
	Dirty   bool `json:"-"`
	monitor TableMonitor
	owner   *GlobalData
}

type GlobalData_Save_Property struct {
	Capacity int32  `json:"C"` //容量
	ConfigId string `json:"I"`
	Name     string //名称
	Test1    string //测试数据1
	Test2    string //测试数据2
}

//保存到DB的数据
type GlobalData_Save struct {
	GlobalData_Save_Property

	TestRec_r GlobalDataTestRec
}

func (s *GlobalData_Save) Base() string {

	return ""

}

func (s *GlobalData_Save) Marshal() (map[string]interface{}, error) {
	data, err := json.Marshal(*s)
	if err != nil {
		return nil, err
	}

	var ret map[string]interface{}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (s *GlobalData_Save) Unmarshal(data map[string]interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(d, s)

	return err
}

func (s *GlobalData_Save) InsertOrUpdate(eq ExecQueryer, insert bool, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) error {
	var sql string
	var args []interface{}
	if insert {
		sql = "INSERT INTO `tbl_globaldata`(`id`,`capacity`,`configid`,`p_name`,`p_test1`,`p_test2`%s ) VALUES(?,?,?,?,?,?%s) "
		args = []interface{}{dbId, s.Capacity, s.ConfigId, s.Name, s.Test1, s.Test2}
		sql = fmt.Sprintf(sql, extfields, extplacehold)
		if extobjs != nil {
			args = append(args, extobjs...)
		}
	} else {
		sql = "UPDATE `tbl_globaldata` SET %s`capacity`=?, `configid`=?,`p_name`=?,`p_test1`=?,`p_test2`=? WHERE `id` = ?"
		if extobjs != nil {
			args = append(args, extobjs...)
		}
		args = append(args, []interface{}{s.Capacity, s.ConfigId, s.Name, s.Test1, s.Test2, dbId}...)
		sql = fmt.Sprintf(sql, extfields)

	}

	if _, err := eq.Exec(sql, args...); err != nil {
		log.LogError("InsertOrUpdate error:", sql, args)
		return err
	}

	return nil
}

func (s *GlobalData_Save) Update(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, false, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	if err = s.TestRec_r.Update(eq, dbId); err != nil {
		return
	}

	return
}

func (s *GlobalData_Save) Insert(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, true, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	if err = s.TestRec_r.Update(eq, dbId); err != nil {
		return
	}

	return
}

func (s *GlobalData_Save) Query(dbId uint64) (sql string, args []interface{}) {
	sql = "SELECT `id`,`capacity`,`configid`,`p_name`,`p_test1`,`p_test2` %s FROM `tbl_globaldata` WHERE `id`=? LIMIT 1"
	args = []interface{}{dbId}
	return
}

func (s *GlobalData_Save) Load(eq ExecQueryer, dbId uint64, extfield string, extobjs ...interface{}) error {
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
	args := []interface{}{&dbId, &s.Capacity, &s.ConfigId, &s.Name, &s.Test1, &s.Test2}
	if extobjs != nil {
		args = append(args, extobjs...)
	}
	if err = r.Scan(args...); err != nil {
		log.LogError("load error:", err)
		return err
	}

	if err = s.TestRec_r.Load(eq, dbId); err != nil {
		log.LogError("load error:", err)
		return err
	}

	return nil
}

type GlobalData_Propertys struct {
	//属性定义

}

type GlobalData_t struct {
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

type GlobalData struct {
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

	GlobalData_t
	GlobalData_Save
	GlobalData_Propertys

	//表格定义
}

func (obj *GlobalData) SetUID(v uint64) {
	obj.uid = v
}

func (obj *GlobalData) UID() uint64 {
	return obj.uid
}

func (obj *GlobalData) SetInBase(v bool) {
	obj.InBase = v
}

func (obj *GlobalData) IsInBase() bool {
	return obj.InBase
}

func (obj *GlobalData) SetInScene(v bool) {
	obj.InScene = v
}

func (obj *GlobalData) IsInScene() bool {
	return obj.InScene
}

func (obj *GlobalData) SetLoading(loading bool) {
	obj.loading = loading
}

func (obj *GlobalData) IsLoading() bool {
	return obj.loading
}

func (obj *GlobalData) SetQuiting() {
	obj.quiting = true
}

func (obj *GlobalData) IsQuiting() bool {
	return obj.quiting
}

func (obj *GlobalData) Config() string {
	return obj.ConfigId
}

func (obj *GlobalData) SetConfig(config string) {
	obj.ConfigId = config
	obj.ConfigIdHash = Hash(config)
}

func (obj *GlobalData) SetSaveFlag() {
	root := obj.Root()
	if root != nil {
		root.SetSaveFlag()
	} else {
		obj.dirty = true
	}
}

func (obj *GlobalData) ClearSaveFlag() {
	obj.dirty = false

	obj.TestRec_r.Dirty = false
}

func (obj *GlobalData) NeedSave() bool {
	if obj.dirty || obj.TestRec_r.Dirty {
		return true
	}
	return false
}

func (obj *GlobalData) ChangeCapacity(capacity int32) error {

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
func (obj *GlobalData) SetCapacity(capacity int32, initcap int32) {

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

func (obj *GlobalData) Caps() int32 {
	return obj.Capacity
}

//获取实际的容量
func (obj *GlobalData) RealCaps() int32 {
	if !obj.ContainerInited {
		return 0
	}
	return int32(len(obj.Childs))
}

func (obj *GlobalData) Root() Entity {
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
func (obj *GlobalData) DBId() uint64 {
	return obj.DbId
}

func (obj *GlobalData) SetDBId(id uint64) {
	obj.DbId = id
}

func (obj *GlobalData) SetParent(p Entity) {
	obj.parent = p
}

func (obj *GlobalData) Parent() Entity {
	return obj.parent
}

func (obj *GlobalData) SetDeleted(d bool) {
	obj.Deleted = d
}

func (obj *GlobalData) IsDeleted() bool {
	return obj.Deleted
}

func (obj *GlobalData) SetObjId(id ObjectID) {
	obj.ObjId = id
}

func (obj *GlobalData) ObjectId() ObjectID {
	return obj.ObjId
}

//设置名字Hash
func (obj *GlobalData) SetNameHash(v int32) {
	obj.NameHash_ = v
}

//获取名字Hash
func (obj *GlobalData) NameHash() int32 {
	return obj.NameHash_
}

//名字比较
func (obj *GlobalData) NameEqual(name string) bool {
	return obj.Name == name
}

//获取ConfigIdHash
func (obj *GlobalData) ConfigHash() int32 {
	return obj.ConfigIdHash
}

//ID比较
func (obj *GlobalData) ConfigIdEqual(id string) bool {
	return obj.ConfigId == id
}

func (obj *GlobalData) ChildCount() int {
	return obj.ChildNum
}

//移除对象
func (obj *GlobalData) RemoveChild(de Entity) error {
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
func (obj *GlobalData) AllChilds() []Entity {
	return obj.Childs
}

//获取容器中索引
func (obj *GlobalData) ChildIndex() int {
	return obj.Index
}

//设置索引，逻辑层不要调用
func (obj *GlobalData) SetIndex(idx int) {
	obj.Index = idx
}

//删除所有的子对象
func (obj *GlobalData) ClearChilds() {
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
func (obj *GlobalData) AddChild(idx int, e Entity) (index int, err error) {
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
func (obj *GlobalData) GetChild(idx int) Entity {
	if !obj.ContainerInited {
		return nil
	}
	if idx < 0 || idx >= len(obj.Childs) {
		return nil
	}
	return obj.Childs[idx]
}

//通过ID获取子对象
func (obj *GlobalData) FindChildByConfigId(id string) Entity {
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
func (obj *GlobalData) FindFirstChildByConfigId(id string) (int, Entity) {
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
func (obj *GlobalData) NextChildByConfigId(start int, id string) (int, Entity) {

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
func (obj *GlobalData) FindChildByName(name string) Entity {
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
func (obj *GlobalData) FindFirstChildByName(name string) (int, Entity) {
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
func (obj *GlobalData) NextChildByName(start int, name string) (int, Entity) {

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
func (obj *GlobalData) SwapChild(src int, dest int) error {
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
func (obj *GlobalData) Base() Entity {
	return nil
}

//获取对象类型
func (obj *GlobalData) ObjType() int {
	return obj.ObjectType
}

//额外的数据
func (obj *GlobalData) SetExtraData(key string, value interface{}) {
	obj.ExtraData[key] = value
}

func (obj *GlobalData) FindExtraData(key string) interface{} {
	if v, ok := obj.ExtraData[key]; ok {
		return v
	}
	return nil
}

func (obj *GlobalData) ExtraDatas() map[string]interface{} {
	return obj.ExtraData
}

func (obj *GlobalData) RemoveExtraData(key string) {
	if _, ok := obj.ExtraData[key]; ok {
		delete(obj.ExtraData, key)
	}
}

func (obj *GlobalData) ClearExtraData() {
	for k := range obj.ExtraData {
		delete(obj.ExtraData, k)
	}
}

//获取对象是否保存
func (obj *GlobalData) IsSave() bool {
	return obj.Save
}

//设置对象是否保存
func (obj *GlobalData) SetSave(s bool) {
	obj.Save = s
}

//获取对象类型名
func (obj *GlobalData) ObjTypeName() string {
	return "GlobalData"
}

func (obj *GlobalData) SetPropUpdate(sync PropUpdater) {
	obj.propupdate = sync
}

func (obj *GlobalData) PropUpdate() PropUpdater {
	return obj.propupdate
}

//属性回调接口
func (obj *GlobalData) SetPropHook(hooker PropChanger) {
	obj.prophooker = hooker
}

func (obj *GlobalData) PropFlag(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propflag[index]&(uint64(1)<<bit) != 0
}

func (obj *GlobalData) SetPropFlag(idx int, flag bool) {
	index := idx / 64
	bit := uint(idx) % 64
	if flag {
		obj.propflag[index] = obj.propflag[index] | (uint64(1) << bit)
		return
	}
	obj.propflag[index] = obj.propflag[index] & ^(uint64(1) << bit)
}

func (obj *GlobalData) IsCritical(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propcritical[index]&(uint64(1)<<bit) != 0
}

func (obj *GlobalData) SetCritical(prop string) {
	idx, err := obj.PropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] | (uint64(1) << bit)
}

func (obj *GlobalData) ClearCritical(prop string) {
	idx, err := obj.PropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] & ^(uint64(1) << bit)
}

//获取所有属性
func (obj *GlobalData) Propertys() []string {
	return []string{
		"Name",
		"Test1",
		"Test2",
	}
}

//获取所有可视属性
func (obj *GlobalData) VisiblePropertys(typ int) []string {
	if typ == 0 {
		return []string{
			"Name",
			"Test1",
			"Test2",
		}
	} else {
		return []string{}
	}

}

//获取属性类型
func (obj *GlobalData) PropertyType(p string) (int, string, error) {
	switch p {
	case "Name":
		return DT_STRING, "string", nil
	case "Test1":
		return DT_STRING, "string", nil
	case "Test2":
		return DT_STRING, "string", nil
	default:
		return DT_NONE, "", ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *GlobalData) PropertyIndex(p string) (int, error) {
	switch p {
	case "Name":
		return 0, nil
	case "Test1":
		return 1, nil
	case "Test2":
		return 2, nil
	default:
		return -1, ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *GlobalData) Inc(p string, v interface{}) error {
	switch p {
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性名设置值
func (obj *GlobalData) Set(p string, v interface{}) error {
	switch p {
	case "Name":
		val, ok := v.(string)
		if ok {
			obj.SetName(val)
		} else {
			return ErrTypeMismatch
		}
	case "Test1":
		val, ok := v.(string)
		if ok {
			obj.SetTest1(val)
		} else {
			return ErrTypeMismatch
		}
	case "Test2":
		val, ok := v.(string)
		if ok {
			obj.SetTest2(val)
		} else {
			return ErrTypeMismatch
		}
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性索引设置值
func (obj *GlobalData) SetByIndex(index int16, v interface{}) error {
	switch index {
	case 0:
		val, ok := v.(string)
		if ok {
			obj.SetName(val)
		} else {
			return ErrTypeMismatch
		}
	case 1:
		val, ok := v.(string)
		if ok {
			obj.SetTest1(val)
		} else {
			return ErrTypeMismatch
		}
	case 2:
		val, ok := v.(string)
		if ok {
			obj.SetTest2(val)
		} else {
			return ErrTypeMismatch
		}
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性名获取值
func (obj *GlobalData) MustGet(p string) interface{} {
	switch p {
	case "Name":
		return obj.Name
	case "Test1":
		return obj.Test1
	case "Test2":
		return obj.Test2
	default:
		return nil
	}
}

//通过属性名获取值
func (obj *GlobalData) Get(p string) (val interface{}, err error) {
	switch p {
	case "Name":
		return obj.Name, nil
	case "Test1":
		return obj.Test1, nil
	case "Test2":
		return obj.Test2, nil
	default:
		return nil, ErrPropertyNotFound
	}
}

//是否需要同步到其它客户端
func (obj *GlobalData) PropertyIsPublic(p string) bool {
	switch p {
	case "Name":
		return false
	case "Test1":
		return false
	case "Test2":
		return false
	default:
		return false
	}
}

//是否需要同步到自己的客户端
func (obj *GlobalData) PropertyIsPrivate(p string) bool {
	switch p {
	case "Name":
		return true
	case "Test1":
		return true
	case "Test2":
		return true
	default:
		return false
	}
}

//是否需要存档
func (obj *GlobalData) PropertyIsSave(p string) bool {
	switch p {
	case "Name":
		return true
	case "Test1":
		return true
	case "Test2":
		return true
	default:
		return false
	}
}

//脏标志(数据保存用)
func (obj *GlobalData) setDirty(p string, v interface{}) {
	//obj.Mdirty[p] = v
	obj.SetSaveFlag()
}

func (obj *GlobalData) Dirtys() map[string]interface{} {

	return obj.Mdirty
}

func (obj *GlobalData) ClearDirty() {
	for k := range obj.Mdirty {
		delete(obj.Mdirty, k)
	}
}

//修改标志(数据同步用)
func (obj *GlobalData) setModify(p string, v interface{}) {
	obj.Mmodify[p] = v
}

func (obj *GlobalData) Modifys() map[string]interface{} {
	return obj.Mmodify
}

func (obj *GlobalData) ClearModify() {
	for k := range obj.Mmodify {
		delete(obj.Mmodify, k)
	}
}

//名称
func (obj *GlobalData) SetName(v string) {
	if obj.Name == v {
		return
	}

	old := obj.Name

	obj.Name = v
	if obj.prophooker != nil && obj.IsCritical(0) && !obj.PropFlag(0) {
		obj.SetPropFlag(0, true)
		obj.prophooker.OnPropChange(obj, "Name", old)
		obj.SetPropFlag(0, false)
	}
	obj.NameHash_ = Hash(v)
	if obj.propupdate != nil {
		obj.propupdate.Update(obj, 0, v)
	} else {
		obj.setModify("Name", v)
	}

	obj.setDirty("Name", v)
}
func (obj *GlobalData) GetName() string {
	return obj.Name
}

//测试数据1
func (obj *GlobalData) SetTest1(v string) {
	if obj.Test1 == v {
		return
	}

	old := obj.Test1

	obj.Test1 = v
	if obj.prophooker != nil && obj.IsCritical(1) && !obj.PropFlag(1) {
		obj.SetPropFlag(1, true)
		obj.prophooker.OnPropChange(obj, "Test1", old)
		obj.SetPropFlag(1, false)
	}
	if obj.propupdate != nil {
		obj.propupdate.Update(obj, 1, v)
	} else {
		obj.setModify("Test1", v)
	}

	obj.setDirty("Test1", v)
}
func (obj *GlobalData) GetTest1() string {
	return obj.Test1
}

//测试数据2
func (obj *GlobalData) SetTest2(v string) {
	if obj.Test2 == v {
		return
	}

	old := obj.Test2

	obj.Test2 = v
	if obj.prophooker != nil && obj.IsCritical(2) && !obj.PropFlag(2) {
		obj.SetPropFlag(2, true)
		obj.prophooker.OnPropChange(obj, "Test2", old)
		obj.SetPropFlag(2, false)
	}
	if obj.propupdate != nil {
		obj.propupdate.Update(obj, 2, v)
	} else {
		obj.setModify("Test2", v)
	}

	obj.setDirty("Test2", v)
}
func (obj *GlobalData) GetTest2() string {
	return obj.Test2
}

func (rec *GlobalDataTestRec) Marshal() ([]byte, error) {
	return json.Marshal(rec)
}

func (rec *GlobalDataTestRec) Unmarshal(data []byte) error {
	return json.Unmarshal(data, rec)
}

//DB
func (rec *GlobalDataTestRec) Update(eq ExecQueryer, dbId uint64) error {

	if !rec.Dirty {
		return nil
	}

	data, err := rec.Marshal()
	if err != nil {
		return err
	}
	sql := "UPDATE `tbl_globaldata` SET `r_testrec`=? WHERE `id` = ?"

	if _, err := eq.Exec(sql, data, dbId); err != nil {
		log.LogError("update record GlobalDataTestRec error:", sql, data, dbId)
		return err
	}

	return nil
}

func (rec *GlobalDataTestRec) Load(eq ExecQueryer, dbId uint64) error {

	rec.Rows = rec.Rows[:0]

	sql := "SELECT `r_testrec` FROM `tbl_globaldata` WHERE `id`=? LIMIT 1"
	r, err := eq.Query(sql, dbId)
	if err != nil {
		log.LogError("load record GlobalDataTestRec error:", err)
		return err
	}
	defer r.Close()
	if !r.Next() {
		log.LogError("load record GlobalDataTestRec error:", sql, dbId)
		return ErrSqlRowError
	}
	var json []byte
	if err = r.Scan(&json); err != nil {
		log.LogError("load record GlobalDataTestRec error:", err)
		return err
	}

	if json == nil || len(json) < 2 {
		log.LogWarning("load record GlobalDataTestRec error: nil")
		return nil
	}

	err = rec.Unmarshal(json)
	if err != nil {
		log.LogError("unmarshal record GlobalDataTestRec error:", err)
		return err
	}

	return nil
}

func (rec *GlobalDataTestRec) Name() string {
	return "TestRec"
}

//表格的容量
func (rec *GlobalDataTestRec) Caps() int {
	return rec.MaxRows
}

//表格当前的行数
func (rec *GlobalDataTestRec) RowCount() int {
	return len(rec.Rows)
}

//获取列定义
func (rec *GlobalDataTestRec) ColTypes() ([]int, []string) {
	col := []int{DT_STRING, DT_INT8}
	cols := []string{"string", "int8"}
	return col, cols
}

//获取列数
func (rec *GlobalDataTestRec) ColCount() int {
	return rec.Cols
}

//是否要同步到客户端
func (rec *GlobalDataTestRec) IsVisible() bool {
	return true
}

//脏标志
func (rec *GlobalDataTestRec) IsDirty() bool {
	return rec.Dirty
}

func (rec *GlobalDataTestRec) ClearDirty() {
	rec.Dirty = false
}

func (rec *GlobalDataTestRec) SetMonitor(s TableMonitor) {
	rec.monitor = s
}

func (rec *GlobalDataTestRec) Monitor() TableMonitor {
	return rec.monitor
}

//序列化
func (rec *GlobalDataTestRec) Serial() ([]byte, error) {
	ar := util.NewStoreArchiver(nil)
	for _, v := range rec.Rows {
		ar.WriteString(v.ID)
		ar.Write(v.Flag)
	}
	return ar.Data(), nil
}

//序列化一行
func (rec *GlobalDataTestRec) SerialRow(row int) ([]byte, error) {
	if row < 0 || row >= len(rec.Rows) {
		return nil, ErrRowError
	}
	ar := util.NewStoreArchiver(nil)
	v := rec.Rows[row]
	ar.WriteString(v.ID)
	ar.Write(v.Flag)
	return ar.Data(), nil
}

//通过行列设置值
func (rec *GlobalDataTestRec) Set(row, col int, val interface{}) error {

	if row < 0 || row >= len(rec.Rows) {
		return ErrRowError
	}

	if col < 0 || col >= 2 {
		return ErrColError
	}

	r := &rec.Rows[row]

	switch col {
	case 0:
		val, ok := val.(string)
		if ok {
			r.ID = val
		} else {
			return ErrTypeMismatch
		}
	case 1:
		val, ok := val.(int8)
		if ok {
			r.Flag = val
		} else {
			return ErrTypeMismatch
		}
	}
	if rec.monitor != nil {
		rec.monitor.RecModify(rec.owner, rec, row, col)
	}
	rec.Dirty = true
	return nil
}

//通过行列获取值
func (rec *GlobalDataTestRec) Get(row, col int) (val interface{}, err error) {
	if row < 0 || row >= len(rec.Rows) {
		err = ErrRowError
		return
	}

	if col < 0 || col >= 2 {
		err = ErrColError
		return
	}

	r := rec.Rows[row]

	switch col {
	case 0:
		val = r.ID
	case 1:
		val = r.Flag
	}

	return
}

//查找Test1
func (rec *GlobalDataTestRec) FindID(v string) int {
	for idx, row := range rec.Rows {
		if row.ID == v {
			return idx
		}
	}
	return -1
}

//查找Test1
func (rec *GlobalDataTestRec) FindNextID(v string, itr int) int {
	itr++
	if itr+1 >= len(rec.Rows) {
		return -1
	}
	for idx, row := range rec.Rows[itr:] {
		if row.ID == v {
			return idx
		}
	}
	return -1
}

//设置Test1
func (rec *GlobalDataTestRec) SetID(row int, v string) error {

	if row < 0 || row >= len(rec.Rows) {
		return ErrRowError
	}

	rec.Rows[row].ID = v
	rec.Dirty = true
	if rec.monitor != nil {
		rec.monitor.RecModify(rec.owner, rec, row, 0)
	}
	return nil
}

//获取Test1
func (rec *GlobalDataTestRec) GetID(row int) (val string, err error) {
	if row < 0 || row >= len(rec.Rows) {
		err = ErrRowError
		return
	}

	val = rec.Rows[row].ID
	return
}

//查找Test2
func (rec *GlobalDataTestRec) FindFlag(v int8) int {
	for idx, row := range rec.Rows {
		if row.Flag == v {
			return idx
		}
	}
	return -1
}

//查找Test2
func (rec *GlobalDataTestRec) FindNextFlag(v int8, itr int) int {
	itr++
	if itr+1 >= len(rec.Rows) {
		return -1
	}
	for idx, row := range rec.Rows[itr:] {
		if row.Flag == v {
			return idx
		}
	}
	return -1
}

//设置Test2
func (rec *GlobalDataTestRec) SetFlag(row int, v int8) error {

	if row < 0 || row >= len(rec.Rows) {
		return ErrRowError
	}

	rec.Rows[row].Flag = v
	rec.Dirty = true
	if rec.monitor != nil {
		rec.monitor.RecModify(rec.owner, rec, row, 1)
	}
	return nil
}

//获取Test2
func (rec *GlobalDataTestRec) GetFlag(row int) (val int8, err error) {
	if row < 0 || row >= len(rec.Rows) {
		err = ErrRowError
		return
	}

	val = rec.Rows[row].Flag
	return
}

//设置一行的值
func (rec *GlobalDataTestRec) SetRow(row int, args ...interface{}) error {

	if _, ok := args[0].(string); !ok {
		return ErrColTypeError
	}
	if _, ok := args[1].(int8); !ok {
		return ErrColTypeError
	}
	return rec.SetRowValue(row, args[0].(string), args[1].(int8))
}

func (rec *GlobalDataTestRec) SetRowInterface(row int, rowvalue interface{}) error {

	if row < 0 || row >= len(rec.Rows) {
		return ErrRowError
	}

	if value, ok := rowvalue.(GlobalDataTestRecRow); ok {
		rec.Rows[row] = value
		if rec.monitor != nil {
			rec.monitor.RecSetRow(rec.owner, rec, row)
		}
		rec.Dirty = true
		return nil
	}

	return ErrColTypeError
}

func (rec *GlobalDataTestRec) SetRowByBytes(row int, rowdata []byte) error {
	if rec.owner.InBase && rec.owner.InScene { //当玩家在场景中时，在base中修改scenedata，在同步时会被覆盖.
		log.LogError("the TaskAccepted will be overwritten by scenedata")
	}

	lr := util.NewLoadArchiver(rowdata)

	var id string
	var flag int8

	if err := lr.Read(&id); err != nil {
		return err
	}
	if err := lr.Read(&flag); err != nil {
		return err
	}

	return rec.SetRowValue(row, id, flag)
}

//设置一行的值
func (rec *GlobalDataTestRec) SetRowValue(row int, id string, flag int8) error {
	if rec.owner.InBase && rec.owner.InScene { //当玩家在场景中时，在base中修改scenedata，在同步时会被覆盖.
		log.LogError("the TestRec will be overwritten by scenedata")
	}

	if row < 0 || row >= len(rec.Rows) {
		return ErrRowError
	}

	rec.Rows[row].ID = id
	rec.Rows[row].Flag = flag

	if rec.monitor != nil {
		rec.monitor.RecSetRow(rec.owner, rec, row)
	}
	rec.Dirty = true
	return nil
}

//增加一行,row=-1时,在表格最后面插入一行,否则在row处插入,返回-1插入失败,否则返回插入的行号
func (rec *GlobalDataTestRec) Add(row int, args ...interface{}) int {

	if len(args) != rec.Cols {
		return -1
	}

	if _, ok := args[0].(string); !ok {
		return -1
	}
	if _, ok := args[1].(int8); !ok {
		return -1
	}
	return rec.AddRowValue(row, args[0].(string), args[1].(int8))
}

//增加一行,row=-1时,在表格最后面插入一行,否则在row处插入,返回-1插入失败,否则返回插入的行号
func (rec *GlobalDataTestRec) AddByBytes(row int, rowdata []byte) int {
	if rec.owner.InBase && rec.owner.InScene { //当玩家在场景中时，在base中修改scenedata，在同步时会被覆盖.
		log.LogError("the TaskAccepted will be overwritten by scenedata")
	}

	lr := util.NewLoadArchiver(rowdata)

	var id string
	var flag int8

	if err := lr.Read(&id); err != nil {
		return -1
	}
	if err := lr.Read(&flag); err != nil {
		return -1
	}

	return rec.AddRowValue(row, id, flag)
}

//增加一行,row=-1时,在表格最后面插入一行,否则在row处插入,返回-1插入失败,否则返回插入的行号
func (rec *GlobalDataTestRec) AddRow(row int) int {

	add := -1

	if len(rec.Rows) >= rec.MaxRows {
		return add
	}

	r := GlobalDataTestRecRow{}

	return rec.AddRowValue(row, r.ID, r.Flag)

}

//增加一行,row=-1时,在表格最后面插入一行,否则在row处插入,返回-1插入失败,否则返回插入的行号
func (rec *GlobalDataTestRec) AddRowValue(row int, id string, flag int8) int {

	add := -1

	if len(rec.Rows) >= rec.MaxRows {
		return add
	}

	r := GlobalDataTestRecRow{id, flag}

	if row == -1 {
		rec.Rows = append(rec.Rows, r)
		add = len(rec.Rows) - 1
	} else {
		if row >= 0 && row < len(rec.Rows) {
			rec.Rows = append(rec.Rows, GlobalDataTestRecRow{})
			copy(rec.Rows[row+1:], rec.Rows[row:])
			rec.Rows[row] = r
			add = row
		} else {
			rec.Rows = append(rec.Rows, r)
			add = len(rec.Rows) - 1
		}

	}
	if add != -1 {
		if rec.monitor != nil {
			rec.monitor.RecAppend(rec.owner, rec, add)
		}
		rec.Dirty = true
	}
	return add
}

//获取一行数据
func (rec *GlobalDataTestRec) GetRow(row int) (id string, flag int8, err error) {

	if row < 0 || row >= len(rec.Rows) {
		err = ErrRowError
		return
	}

	r := rec.Rows[row]
	id = r.ID
	flag = r.Flag

	return
}

//获取一行数据
func (rec *GlobalDataTestRec) FindRowInterface(row int) (rowvalue interface{}, err error) {

	if row < 0 || row >= len(rec.Rows) {
		err = ErrRowError
		return
	}

	rowvalue = rec.Rows[row]
	return
}

//获取数据
func (rec *GlobalDataTestRec) Scan(row int, id *string, flag *int8) bool {

	if row < 0 || row >= len(rec.Rows) {
		return false
	}

	r := rec.Rows[row]
	*id = r.ID
	*flag = r.Flag

	return true
}

//删除一行
func (rec *GlobalDataTestRec) Del(row int) {

	if row < 0 || row >= len(rec.Rows) {
		return
	}

	copy(rec.Rows[row:], rec.Rows[row+1:])
	rec.Rows = rec.Rows[:len(rec.Rows)-1]
	rec.Dirty = true

	if rec.monitor != nil {
		rec.monitor.RecDelete(rec.owner, rec, row)
	}
}

//清空表格
func (rec *GlobalDataTestRec) Clear() {

	rec.Rows = rec.Rows[:0]
	rec.Dirty = true
	if rec.monitor != nil {
		rec.monitor.RecClear(rec.owner, rec)
	}
}

//是否保存
func (rec *GlobalDataTestRec) IsSave() bool {
	return true
}

//初始化GlobalDataTestRec表
func (obj *GlobalData) initGlobalDataTestRec() {
	obj.TestRec_r.MaxRows = 1024
	obj.TestRec_r.Cols = 2
	obj.TestRec_r.Rows = make([]GlobalDataTestRecRow, 0, 1024)
	obj.TestRec_r.owner = obj
}

//获取GlobalDataTestRec表
func (obj *GlobalData) GetGlobalDataTestRec() *GlobalDataTestRec {
	return &obj.TestRec_r
}

//初始化所有的表格
func (obj *GlobalData) initRec() {

	obj.initGlobalDataTestRec()
}

//获取某个表格
func (obj *GlobalData) FindRec(rec string) Record {
	switch rec {
	case "TestRec":
		return &obj.TestRec_r
	default:
		return nil
	}
}

//获取所有表格名称
func (obj *GlobalData) RecordNames() []string {
	return []string{"TestRec"}
}

func (obj *GlobalData) GlobalDataInit() {
	obj.quiting = false
	obj.Save = true
	obj.ObjectType = HELPER
	obj.InBase = false
	obj.InScene = false
	obj.uid = 0
}

//重置
func (obj *GlobalData) Reset() {

	//属性初始化
	obj.GlobalData_t = GlobalData_t{}
	obj.GlobalData_Save.GlobalData_Save_Property = GlobalData_Save_Property{}
	obj.GlobalData_Propertys = GlobalData_Propertys{}
	obj.GlobalDataInit()
	//表格初始化
	obj.TestRec_r.Clear()
	obj.TestRec_r.monitor = nil

	obj.ClearDirty()
	obj.ClearModify()
	obj.ClearExtraData()
	for k := range obj.propcritical {
		obj.propcritical[k] = 0
		obj.propflag[k] = 0
	}

}

//对象拷贝
func (obj *GlobalData) Copy(other Entity) error {
	if t, ok := other.(*GlobalData); ok {
		//属性复制
		obj.DbId = t.DbId
		obj.NameHash_ = t.NameHash_
		obj.ConfigIdHash = t.ConfigIdHash
		obj.uid = t.uid

		obj.GlobalData_t = t.GlobalData_t
		obj.GlobalData_Save.GlobalData_Save_Property = t.GlobalData_Save_Property
		obj.GlobalData_Propertys = t.GlobalData_Propertys

		var l int
		//表格复制
		obj.TestRec_r.Clear()
		l = len(t.TestRec_r.Rows)
		for i := 0; i < l; i++ {
			obj.TestRec_r.AddRowValue(-1, t.TestRec_r.Rows[i].ID, t.TestRec_r.Rows[i].Flag)
		}

		return nil
	}

	return ErrCopyObjError
}

//DB相关
//同步到数据库
func (obj *GlobalData) SyncToDb() {

}

//从data中取出configid
func (obj *GlobalData) GetConfigFromDb(data interface{}) string {
	if v, ok := data.(*GlobalData_Save); ok {
		return v.ConfigId
	}

	return ""
}

//从数据库恢复
func (obj *GlobalData) SyncFromDb(data interface{}) bool {
	if v, ok := data.(*GlobalData_Save); ok {
		obj.GlobalData_Save.GlobalData_Save_Property = v.GlobalData_Save_Property

		obj.TestRec_r.Clear()
		if l := len(v.TestRec_r.Rows); l > 0 {
			for i := 0; i < l; i++ {
				obj.TestRec_r.AddRowValue(-1, v.TestRec_r.Rows[i].ID, v.TestRec_r.Rows[i].Flag)
			}
		}

		obj.NameHash_ = Hash(obj.Name)
		obj.ConfigIdHash = Hash(obj.ConfigId)
		return true
	}

	return false
}

func (obj *GlobalData) SaveLoader() DBSaveLoader {
	return &obj.GlobalData_Save
}

func (obj *GlobalData) GobEncode() ([]byte, error) {
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

	err = encoder.Encode(obj.GlobalData_Save)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.GlobalData_Propertys)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (obj *GlobalData) GobDecode(buf []byte) error {
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

	err = decoder.Decode(&obj.GlobalData_Save)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.GlobalData_Propertys)
	if err != nil {
		return err
	}

	return nil
}

//由子类调用的初始化函数
func (obj *GlobalData) baseInit(dirty, modify, extra map[string]interface{}) {
	//初始化表格
	obj.initRec()
	obj.Mdirty = dirty
	obj.Mmodify = modify
	obj.ExtraData = extra
}

func (obj *GlobalData) Serial() ([]byte, error) {
	ar := util.NewStoreArchiver(nil)
	ps := obj.VisiblePropertys(0)
	ar.Write(int16(len(ps)))

	ar.Write(int16(0))
	ar.Write(obj.Name)
	ar.Write(int16(1))
	ar.Write(obj.Test1)
	ar.Write(int16(2))
	ar.Write(obj.Test2)
	return ar.Data(), nil
}

func (obj *GlobalData) SerialModify() ([]byte, error) {
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

func (obj *GlobalData) IsSceneData(prop string) bool {
	idx, err := obj.PropertyIndex(prop)
	if err != nil {
		return false
	}

	return IsGlobalDataSceneData(idx)
}

//通过scenedata同步
func (obj *GlobalData) SyncFromSceneData(val interface{}) error {
	var sd *GlobalDataSceneData
	var ok bool
	if sd, ok = val.(*GlobalDataSceneData); !ok {
		return fmt.Errorf("type not GlobalDataSceneData", sd)
	}

	if sd.TestRec_r.Dirty {
		obj.TestRec_r.Rows = obj.TestRec_r.Rows[:0]
		obj.TestRec_r.Rows = append(obj.TestRec_r.Rows, sd.TestRec_r.Rows...) //Test任务表
		obj.TestRec_r.Dirty = true
	}

	return nil
}

func (obj *GlobalData) SceneData() interface{} {
	sd := &GlobalDataSceneData{}

	//属性

	//表格
	sd.TestRec_r = obj.TestRec_r
	return sd
}

//创建函数
func CreateGlobalData() *GlobalData {
	obj := &GlobalData{}

	obj.ObjectType = HELPER
	obj.initRec()

	obj.propcritical = make([]uint64, int(math.Ceil(float64(3)/64)))
	obj.propflag = make([]uint64, int(math.Ceil(float64(3)/64)))
	obj.GlobalDataInit()

	obj.Mdirty = make(map[string]interface{}, 32)
	obj.Mmodify = make(map[string]interface{}, 32)
	obj.ExtraData = make(map[string]interface{}, 16)

	return obj
}

type GlobalDataSceneData struct {
	TestRec_r GlobalDataTestRec //Test任务表
}

func IsGlobalDataSceneData(idx int) bool {
	switch idx {
	case 0: //名称
		return false
	case 1: //测试数据1
		return false
	case 2: //测试数据2
		return false
	}
	return false
}

func GlobalDataInit() {
	gob.Register(&GlobalData_Save{})
	gob.Register(&GlobalData{})
	gob.Register(&GlobalDataSceneData{})
}
