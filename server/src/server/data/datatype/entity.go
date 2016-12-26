package datatype

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"server/libs/log"
)

var (
	objects                     = make(map[string]func() Entity)
	ErrRowError                 = errors.New("row index out of range")
	ErrColError                 = errors.New("col index out of range")
	ErrTypeMismatch             = errors.New("val type mismatch")
	ErrColTypeError             = errors.New("column type error")
	ErrPropertyNotFound         = errors.New("property not found")
	ErrSqlRowError              = errors.New("sql query not found")
	ErrSqlUpdateError           = errors.New("update id not found")
	ErrContainerFull            = errors.New("container is full")
	ErrContainerIndexHasChild   = errors.New("container index not empty")
	ErrContainerIndexOutOfRange = errors.New("container index out of range")
	ErrContainerNotInit         = errors.New("container not init")
	ErrContainerCapacity        = errors.New("capacity illegal")
	ErrChildObjectNotFound      = errors.New("child obj not found")
	ErrCopyObjError             = errors.New("type not equal")
	ErrExtraDataError           = errors.New("extra data not found")
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

type Record interface {
	//获取表格名
	Name() string
	//获取表最大行数
	Caps() int
	//获取表格列数
	ColCount() int
	//获取表格行数
	RowCount() int
	//获取列类型
	ColTypes() ([]int, []string)
	//脏标志
	IsDirty() bool
	//清理脏标志
	ClearDirty()
	//是否保存
	IsSave() bool
	//是否可视
	IsVisible() bool
	//设置单元格值
	Set(row, col int, val interface{}) error
	//获取单元格值
	Get(row, col int) (val interface{}, err error)
	//设置一行的值
	SetRow(row int, args ...interface{}) error
	//通过rowdata设置一行的值
	SetRowByBytes(row int, rowdata []byte) error
	//通过行类型进行设置一行的值
	SetRowInterface(row int, rowvalue interface{}) error
	//获取一行
	FindRowInterface(row int) (rowvalue interface{}, err error)

	//增加一行数据,row插入的位置，-1表示插入在最后
	Add(row int, args ...interface{}) int
	//增加一行数据,row插入的位置，-1表示插入在最后
	AddByBytes(row int, rowdata []byte) int
	//增加一行
	AddRow(row int) int
	//删除一行
	Del(row int)
	//清除表格内容
	Clear()
	//表格监视
	SetMonitor(s TableMonitor)
	Monitor() TableMonitor
	//序列号表格
	Serial() ([]byte, error)
	//序列号一行
	SerialRow(row int) ([]byte, error)
}

//表格变动监视
type TableMonitor interface {
	RecAppend(self Entity, rec Record, row int)
	RecDelete(self Entity, rec Record, row int)
	RecClear(self Entity, rec Record)
	RecModify(self Entity, rec Record, row, col int)
	RecSetRow(self Entity, rec Record, row int)
}

//属性更新回调，主要针对客户端可视的属性，用来向客户端同步数据
type PropUpdater interface {
	Update(self Entity, index int16, value interface{})
}

//属性变动回调，逻辑层挂勾后的属性回调的接口
type PropChanger interface {
	OnPropChange(object Entity, prop string, value interface{})
}

//entity序列化数据
type EntityInfo struct {
	Type   string
	Caps   int32
	DbId   uint64
	ObjId  ObjectID
	Index  int
	Data   []byte
	Childs []*EntityInfo
}

type Entity interface {
	//唯一标识，也就是mailbox
	UID() uint64
	SetUID(v uint64)
	//是否在base中
	SetInBase(v bool)
	IsInBase() bool
	//是否在场景中
	SetInScene(v bool)
	IsInScene() bool
	//属性同步模块
	SetPropUpdate(sync PropUpdater)
	PropUpdate() PropUpdater
	//属性回调挂钩
	SetPropHook(hooker PropChanger)
	//设置属性标志(内部使用)
	PropFlag(idx int) bool
	SetPropFlag(idx int, flag bool)
	//设置关键属性(回调标志)
	IsCritical(idx int) bool
	SetCritical(prop string)
	ClearCritical(prop string)
	//加载标志
	SetLoading(loading bool)
	IsLoading() bool
	//退出标志
	SetQuiting()
	IsQuiting() bool
	//获取配置编号
	Config() string
	SetConfig(config string)
	ConfigHash() int32
	//设置存档标志
	SetSaveFlag()
	ClearSaveFlag()
	//是否需要保存
	NeedSave() bool
	//获取根对象
	Root() Entity
	//设置父对象
	SetParent(p Entity)
	//获取父对象
	Parent() Entity
	//删除标志
	SetDeleted(d bool)
	IsDeleted() bool
	//设置对象号
	SetObjId(id ObjectID)
	//获取对象号
	ObjectId() ObjectID
	//设置名字hash
	SetNameHash(v int32)
	//判断两个对象的configid是否相等
	ConfigIdEqual(id string) bool
	NameHash() int32
	//判断名字是否相等
	NameEqual(name string) bool
	//设置容量(-1无限)
	SetCapacity(capacity int32, initcap int32)
	//修改容量
	ChangeCapacity(capacity int32) error
	//获取容量
	Caps() int32
	//获取实际的容量
	RealCaps() int32
	//子对象数量
	ChildCount() int
	//获取所有的子对象
	AllChilds() []Entity
	//获取在父对象中的索引
	ChildIndex() int
	//设置索引(由引擎自己设置，不要手动设置)
	SetIndex(idx int)
	//清除所有子对象
	ClearChilds()
	//增加一个子对象
	AddChild(idx int, e Entity) (index int, err error)
	//删除一个子对象
	RemoveChild(e Entity) error
	//通过索引获取一个子对象
	GetChild(idx int) Entity
	//通过配置ID获取一个子对象
	FindChildByConfigId(id string) Entity
	//通过配置ID获取第一个子对象
	FindFirstChildByConfigId(id string) (int, Entity)
	//通过配置ID获取从start开始的下一个子对象
	NextChildByConfigId(start int, id string) (int, Entity)
	//获取名字获取子对象
	FindChildByName(name string) Entity
	//通过名字获取第一个子对象
	FindFirstChildByName(name string) (int, Entity)
	//通过名字获取从start开始的下一下子对象
	NextChildByName(start int, name string) (int, Entity)
	//交换两个子对象的位置
	SwapChild(src int, dest int) error
	//设置data
	SetExtraData(key string, value interface{})
	//获取data
	FindExtraData(key string) interface{}
	//获取所有data
	ExtraDatas() map[string]interface{}
	//通过key移除data
	RemoveExtraData(key string)
	//移除所有data
	ClearExtraData()
	//对象类型枚举
	ObjType() int
	//对象类型字符串
	ObjTypeName() string
	//父类型(暂时未用)
	Base() Entity
	//数据库ID
	DBId() uint64
	//设置数据库ID(!!!不要手动设置)
	SetDBId(id uint64)
	//是否保存
	IsSave() bool
	SetSave(s bool)
	//获取所有属性名
	Propertys() []string
	//获取所有可视属性名
	VisiblePropertys(typ int) []string
	//获取所有属性类型
	PropertyType(p string) (int, string, error)
	//获取属性索引
	PropertyIndex(p string) (int, error)
	//属性自增
	Inc(p string, v interface{}) error
	//设置属性(通用接口)
	Set(p string, v interface{}) error
	//通过属性索引设置
	SetByIndex(index int16, v interface{}) error
	//通过属性名获取属性不抛出异常(在确定属性存在的情况下使用)
	MustGet(p string) interface{}
	//通过属性名获取属性
	Get(p string) (val interface{}, err error)
	//属性是否别人可见(同步到别人的客户端)
	PropertyIsPublic(p string) bool
	//属性是否自己可见(同步到自己的客户端)
	PropertyIsPrivate(p string) bool
	//属性是否保存
	PropertyIsSave(p string) bool
	//获取所有脏数据(保存用)
	Dirtys() map[string]interface{}
	//清除脏标志
	ClearDirty()
	//获取所有被修改的属性(同步用)
	Modifys() map[string]interface{}
	//清除所有修改标志
	ClearModify()
	//通过表格名获取表格
	FindRec(rec string) Record
	//获取所有表格的名字
	RecordNames() []string
	//清空对象所有数据
	Reset()
	//复制另一个对象数据
	Copy(other Entity) error
	//DB
	SyncToDb()
	//获取保存对象的配置ID
	GetConfigFromDb(data interface{}) string
	//从数据库加载
	SyncFromDb(data interface{}) bool
	//获取数据库操作接口
	SaveLoader() DBSaveLoader
	//序列化
	Serial() ([]byte, error)
	//序列化变动数据
	SerialModify() ([]byte, error)
	//是否是场景数据(跟随玩家进入场景的数据)
	IsSceneData(prop string) bool
	//从scenedata同步
	SyncFromSceneData(val interface{}) error
	//获取scenedata
	SceneData() interface{}
}

//注册函数
func Register(name string, createfunc func() Entity) {
	if _, dup := objects[name]; dup {
		panic("entity: Register called twice for object " + name)
	}
	log.LogMessage("register entity:", name)
	objects[name] = createfunc
}

//创建数据对象
func Create(name string) Entity {
	if create, exist := objects[name]; exist {
		return create()
	}

	return nil
}
