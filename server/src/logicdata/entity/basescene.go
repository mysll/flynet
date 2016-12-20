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

type BaseScene_Save_Property struct {
	Capacity int32  `json:"C"` //容量
	ConfigId string `json:"I"`
	Name     string //名称
}

//保存到DB的数据
type BaseScene_Save struct {
	BaseScene_Save_Property
}

func (s *BaseScene_Save) Base() string {

	return ""

}

func (s *BaseScene_Save) Marshal() (map[string]interface{}, error) {
	data, err := json.Marshal(*s)
	if err != nil {
		return nil, err
	}

	var ret map[string]interface{}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (s *BaseScene_Save) Unmarshal(data map[string]interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(d, s)

	return err
}

func (s *BaseScene_Save) InsertOrUpdate(eq ExecQueryer, insert bool, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) error {
	var sql string
	var args []interface{}
	if insert {
		sql = "INSERT INTO `tbl_basescene`(`id`,`capacity`,`configid`,`p_name`%s ) VALUES(?,?,?,?%s) "
		args = []interface{}{dbId, s.Capacity, s.ConfigId, s.Name}
		sql = fmt.Sprintf(sql, extfields, extplacehold)
		if extobjs != nil {
			args = append(args, extobjs...)
		}
	} else {
		sql = "UPDATE `tbl_basescene` SET %s`capacity`=?, `configid`=?,`p_name`=? WHERE `id` = ?"
		if extobjs != nil {
			args = append(args, extobjs...)
		}
		args = append(args, []interface{}{s.Capacity, s.ConfigId, s.Name, dbId}...)
		sql = fmt.Sprintf(sql, extfields)

	}

	if _, err := eq.Exec(sql, args...); err != nil {
		log.LogError("InsertOrUpdate error:", sql, args)
		return err
	}

	return nil
}

func (s *BaseScene_Save) Update(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, false, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	return
}

func (s *BaseScene_Save) Insert(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, true, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	return
}

func (s *BaseScene_Save) Query(dbId uint64) (sql string, args []interface{}) {
	sql = "SELECT `id`,`capacity`,`configid`,`p_name` %s FROM `tbl_basescene` WHERE `id`=? LIMIT 1"
	args = []interface{}{dbId}
	return
}

func (s *BaseScene_Save) Load(eq ExecQueryer, dbId uint64, extfield string, extobjs ...interface{}) error {
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
	args := []interface{}{&dbId, &s.Capacity, &s.ConfigId, &s.Name}
	if extobjs != nil {
		args = append(args, extobjs...)
	}
	if err = r.Scan(args...); err != nil {
		log.LogError("load error:", err)
		return err
	}

	return nil
}

type BaseScene_Propertys struct {
	//属性定义

}

type BaseScene_t struct {
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

type BaseScene struct {
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

	BaseScene_t
	BaseScene_Save
	BaseScene_Propertys
}

func (obj *BaseScene) SetUID(v uint64) {
	obj.uid = v
}

func (obj *BaseScene) UID() uint64 {
	return obj.uid
}

func (obj *BaseScene) SetInBase(v bool) {
	obj.InBase = v
}

func (obj *BaseScene) IsInBase() bool {
	return obj.InBase
}

func (obj *BaseScene) SetInScene(v bool) {
	obj.InScene = v
}

func (obj *BaseScene) IsInScene() bool {
	return obj.InScene
}

func (obj *BaseScene) SetLoading(loading bool) {
	obj.loading = loading
}

func (obj *BaseScene) IsLoading() bool {
	return obj.loading
}

func (obj *BaseScene) SetQuiting() {
	obj.quiting = true
}

func (obj *BaseScene) IsQuiting() bool {
	return obj.quiting
}

func (obj *BaseScene) Config() string {
	return obj.ConfigId
}

func (obj *BaseScene) SetConfig(config string) {
	obj.ConfigId = config
	obj.ConfigIdHash = Hash(config)
}

func (obj *BaseScene) SetSaveFlag() {
	root := obj.Root()
	if root != nil {
		root.SetSaveFlag()
	} else {
		obj.dirty = true
	}
}

func (obj *BaseScene) ClearSaveFlag() {
	obj.dirty = false

}

func (obj *BaseScene) NeedSave() bool {
	if obj.dirty {
		return true
	}
	return false
}

func (obj *BaseScene) ChangeCapacity(capacity int32) error {

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
func (obj *BaseScene) SetCapacity(capacity int32, initcap int32) {

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

func (obj *BaseScene) Caps() int32 {
	return obj.Capacity
}

//获取实际的容量
func (obj *BaseScene) RealCaps() int32 {
	if !obj.ContainerInited {
		return 0
	}
	return int32(len(obj.Childs))
}

func (obj *BaseScene) Root() Entity {
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
func (obj *BaseScene) DBId() uint64 {
	return obj.DbId
}

func (obj *BaseScene) SetDBId(id uint64) {
	obj.DbId = id
}

func (obj *BaseScene) SetParent(p Entity) {
	obj.parent = p
}

func (obj *BaseScene) Parent() Entity {
	return obj.parent
}

func (obj *BaseScene) SetDeleted(d bool) {
	obj.Deleted = d
}

func (obj *BaseScene) IsDeleted() bool {
	return obj.Deleted
}

func (obj *BaseScene) SetObjId(id ObjectID) {
	obj.ObjId = id
}

func (obj *BaseScene) ObjectId() ObjectID {
	return obj.ObjId
}

//设置名字Hash
func (obj *BaseScene) SetNameHash(v int32) {
	obj.NameHash_ = v
}

//获取名字Hash
func (obj *BaseScene) NameHash() int32 {
	return obj.NameHash_
}

//名字比较
func (obj *BaseScene) NameEqual(name string) bool {
	return obj.Name == name
}

//获取ConfigIdHash
func (obj *BaseScene) ConfigHash() int32 {
	return obj.ConfigIdHash
}

//ID比较
func (obj *BaseScene) ConfigIdEqual(id string) bool {
	return obj.ConfigId == id
}

func (obj *BaseScene) ChildCount() int {
	return obj.ChildNum
}

//移除对象
func (obj *BaseScene) RemoveChild(de Entity) error {
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
func (obj *BaseScene) AllChilds() []Entity {
	return obj.Childs
}

//获取容器中索引
func (obj *BaseScene) ChildIndex() int {
	return obj.Index
}

//设置索引，逻辑层不要调用
func (obj *BaseScene) SetIndex(idx int) {
	obj.Index = idx
}

//删除所有的子对象
func (obj *BaseScene) ClearChilds() {
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
func (obj *BaseScene) AddChild(idx int, e Entity) (index int, err error) {
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
func (obj *BaseScene) GetChild(idx int) Entity {
	if !obj.ContainerInited {
		return nil
	}
	if idx < 0 || idx >= len(obj.Childs) {
		return nil
	}
	return obj.Childs[idx]
}

//通过ID获取子对象
func (obj *BaseScene) FindChildByConfigId(id string) Entity {
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
func (obj *BaseScene) FindFirstChildByConfigId(id string) (int, Entity) {
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
func (obj *BaseScene) NextChildByConfigId(start int, id string) (int, Entity) {

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
func (obj *BaseScene) FindChildByName(name string) Entity {
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
func (obj *BaseScene) FindFirstChildByName(name string) (int, Entity) {
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
func (obj *BaseScene) NextChildByName(start int, name string) (int, Entity) {

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
func (obj *BaseScene) SwapChild(src int, dest int) error {
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
func (obj *BaseScene) Base() Entity {
	return nil
}

//获取对象类型
func (obj *BaseScene) ObjType() int {
	return obj.ObjectType
}

//额外的数据
func (obj *BaseScene) SetExtraData(key string, value interface{}) {
	obj.ExtraData[key] = value
}

func (obj *BaseScene) FindExtraData(key string) interface{} {
	if v, ok := obj.ExtraData[key]; ok {
		return v
	}
	return nil
}

func (obj *BaseScene) ExtraDatas() map[string]interface{} {
	return obj.ExtraData
}

func (obj *BaseScene) RemoveExtraData(key string) {
	if _, ok := obj.ExtraData[key]; ok {
		delete(obj.ExtraData, key)
	}
}

func (obj *BaseScene) ClearExtraData() {
	for k := range obj.ExtraData {
		delete(obj.ExtraData, k)
	}
}

//获取对象是否保存
func (obj *BaseScene) IsSave() bool {
	return obj.Save
}

//设置对象是否保存
func (obj *BaseScene) SetSave(s bool) {
	obj.Save = s
}

//获取对象类型名
func (obj *BaseScene) ObjTypeName() string {
	return "BaseScene"
}

func (obj *BaseScene) SetPropUpdate(sync PropUpdater) {
	obj.propupdate = sync
}

func (obj *BaseScene) PropUpdate() PropUpdater {
	return obj.propupdate
}

//属性回调接口
func (obj *BaseScene) SetPropHook(hooker PropChanger) {
	obj.prophooker = hooker
}

func (obj *BaseScene) PropFlag(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propflag[index]&(uint64(1)<<bit) != 0
}

func (obj *BaseScene) SetPropFlag(idx int, flag bool) {
	index := idx / 64
	bit := uint(idx) % 64
	if flag {
		obj.propflag[index] = obj.propflag[index] | (uint64(1) << bit)
		return
	}
	obj.propflag[index] = obj.propflag[index] & ^(uint64(1) << bit)
}

func (obj *BaseScene) IsCritical(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propcritical[index]&(uint64(1)<<bit) != 0
}

func (obj *BaseScene) SetCritical(prop string) {
	idx, err := obj.PropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] | (uint64(1) << bit)
}

func (obj *BaseScene) ClearCritical(prop string) {
	idx, err := obj.PropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] & ^(uint64(1) << bit)
}

//获取所有属性
func (obj *BaseScene) Propertys() []string {
	return []string{
		"Name",
	}
}

//获取所有可视属性
func (obj *BaseScene) VisiblePropertys(typ int) []string {
	if typ == 0 {
		return []string{}
	} else {
		return []string{}
	}

}

//获取属性类型
func (obj *BaseScene) PropertyType(p string) (int, string, error) {
	switch p {
	case "Name":
		return DT_STRING, "string", nil
	default:
		return DT_NONE, "", ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *BaseScene) PropertyIndex(p string) (int, error) {
	switch p {
	case "Name":
		return 0, nil
	default:
		return -1, ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *BaseScene) Inc(p string, v interface{}) error {
	switch p {
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性名设置值
func (obj *BaseScene) Set(p string, v interface{}) error {
	switch p {
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
func (obj *BaseScene) SetByIndex(index int16, v interface{}) error {
	switch index {
	case 0:
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
func (obj *BaseScene) MustGet(p string) interface{} {
	switch p {
	case "Name":
		return obj.Name
	default:
		return nil
	}
}

//通过属性名获取值
func (obj *BaseScene) Get(p string) (val interface{}, err error) {
	switch p {
	case "Name":
		return obj.Name, nil
	default:
		return nil, ErrPropertyNotFound
	}
}

//是否需要同步到其它客户端
func (obj *BaseScene) PropertyIsPublic(p string) bool {
	switch p {
	case "Name":
		return false
	default:
		return false
	}
}

//是否需要同步到自己的客户端
func (obj *BaseScene) PropertyIsPrivate(p string) bool {
	switch p {
	case "Name":
		return false
	default:
		return false
	}
}

//是否需要存档
func (obj *BaseScene) PropertyIsSave(p string) bool {
	switch p {
	case "Name":
		return true
	default:
		return false
	}
}

//脏标志(数据保存用)
func (obj *BaseScene) setDirty(p string, v interface{}) {
	//obj.Mdirty[p] = v
	obj.SetSaveFlag()
}

func (obj *BaseScene) Dirtys() map[string]interface{} {

	return obj.Mdirty
}

func (obj *BaseScene) ClearDirty() {
	for k := range obj.Mdirty {
		delete(obj.Mdirty, k)
	}
}

//修改标志(数据同步用)
func (obj *BaseScene) setModify(p string, v interface{}) {
	obj.Mmodify[p] = v
}

func (obj *BaseScene) Modifys() map[string]interface{} {
	return obj.Mmodify
}

func (obj *BaseScene) ClearModify() {
	for k := range obj.Mmodify {
		delete(obj.Mmodify, k)
	}
}

//名称
func (obj *BaseScene) SetName(v string) {
	if obj.Name == v {
		return
	}

	old := obj.Name

	if !obj.InBase { //只有base能够修改自身的数据
		log.LogError("can't change base data")
	}

	obj.Name = v
	if obj.prophooker != nil && obj.IsCritical(0) && !obj.PropFlag(0) {
		obj.SetPropFlag(0, true)
		obj.prophooker.OnPropChange(obj, "Name", old)
		obj.SetPropFlag(0, false)
	}
	obj.NameHash_ = Hash(v)

	obj.setDirty("Name", v)
}
func (obj *BaseScene) GetName() string {
	return obj.Name
}

//初始化所有的表格
func (obj *BaseScene) initRec() {

}

//获取某个表格
func (obj *BaseScene) FindRec(rec string) Record {
	switch rec {
	default:
		return nil
	}
}

//获取所有表格名称
func (obj *BaseScene) RecordNames() []string {
	return []string{}
}

func (obj *BaseScene) BaseSceneInit() {
	obj.quiting = false
	obj.Save = true
	obj.ObjectType = SCENE
	obj.InBase = false
	obj.InScene = false
	obj.uid = 0
}

//重置
func (obj *BaseScene) Reset() {

	//属性初始化
	obj.BaseScene_t = BaseScene_t{}
	obj.BaseScene_Save.BaseScene_Save_Property = BaseScene_Save_Property{}
	obj.BaseScene_Propertys = BaseScene_Propertys{}
	obj.BaseSceneInit()
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
func (obj *BaseScene) Copy(other Entity) error {
	if t, ok := other.(*BaseScene); ok {
		//属性复制
		obj.DbId = t.DbId
		obj.NameHash_ = t.NameHash_
		obj.ConfigIdHash = t.ConfigIdHash
		obj.uid = t.uid

		obj.BaseScene_t = t.BaseScene_t
		obj.BaseScene_Save.BaseScene_Save_Property = t.BaseScene_Save_Property
		obj.BaseScene_Propertys = t.BaseScene_Propertys

		//表格复制

		return nil
	}

	return ErrCopyObjError
}

//DB相关
//同步到数据库
func (obj *BaseScene) SyncToDb() {

}

//从data中取出configid
func (obj *BaseScene) GetConfigFromDb(data interface{}) string {
	if v, ok := data.(*BaseScene_Save); ok {
		return v.ConfigId
	}

	return ""
}

//从数据库恢复
func (obj *BaseScene) SyncFromDb(data interface{}) bool {
	if v, ok := data.(*BaseScene_Save); ok {
		obj.BaseScene_Save.BaseScene_Save_Property = v.BaseScene_Save_Property

		obj.NameHash_ = Hash(obj.Name)
		obj.ConfigIdHash = Hash(obj.ConfigId)
		return true
	}

	return false
}

func (obj *BaseScene) SaveLoader() DBSaveLoader {
	return &obj.BaseScene_Save
}

func (obj *BaseScene) GobEncode() ([]byte, error) {
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

	err = encoder.Encode(obj.BaseScene_Save)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.BaseScene_Propertys)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (obj *BaseScene) GobDecode(buf []byte) error {
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

	err = decoder.Decode(&obj.BaseScene_Save)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.BaseScene_Propertys)
	if err != nil {
		return err
	}

	return nil
}

//由子类调用的初始化函数
func (obj *BaseScene) baseInit(dirty, modify, extra map[string]interface{}) {
	//初始化表格
	obj.initRec()
	obj.Mdirty = dirty
	obj.Mmodify = modify
	obj.ExtraData = extra
}

func (obj *BaseScene) Serial() ([]byte, error) {
	ar := util.NewStoreArchiver(nil)
	ps := obj.VisiblePropertys(0)
	ar.Write(int16(len(ps)))

	return ar.Data(), nil
}

func (obj *BaseScene) SerialModify() ([]byte, error) {
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

func (obj *BaseScene) IsSceneData(prop string) bool {
	idx, err := obj.PropertyIndex(prop)
	if err != nil {
		return false
	}

	return IsBaseSceneSceneData(idx)
}

//通过scenedata同步
func (obj *BaseScene) SyncFromSceneData(val interface{}) error {
	var sd *BaseSceneSceneData
	var ok bool
	if sd, ok = val.(*BaseSceneSceneData); !ok {
		return fmt.Errorf("type not BaseSceneSceneData", sd)
	}

	return nil
}

func (obj *BaseScene) SceneData() interface{} {
	sd := &BaseSceneSceneData{}

	//属性

	//表格
	return sd
}

//创建函数
func CreateBaseScene() *BaseScene {
	obj := &BaseScene{}

	obj.ObjectType = SCENE
	obj.initRec()

	obj.propcritical = make([]uint64, int(math.Ceil(float64(1)/64)))
	obj.propflag = make([]uint64, int(math.Ceil(float64(1)/64)))
	obj.BaseSceneInit()

	obj.Mdirty = make(map[string]interface{}, 32)
	obj.Mmodify = make(map[string]interface{}, 32)
	obj.ExtraData = make(map[string]interface{}, 16)

	return obj
}

type BaseSceneSceneData struct {
}

func IsBaseSceneSceneData(idx int) bool {
	switch idx {
	case 0: //名称
		return false
	}
	return false
}

func BaseSceneInit() {
	gob.Register(&BaseScene_Save{})
	gob.Register(&BaseScene{})
	gob.Register(&BaseSceneSceneData{})
}
