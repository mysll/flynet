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

type Container_Save_Property struct {
	Capacity      int32  `json:"C"` //容量
	ConfigId      string `json:"I"`
	Name          string `json:"1"` //名称
	ContainerType int32  `json:"2"` //类型
}

//保存到DB的数据
type Container_Save struct {
	Container_Save_Property
}

func (s *Container_Save) Base() string {

	return ""

}

func (s *Container_Save) Marshal() (map[string]interface{}, error) {
	data, err := json.Marshal(*s)
	if err != nil {
		return nil, err
	}

	var ret map[string]interface{}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (s *Container_Save) Unmarshal(data map[string]interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(d, s)

	return err
}

func (s *Container_Save) InsertOrUpdate(eq ExecQueryer, insert bool, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) error {
	var sql string
	var args []interface{}
	if insert {
		sql = "INSERT INTO `tbl_container`(`id`,`capacity`,`configid`,`p_name`,`p_containertype`%s ) VALUES(?,?,?,?,?%s) "
		args = []interface{}{dbId, s.Capacity, s.ConfigId, s.Name, s.ContainerType}
		sql = fmt.Sprintf(sql, extfields, extplacehold)
		if extobjs != nil {
			args = append(args, extobjs...)
		}
	} else {
		sql = "UPDATE `tbl_container` SET %s`capacity`=?, `configid`=?,`p_name`=?,`p_containertype`=? WHERE `id` = ?"
		if extobjs != nil {
			args = append(args, extobjs...)
		}
		args = append(args, []interface{}{s.Capacity, s.ConfigId, s.Name, s.ContainerType, dbId}...)
		sql = fmt.Sprintf(sql, extfields)

	}

	if _, err := eq.Exec(sql, args...); err != nil {
		log.LogError("InsertOrUpdate error:", sql, args)
		return err
	}

	return nil
}

func (s *Container_Save) Update(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, false, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	return
}

func (s *Container_Save) Insert(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, true, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	return
}

func (s *Container_Save) Query(dbId uint64) (sql string, args []interface{}) {
	sql = "SELECT `id`,`capacity`,`configid`,`p_name`,`p_containertype` %s FROM `tbl_container` WHERE `id`=? LIMIT 1"
	args = []interface{}{dbId}
	return
}

func (s *Container_Save) Load(eq ExecQueryer, dbId uint64, extfield string, extobjs ...interface{}) error {
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
	args := []interface{}{&dbId, &s.Capacity, &s.ConfigId, &s.Name, &s.ContainerType}
	if extobjs != nil {
		args = append(args, extobjs...)
	}
	if err = r.Scan(args...); err != nil {
		log.LogError("load error:", err)
		return err
	}

	return nil
}

type Container_Propertys struct {
	//属性定义

}

type Container_t struct {
	dirty           bool
	ObjectType      int
	DbId            uint64
	parent          Entity
	ObjId           ObjectID
	Deleted         bool
	NameHash        int32
	IDHash          int32
	ContainerInited bool
	Index           int //在容器中的位置
	Childs          []Entity
	ChildNum        int

	Save bool //是否保存
}

type Container struct {
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

	Container_t
	Container_Save
	Container_Propertys
}

func (obj *Container) SetUID(v uint64) {
	obj.uid = v
}

func (obj *Container) UID() uint64 {
	return obj.uid
}

func (obj *Container) SetInBase(v bool) {
	obj.InBase = v
}

func (obj *Container) IsInBase() bool {
	return obj.InBase
}

func (obj *Container) SetInScene(v bool) {
	obj.InScene = v
}

func (obj *Container) IsInScene() bool {
	return obj.InScene
}

func (obj *Container) SetLoading(loading bool) {
	obj.loading = loading
}

func (obj *Container) IsLoading() bool {
	return obj.loading
}

func (obj *Container) SetQuiting() {
	obj.quiting = true
}

func (obj *Container) IsQuiting() bool {
	return obj.quiting
}

func (obj *Container) GetConfig() string {
	return obj.ConfigId
}

func (obj *Container) SetConfig(config string) {
	obj.ConfigId = config
	obj.IDHash = Hash(config)
}

func (obj *Container) SetSaveFlag() {
	root := obj.GetRoot()
	if root != nil {
		root.SetSaveFlag()
	} else {
		obj.dirty = true
	}
}

func (obj *Container) ClearSaveFlag() {
	obj.dirty = false

}

func (obj *Container) NeedSave() bool {
	if obj.dirty {
		return true
	}
	return false
}

func (obj *Container) ChangeCapacity(capacity int32) error {

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
func (obj *Container) SetCapacity(capacity int32, initcap int32) {

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

func (obj *Container) GetCapacity() int32 {
	return obj.Capacity
}

//获取实际的容量
func (obj *Container) GetRealCap() int32 {
	if !obj.ContainerInited {
		return 0
	}
	return int32(len(obj.Childs))
}

func (obj *Container) GetRoot() Entity {
	var ent Entity
	if obj.GetParent() == nil {
		return nil
	}
	ent = obj
	for {
		if ent.GetParent() == nil {
			break
		}
		if ent.GetParent().ObjType() == SCENE {
			break
		}
		ent = ent.GetParent()
	}
	return ent
}

//获取数据库id
func (obj *Container) GetDbId() uint64 {
	return obj.DbId
}

func (obj *Container) SetDbId(id uint64) {
	obj.DbId = id
}

func (obj *Container) SetParent(p Entity) {
	obj.parent = p
}

func (obj *Container) GetParent() Entity {
	return obj.parent
}

func (obj *Container) SetDeleted(d bool) {
	obj.Deleted = d
}

func (obj *Container) GetDeleted() bool {
	return obj.Deleted
}

func (obj *Container) SetObjId(id ObjectID) {
	obj.ObjId = id
}

func (obj *Container) GetObjId() ObjectID {
	return obj.ObjId
}

//设置名字Hash
func (obj *Container) SetNameHash(v int32) {
	obj.NameHash = v
}

//获取名字Hash
func (obj *Container) GetNameHash() int32 {
	return obj.NameHash
}

//名字比较
func (obj *Container) NameEqual(name string) bool {
	return obj.Name == name
}

//获取IDHash
func (obj *Container) GetIDHash() int32 {
	return obj.IDHash
}

//ID比较
func (obj *Container) IDEqual(id string) bool {
	return obj.ConfigId == id
}

func (obj *Container) ChildCount() int {
	return obj.ChildNum
}

//移除对象
func (obj *Container) RemoveChild(de Entity) error {
	idx := de.GetIndex()
	e := obj.GetChild(idx)

	if e != nil && e.GetObjId().Equal(de.GetObjId()) {
		obj.Childs[idx] = nil
		de.SetParent(nil)
		obj.ChildNum--
		return nil
	}

	return ErrChildObjectNotFound
}

//获取子对象
func (obj *Container) GetChilds() []Entity {
	return obj.Childs
}

//获取容器中索引
func (obj *Container) GetIndex() int {
	return obj.Index
}

//设置索引，逻辑层不要调用
func (obj *Container) SetIndex(idx int) {
	obj.Index = idx
}

//删除所有的子对象
func (obj *Container) ClearChilds() {
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
func (obj *Container) AddChild(idx int, e Entity) (index int, err error) {
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
				index = e.GetIndex()
				return
			}
		}
		obj.Childs = append(obj.Childs, e)
		e.SetIndex(len(obj.Childs) - 1)
		e.SetParent(obj)
		index = e.GetIndex()
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
	index = e.GetIndex()
	return

}

//获取子对象
func (obj *Container) GetChild(idx int) Entity {
	if !obj.ContainerInited {
		return nil
	}
	if idx < 0 || idx >= len(obj.Childs) {
		return nil
	}
	return obj.Childs[idx]
}

//通过ID获取子对象
func (obj *Container) GetChildByConfigId(id string) Entity {
	if !obj.ContainerInited {
		return nil
	}
	h := Hash(id)
	for _, v := range obj.Childs {
		if (v != nil) && (v.GetIDHash() == h) && v.IDEqual(id) {
			return v
		}
	}
	return nil
}
func (obj *Container) GetFirstChildByConfigId(id string) (int, Entity) {
	if !obj.ContainerInited {
		return -1, nil
	}
	h := Hash(id)
	for k, v := range obj.Childs {
		if (v != nil) && (v.GetIDHash() == h) && v.IDEqual(id) {
			return k + 1, v
		}
	}
	return -1, nil
}
func (obj *Container) GetNextChildByConfigId(start int, id string) (int, Entity) {

	if !obj.ContainerInited || start == -1 || start >= len(obj.Childs) {
		return -1, nil
	}
	h := Hash(id)
	for k, v := range obj.Childs[start:] {
		if (v != nil) && (v.GetIDHash() == h) && v.IDEqual(id) {
			return start + k + 1, v
		}
	}
	return -1, nil
}

//通过名称获取子对象
func (obj *Container) GetChildByName(name string) Entity {
	if !obj.ContainerInited {
		return nil
	}
	h := Hash(name)
	for _, v := range obj.Childs {
		if (v != nil) && (v.GetNameHash() == h) && v.NameEqual(name) {
			return v
		}
	}
	return nil
}
func (obj *Container) GetFirstChild(name string) (int, Entity) {
	if !obj.ContainerInited {
		return -1, nil
	}
	h := Hash(name)
	for k, v := range obj.Childs {
		if (v != nil) && (v.GetNameHash() == h) && v.NameEqual(name) {
			return k + 1, v
		}
	}
	return -1, nil
}
func (obj *Container) GetNextChild(start int, name string) (int, Entity) {

	if !obj.ContainerInited || start == -1 || start >= len(obj.Childs) {
		return -1, nil
	}
	h := Hash(name)
	for k, v := range obj.Childs[start:] {
		if (v != nil) && (v.GetNameHash() == h) && v.NameEqual(name) {
			return start + k + 1, v
		}
	}
	return -1, nil
}

//交换子对象的位置
func (obj *Container) SwapChild(src int, dest int) error {
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
func (obj *Container) Base() Entity {
	return nil
}

//获取对象类型
func (obj *Container) ObjType() int {
	return obj.ObjectType
}

//额外的数据
func (obj *Container) SetExtraData(key string, value interface{}) {
	obj.ExtraData[key] = value
}

func (obj *Container) GetExtraData(key string) interface{} {
	if v, ok := obj.ExtraData[key]; ok {
		return v
	}
	return nil
}

func (obj *Container) GetAllExtraData() map[string]interface{} {
	return obj.ExtraData
}

func (obj *Container) RemoveExtraData(key string) {
	if _, ok := obj.ExtraData[key]; ok {
		delete(obj.ExtraData, key)
	}
}

func (obj *Container) ClearExtraData() {
	for k := range obj.ExtraData {
		delete(obj.ExtraData, k)
	}
}

//获取对象是否保存
func (obj *Container) IsSave() bool {
	return obj.Save
}

//设置对象是否保存
func (obj *Container) SetSave(s bool) {
	obj.Save = s
}

//获取对象类型名
func (obj *Container) ObjTypeName() string {
	return "Container"
}

func (obj *Container) SetPropUpdate(sync PropUpdater) {
	obj.propupdate = sync
}

func (obj *Container) PropUpdate() PropUpdater {
	return obj.propupdate
}

//属性回调接口
func (obj *Container) SetPropHook(hooker PropChanger) {
	obj.prophooker = hooker
}

func (obj *Container) GetPropFlag(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propflag[index]&(uint64(1)<<bit) != 0
}

func (obj *Container) SetPropFlag(idx int, flag bool) {
	index := idx / 64
	bit := uint(idx) % 64
	if flag {
		obj.propflag[index] = obj.propflag[index] | (uint64(1) << bit)
		return
	}
	obj.propflag[index] = obj.propflag[index] & ^(uint64(1) << bit)
}

func (obj *Container) IsCritical(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propcritical[index]&(uint64(1)<<bit) != 0
}

func (obj *Container) SetCritical(prop string) {
	idx, err := obj.GetPropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] | (uint64(1) << bit)
}

func (obj *Container) ClearCritical(prop string) {
	idx, err := obj.GetPropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] & ^(uint64(1) << bit)
}

//获取所有属性
func (obj *Container) GetPropertys() []string {
	return []string{
		"Name",
		"ContainerType",
	}
}

//获取所有可视属性
func (obj *Container) GetVisiblePropertys(typ int) []string {
	if typ == 0 {
		return []string{}
	} else {
		return []string{}
	}

}

//获取属性类型
func (obj *Container) GetPropertyType(p string) (int, string, error) {
	switch p {
	case "Name":
		return DT_STRING, "string", nil
	case "ContainerType":
		return DT_INT32, "int32", nil
	default:
		return DT_NONE, "", ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *Container) GetPropertyIndex(p string) (int, error) {
	switch p {
	case "Name":
		return 0, nil
	case "ContainerType":
		return 1, nil
	default:
		return -1, ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *Container) Inc(p string, v interface{}) error {
	switch p {
	case "ContainerType":
		var dst int32
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.IncContainerType(dst)
		}
		return err
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性名设置值
func (obj *Container) Set(p string, v interface{}) error {
	switch p {
	case "Name":
		val, ok := v.(string)
		if ok {
			obj.SetName(val)
		} else {
			return ErrTypeMismatch
		}
	case "ContainerType":
		var dst int32
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.SetContainerType(dst)
		}
		return err
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性索引设置值
func (obj *Container) SetByIndex(index int16, v interface{}) error {
	switch index {
	case 0:
		val, ok := v.(string)
		if ok {
			obj.SetName(val)
		} else {
			return ErrTypeMismatch
		}
	case 1:
		var dst int32
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.SetContainerType(dst)
		}
		return err
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性名获取值
func (obj *Container) MustGet(p string) interface{} {
	switch p {
	case "Name":
		return obj.Name
	case "ContainerType":
		return obj.ContainerType
	default:
		return nil
	}
}

//通过属性名获取值
func (obj *Container) Get(p string) (val interface{}, err error) {
	switch p {
	case "Name":
		return obj.Name, nil
	case "ContainerType":
		return obj.ContainerType, nil
	default:
		return nil, ErrPropertyNotFound
	}
}

//是否需要同步到其它客户端
func (obj *Container) PropertyIsPublic(p string) bool {
	switch p {
	case "Name":
		return false
	case "ContainerType":
		return false
	default:
		return false
	}
}

//是否需要同步到自己的客户端
func (obj *Container) PropertyIsPrivate(p string) bool {
	switch p {
	case "Name":
		return false
	case "ContainerType":
		return false
	default:
		return false
	}
}

//是否需要存档
func (obj *Container) PropertyIsSave(p string) bool {
	switch p {
	case "Name":
		return true
	case "ContainerType":
		return true
	default:
		return false
	}
}

//脏标志(数据保存用)
func (obj *Container) setDirty(p string, v interface{}) {
	//obj.Mdirty[p] = v
	obj.SetSaveFlag()
}

func (obj *Container) GetDirty() map[string]interface{} {

	return obj.Mdirty
}

func (obj *Container) ClearDirty() {
	for k := range obj.Mdirty {
		delete(obj.Mdirty, k)
	}
}

//修改标志(数据同步用)
func (obj *Container) setModify(p string, v interface{}) {
	obj.Mmodify[p] = v
}

func (obj *Container) GetModify() map[string]interface{} {
	return obj.Mmodify
}

func (obj *Container) ClearModify() {
	for k := range obj.Mmodify {
		delete(obj.Mmodify, k)
	}
}

//名称
func (obj *Container) SetName(v string) {
	if obj.Name == v {
		return
	}

	old := obj.Name

	if !obj.InBase { //只有base能够修改自身的数据
		log.LogError("can't change base data")
	}

	obj.Name = v
	if obj.prophooker != nil && obj.IsCritical(0) && !obj.GetPropFlag(0) {
		obj.SetPropFlag(0, true)
		obj.prophooker.OnPropChange(obj, "Name", old)
		obj.SetPropFlag(0, false)
	}
	obj.NameHash = Hash(v)

	obj.setDirty("Name", v)
}
func (obj *Container) GetName() string {
	return obj.Name
}

//类型
func (obj *Container) SetContainerType(v int32) {
	if obj.ContainerType == v {
		return
	}

	old := obj.ContainerType

	if !obj.InBase { //只有base能够修改自身的数据
		log.LogError("can't change base data")
	}

	obj.ContainerType = v
	if obj.prophooker != nil && obj.IsCritical(1) && !obj.GetPropFlag(1) {
		obj.SetPropFlag(1, true)
		obj.prophooker.OnPropChange(obj, "ContainerType", old)
		obj.SetPropFlag(1, false)
	}

	obj.setDirty("ContainerType", v)
}
func (obj *Container) GetContainerType() int32 {
	return obj.ContainerType
}
func (obj *Container) IncContainerType(v int32) {
	obj.SetContainerType(obj.ContainerType + v)
}

//初始化所有的表格
func (obj *Container) initRec() {

}

//获取某个表格
func (obj *Container) GetRec(rec string) Record {
	switch rec {
	default:
		return nil
	}
}

//获取所有表格名称
func (obj *Container) GetRecNames() []string {
	return []string{}
}

func (obj *Container) ContainerInit() {
	obj.quiting = false
	obj.Save = true
	obj.ObjectType = ITEM
	obj.InBase = false
	obj.InScene = false
	obj.uid = 0
}

//重置
func (obj *Container) Reset() {

	//属性初始化
	obj.Container_t = Container_t{}
	obj.Container_Save.Container_Save_Property = Container_Save_Property{}
	obj.Container_Propertys = Container_Propertys{}
	obj.ContainerInit()
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
func (obj *Container) Copy(other Entity) error {
	if t, ok := other.(*Container); ok {
		//属性复制
		obj.DbId = t.DbId
		obj.NameHash = t.NameHash
		obj.IDHash = t.IDHash
		obj.uid = t.uid

		obj.Container_t = t.Container_t
		obj.Container_Save.Container_Save_Property = t.Container_Save_Property
		obj.Container_Propertys = t.Container_Propertys

		//表格复制

		return nil
	}

	return ErrCopyObjError
}

//DB相关
//同步到数据库
func (obj *Container) SyncToDb() {

}

//从data中取出configid
func (obj *Container) GetConfigFromDb(data interface{}) string {
	if v, ok := data.(*Container_Save); ok {
		return v.ConfigId
	}

	return ""
}

//从数据库恢复
func (obj *Container) SyncFromDb(data interface{}) bool {
	if v, ok := data.(*Container_Save); ok {
		obj.Container_Save.Container_Save_Property = v.Container_Save_Property

		obj.NameHash = Hash(obj.Name)
		obj.IDHash = Hash(obj.ConfigId)
		return true
	}

	return false
}

func (obj *Container) GetSaveLoader() DBSaveLoader {
	return &obj.Container_Save
}

func (obj *Container) GobEncode() ([]byte, error) {
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
	err = encoder.Encode(obj.NameHash)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.IDHash)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.Index)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(obj.Container_Save)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.Container_Propertys)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (obj *Container) GobDecode(buf []byte) error {
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
	err = decoder.Decode(&obj.NameHash)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.IDHash)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.Index)
	if err != nil {
		return err
	}

	err = decoder.Decode(&obj.Container_Save)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.Container_Propertys)
	if err != nil {
		return err
	}

	return nil
}

//由子类调用的初始化函数
func (obj *Container) baseInit(dirty, modify, extra map[string]interface{}) {
	//初始化表格
	obj.initRec()
	obj.Mdirty = dirty
	obj.Mmodify = modify
	obj.ExtraData = extra
}

func (obj *Container) Serial() ([]byte, error) {
	ar := util.NewStoreArchiver(nil)
	ps := obj.GetVisiblePropertys(0)
	ar.Write(int16(len(ps)))

	return ar.Data(), nil
}

func (obj *Container) SerialModify() ([]byte, error) {
	if len(obj.Mmodify) == 0 {
		return nil, nil
	}
	ar := util.NewStoreArchiver(nil)
	ar.Write(int16(len(obj.Mmodify)))
	for k, v := range obj.Mmodify {
		if !obj.PropertyIsPrivate(k) {
			continue
		}
		idx, _ := obj.GetPropertyIndex(k)

		ar.Write(int16(idx))
		ar.Write(v)
	}

	return ar.Data(), nil
}

func (obj *Container) IsSceneData(prop string) bool {
	idx, err := obj.GetPropertyIndex(prop)
	if err != nil {
		return false
	}

	return IsContainerSceneData(idx)
}

//通过scenedata同步
func (obj *Container) SyncFromSceneData(val interface{}) error {
	var sd *ContainerSceneData
	var ok bool
	if sd, ok = val.(*ContainerSceneData); !ok {
		return fmt.Errorf("type not ContainerSceneData", sd)
	}

	return nil
}

func (obj *Container) GetSceneData() interface{} {
	sd := &ContainerSceneData{}

	//属性

	//表格
	return sd
}

//创建函数
func CreateContainer() *Container {
	obj := &Container{}

	obj.ObjectType = ITEM
	obj.initRec()

	obj.propcritical = make([]uint64, int(math.Ceil(float64(2)/64)))
	obj.propflag = make([]uint64, int(math.Ceil(float64(2)/64)))
	obj.ContainerInit()

	obj.Mdirty = make(map[string]interface{}, 32)
	obj.Mmodify = make(map[string]interface{}, 32)
	obj.ExtraData = make(map[string]interface{}, 16)

	return obj
}

type ContainerSceneData struct {
}

func IsContainerSceneData(idx int) bool {
	switch idx {
	case 0: //名称
		return false
	case 1: //类型
		return false
	}
	return false
}

func ContainerInit() {
	gob.Register(&Container_Save{})
	gob.Register(&Container{})
	gob.Register(&ContainerSceneData{})
}
