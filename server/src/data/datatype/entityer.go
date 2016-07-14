package datatype

import (
	"database/sql"
	"database/sql/driver"
)

type ExecQueryer interface {
	Exec(query string, args ...interface{}) (result driver.Result, err error)
	Query(query string, args ...interface{}) (row *sql.Rows, err error)
	GetDB() *sql.DB
}

type DBSaveLoader interface {
	Base() string
	Update(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) error
	Insert(eq ExecQueryer, dbId uint64, extfields string, extplacehold string, extobjs ...interface{}) error
	Load(eq ExecQueryer, dbId uint64, extfield string, extobjs ...interface{}) error
	Marshal() (map[string]interface{}, error)
	Unmarshal(data map[string]interface{}) error
}

type Recorder interface {
	GetName() string
	GetCap() int
	GetCols() int
	GetRows() int
	ColTypes() ([]int, []string)
	IsDirty() bool
	IsSave() bool
	IsVisible() bool
	ClearDirty()
	Set(row, col int, val interface{}) error
	Get(row, col int) (val interface{}, err error)
	SetRow(row int, args ...interface{}) error
	GetRowInterface(row int) (rowvalue interface{}, err error)
	Add(row int, args ...interface{}) int
	AddRow(row int) int
	Del(row int)
	Clear()
	SetSyncer(s TableSyncer)
	GetSyncer() TableSyncer
	Serial() ([]byte, error)
	SerialRow(row int) ([]byte, error)
}

type TableSyncer interface {
	RecAppend(rec Recorder, row int)
	RecDelete(rec Recorder, row int)
	RecClear(rec Recorder)
	RecModify(rec Recorder, row, col int)
	RecSetRow(rec Recorder, row int)
}

type PropSyncer interface {
	Update(index int16, value interface{})
}

type PropHooker interface {
	OnPropChange(object Entityer, prop string, value interface{})
}

type EntityInfo struct {
	Type   string
	Caps   int32
	DbId   uint64
	ObjId  ObjectID
	Index  int
	Data   []byte
	Childs []*EntityInfo
}

type Entityer interface {
	SetPropSyncer(sync PropSyncer)
	GetPropSyncer() PropSyncer
	SetPropHooker(hooker PropHooker)
	GetPropFlag(idx int) bool
	SetPropFlag(idx int, flag bool)
	IsCritical(idx int) bool
	SetCritical(prop string)
	ClearCritical(prop string)
	SetLoading(loading bool)
	IsLoading() bool
	SetQuiting()
	IsQuiting() bool
	GetConfig() string
	SetConfig(config string)
	SetSaveFlag()
	ClearSaveFlag()
	NeedSave() bool
	GetRoot() Entityer
	SetParent(p Entityer)
	GetParent() Entityer
	SetDeleted(d bool)
	GetDeleted() bool
	SetObjId(id ObjectID)
	GetObjId() ObjectID
	SetNameHash(v int32)
	GetIDHash() int32
	IDEqual(id string) bool
	GetNameHash() int32
	NameEqual(name string) bool
	SetCapacity(capacity int32, initcap int32)
	ChangeCapacity(capacity int32) error
	GetCapacity() int32
	GetRealCap() int32
	ChildCount() int
	GetChilds() []Entityer
	GetIndex() int
	SetIndex(idx int)
	ClearChilds()
	AddChild(idx int, e Entityer) (index int, err error)
	RemoveChild(e Entityer) error
	GetChild(idx int) Entityer
	GetChildByConfigId(id string) Entityer
	GetFirstChildByConfigId(id string) (int, Entityer)
	GetNextChildByConfigId(start int, id string) (int, Entityer)
	GetChildByName(name string) Entityer
	GetFirstChild(name string) (int, Entityer)
	GetNextChild(start int, name string) (int, Entityer)
	SwapChild(src int, dest int) error
	SetExtraData(key string, value interface{})
	GetExtraData(key string) interface{}
	GetAllExtraData() map[string]interface{}
	RemoveExtraData(key string)
	ClearExtraData()
	ObjType() int
	Base() Entityer
	GetDbId() uint64
	SetDbId(id uint64)
	IsSave() bool
	SetSave(s bool)
	ObjTypeName() string
	//property
	GetPropertys() []string
	GetVisiblePropertys(typ int) []string
	GetPropertyType(p string) (int, string, error)
	GetPropertyIndex(p string) (int, error)
	Inc(p string, v interface{}) error
	Set(p string, v interface{}) error
	MustGet(p string) interface{}
	Get(p string) (val interface{}, err error)
	PropertyIsPublic(p string) bool
	PropertyIsPrivate(p string) bool
	PropertyIsSave(p string) bool
	GetDirty() map[string]interface{}
	ClearDirty()
	GetModify() map[string]interface{}
	ClearModify()
	//record
	GetRec(rec string) Recorder
	GetRecNames() []string
	Reset()
	Copy(other Entityer) error
	//DB
	SyncToDb()
	GetConfigFromDb(data interface{}) string
	SyncFromDb(data interface{}) bool
	GetSaveLoader() DBSaveLoader
	Serial() ([]byte, error)
	SerialModify() ([]byte, error)
}
