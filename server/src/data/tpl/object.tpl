// Code generated by data parser.
// DO NOT EDIT!
package entity
{{$Obj := .Name}}
import (
	. "data/datatype"
	"encoding/gob"
	"libs/log"
	"bytes"
	"encoding/json"
	{{if len .Interfaces}}."data/inter"{{end}}
	"fmt"
	"util"
	"math"
)

{{range .Records}}
//{{.Comment}}行定义
type {{$Obj}}{{.Name}}Row struct { {{range .Columns}}
	{{.Name}} {{.Type}} {{.Tag}} //{{.Comment}}{{end}}
}

//{{.Comment}}
type {{$Obj}}{{.Name}} struct {
	MaxRows int `json:"-"`
	Cols    int `json:"-"`
	Rows    []{{$Obj}}{{.Name}}Row
	Dirty   bool	`json:"-"`
	syncer  TableSyncer
}
{{end}}

type {{$Obj}}_Save_Property struct{
	Capacity    int32   `json:"C"` //容量 
	ConfigId	string  `json:"I"`{{range .Propertys}}{{if eq .Save "true"}}
	{{.Name}} {{.Type}} {{.Tag}}//{{.Comment}}{{end}}{{end}}
}

//保存到DB的数据
type {{$Obj}}_Save struct {
	{{$Obj}}_Save_Property

	{{range .Records}}{{ if eq .Save "true"}}
	{{.Name}}_r {{$Obj}}{{.Name}}{{end}}{{end}}
}

func (s *{{$Obj}}_Save) Base() string{
	{{if .Base}}
	return "{{.Base}}"{{else}}
	return ""
	{{end}}
}

