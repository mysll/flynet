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

type GlobalSet_Save_Property struct {
	Capacity int32  `json:"C"` //容量
	ConfigId string `json:"I"`
	Name     string //名称
}

//保存到DB的数据
type GlobalSet_Save struct {
	GlobalSet_Save_Property
}

func (s *GlobalSet_Save) Base() string {

	return ""

}

func (s *GlobalSet_Save) Marshal() (map[string]interface{}, error) {
	data, err := json.Marshal(*s)
	if err != nil {
		return nil, err
	}

	var ret map[string]interface{}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (s *GlobalSet_Save) Unmarshal(data map[string]interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(d, s)

	return err
}

func (s *GlobalSet_Save) InsertOrUpdate(eq ExecQueryer, insert bool, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) error {
	var sql string
	var args []interface{}
	if insert {
		sql = "INSERT INTO `tbl_globalset`(`id`,`capacity`,`configid`,`p_name`%s ) VALUES(?,?,?,?%s) "
		args = []interface{}{dbId, s.Capacity, s.ConfigId, s.Name}
		sql = fmt.Sprintf(sql, extfields, extplacehold)
		if extobjs != nil {
			args = append(args, extobjs...)
		}
	} else {
		sql = "UPDATE `tbl_globalset` SET %s`capacity`=?, `configid`=?,`p_name`=? WHERE `id` = ?"
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

func (s *GlobalSet_Save) Update(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, false, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	return
}

func (s *GlobalSet_Save) Insert(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, true, dbId, extfields, extplacehold, extobjs...); err != nil {
		return
	}

	return
}

func (s *GlobalSet_Save) Query(dbId uint64) (sql string, args []interface{}) {
	sql = "SELECT `id`,`capacity`,`configid`,`p_name` %s FROM `tbl_globalset` WHERE `id`=? LIMIT 1"
	args = []interface{}{dbId}
	return
}

func (s *GlobalSet_Save) Load(eq ExecQueryer, dbId uint64, extfield string, extobjs ...interface{}) error {
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

type GlobalSet_Propertys struct {
	//属性定义

}

type GlobalSet_t struct {
	dirty           bool
	ObjectType      int
	DbId            uint64
	parent          Entityer
	ObjId           ObjectID
	Deleted         bool
	NameHash        int32
	IDHash          int32
	ContainerInited bool
	Index           int //在容器中的位置
	Childs          []Entityer
	ChildNum        int

	Save bool //是否保存
}

type GlobalSet struct {
	InScene      bool //是否在场景中
	InBase       bool //是否在base中
	Mdirty       map[string]interface{}
	Mmodify      map[string]interface{}
	ExtraData    map[string]interface{}
	loading      bool
	quiting      bool
	propsyncer   PropSyncer
	prophooker   PropHooker
	propcritical []uint64
	propflag     []uint64

	GlobalSet_t
	GlobalSet_Save
	GlobalSet_Propertys
}

func (obj *GlobalSet) SetInBase(v bool) {
	obj.InBase = v
}

func (obj *GlobalSet) IsInBase() bool {
	return obj.InBase
}

func (obj *GlobalSet) SetInScene(v bool) {
	obj.InScene = v
}

func (obj *GlobalSet) IsInScene() bool {
	return obj.InScene
}

func (obj *GlobalSet) SetLoading(loading bool) {
	obj.loading = loading
}

func (obj *GlobalSet) IsLoading() bool {
	return obj.loading
}

func (obj *GlobalSet) SetQuiting() {
	obj.quiting = true
}

func (obj *GlobalSet) IsQuiting() bool {
	return obj.quiting
}

func (obj *GlobalSet) GetConfig() string {
	return obj.ConfigId
}

func (obj *GlobalSet) SetConfig(config string) {
	obj.ConfigId = config
	obj.IDHash = Hash(config)
}

func (obj *GlobalSet) SetSaveFlag() {
	root := obj.GetRoot()
	if root != nil {
		root.SetSaveFlag()
	} else {
		obj.dirty = true
	}
}

func (obj *GlobalSet) ClearSaveFlag() {
	obj.dirty = false

}

func (obj *GlobalSet) NeedSave() bool {
	if obj.dirty {
		return true
	}
	return false
}

func (obj *GlobalSet) ChangeCapacity(capacity int32) error {

	if !obj.ContainerInited {
		return ErrContainerNotInit
	}

	if capacity == -1 || capacity == obj.Capacity {
		return ErrContainerCapacity
	}

	if capacity < int32(obj.ChildCount()) {
		return ErrContainerCapacity
	}

	newchilds := make([]Entityer, capacity)
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
func (obj *GlobalSet) SetCapacity(capacity int32, initcap int32) {

	if obj.ContainerInited {
		return
	}

	if capacity == -1 {
		obj.Childs = make([]Entityer, 0, initcap)
	} else if capacity > 0 {
		obj.Childs = make([]Entityer, capacity)
	} else {
		obj.Childs = nil
	}
	obj.Capacity = capacity
	obj.ContainerInited = true

}

func (obj *GlobalSet) GetCapacity() int32 {
	return obj.Capacity
}

//获取实际的容量
func (obj *GlobalSet) GetRealCap() int32 {
	if !obj.ContainerInited {
		return 0
	}
	return int32(len(obj.Childs))
}

func (obj *GlobalSet) GetRoot() Entityer {
	var ent Entityer
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
func (obj *GlobalSet) GetDbId() uint64 {
	return obj.DbId
}

func (obj *GlobalSet) SetDbId(id uint64) {
	obj.DbId = id
}

func (obj *GlobalSet) SetParent(p Entityer) {
	obj.parent = p
}

func (obj *GlobalSet) GetParent() Entityer {
	return obj.parent
}

func (obj *GlobalSet) SetDeleted(d bool) {
	obj.Deleted = d
}

func (obj *GlobalSet) GetDeleted() bool {
	return obj.Deleted
}

func (obj *GlobalSet) SetObjId(id ObjectID) {
	obj.ObjId = id
}

func (obj *GlobalSet) GetObjId() ObjectID {
	return obj.ObjId
}

//设置名字Hash
func (obj *GlobalSet) SetNameHash(v int32) {
	obj.NameHash = v
}

//获取名字Hash
func (obj *GlobalSet) GetNameHash() int32 {
	return obj.NameHash
}

//名字比较
func (obj *GlobalSet) NameEqual(name string) bool {
	return obj.Name == name
}

//获取IDHash
func (obj *GlobalSet) GetIDHash() int32 {
	return obj.IDHash
}

//ID比较
func (obj *GlobalSet) IDEqual(id string) bool {
	return obj.ConfigId == id
}

func (obj *GlobalSet) ChildCount() int {
	return obj.ChildNum
}

//移除对象
func (obj *GlobalSet) RemoveChild(de Entityer) error {
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
func (obj *GlobalSet) GetChilds() []Entityer {
	return obj.Childs
}

//获取容器中索引
func (obj *GlobalSet) GetIndex() int {
	return obj.Index
}

//设置索引，逻辑层不要调用
func (obj *GlobalSet) SetIndex(idx int) {
	obj.Index = idx
}

//删除所有的子对象
func (obj *GlobalSet) ClearChilds() {
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
func (obj *GlobalSet) AddChild(idx int, e Entityer) (index int, err error) {
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
func (obj *GlobalSet) GetChild(idx int) Entityer {
	if !obj.ContainerInited {
		return nil
	}
	if idx < 0 || idx >= len(obj.Childs) {
		return nil
	}
	return obj.Childs[idx]
}

//通过ID获取子对象
func (obj *GlobalSet) GetChildByConfigId(id string) Entityer {
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
func (obj *GlobalSet) GetFirstChildByConfigId(id string) (int, Entityer) {
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
func (obj *GlobalSet) GetNextChildByConfigId(start int, id string) (int, Entityer) {

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
func (obj *GlobalSet) GetChildByName(name string) Entityer {
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
func (obj *GlobalSet) GetFirstChild(name string) (int, Entityer) {
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
func (obj *GlobalSet) GetNextChild(start int, name string) (int, Entityer) {

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
func (obj *GlobalSet) SwapChild(src int, dest int) error {
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
func (obj *GlobalSet) Base() Entityer {
	return nil
}

//获取对象类型
func (obj *GlobalSet) ObjType() int {
	return obj.ObjectType
}

//额外的数据
func (obj *GlobalSet) SetExtraData(key string, value interface{}) {
	obj.ExtraData[key] = value
}

func (obj *GlobalSet) GetExtraData(key string) interface{} {
	if v, ok := obj.ExtraData[key]; ok {
		return v
	}
	return nil
}

func (obj *GlobalSet) GetAllExtraData() map[string]interface{} {
	return obj.ExtraData
}

func (obj *GlobalSet) RemoveExtraData(key string) {
	if _, ok := obj.ExtraData[key]; ok {
		delete(obj.ExtraData, key)
	}
}

func (obj *GlobalSet) ClearExtraData() {
	for k := range obj.ExtraData {
		delete(obj.ExtraData, k)
	}
}

//获取对象是否保存
func (obj *GlobalSet) IsSave() bool {
	return obj.Save
}

//设置对象是否保存
func (obj *GlobalSet) SetSave(s bool) {
	obj.Save = s
}

//获取对象类型名
func (obj *GlobalSet) ObjTypeName() string {
	return "GlobalSet"
}

func (obj *GlobalSet) SetPropSyncer(sync PropSyncer) {
	obj.propsyncer = sync
}

func (obj *GlobalSet) GetPropSyncer() PropSyncer {
	return obj.propsyncer
}

//属性回调接口
func (obj *GlobalSet) SetPropHooker(hooker PropHooker) {
	obj.prophooker = hooker
}

func (obj *GlobalSet) GetPropFlag(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propflag[index]&(uint64(1)<<bit) != 0
}

func (obj *GlobalSet) SetPropFlag(idx int, flag bool) {
	index := idx / 64
	bit := uint(idx) % 64
	if flag {
		obj.propflag[index] = obj.propflag[index] | (uint64(1) << bit)
		return
	}
	obj.propflag[index] = obj.propflag[index] & ^(uint64(1) << bit)
}

func (obj *GlobalSet) IsCritical(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propcritical[index]&(uint64(1)<<bit) != 0
}

func (obj *GlobalSet) SetCritical(prop string) {
	idx, err := obj.GetPropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] | (uint64(1) << bit)
}

func (obj *GlobalSet) ClearCritical(prop string) {
	idx, err := obj.GetPropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] & ^(uint64(1) << bit)
}

//获取所有属性
func (obj *GlobalSet) GetPropertys() []string {
	return []string{
		"Name",
	}
}

//获取所有可视属性
func (obj *GlobalSet) GetVisiblePropertys(typ int) []string {
	if typ == 0 {
		return []string{}
	} else {
		return []string{}
	}

}

//获取属性类型
func (obj *GlobalSet) GetPropertyType(p string) (int, string, error) {
	switch p {
	case "Name":
		return DT_STRING, "string", nil
	default:
		return DT_NONE, "", ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *GlobalSet) GetPropertyIndex(p string) (int, error) {
	switch p {
	case "Name":
		return 0, nil
	default:
		return -1, ErrPropertyNotFound
	}
}

//通过属性名设置值
func (obj *GlobalSet) Inc(p string, v interface{}) error {
	switch p {
	default:
		return ErrPropertyNotFound
	}
	return nil
}

//通过属性名设置值
func (obj *GlobalSet) Set(p string, v interface{}) error {
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
func (obj *GlobalSet) SetByIndex(index int16, v interface{}) error {
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
func (obj *GlobalSet) MustGet(p string) interface{} {
	switch p {
	case "Name":
		return obj.Name
	default:
		return nil
	}
}

//通过属性名获取值
func (obj *GlobalSet) Get(p string) (val interface{}, err error) {
	switch p {
	case "Name":
		return obj.Name, nil
	default:
		return nil, ErrPropertyNotFound
	}
}

//是否需要同步到其它客户端
func (obj *GlobalSet) PropertyIsPublic(p string) bool {
	switch p {
	case "Name":
		return false
	default:
		return false
	}
}

//是否需要同步到自己的客户端
func (obj *GlobalSet) PropertyIsPrivate(p string) bool {
	switch p {
	case "Name":
		return false
	default:
		return false
	}
}

//是否需要存档
func (obj *GlobalSet) PropertyIsSave(p string) bool {
	switch p {
	case "Name":
		return true
	default:
		return false
	}
}

//脏标志(数据保存用)
func (obj *GlobalSet) setDirty(p string, v interface{}) {
	//obj.Mdirty[p] = v
	obj.SetSaveFlag()
}

func (obj *GlobalSet) GetDirty() map[string]interface{} {

	return obj.Mdirty
}

func (obj *GlobalSet) ClearDirty() {
	for k := range obj.Mdirty {
		delete(obj.Mdirty, k)
	}
}

//修改标志(数据同步用)
func (obj *GlobalSet) setModify(p string, v interface{}) {
	obj.Mmodify[p] = v
}

func (obj *GlobalSet) GetModify() map[string]interface{} {
	return obj.Mmodify
}

func (obj *GlobalSet) ClearModify() {
	for k := range obj.Mmodify {
		delete(obj.Mmodify, k)
	}
}

//名称
func (obj *GlobalSet) SetName(v string) {
	if obj.Name == v {
		return
	}

	old := obj.Name

	obj.Name = v
	if obj.prophooker != nil && obj.IsCritical(0) && !obj.GetPropFlag(0) {
		obj.SetPropFlag(0, true)
		obj.prophooker.OnPropChange(obj, "Name", old)
		obj.SetPropFlag(0, false)
	}
	obj.NameHash = Hash(v)

	obj.setDirty("Name", v)
}
func (obj *GlobalSet) GetName() string {
	return obj.Name
}

//初始化所有的表格
func (obj *GlobalSet) initRec() {

}

//获取某个表格
func (obj *GlobalSet) GetRec(rec string) Recorder {
	switch rec {
	default:
		return nil
	}
}

//获取所有表格名称
func (obj *GlobalSet) GetRecNames() []string {
	return []string{}
}

func (obj *GlobalSet) GlobalSetInit() {
	obj.quiting = false
	obj.Save = true
	obj.ObjectType = HELPER
	obj.InBase = false
	obj.InScene = false
}

//重置
func (obj *GlobalSet) Reset() {

	//属性初始化
	obj.GlobalSet_t = GlobalSet_t{}
	obj.GlobalSet_Save.GlobalSet_Save_Property = GlobalSet_Save_Property{}
	obj.GlobalSet_Propertys = GlobalSet_Propertys{}
	obj.GlobalSetInit()
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
func (obj *GlobalSet) Copy(other Entityer) error {
	if t, ok := other.(*GlobalSet); ok {
		//属性复制
		obj.DbId = t.DbId
		obj.NameHash = t.NameHash
		obj.IDHash = t.IDHash

		obj.GlobalSet_t = t.GlobalSet_t
		obj.GlobalSet_Save.GlobalSet_Save_Property = t.GlobalSet_Save_Property
		obj.GlobalSet_Propertys = t.GlobalSet_Propertys

		//表格复制

		return nil
	}

	return ErrCopyObjError
}

//DB相关
//同步到数据库
func (obj *GlobalSet) SyncToDb() {

}

//从data中取出configid
func (obj *GlobalSet) GetConfigFromDb(data interface{}) string {
	if v, ok := data.(*GlobalSet_Save); ok {
		return v.ConfigId
	}

	return ""
}

//从数据库恢复
func (obj *GlobalSet) SyncFromDb(data interface{}) bool {
	if v, ok := data.(*GlobalSet_Save); ok {
		obj.GlobalSet_Save.GlobalSet_Save_Property = v.GlobalSet_Save_Property

		obj.NameHash = Hash(obj.Name)
		obj.IDHash = Hash(obj.ConfigId)
		return true
	}

	return false
}

func (obj *GlobalSet) GetSaveLoader() DBSaveLoader {
	return &obj.GlobalSet_Save
}

func (obj *GlobalSet) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	var err error

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

	err = encoder.Encode(obj.GlobalSet_Save)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.GlobalSet_Propertys)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (obj *GlobalSet) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	var err error

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

	err = decoder.Decode(&obj.GlobalSet_Save)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.GlobalSet_Propertys)
	if err != nil {
		return err
	}

	return nil
}

//由子类调用的初始化函数
func (obj *GlobalSet) baseInit(dirty, modify, extra map[string]interface{}) {
	//初始化表格
	obj.initRec()
	obj.Mdirty = dirty
	obj.Mmodify = modify
	obj.ExtraData = extra
}

func (obj *GlobalSet) Serial() ([]byte, error) {
	ar := util.NewStoreArchiver(nil)
	ps := obj.GetVisiblePropertys(0)
	ar.Write(int16(len(ps)))

	return ar.Data(), nil
}

func (obj *GlobalSet) SerialModify() ([]byte, error) {
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

func (obj *GlobalSet) IsSceneData(prop string) bool {
	idx, err := obj.GetPropertyIndex(prop)
	if err != nil {
		return false
	}

	return IsGlobalSetSceneData(idx)
}

//通过scenedata同步
func (obj *GlobalSet) SyncFromSceneData(val interface{}) error {
	var sd *GlobalSetSceneData
	var ok bool
	if sd, ok = val.(*GlobalSetSceneData); !ok {
		return fmt.Errorf("type not GlobalSetSceneData", sd)
	}

	return nil
}

func (obj *GlobalSet) GetSceneData() interface{} {
	sd := &GlobalSetSceneData{}

	//属性

	//表格
	return sd
}

//创建函数
func CreateGlobalSet() *GlobalSet {
	obj := &GlobalSet{}

	obj.ObjectType = HELPER
	obj.initRec()

	obj.propcritical = make([]uint64, int(math.Ceil(float64(1)/64)))
	obj.propflag = make([]uint64, int(math.Ceil(float64(1)/64)))
	obj.GlobalSetInit()

	obj.Mdirty = make(map[string]interface{}, 32)
	obj.Mmodify = make(map[string]interface{}, 32)
	obj.ExtraData = make(map[string]interface{}, 16)

	return obj
}

type GlobalSetSceneData struct {
}

func IsGlobalSetSceneData(idx int) bool {
	switch idx {
	case 0: //名称
		return false
	}
	return false
}

func GlobalSetInit() {
	gob.Register(&GlobalSet_Save{})
	gob.Register(&GlobalSet{})
	gob.Register(&GlobalSetSceneData{})
}