func (s *{{$Obj}}_Save) Marshal() (map[string]interface{}, error) {
	data, err := json.Marshal(*s)
	if err != nil {
		return nil, err
	}

	var ret  map[string]interface{}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (s *{{$Obj}}_Save) Unmarshal(data map[string]interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(d, s)

	return err
}

func (s *{{$Obj}}_Save) InsertOrUpdate(eq ExecQueryer, insert bool, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) error{
	var sql string
	var args []interface{}
	if insert {
		sql = "INSERT INTO `tbl_{{tolower $Obj}}`(`id`,`capacity`,`configid`{{range $idx, $ele := .Propertys}}{{if eq $ele.Save "true"}},{{insertname $ele.Type $ele.Name}}{{end}}{{end}}%s ) VALUES(?,?,?{{range $idx, $ele := .Propertys}}{{if eq $ele.Save "true"}},{{placehold $ele.Type $ele.Name}}{{end}}{{end}}%s) "
		args = []interface{}{dbId, s.Capacity, s.ConfigId{{range $idx, $ele := .Propertys}}{{if eq $ele.Save "true"}},{{valuestr $ele.Type "s" $ele.Name}} {{end}}{{end}}}
		sql = fmt.Sprintf(sql, extfields, extplacehold)
		if extobjs != nil {
			args = append(args, extobjs...)
		}
	} else {
		sql = "UPDATE `tbl_{{tolower $Obj}}` SET %s`capacity`=?, `configid`=?{{range $idx, $ele := .Propertys}}{{if eq $ele.Save "true"}},{{updatename $ele.Type $ele.Name}}{{end}}{{end}} WHERE `id` = ?"
		if extobjs != nil {
			args = append(args, extobjs...)
		}
		args=append(args, []interface{}{s.Capacity, s.ConfigId{{range $idx, $ele := .Propertys}}{{if eq $ele.Save "true"}}, {{valuestr $ele.Type "s" $ele.Name}}{{end}}{{end}},dbId} ...)
		sql = fmt.Sprintf(sql, extfields)
		
	}
	
	
	if _, err := eq.Exec(sql, args...);err != nil {
		log.LogError("InsertOrUpdate error:",sql, args)
		return err
	} 

	return nil
}

func (s *{{$Obj}}_Save) Update(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) { {{$count := len .Propertys}}{{if gt $count 0}}
	if err = s.InsertOrUpdate(eq, false, dbId, extfields, extplacehold, extobjs...);err != nil {
		return
	}{{end}} 
	{{range .Records}}{{ if eq .Save "true"}}
	if err = s.{{.Name}}_r.Update(eq, dbId); err != nil {
		return
	}{{end}}
	{{end}}
	return
}

func (s *{{$Obj}}_Save) Insert(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) (err error) {
	if err = s.InsertOrUpdate(eq, true, dbId, extfields, extplacehold, extobjs...);err != nil {
		return 
	}
	{{range .Records}}{{ if eq .Save "true"}}
	if err = s.{{.Name}}_r.Update(eq, dbId); err != nil {
		return
	}{{end}}
	{{end}}
	return
}

func (s *{{$Obj}}_Save) Query(dbId uint64) (sql string, args []interface{}) {
	sql = "SELECT `id`,`capacity`,`configid`{{range $idx, $ele := .Propertys}}{{if eq $ele.Save "true"}},{{insertname $ele.Type $ele.Name}}{{end}}{{end}} %s FROM `tbl_{{tolower $Obj}}` WHERE `id`=? LIMIT 1"
	args=[]interface{}{dbId}
	return
}

func (s *{{$Obj}}_Save) Load(eq ExecQueryer, dbId uint64, extfield string, extobjs ...interface{}) error {
	sql, a := s.Query(dbId)
	sql = fmt.Sprintf(sql, extfield)
	r, err := eq.Query(sql, a...)
	if err != nil {
		log.LogError("load error:",err)
		return err
	}
	defer r.Close()
	if !r.Next() {
		log.LogError("load error:", sql, a)
		return ErrSqlRowError
	}
	args:=[]interface{}{&dbId, &s.Capacity, &s.ConfigId{{range $idx, $ele := .Propertys}}{{if eq $ele.Save "true"}},{{valuestr $ele.Type "&s" $ele.Name}} {{end}}{{end}} }
	if extobjs != nil {
		args = append(args, extobjs...)
	}
	if err = r.Scan(args...); err != nil {
		log.LogError("load error:",err)
		return err
	}
	{{range .Records}}{{ if eq .Save "true"}}
	if err = s.{{.Name}}_r.Load(eq, dbId); err != nil {
		log.LogError("load error:",err)
		return err
	}{{end}}{{end}}

	return nil
}

type {{.Name}}_Propertys struct {
	{{if .Propertys}}//属性定义{{range .Propertys}}{{if not (eq .Save "true")}}
	{{.Name}} {{.Type}} `type:"property",name:"{{.Name}}",datatype:"{{.Type}}"` //{{.Comment}}{{end}}{{end}}
	{{end}}	
}

type {{.Name}}_t struct {
	{{if not .Base}}
	dirty bool
	ObjectType int
	DbId uint64
	parent Entityer
	ObjId ObjectID
	Deleted bool
	NameHash int32
	IDHash int32
	ContainerInited bool 
	Index int //在容器中的位置
	Childs []Entityer
	ChildNum int
	{{end}}

	Save bool //是否保存
}

type {{.Name}} struct {
	{{if .Base}}{{.Base}}{{end}}
	{{range $k, $v := .Interfaces}}
	*{{$v}}{{end}}
	
	{{if not .Base}}
	Mdirty  map[string]interface{}
	Mmodify map[string]interface{}
	ExtraData map[string]interface{}
	loading bool
	quiting bool
	propsyncer PropSyncer
	prophooker PropHooker
	propcritical []uint64
	propflag []uint64
	{{end}}

	{{.Name}}_t
	{{$Obj}}_Save
	{{$Obj}}_Propertys
	
	{{if .Records}}
	//表格定义{{range .Records}}{{ if not (eq .Save "true")}}
	{{.Name}}_r {{$Obj}}{{.Name}} //{{.Comment}}{{end}}{{end}} {{end}}
}

func (obj *{{.Name}})SetLoading(loading bool){
	obj.loading = loading
}

func (obj *{{.Name}})IsLoading() bool{
	return obj.loading
}

func (obj *{{.Name}})SetQuiting() {
	obj.quiting = true
}

func (obj *{{.Name}})	IsQuiting() bool{
	return obj.quiting
}

func (obj *{{.Name}}) GetConfig() string {
	return obj.ConfigId
}

func (obj *{{.Name}}) SetConfig(config string) {
	obj.ConfigId = config
	obj.IDHash = Hash(config)
}

func (obj *{{.Name}}) SetSaveFlag() {
	root := obj.GetRoot()
	if root != nil {
		root.SetSaveFlag()
	} else {
		obj.dirty = true
	}
}

func (obj *{{.Name}}) ClearSaveFlag() {
	obj.dirty = false
	{{if .Records}}{{range .Records}}{{ if eq .Save "true"}}
	 obj.{{.Name}}_r.Dirty = false{{end}}{{end}}{{end}}
}

func (obj *{{.Name}}) NeedSave() bool {
	if obj.dirty  {{if .Records}}{{range .Records}}{{ if eq .Save "true"}} || obj.{{.Name}}_r.Dirty{{end}}{{end}}{{end}} {
		return true
	}
	return false
}

func (obj *{{.Name}}) ChangeCapacity(capacity int32) error {
	
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
func (obj *{{.Name}}) SetCapacity(capacity int32, initcap int32) {
	{{if .Base}}	
	obj.{{.Base}}.SetCapacity(capacity, initcap)
	obj.Capacity = obj.{{.Base}}.Capacity{{else}}
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
	{{end}}
}

func (obj *{{.Name}}) GetCapacity() int32 {
	return obj.Capacity
}

//获取实际的容量
func (obj *{{.Name}}) GetRealCap() int32 {
	if !obj.ContainerInited {
		return 0
	}
	return int32(len(obj.Childs))
}

{{if not .Base}}
func (obj *{{$Obj}}) GetRoot() Entityer {
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
func (obj *{{$Obj}}) GetDbId() uint64 {
	return obj.DbId
}

func (obj *{{$Obj}}) SetDbId(id uint64) {
	obj.DbId = id
}

func (obj *{{.Name}}) SetParent(p Entityer) {
	obj.parent = p
}

func (obj *{{.Name}}) GetParent() Entityer {
	return obj.parent
}

func (obj *{{.Name}}) SetDeleted(d bool) {
	obj.Deleted = d
}

func (obj *{{.Name}}) GetDeleted() bool {
	return obj.Deleted
}

func (obj *{{.Name}}) SetObjId(id ObjectID) {
	obj.ObjId = id
}

func (obj *{{.Name}}) GetObjId() ObjectID{
	return obj.ObjId
}


//设置名字Hash
func (obj *{{.Name}}) SetNameHash(v int32) {
	obj.NameHash = v
}

//获取名字Hash
func (obj *{{.Name}}) GetNameHash() int32 {
	return obj.NameHash 
}

//名字比较
func (obj *{{.Name}}) NameEqual(name string) bool{
	return obj.Name == name
}

//获取IDHash
func (obj *{{.Name}}) GetIDHash() int32 {
	return obj.IDHash 
}

//ID比较
func (obj *{{.Name}}) IDEqual(id string) bool{
	return obj.ConfigId == id
}

func (obj *{{.Name}}) ChildCount() int {
	return obj.ChildNum
}

//移除对象
func (obj *{{.Name}}) RemoveChild(de Entityer) error {
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
func (obj *{{.Name}}) GetChilds() []Entityer {
	return obj.Childs
}

//获取容器中索引
func (obj *{{.Name}}) GetIndex() int{
	return obj.Index
}

//设置索引，逻辑层不要调用
func (obj *{{.Name}}) SetIndex(idx int){
	obj.Index = idx
}

//删除所有的子对象
func (obj *{{.Name}}) ClearChilds() {
	for _, c := range obj.Childs {
		if c != nil{
			obj.RemoveChild(c)
		}
	}

	if obj.Capacity == -1 {		
		obj.Childs = obj.Childs[:0]
	}
	obj.ChildNum = 0
}

//增加子对象
func (obj *{{.Name}}) AddChild(idx int, e Entityer) (index int, err error){
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
		e.SetIndex(len(obj.Childs) -1 )
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
		log.LogError("out of range, ", idx,",", len(obj.Childs))
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
func (obj *{{.Name}}) GetChild(idx int) Entityer{
	if !obj.ContainerInited {
		return nil
	}
	if idx <0 || idx >= len(obj.Childs) {
		return nil
	}
	return obj.Childs[idx]
}

//通过ID获取子对象
func (obj *{{.Name}}) GetChildByConfigId(id string) Entityer{
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
func (obj *{{.Name}}) GetFirstChildByConfigId(id string) (int, Entityer){
	if !obj.ContainerInited {
		return -1, nil
	}
	h := Hash(id)
	for k, v := range obj.Childs {
		if (v != nil) && (v.GetIDHash() == h) && v.IDEqual(id) {
			return k+1, v
		}
	}
	return -1, nil
}
func (obj *{{.Name}}) GetNextChildByConfigId(start int, id string) (int, Entityer){

	if !obj.ContainerInited || start == -1 || start >= len(obj.Childs) {
		return -1, nil
	}
	h := Hash(id)
	for k, v := range obj.Childs[start:] {
		if (v != nil) && (v.GetIDHash() == h) && v.IDEqual(id) {
			return start+k+1, v
		}
	}
	return -1, nil
}

//通过名称获取子对象
func (obj *{{.Name}}) GetChildByName(name string) Entityer{
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
func (obj *{{.Name}}) GetFirstChild(name string) (int, Entityer){
	if !obj.ContainerInited {
		return -1, nil
	}
	h := Hash(name)
	for k, v := range obj.Childs {
		if (v != nil) && (v.GetNameHash() == h) && v.NameEqual(name) {
			return k+1, v
		}
	}
	return -1, nil
}
func (obj *{{.Name}}) GetNextChild(start int, name string) (int, Entityer){

	if !obj.ContainerInited || start == -1 || start >= len(obj.Childs) {
		return -1, nil
	}
	h := Hash(name)
	for k, v := range obj.Childs[start:] {
		if (v != nil) && (v.GetNameHash() == h) && v.NameEqual(name) {
			return start+k+1, v
		}
	}
	return -1, nil
}

//交换子对象的位置
func (obj *{{.Name}}) SwapChild(src int, dest int) error{
	if !obj.ContainerInited {
		return ErrContainerNotInit
	}
	if src < 0 || src >= len(obj.Childs) || dest <0 || dest >= len(obj.Childs) {
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
func (obj *{{.Name}}) Base () Entityer {	 
	 return nil
}

//获取对象类型
func (obj *{{.Name}}) ObjType() int {
	return obj.ObjectType
}

//额外的数据
func (obj *{{.Name}}) SetExtraData(key string, value interface{}) {
	obj.ExtraData[key] = value
}

func (obj *{{.Name}}) GetExtraData(key string) interface{}{
	if v, ok := obj.ExtraData[key] ; ok {
		return v
	}
	return nil
}

func (obj *{{.Name}}) GetAllExtraData() map[string]interface{} {
	return obj.ExtraData
}

func (obj *{{.Name}}) RemoveExtraData(key string) {
	if _, ok := obj.ExtraData[key] ; ok {			
		delete(obj.ExtraData, key)
	}
}

func (obj *{{.Name}}) ClearExtraData() {
	for k := range obj.ExtraData {
		delete(obj.ExtraData, k)
	}
}


{{else}}
//获取基类
func (obj *{{.Name}}) Base () Entityer {	 
	 return &obj.{{.Base}}
}
{{end}}

//获取对象是否保存
func (obj *{{.Name}}) IsSave() bool {
	return obj.Save
}

//设置对象是否保存
func (obj *{{.Name}}) SetSave(s bool)  {
	 obj.Save = s
}

//获取对象类型名
func (obj *{{.Name}}) ObjTypeName() string {
	return "{{.Name}}"
}

func (obj *{{.Name}}) SetPropSyncer(sync PropSyncer) {
	obj.propsyncer = sync
}

func (obj *{{.Name}}) GetPropSyncer() PropSyncer {
	return obj.propsyncer
}

//属性回调接口
func (obj *{{.Name}}) SetPropHooker(hooker PropHooker) {
	obj.prophooker = hooker	
}

func (obj *{{.Name}}) GetPropFlag(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propflag[index] & (uint64(1) << bit) != 0
}

func (obj *{{.Name}}) SetPropFlag(idx int, flag bool) {
	index := idx / 64
	bit := uint(idx) % 64
	if flag {
		obj.propflag[index] = obj.propflag[index] | (uint64(1) << bit)
		return
	}
	obj.propflag[index] = obj.propflag[index] & ^(uint64(1) << bit)
}

func (obj *{{.Name}}) IsCritical(idx int) bool {
	index := idx / 64
	bit := uint(idx) % 64
	return obj.propcritical[index] & (uint64(1)<<bit) != 0
}

func (obj *{{.Name}}) SetCritical(prop string) {
	idx, err := obj.GetPropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] | (uint64(1) << bit)
}

func (obj *{{.Name}}) ClearCritical(prop string) {
	idx, err := obj.GetPropertyIndex(prop)
	if err != nil {
		return
	}

	index := int(idx) / 64
	bit := uint(idx) % 64

	obj.propcritical[index] = obj.propcritical[index] & ^(uint64(1) << bit)
}

//获取所有属性
func (obj *{{.Name}}) GetPropertys() []string {	
	return []string{ {{range  .Propertys}}
	"{{.Name}}", {{end}}
	}		
}

//获取所有可视属性
func (obj *{{.Name}}) GetVisiblePropertys(typ int) []string {
	if typ == 0 {
		return []string{ {{range  .Propertys}}{{if eq (isprivate .Public) "true"}}
		"{{.Name}}", {{end}} {{end}}
		}
	} else {
		return []string{ {{range  .Propertys}}{{if eq (ispublic .Public) "true"}}
		"{{.Name}}", {{end}} {{end}}
		}
	}
	
}
//获取属性类型
func (obj *{{.Name}}) GetPropertyType(p string) (int, string, error) {
	switch p {
	{{range .Propertys}}case "{{.Name}}":
		return DT_{{toupper .Type}}, "{{.Type}}", nil
	{{end}}default:
	{{if .Base}}	return obj.BaseObj.GetPropertyType(p){{else}}	return DT_NONE, "", ErrPropertyNotFound{{end}}
	}
}

//通过属性名设置值
func (obj *{{.Name}}) GetPropertyIndex(p string) (int, error) {
	switch p {
	{{range $idx, $val := .Propertys}}case "{{$val.Name}}":
		return {{$idx}}, nil
	{{end}}default:
	{{if .Base}}	return obj.{{.Base}}.GetPropertyIndex(p){{else}}	return -1, ErrPropertyNotFound{{end}}
	}
}

//通过属性名设置值
func (obj *{{.Name}}) Inc(p string, v interface{}) error {
	switch p {
	{{range .Propertys}}{{if eq .Type "int8" "uint8" "int16" "uint16" "int32" "uint32" "int64" "uint64" "int" "int64" "float32" "float64"}}case "{{.Name}}":
		var dst {{.Type}}
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.Inc{{.Name}}(dst)
		} 
		return err
	{{end}}{{end}}default:
	{{if .Base}}	return obj.{{.Base}}.Set(p, v){{else}}	return ErrPropertyNotFound{{end}}
	}
	return nil
}

//通过属性名设置值
func (obj *{{.Name}}) Set(p string, v interface{}) error {
	switch p {
	{{range .Propertys}}case "{{.Name}}":{{if eq .Type "int8" "uint8" "int16" "uint16" "int32" "uint32" "int64" "uint64" "int" "int64" "float32" "float64"}}
		var dst {{.Type}}
		err := ParseNumber(v, &dst)
		if err == nil {
			obj.Set{{.Name}}(dst)
		} 
		return err {{else}}
		val, ok := v.({{.Type}})
		if ok {
			obj.Set{{.Name}}(val)
		} else { 
			return ErrTypeMismatch 
		}{{end}}
	{{end}}default:
	{{if .Base}}	return obj.{{.Base}}.Set(p, v){{else}}	return ErrPropertyNotFound{{end}}
	}
	return nil
}

//通过属性名获取值
func (obj *{{.Name}}) MustGet(p string) interface{} {
	switch p {
	{{range .Propertys}}case "{{.Name}}":
		return obj.{{.Name}}
	{{end}}default:
	{{if .Base}}	return obj.{{.Base}}.MustGet(p){{else}}	return nil{{end}}
	}
}

//通过属性名获取值
func (obj *{{.Name}}) Get(p string) (val interface{}, err error) {
	switch p {
	{{range .Propertys}}case "{{.Name}}":
		return obj.{{.Name}}, nil
	{{end}}default:
	{{if .Base}}	return obj.{{.Base}}.Get(p){{else}}	return nil, ErrPropertyNotFound{{end}}
	}
}

//是否需要同步到其它客户端
func (obj *{{.Name}}) PropertyIsPublic(p string) bool {
	switch p {
	{{range .Propertys}}case "{{.Name}}":
		return {{ispublic .Public}}
	{{end}}default:
	{{if .Base}}	return obj.{{.Base}}.PropertyIsPublic(p){{else}}	return false{{end}}
	}
}

//是否需要同步到自己的客户端
func (obj *{{.Name}}) PropertyIsPrivate(p string) bool {
	switch p {
	{{range .Propertys}}case "{{.Name}}":
		return {{isprivate .Public}}
	{{end}}default:
	{{if .Base}}	return obj.{{.Base}}.PropertyIsPrivate(p){{else}}	return false{{end}}
	}
}

//是否需要存档
func (obj *{{.Name}}) PropertyIsSave(p string) bool {
	switch p {
	{{range .Propertys}}case "{{.Name}}":
		return {{if eq .Save "true"}}true{{else}}false{{end}}
	{{end}}default:
	{{if .Base}}	return obj.{{.Base}}.PropertyIsSave(p){{else}}	return false{{end}}
	}
}{{if eq .Base ""}}

//脏标志(数据保存用)
func (obj *{{.Name}}) setDirty(p string, v interface{}) {
	//obj.Mdirty[p] = v
	obj.SetSaveFlag()
}

func (obj *{{.Name}}) GetDirty() map[string]interface{} {

	return obj.Mdirty
}

func (obj *{{.Name}}) ClearDirty() {
	for k := range obj.Mdirty {
		delete(obj.Mdirty, k)
	}
}

//修改标志(数据同步用)
func (obj *{{.Name}}) setModify(p string, v interface{}) {
	obj.Mmodify[p] = v
}

func (obj *{{.Name}}) GetModify() map[string]interface{} {
	return obj.Mmodify
}

func (obj *{{.Name}}) ClearModify() {
	for k := range obj.Mmodify {
		delete(obj.Mmodify, k)
	}
}{{end}}
{{range $idx, $prop := .Propertys}}
//{{$prop.Comment}}
func (obj *{{$Obj}}) Set{{$prop.Name}}(v {{$prop.Type}}) {
	if obj.{{$prop.Name}} == v {
		return
	}
	old := obj.{{$prop.Name}}
	obj.{{$prop.Name}} = v	
	if obj.prophooker != nil && obj.IsCritical({{$idx}}) && !obj.GetPropFlag({{$idx}}){
		obj.SetPropFlag({{$idx}}, true)
		obj.prophooker.OnPropChange(obj, "{{$prop.Name}}", old)
		obj.SetPropFlag({{$idx}}, false)
	}{{if eq $prop.Name "Name"}}
	obj.NameHash = Hash(v){{end}}{{if eq $prop.Realtime "true"}}{{if eq (isprivate $prop.Public) "true"}}
	if obj.propsyncer != nil {
		obj.propsyncer.Update({{$idx}}, v)
	} else {
		obj.setModify("{{$prop.Name}}", v)
	}{{end}}{{else if $prop.Public}}
	obj.setModify("{{$prop.Name}}", v){{end}}
	{{if eq $prop.Save "true"}}
	obj.setDirty("{{$prop.Name}}", v){{end}}
}
func (obj *{{$Obj}}) Get{{$prop.Name}}() {{$prop.Type}} {
	return obj.{{$prop.Name}}
}{{if eq $prop.Type "int8" "uint8" "int16" "uint16" "int32" "uint32" "int64" "uint64" "int" "int64" "float32" "float64"}}
func (obj *{{$Obj}}) Inc{{$prop.Name}}(v {{$prop.Type}}) {
	obj.Set{{$prop.Name}}(obj.{{$prop.Name}} + v)
}{{end}}{{end}}
{{range .Records}}

{{/* func (r *{{$Obj}}{{.Name}}Row) MarshalJSON() ([]byte, error) {

	w := bytes.NewBuffer(nil)
	enc := json.NewEncoder(w)
	err := enc.Encode([]interface{}{ {{range .Columns}}
		r.{{.Name}},{{end}} 
		})

	return w.Bytes(), err
}

func (r *{{$Obj}}{{.Name}}Row) UnmarshalJSON(data []byte) error {
	defer func() {
		if e := recover(); e != nil {
			log.LogError(e)
		}
	}()
	var row  []interface{}
	reader := bytes.NewReader(data)
	dec := json.NewDecoder(reader)
	err := dec.Decode(&row)
	if err != nil {
		return err
	}
	count := len(row)
	{{range $index, $ele := .Columns}}
	if count > {{$index}} {
	r.{{$ele.Name}} = {{$ele.Type}}(row[{{$index}}].({{if eq $ele.Type "string"}}string{{else}}float64{{end}}))
	} {{end}} 
	return nil
} */}}

func (rec *{{$Obj}}{{.Name}}) Marshal() ([]byte, error) {
	 return json.Marshal(rec)
}

func (rec *{{$Obj}}{{.Name}}) Unmarshal(data []byte) error {
	return json.Unmarshal(data, rec)
}

//DB
func (rec *{{$Obj}}{{.Name}}) Update(eq ExecQueryer, dbId uint64) error {
	{{if not (eq .Save "true")}}
	return nil{{else}}
	if !rec.Dirty {
		return nil
	}
	{{if eq .Type ""}}
	_, err := eq.Exec("DELETE FROM `tbl_{{tolower $Obj}}_{{tolower .Name}}` WHERE `id`=?", dbId)
	if err != nil {
		return err
	}
	rows := len(rec.Rows)
	sql := "INSERT INTO `tbl_{{tolower $Obj}}_{{tolower .Name}}` (`id`,`index`,`delete`{{range $idx, $ele := .Columns}},{{insertname $ele.Type $ele.Name}}{{end}}) VALUES(?,?,?{{range $idx, $ele := .Columns}},{{placehold $ele.Type $ele.Name}}{{end}})"
	db := eq.GetDB()
	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for i :=0; i < rows; i++ {	
		_, err := stmt.Exec(dbId, i, 0 {{range $idx, $ele := .Columns}},{{valuestr $ele.Type "rec.Rows[i]" $ele.Name}}{{end}})
		if err != nil {
			return err
		}
	}{{else}}
	data, err := rec.Marshal()
	if err != nil {
		return err
	}
	sql := "UPDATE `tbl_{{tolower $Obj}}` SET `r_{{tolower .Name}}`=? WHERE `id` = ?"
	
	if _, err := eq.Exec(sql, data, dbId);err != nil {
		log.LogError("update record {{$Obj}}{{.Name}} error:",sql, data, dbId)
		return err
	} 
	{{end}}

	return nil{{end}}
}

func (rec *{{$Obj}}{{.Name}}) Load(eq ExecQueryer, dbId uint64) error {
	{{if not (eq .Save "true")}}
	return nil{{else}}
	rec.Rows = rec.Rows[:0]
	{{if eq .Type ""}}
	sql := "SELECT `index`{{range $idx, $ele := .Columns}},{{insertname $ele.Type $ele.Name}}{{end}} FROM `tbl_{{tolower $Obj}}_{{tolower .Name}}` WHERE `id`=? and `delete`=0 ORDER BY `index`"
	r, err := eq.Query(sql, dbId)
	if err != nil {
		return err
	}
	defer r.Close()
	var index int
	for r.Next() {	
		row := {{$Obj}}{{.Name}}Row{}
		if err = r.Scan(&index{{range .Columns}},{{valuestr .Type "&row" .Name}}{{end}} ); err != nil {
			return err
		}
		rec.Rows = append(rec.Rows, row)
		r.Next()
	}{{else}}
	sql := "SELECT `r_{{tolower .Name}}` FROM `tbl_{{tolower $Obj}}` WHERE `id`=? LIMIT 1"
	r, err := eq.Query(sql, dbId)
	if err != nil {
		log.LogError("load record {{$Obj}}{{.Name}} error:", err)
		return err
	}
	defer r.Close()
	if !r.Next() {
		log.LogError("load record {{$Obj}}{{.Name}} error:", sql, dbId)
		return ErrSqlRowError
	}
	var json []byte
	if err = r.Scan(&json); err != nil {
		log.LogError("load record {{$Obj}}{{.Name}} error:", err)
		return err
	}

	if json == nil || len(json) < 2 {
		log.LogWarning("load record {{$Obj}}{{.Name}} error: nil")
		return nil
	}
	
	err = rec.Unmarshal(json)
	if err != nil {
		log.LogError("unmarshal record {{$Obj}}{{.Name}} error:", err)
		return err
	}
	{{end}}

	return nil{{end}}
}

func (rec *{{$Obj}}{{.Name}}) GetName() string {
	return "{{.Name}}"
}
//表格的容量
func (rec *{{$Obj}}{{.Name}}) GetCap() int {
	return rec.MaxRows
}

//表格当前的行数
func (rec *{{$Obj}}{{.Name}}) GetRows() int {
	return len(rec.Rows)
}

//获取列定义
func (rec *{{$Obj}}{{.Name}}) ColTypes() ([]int, []string) {
	col := []int{ {{range $idx, $ele := .Columns}}{{if eq $idx 0}}DT_{{toupper $ele.Type}}{{else}}, DT_{{toupper $ele.Type}}{{end}}{{end}} }
	cols := []string{ {{range $idx, $ele := .Columns}}{{if eq $idx 0}}"{{$ele.Type}}"{{else}}, "{{$ele.Type}}"{{end}}{{end}} }
	return col, cols
}

//获取列数
func (rec *{{$Obj}}{{.Name}}) GetCols() int {
	return rec.Cols
}

//是否要同步到客户端
func (rec *{{$Obj}}{{.Name}}) IsVisible() bool {
	return {{if eq .Visible "true"}}true{{else}}false{{end}}
}

//脏标志
func (rec *{{$Obj}}{{.Name}}) IsDirty() bool {
	return rec.Dirty
}

func (rec *{{$Obj}}{{.Name}}) ClearDirty() {
	rec.Dirty = false
}

func (rec *{{$Obj}}{{.Name}}) SetSyncer(s TableSyncer){
	rec.syncer = s
}

func (rec *{{$Obj}}{{.Name}}) GetSyncer() TableSyncer{
	return rec.syncer
}

//序列化
func (rec *{{$Obj}}{{.Name}}) Serial()([]byte, error){
	ar := util.NewStoreArchive()
	for _, v := range rec.Rows { {{range .Columns}}
		{{if eq .Type "string"}}ar.WriteString(v.{{.Name}}){{else if eq .Type "ObjectID"}}ar.WriteObject(v.{{.Name}}){{else}}ar.Write(v.{{.Name}}){{end}}{{end}}
	}
	return ar.Data(), nil
}

//序列化一行
func (rec *{{$Obj}}{{.Name}}) SerialRow(row int)([]byte, error){
	if row < 0 || row >= len(rec.Rows) {
		return nil, ErrRowError
	}
	ar := util.NewStoreArchive()
	v := rec.Rows[row] {{range .Columns}}
	{{if eq .Type "string"}}ar.WriteString(v.{{.Name}}){{else if eq .Type "ObjectID"}}ar.WriteObject(v.{{.Name}}){{else}}ar.Write(v.{{.Name}}){{end}}{{end}}
	return ar.Data(), nil
}

//通过行列设置值
func (rec *{{$Obj}}{{.Name}}) Set(row, col int, val interface{}) error {
	if row < 0 || row >= len(rec.Rows) {
		return ErrRowError
	}

	if col < 0 || col >= {{len .Columns}} {
		return ErrColError
	}

	r := rec.Rows[row]

	switch col { {{range $idx, $ele := .Columns}}
	case {{$idx}}:
		val, ok := val.({{$ele.Type}})
		if ok {
			r.{{$ele.Name}} = val
		} else {
			return ErrTypeMismatch
		}{{end}}
	}
	if rec.syncer != nil {
		rec.syncer.RecModify(rec, row, col)
	}
	{{if eq .Save "true"}}rec.Dirty = true{{end}}
	return nil
}

//通过行列获取值
func (rec *{{$Obj}}{{.Name}}) Get(row, col int) (val interface{}, err error) {
	if row < 0 || row >= len(rec.Rows) {
		err = ErrRowError
		return
	}

	if col < 0 || col >= {{len .Columns}} {
		err = ErrColError
		return
	}

	r := rec.Rows[row]

	switch col { {{range $idx, $ele := .Columns}}
	case {{$idx}}:
		val = r.{{$ele.Name}}{{end}}
	}

	return
}{{$rec := .Name}}
{{range $idx, $col := .Columns}}
//查找{{$col.Comment}}
func (rec *{{$Obj}}{{$rec}}) Find{{$col.Name}}(v {{$col.Type}}) int {
	for idx, row := range rec.Rows {
		if row.{{$col.Name}} == v {
			return idx
		}
	}
	return -1
}

//查找{{$col.Comment}}
func (rec *{{$Obj}}{{$rec}}) FindNext{{$col.Name}}(v {{$col.Type}}, itr int) int {
	itr++
	if itr+1 >= len(rec.Rows) {
		return -1
	}
	for idx, row := range rec.Rows[itr:] {
		if row.{{$col.Name}} == v {
			return idx
		}
	}
	return -1
}

//设置{{$col.Comment}}
func (rec *{{$Obj}}{{$rec}}) Set{{$col.Name}}(row int, v {{$col.Type}}) error {
	if row < 0 || row >= len(rec.Rows) {
		return ErrRowError
	}

	rec.Rows[row].{{$col.Name}} = v
	rec.Dirty = true
	if rec.syncer != nil {
		rec.syncer.RecModify(rec, row, {{$idx}})
	}
	return nil
}

//获取{{$col.Comment}}
func (rec *{{$Obj}}{{$rec}}) Get{{$col.Name}}(row int) (val {{$col.Type}}, err error) {
	if row < 0 || row >= len(rec.Rows) {
		err = ErrRowError
		return
	}

	val = rec.Rows[row].{{$col.Name}}
	return
}{{end}}

//设置一行的值
func (rec *{{$Obj}}{{.Name}}) SetRow(row int, args ...interface{} ) error {
	
	{{range $idx, $c := .Columns}}
	if _, ok := args[{{$idx}}].({{$c.Type}}); !ok {
		return ErrColTypeError
	}{{end}}
	return rec.SetRowValue(row,{{range $idx, $c := .Columns}}{{if eq 0 $idx}}args[{{$idx}}].({{$c.Type}}){{else}},args[{{$idx}}].({{$c.Type}}){{end}}{{end}})
}

//设置一行的值
func (rec *{{$Obj}}{{.Name}}) SetRowValue(row int{{range .Columns}}, {{tolower .Name}} {{.Type}}{{end}} ) error {
	if row < 0 || row >= len(rec.Rows) {
		return ErrRowError
	}
	{{range $idx, $c := .Columns}}
	rec.Rows[row].{{$c.Name}}={{tolower $c.Name}}{{end}}

	if rec.syncer != nil {
		rec.syncer.RecSetRow(rec, row)
	}
	rec.Dirty = true
	return nil
}

//增加一行,row=-1时,在表格最后面插入一行,否则在row处插入,返回-1插入失败,否则返回插入的行号
func (rec *{{$Obj}}{{.Name}}) Add(row int, args ...interface{} ) int {
	if len(args) != rec.Cols {
		return -1
	}
	{{range $idx, $c := .Columns}}
	if _, ok := args[{{$idx}}].({{$c.Type}}); !ok {
		return -1
	}{{end}}
	return rec.AddRowValue(row,{{range $idx, $c := .Columns}}{{if eq 0 $idx}}args[{{$idx}}].({{$c.Type}}){{else}},args[{{$idx}}].({{$c.Type}}){{end}}{{end}})
}

//增加一行,row=-1时,在表格最后面插入一行,否则在row处插入,返回-1插入失败,否则返回插入的行号
func (rec *{{$Obj}}{{.Name}}) AddRow(row int) int {
	add := -1

	if len(rec.Rows) >= rec.MaxRows {
		return add
	}

	r := {{$Obj}}{{.Name}}Row{}

	return rec.AddRowValue(row{{range $idx, $c := .Columns}},r.{{$c.Name}}{{end}})
	{{/*if row == -1 {
		rec.Rows = append(rec.Rows, r)
		add = len(rec.Rows) - 1
	} else {
		if row >= 0 && row < len(rec.Rows) {
			rec.Rows = append(rec.Rows, {{$Obj}}{{.Name}}Row{})
			copy(rec.Rows[row+1:], rec.Rows[row:])
			rec.Rows[row] = r
			add = row
		} else {
			rec.Rows = append(rec.Rows, r)
			add = len(rec.Rows) - 1
		}

	}
	if add != -1 {
		if rec.syncer != nil {
			rec.syncer.RecAppend(rec, add)
		}
		rec.Dirty = true
	}
	return add*/}}
}

//增加一行,row=-1时,在表格最后面插入一行,否则在row处插入,返回-1插入失败,否则返回插入的行号
func (rec *{{$Obj}}{{.Name}}) AddRowValue(row int {{range .Columns}}, {{tolower .Name}} {{.Type}}{{end}} ) int {
	add := -1

	if len(rec.Rows) >= rec.MaxRows {
		return add
	}

	r := {{$Obj}}{{.Name}}Row{ {{range $idx, $ele := .Columns}}{{if eq $idx 0}}{{tolower $ele.Name}}{{else}}, {{tolower $ele.Name}}{{end}}{{end}}}

	if row == -1 {
		rec.Rows = append(rec.Rows, r)
		add = len(rec.Rows) - 1
	} else {
		if row >= 0 && row < len(rec.Rows) {
			rec.Rows = append(rec.Rows, {{$Obj}}{{.Name}}Row{})
			copy(rec.Rows[row+1:], rec.Rows[row:])
			rec.Rows[row] = r
			add = row
		} else {
			rec.Rows = append(rec.Rows, r)
			add = len(rec.Rows) - 1
		}

	}
	if add != -1 {
		if rec.syncer != nil {
			rec.syncer.RecAppend(rec, add)
		}
		rec.Dirty = true
	}
	return add
}

//获取一行数据
func (rec *{{$Obj}}{{.Name}}) GetRow(row int)({{range .Columns}}{{tolower .Name}} {{.Type}},{{end}} err error) {

	if row < 0 || row >= len(rec.Rows) {
		err = ErrRowError
		return
	}

	r := rec.Rows[row]
	{{range .Columns}}{{tolower .Name}} = r.{{.Name}} 
	{{end}}
	return
}

//获取一行数据
func (rec *{{$Obj}}{{.Name}}) GetRowInterface(row int)(rowvalue interface{}, err error) {

	if row < 0 || row >= len(rec.Rows) {
		err = ErrRowError
		return
	}

	rowvalue = rec.Rows[row]
	return
}

//获取数据
func (rec *{{$Obj}}{{.Name}}) Scan(row int {{range .Columns}}, {{tolower .Name}} *{{.Type}}{{end}}) bool {

	if row < 0 || row >= len(rec.Rows) {
		return false
	}

	r := rec.Rows[row]
	{{range .Columns}}*{{tolower .Name}} = r.{{.Name}} 
	{{end}}
	return true
}

//删除一行
func (rec *{{$Obj}}{{.Name}}) Del(row int) {
	if row < 0 || row >= len(rec.Rows) {
		return
	}

	copy(rec.Rows[row:], rec.Rows[row+1:])
	rec.Rows = rec.Rows[:len(rec.Rows)-1]
	rec.Dirty = true

	if rec.syncer != nil {
		rec.syncer.RecDelete(rec, row)
	}
}

//清空表格
func (rec *{{$Obj}}{{.Name}}) Clear() {
	rec.Rows = rec.Rows[:0]
	rec.Dirty = true
	if rec.syncer != nil {
		rec.syncer.RecClear(rec)
	}
}

//是否保存
func (rec *{{$Obj}}{{.Name}}) IsSave() bool{
	return {{.Save}}
}

//初始化{{$Obj}}{{.Name}}表
func (obj *{{$Obj}}) init{{$Obj}}{{.Name}}() {
	obj.{{.Name}}_r.MaxRows = {{.Maxrows}}
	obj.{{.Name}}_r.Cols = {{len .Columns}}
	obj.{{.Name}}_r.Rows = make([]{{$Obj}}{{.Name}}Row, 0, {{.Maxrows}})
}

//获取{{$Obj}}{{.Name}}表
func (obj *{{$Obj}}) Get{{$Obj}}{{.Name}}() *{{$Obj}}{{.Name}} {
	return &obj.{{.Name}}_r
}{{end}}

//初始化所有的表格
func (obj *{{$Obj}}) initRec() {
	{{range .Records}}
	obj.init{{$Obj}}{{.Name}}(){{end}}
}

//获取某个表格
func (obj *{{$Obj}}) GetRec(rec string) Recorder {
	switch rec { {{range .Records}}
	case "{{.Name}}":
		return &obj.{{.Name}}_r{{end}}
	default:{{if .Base}}
		return obj.BaseObj.GetRec(rec){{else}}
		return nil{{end}}
	}
}

//获取所有表格名称
func (obj *{{$Obj}}) GetRecNames() []string {
	return []string{ {{range $index, $ele := .Records}}{{if $index}},"{{$ele.Name}}"{{else}}"{{$ele.Name}}"{{end}}{{end}} }
}

func (obj *{{$Obj}}){{$Obj}}Init() {
	obj.quiting = false
	obj.Save = true
	obj.ObjectType = {{.Type}}
}

//重置
func (obj *{{$Obj}}) Reset() {
	{{range $k, $v := .Interfaces}}
	obj.{{$v}}.Clear(){{end}}
	//属性初始化
	obj.{{$Obj}}_t = {{$Obj}}_t{}
	obj.{{$Obj}}_Save.{{$Obj}}_Save_Property = {{$Obj}}_Save_Property{}
	obj.{{$Obj}}_Propertys = {{$Obj}}_Propertys{}
	obj.{{$Obj}}Init()
	//表格初始化{{range .Records}}	
	obj.{{.Name}}_r.Clear()
	obj.{{.Name}}_r.syncer = nil{{end}}

	{{if .Base}}
	//基类初始化
	obj.{{.Base}}.Reset(){{else}}
	obj.ClearDirty()
	obj.ClearModify()
	obj.ClearExtraData()
	for k :=  range obj.propcritical {
		obj.propcritical[k] = 0
		obj.propflag[k] = 0
	}
	{{end}}
}

//对象拷贝
func (obj *{{$Obj}}) Copy(other Entityer) error {
	if t, ok := other.(*{{$Obj}}); ok {
		//属性复制{{if not .Base}}
		obj.DbId = t.DbId
		obj.NameHash = t.NameHash
		obj.IDHash = t.IDHash{{end}}

		{{range $k, $v := .Interfaces}}
		*obj.{{$v}} = *t.{{$v}}{{end}}
		obj.{{$Obj}}_t = t.{{$Obj}}_t
		obj.{{$Obj}}_Save.{{$Obj}}_Save_Property = t.{{$Obj}}_Save_Property
		obj.{{$Obj}}_Propertys = t.{{$Obj}}_Propertys	

		

		{{with .Records}}var l int{{end}}
		//表格复制{{range .Records}}{{$recname := .Name}}	
		obj.{{.Name}}_r.Clear()
		l = len(t.{{.Name}}_r.Rows)
		for i := 0; i <l; i++ {
		obj.{{.Name}}_r.AddRowValue(-1 {{range $i, $k :=.Columns}}, t.{{$recname}}_r.Rows[i].{{$k.Name}}{{end}} ) 
		}
		{{end}}
		{{if .Base}}
		return obj.{{.Base}}.Copy(&t.{{.Base}}){{else}}
		return nil{{end}}
	}

	return ErrCopyObjError
}

//DB相关
//同步到数据库
func (obj *{{$Obj}}) SyncToDb() {
	
}

//从data中取出configid
func (obj *{{$Obj}}) GetConfigFromDb(data interface{}) string {
	if v, ok := data.(*{{$Obj}}_Save); ok {
		return v.ConfigId
	}

	return ""
}

//从数据库恢复
func (obj *{{$Obj}}) SyncFromDb(data interface{}) bool{
	if v, ok := data.(*{{$Obj}}_Save); ok {
		obj.{{$Obj}}_Save.{{$Obj}}_Save_Property = v.{{$Obj}}_Save_Property

		{{range .Records}}{{ if eq .Save "true"}}
		{{$recname := .Name}}
		obj.{{.Name}}_r.Clear()
		if l := len(v.{{.Name}}_r.Rows); l > 0 {
		for i := 0; i <l; i++ {
		obj.{{.Name}}_r.AddRowValue(-1 {{range $i, $k :=.Columns}}, v.{{$recname}}_r.Rows[i].{{$k.Name}}{{end}} ) 
		}
		}
		{{end}}{{end}}
		{{if not .Base}}obj.NameHash=Hash(obj.Name)
		obj.IDHash=Hash(obj.ConfigId){{end}}
		return true
	}

	return false
}

func (obj *{{$Obj}}) GetSaveLoader() DBSaveLoader{
	return &obj.{{$Obj}}_Save
}

func (obj *{{$Obj}}) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	var err error
	{{if .Base}}
	err = encoder.Encode(&obj.{{.Base}})
	if err != nil {
		return nil, err
	}{{end}}
	{{if not .Base}}	
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
	}{{end}}	

	{{range $k, $v := .Interfaces}}
	err = encoder.Encode(obj.{{$v}})
	if err != nil {
		return nil, err
	}{{end}}
	err = encoder.Encode(obj.{{$Obj}}_Save)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(obj.{{$Obj}}_Propertys)
	if err != nil {
		return nil, err
	}
	{{range .Records}}{{ if not (eq .Save "true")}}
	err = encoder.Encode(obj.{{.Name}}_r)
	if err != nil {
		return nil, err
	}{{end}}{{end}}	
	return w.Bytes(), nil
}

func (obj *{{$Obj}}) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	var err error
	{{if .Base}}
	err = decoder.Decode(&obj.{{.Base}})
	if err != nil {
		return err
	}{{end}}
	{{if not .Base}}
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
		return  err
	}{{end}}	

	{{range $k, $v := .Interfaces}}
	err = decoder.Decode(obj.{{$v}})
	if err != nil {
		return err
	}{{end}}
	err = decoder.Decode(&obj.{{$Obj}}_Save)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.{{$Obj}}_Propertys)
	if err != nil {
		return err
	}
	{{range .Records}}{{ if not (eq .Save "true")}}
	err = decoder.Decode(&obj.{{.Name}}_r)
	if err != nil {
		return err
	}{{end}}{{end}}
	return nil
}

//由子类调用的初始化函数
func (obj *{{$Obj}}) baseInit(dirty, modify, extra map[string]interface{}) {
	//初始化表格
	obj.initRec(){{if .Base}}
	//初始化基类
	obj.{{.Base}}.baseInit(dirty, modify,extra){{else}}
	obj.Mdirty = dirty
	obj.Mmodify = modify
	obj.ExtraData = extra{{end}}
}

func (obj *{{$Obj}}) Serial()([]byte, error) {
	ar := util.NewStoreArchive()
	ps := obj.GetVisiblePropertys(0)
	ar.Write(int16(len(ps)))
	{{range  $idx, $p := .Propertys}}{{if eq (isprivate $p.Public) "true"}}
	ar.Write(int16({{$idx}}))
	ar.Write(obj.{{$p.Name}}){{end}}{{end}}
	return ar.Data(), nil
}

func (obj *{{$Obj}}) SerialModify()([]byte, error) {
	if len(obj.Mmodify) == 0 {
		return nil, nil
	}
	ar := util.NewStoreArchive()
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

//创建函数
func Create{{$Obj}}() *{{$Obj}} {
	obj := &{{$Obj}}{}
	{{range $k, $v := .Interfaces}}
	obj.{{$v}} = &{{$v}}{}
	obj.{{$v}}.Init(){{end}}
	obj.ObjectType = {{.Type}}
	obj.initRec()
	obj.propcritical = make([]uint64, int(math.Ceil(float64({{len .Propertys}}) / 64)))
	obj.propflag = make([]uint64, int(math.Ceil(float64({{len .Propertys}}) / 64)))
	obj.{{$Obj}}Init()
	{{if .Base}}
	obj.{{.Base}}.baseInit(
	make(map[string]interface{},32), 
	make(map[string]interface{},32),
	make(map[string]interface{},16),
	){{else}}
	obj.Mdirty = make(map[string]interface{},32)
	obj.Mmodify = make(map[string]interface{},32)
	obj.ExtraData = make(map[string]interface{},16)
	{{end}}
	return obj
}

func {{$Obj}}Init() {
	gob.Register(&{{$Obj}}_Save{})
	gob.Register(&{{$Obj}}{})
}