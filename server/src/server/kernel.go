package server

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	. "server/data/datatype"
	"server/data/helper"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
)

// Kernel是一个核心API集合，所有游戏性相关的操作都封装在这里
// 主要功能：
// 		对象管理
//		异步IO处理
//		全局数据管理
//		场景心跳处理
//		对象跨服务切换
//		通用定时器处理
//		对象的事件回调管理
type Kernel struct {
	factory   *Factory
	uidSerial uint64
}

var (
	ErrObjNotFound      = errors.New("object not found")
	ErrCalleeAlreadyReg = errors.New("callee already registed")
	ErrContainerCantAdd = errors.New("container can't add")
)

func NewKernel(factory *Factory) *Kernel {
	k := &Kernel{}
	k.factory = factory
	return k
}

//当前是否是Base
func (k *Kernel) CurrentInBase(base bool) {
	k.factory.inBase = base
}

//转储某个对象信息，保存进文件
func (k *Kernel) DumpInfo(obj interface{}, fname string) {
	data, err := json.Marshal(obj)
	if err != nil {
		log.LogError("dump info err:", err)
		return
	}

	f, err := os.Create(fname)
	if err != nil {
		log.LogError("dump info file create failed,", fname)
		return
	}

	f.Write(data)
	f.Close()
}

//通过id获取entity
func (k *Kernel) GetEntity(obj ObjectID) Entityer {
	return k.factory.Find(obj)
}

//通过类型创建对象,将回调Callee.OnCreate。
func (k *Kernel) Create(typ string) (ent Entityer, err error) {
	ent, err = k.CreateContainer(typ, -1)
	return
}

//创建角色，将触发role.OnCreateRole回调，在OnCreateRole回调中，可以对这个角色的数据进行初始化处理
//其中args为创建角色的参数，由客户端提供，其中要包含角色的名称以及创建角色的索引位置，这个参数直接回调给OnCreateRole处理
func (k *Kernel) CreateRole(typ string, args interface{}) (ent Entityer, err error) {
	ent, err = k.factory.Create(typ)
	if err != nil {
		return
	}
	ent.SetPropHooker(k)
	ent.SetCapacity(-1, 16)
	callee := GetCallee("role")
	res := 0
	for _, cl := range callee {
		res = cl.OnCreateRole(ent, args)
		if res == -1 {
			k.factory.Destroy(ent.GetObjId())
			err = errors.New("create role failed")
			return
		}
		if res == 0 {
			break
		}
	}
	k.PreSave(ent, true)
	return
}

//创建一个容器,将回调Callee.OnCreate。
func (k *Kernel) CreateContainer(typ string, caps int) (ent Entityer, err error) {
	ent, err = k.factory.Create(typ)
	if err != nil {
		return
	}
	ent.SetPropHooker(k)
	if caps > 0 {
		ent.SetCapacity(int32(caps), int32(caps))
	} else {
		ent.SetCapacity(-1, 16)
	}

	callee := GetCallee(typ)
	res := 0
	for _, cl := range callee {
		res = cl.OnCreate(ent, nil)
		if res == 0 {
			break
		}
	}
	return
}

//创建一个子对象,将回调Callee.OnCreate。
func (k *Kernel) CreateChild(parent ObjectID, typ string, index int) (ent Entityer, err error) {
	ent, err = k.CreateChildContainer(parent, typ, -1, index)
	return
}

//创建一个子容器,将回调Callee.OnCreate。
func (k *Kernel) CreateChildContainer(parent ObjectID, typ string, caps int, index int) (ent Entityer, err error) {
	p := k.factory.Find(parent)
	if p == nil {
		err = errors.New("parent not found")
		return
	}

	ent, err = k.factory.Create(typ)
	if err != nil {
		return
	}
	ent.SetPropHooker(k)
	if caps > 0 {
		ent.SetCapacity(int32(caps), int32(caps))
	} else {
		ent.SetCapacity(-1, 16)
	}

	callee := GetCallee(typ)
	res := 0
	for _, cl := range callee {
		res = cl.OnCreate(ent, p)
		if res == 0 {
			break
		}
	}
	_, err = k.addChild(p, ent, index)
	if err != nil {
		k.factory.Destroy(ent.GetObjId())
	}
	return
}

//增加一个子对象，处理流程：
//		1.首先回调Callee.OnTestAdd,如果逻辑返回值为0，则你对象不允许增加，流程结束，插入失败
//		2.如果子对象是由另一个容器移动过来的，则会先触发父对象的删除当前对象操作
//		3.父对象将回调Callee.OnAdd,OnAfterAdd,当前子对象将回调Callee.OnEntry
//		4.如果当前你对象是一个视图，则会自动同步操作到客户端
func (k *Kernel) AddChild(parent ObjectID, child ObjectID, index int) (destindex int, err error) {
	p := k.factory.Find(parent)
	c := k.factory.Find(child)
	if p == nil {
		return -1, errors.New("parent not found")
	}
	if c == nil {
		return -1, errors.New("object not found")
	}
	return k.addChild(p, c, index)
}

func (k *Kernel) addChild(parent Entityer, child Entityer, index int) (destindex int, err error) {

	calleeparent := GetCallee(parent.ObjTypeName())
	calleeself := GetCallee(child.ObjTypeName())

	for _, cl := range calleeparent {
		if cl.OnTestAdd(parent, child, index) == 0 {
			err = ErrContainerCantAdd
			return
		}
	}

	if child.GetParent() != nil {
		k.removeChild(child.GetParent(), child, true)
	}

	res := 0
	for _, cl := range calleeparent {
		res, destindex = cl.OnAdd(parent, child, index)
		if res&8 != 0 {
			break
		}
	}

	if res&4 != 0 {
		err = errors.New("add child failed")
		return
	} else if res&2 != 0 {
		newchild := parent.GetChild(destindex)
		if newchild == nil || child.ObjTypeName() != newchild.ObjTypeName() {
			err = errors.New("combine child failed")
			return
		}
		k.factory.Destroy(child.GetObjId())
		child = newchild
	} else {
		destindex, err = parent.AddChild(index, child)
		if err == nil {
			if viewid := parent.GetExtraData("viewportid"); viewid != nil {
				root := parent.GetRoot()
				vp := k.FindViewport(root)
				if vp != nil {
					vp.ViewportNotifyAdd(viewid.(int32), int32(destindex))
				}
			}
		}
	}

	if err != nil {
		return
	}

	for _, cl := range calleeself {
		res = cl.OnEntry(child, parent)
		if res == 0 {
			break
		}
	}

	for _, cl := range calleeparent {
		res = cl.OnAfterAdd(parent, child, destindex)
		if res == 0 {
			break
		}
	}

	return
}

//交换子对象的位置
func (k *Kernel) Exchange(src Entityer, dest Entityer) bool {
	if src == nil || dest == nil ||
		src.GetParent() == nil ||
		dest.GetParent() == nil ||
		!src.GetParent().GetObjId().Equal(dest.GetParent().GetObjId()) {
		log.LogError("parent not equal")
		return false
	}

	parent := src.GetParent()
	err := parent.SwapChild(src.GetIndex(), dest.GetIndex())
	if err != nil {
		log.LogError(err)
		return false
	}

	if viewid := parent.GetExtraData("viewportid"); viewid != nil {
		root := parent.GetRoot()
		vp := k.FindViewport(root)
		if vp != nil {
			vp.ViewportNotifyExchange(viewid.(int32), int32(dest.GetIndex()), viewid.(int32), int32(src.GetIndex()))
		}
	}

	return true
}

//移除一个子对象，父对象将回调Callee.OnBeforeRemove,Callee.OnRemove。子对象将回调Callee.OnLeave
func (k *Kernel) RemoveChild(parent Entityer, child Entityer) error {
	if parent != nil && child != nil && !child.GetParent().GetObjId().Equal(parent.GetObjId()) {
		return errors.New("parent not equal")
	}

	if parent != nil && child != nil {
		return k.removeChild(parent, child, true)
	}

	return ErrChildObjectNotFound
}

func (k *Kernel) removeChild(parent Entityer, child Entityer, needcallback bool) error {
	cself := GetCallee(child.ObjTypeName())

	res := 0

	index := child.GetIndex()
	if c := parent.GetChild(index); c != nil {
		if !c.GetObjId().Equal(child.GetObjId()) {
			return ErrChildObjectNotFound
		}
	} else {
		return ErrChildObjectNotFound
	}
	cparent := GetCallee(parent.ObjTypeName())
	if needcallback {
		for _, cl := range cparent {
			res = cl.OnBeforeRemove(parent, child)
			if res == 0 {
				break
			}
		}
		for _, cl := range cself {
			res = cl.OnLeave(child, parent)
			if res == 0 {
				break
			}
		}
	}

	parent.RemoveChild(child)
	if needcallback {
		for _, cl := range cparent {
			res = cl.OnRemove(parent, child, index)
			if res == 0 {
				break
			}
		}
	}

	if viewid := parent.GetExtraData("viewportid"); viewid != nil {
		root := parent.GetRoot()
		vp := k.FindViewport(root)
		if vp != nil {
			vp.ViewportNotifyRemove(viewid.(int32), int32(index))
		}
	}

	return nil
}

func (k *Kernel) loadArchive(data *EntityInfo) (ent Entityer, err error) {

	ent, err = k.CreateContainer(data.Type, int(data.Caps))
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(data.Data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(ent)
	if err != nil {
		k.destroyObj(ent, false)
		ent = nil
		return
	}

	/*
		if data.BaseData != nil && ent.Base() != nil {
			err = k.loadBaseArchive(ent.Base(), data.BaseData)
			if err != nil {
				k.destroyObj(ent, false)
				ent = nil
				return
			}
		}*/

	ent.SetExtraData("linkObj", data.ObjId)
	ent.SetExtraData("isLoad", true)
	var index int
	for _, c := range data.Childs {
		if c != nil {
			child, e := k.loadArchive(c)
			if e != nil {
				err = e
				k.destroyObj(ent, false)
				ent = nil
				return
			}
			index = c.Index
			if ent.GetCapacity() == -1 {
				index = -1
			}
			_, err = k.addChild(ent, child, index)
			if err != nil {
				log.LogError("add failed:", ent.GetCapacity(), ",", index)
				k.destroyObj(child, false)
				k.destroyObj(ent, false)
				ent = nil
				return
			}
		}
	}
	ent.RemoveExtraData("isLoad")
	return
}

//通过数据创建对象,对象将回调Callee.OnLoad
func (k *Kernel) CreateFromArchive(data *EntityInfo, extra map[string]interface{}) (ent Entityer, err error) {
	ent, err = k.loadArchive(data)
	if err != nil {
		return
	}
	if extra != nil {
		for k, v := range extra {
			ent.SetExtraData(k, v)
		}
	}

	callee := GetCallee(ent.ObjTypeName())
	res := 0
	for _, cl := range callee {
		res = cl.OnLoad(ent, share.LOAD_ARCHIVE)
		if res == 0 {
			break
		}
	}
	return
}

func (k *Kernel) loadObj(parent Entityer, data *share.SaveEntity) (ent Entityer, err error) {
	ent, err = k.factory.Create(data.Typ)
	if err != nil {
		log.LogError("load object failed:", err)
		return
	}
	ent.SetPropHooker(k)
	if parent == nil {
		ent.SetLoading(true)
	}

	callee := GetCallee(data.Typ)
	res := 0
	for _, cl := range callee {
		res = cl.OnCreate(ent, parent)
		if res == 0 {
			break
		}
	}

	ent.SetDbId(data.DBId)

	configid := ent.GetConfigFromDb(data.Obj)
	if configid != "" {
		helper.LoadFromConfig(configid, ent)
	}

	ent.SyncFromDb(data.Obj)
	//加载父类
	if ent.Base() != nil {
		err = k.loadBase(ent.Base(), data.Base)
		if err != nil {
			log.LogMessage("load object base failed:", err)
			return
		}
	}

	caps := ent.GetCapacity()
	if caps == -1 {
		ent.SetCapacity(-1, 16)
	} else {
		ent.SetCapacity(caps, caps)
	}

	if parent != nil {
		if parent.GetCapacity() == -1 {
			_, err = k.addChild(parent, ent, -1)
		} else {
			_, err = k.addChild(parent, ent, data.Index)
		}
		if err != nil {
			log.LogError(err)
		}
	}

	//加载子对象
	if data.Childs != nil && len(data.Childs) > 0 {
		for _, c := range data.Childs {
			if c != nil {
				_, err = k.loadObj(ent, c)
				if err != nil {
					log.LogError("load object child err:", err)
				}
			}
		}

	}

	if parent == nil {
		ent.SetLoading(false)
	}

	ent.ClearDirty()
	ent.ClearModify()
	return
}

func (k *Kernel) loadBase(object Entityer, data *share.SaveEntity) error {
	if data == nil {
		return errors.New("base data is nil")
	}
	object.SyncFromDb(data.Obj)
	if object.Base() != nil {
		err := k.loadBase(object.Base(), data.Base)
		if err != nil {
			return err
		}
	}
	return nil
}

//通过存档创建,对象将回调Callee.OnCreate,Callee.OnLoad
func (k *Kernel) CreateFromDb(data *share.DbSave) (ent Entityer, err error) {
	if data == nil || data.Data == nil {
		err = errors.New("data is nil")
	}
	ent, err = k.loadObj(nil, data.Data)
	if err != nil {
		return
	}
	res := 0
	callee := GetCallee(ent.ObjTypeName())
	for _, cl := range callee {
		res = cl.OnLoad(ent, share.LOAD_DB)
		if res == 0 {
			break
		}
	}
	return
}

//通过配置文件加载,对象将回调Callee.OnLoad
func (k *Kernel) LoadFromConfig(object Entityer, configid string) error {
	err := helper.LoadFromConfig(configid, object)
	res := 0
	callee := GetCallee(object.ObjTypeName())
	if err == nil {
		for _, cl := range callee {
			res = cl.OnLoad(object, share.LOAD_CONFIG)
			if res == 0 {
				break
			}
		}
		object.SetConfig(configid)
	}
	return err
}

//通过配置文件创建对象Callee.OnCreate,Callee.OnLoad
func (k *Kernel) CreateFromConfig(configid string) (ent Entityer, err error) {
	typ, err := helper.GetEntityByConfig(configid)
	if err != nil {
		return nil, err
	}

	ent, err = k.factory.Create(typ)
	if err != nil {
		return
	}
	ent.SetPropHooker(k)
	res := 0
	callee := GetCallee(typ)

	for _, cl := range callee {
		res = cl.OnCreate(ent, nil)
		if res == 0 {
			break
		}
	}

	err = k.LoadFromConfig(ent, configid)
	return ent, err
}

//销毁一个对象,将收到回调Callee.OnDestroy
func (k *Kernel) Destroy(obj ObjectID) (err error) {
	object := k.factory.Find(obj)
	if object == nil {
		err = ErrObjNotFound
		return
	}

	k.destroyObj(object, true)
	return
}

func (k *Kernel) destroyObj(object Entityer, needcallback bool) {
	if !object.IsSave() {
		return
	}

	//从视图中移除
	if viewid := object.GetExtraData("viewportid"); viewid != nil {
		root := object.GetRoot()
		if root != nil {
			vp := k.FindViewport(root)
			if vp != nil {
				vp.RemoveViewport(viewid.(int32))
			}
		}
	}

	chs := object.GetChilds()
	for _, c := range chs {
		if c != nil {
			k.destroyObj(c, needcallback)
		}
	}

	object.ClearChilds()

	parent := object.GetParent()
	if parent != nil {
		k.removeChild(parent, object, needcallback)
	}

	callee := GetCallee(object.ObjTypeName())
	res := 0
	if needcallback {
		for _, cl := range callee {
			res = cl.OnDestroy(object, parent)
			if res == 0 {
				break
			}
		}
	}

	//删除视图
	if data := object.GetExtraData("viewportlinkid"); data != nil {
		sid := data.(int32)
		k.RemoveSchedulerById(sid)
		object.RemoveExtraData("viewportlinkid")
	}

	//清除所有的心跳
	k.RemoveObjectHeartbeat(object.GetObjId())
	k.factory.destroySelf(object)
}

type Transform struct {
	Scene string
	Pos   Vector3
	Dir   float32
}

//设置地图登录点
func (k *Kernel) SetLandpos(object Entityer, trans Transform) {
	object.SetExtraData("landpos", trans)
}

//获取地图登录点
func (k *Kernel) GetLandpos(object Entityer) Transform {
	if trans := object.GetExtraData("landpos"); trans != nil {
		if tr, ok := trans.(Transform); ok {
			return tr
		}
	}
	return Transform{}
}

//设置角色信息,对应数据库中的roleinfo
func (k *Kernel) SetRoleInfo(object Entityer, info string) {
	object.SetExtraData("roleinfo", info)
}

//预保存，设置唯一ID
func (k *Kernel) PreSave(object Entityer, ignore bool) {
	if object.GetDbId() == 0 && !ignore && object.IsSave() {
		object.SetDbId(k.GetUid())
	}

	cds := object.GetChilds()
	for _, c := range cds {
		if c == nil {
			continue
		}
		k.PreSave(c, false)
	}
}

//对象保存，将回调Callee.OnStore
func (k *Kernel) Save(object Entityer, typ int) (err error) {
	if object == nil {
		err = ErrObjNotFound
		return
	}
	k.PreSave(object, true)
	callee := GetCallee(object.ObjTypeName())
	res := 0
	for _, cl := range callee {
		res = cl.OnStore(object, typ)
		if res == 0 {
			break
		}
	}

	return
}

func (k *Kernel) setSerial(serial uint64) {
	k.uidSerial = serial
}

//获取一个唯一序列号
func (k *Kernel) GetSerial() uint64 {
	return k.uidSerial
}

//获取一个UID
func (k *Kernel) GetUid() uint64 {
	k.uidSerial++
	return k.uidSerial
}

//玩家断开，收到Callee.OnDisconnect
func (k *Kernel) Disconnect(object Entityer) {
	if object == nil {
		return
	}
	callee := GetCallee(object.ObjTypeName())
	res := 0
	for _, cl := range callee {
		res = cl.OnDisconnect(object)
		if res == 0 {
			break
		}
	}
}

//玩家进入场景前,收到Callee.OnEnterScene
func (k *Kernel) EntryScene(object Entityer) {
	if object == nil {
		return
	}
	callee := GetCallee(object.ObjTypeName())
	res := 0
	for _, cl := range callee {
		res = cl.OnEnterScene(object)
		if res == 0 {
			break
		}
	}
}

//玩家进入场景,收到Callee.OnEnterScene
func (k *Kernel) EnterScene(object Entityer) {
	if object == nil {
		return
	}
	callee := GetCallee(object.ObjTypeName())
	res := 0
	for _, cl := range callee {
		res = cl.OnEnterScene(object)
		if res == 0 {
			break
		}
	}
}

type Motioner interface {
	SetPos(pos Vector3)
	SetOrient(dir float32)
}

//场景里放置一个对象
func (k *Kernel) PlaceObj(scene Entityer, object Entityer, pos Vector3, orient float32) bool {
	if scene == nil || object == nil {
		return false
	}

	if mover, ok := object.(Motioner); ok {
		mover.SetPos(pos)
		mover.SetOrient(orient)
	}

	_, err := k.addChild(scene, object, -1)
	if err != nil {
		return false
	}
	return true
}

//sender给target发送消息，接收方回调Callee.OnCommand
func (k *Kernel) Command(src, dest ObjectID, msgid int, msg interface{}) bool {
	sender := k.factory.Find(src)
	target := k.factory.Find(dest)
	if target == nil {
		return false
	}

	callee := GetCallee(target.ObjTypeName())
	var res int
	for _, cl := range callee {
		res = cl.OnCommand(target, sender, msgid, msg)
		if res == 0 {
			break
		}
	}

	return true
}

//self对target使用item.物品对象回调Callee.OnUseTo
func (k *Kernel) UseTo(self, target, item ObjectID) bool {
	sender := k.factory.Find(self)
	if sender == nil {
		return false
	}
	other := k.factory.Find(target)
	if other == nil {
		return false
	}
	object := k.factory.Find(item)
	if object == nil || object.ObjType() != ITEM {
		return false
	}

	callee := GetCallee(object.ObjTypeName())
	var res int
	for _, cl := range callee {
		res = cl.OnUseTo(object, sender, other)
		if res == 0 {
			break
		}
	}
	return true
}

//self使用item，物品对象回调Callee.OnUse
func (k *Kernel) Use(self, item ObjectID) bool {
	sender := k.factory.Find(self)
	if sender == nil {
		return false
	}
	object := k.factory.Find(item)
	if object == nil || object.ObjType() != ITEM {
		return false
	}

	callee := GetCallee(object.ObjTypeName())
	var res int
	for _, cl := range callee {
		res = cl.OnUse(object, sender)
		if res == 0 {
			break
		}
	}
	return true
}

//self装备equip，物品对象回调Callee.OnEquip
func (k *Kernel) Equip(self, equip ObjectID, idx int) bool {
	sender := k.factory.Find(self)
	if sender == nil {
		return false
	}
	object := k.factory.Find(equip)
	if object == nil || object.ObjType() != ITEM {
		return false
	}

	callee := GetCallee(object.ObjTypeName())
	var res int
	for _, cl := range callee {
		res = cl.OnEquip(object, sender, idx)
		if res == 0 {
			break
		}
	}
	return true
}

//self卸下equip，物品对象回调Callee.OnUnEquip
func (k *Kernel) UnEquip(self, equip ObjectID, idx int) bool {
	sender := k.factory.Find(self)
	if sender == nil {
		return false
	}
	object := k.factory.Find(equip)
	if object == nil || object.ObjType() != ITEM {
		return false
	}
	callee := GetCallee(object.ObjTypeName())
	var res int
	for _, cl := range callee {
		res = cl.OnUnEquip(object, sender, idx)
		if res == 0 {
			break
		}
	}
	return true
}

//对象属性变动，对象回调Callee.OnPropertyChange
func (k *Kernel) OnPropChange(object Entityer, prop string, value interface{}) {
	callee := GetCallee(object.ObjTypeName())
	var res int
	for _, cl := range callee {
		res = cl.OnPropertyChange(object, prop, value)
		if res == 0 {
			break
		}
	}
}

//玩家已经就绪,玩家对象将回调Callee.OnReady
func (k *Kernel) PlayerReady(player Entityer, first bool) {
	if player.ObjType() != PLAYER {
		return
	}
	callee := GetCallee(player.ObjTypeName())
	var res int
	for _, cl := range callee {
		res = cl.OnReady(player, first)
		if res == 0 {
			break
		}
	}
}

func (k *Kernel) SetPropertyEx(object Entityer, prop string, val string, opt int) error {
	old, err := object.Get(prop)
	if err != nil {
		return err
	}

	opval := old

	err = ParseStrNumberEx(val, &opval, old, opt)
	if err != nil {
		return err
	}

	return object.Set(prop, opval)
}

func (k *Kernel) clearProperty(object Entityer, prop string, val string, opt int) error {
	return nil
}

func (k *Kernel) CallScript(object Entityer, id string, prop string, revert bool) error {
	defer func() {
		if err := recover(); err != nil {
			log.LogError(err)
		}
	}()

	ops := helper.GetPropOpt(id, prop)
	if ops != nil {
		if revert {
			for _, op := range ops {
				k.clearProperty(object, op.Prop, op.Value, op.Option)
			}
		} else {
			for _, op := range ops {
				k.SetPropertyEx(object, op.Prop, op.Value, op.Option)
			}
		}

	}
	return nil
}

//查找一个对象的视图
func (k *Kernel) FindViewport(player Entityer) *Viewport {
	if data := player.GetExtraData("viewportlinkid"); data != nil {
		sid := data.(int32)
		scheduler := k.GetScheduler(sid)
		if scheduler != nil {
			switch inst := scheduler.(type) {
			case *Viewport:
				return inst
			}
		}
	}
	return nil
}

//Player增加一个视图
func (k *Kernel) AddViewport(player Entityer, mailbox rpc.Mailbox, viewid int32, container Entityer) error {
	if player == nil || player.ObjType() != PLAYER || container == nil {
		return fmt.Errorf("type is illegality")
	}
	if vp := k.FindViewport(player); vp != nil {
		return vp.AddViewport(viewid, container)
	}

	vp := NewViewport(player, mailbox)
	k.AddScheduler(vp)
	player.SetExtraData("viewportlinkid", vp.GetSchedulerID())
	return vp.AddViewport(viewid, container)
}

//Player删除一个视图
func (k *Kernel) DeleteViewport(player Entityer, viewid int32) {
	if player == nil {
		return
	}

	if vp := k.FindViewport(player); vp != nil {
		vp.RemoveViewport(viewid)
	}
}

//和client进行绑定
func (k *Kernel) AttachPlayer(player Entityer, mailbox rpc.Mailbox) error {
	if player.ObjType() != PLAYER {
		return fmt.Errorf("object is not player")
	}

	create := core.rpcProto.CreateObjectMessage(player, true, mailbox)
	err := MailTo(nil, &mailbox, "Create", create)
	if err != nil {
		log.LogError(err)
		return err
	}

	if player.GetPropSyncer() != nil {
		k.RemoveScheduler(player.GetPropSyncer().(*PropSync))
	}
	propsync := NewPropSync(mailbox, player.GetObjId())
	player.SetPropSyncer(propsync)
	recs := player.GetRecNames()
	tablesyncer := NewTableSync(mailbox)
	for _, r := range recs {
		rec := player.GetRec(r)
		if rec.IsVisible() {
			rec.SetSyncer(tablesyncer)
		}
	}
	tablesyncer.SyncTable(player)

	k.AddScheduler(propsync)
	return nil
}

//和client解绑
func (k *Kernel) DetachPlayer(player Entityer) {
	if player.GetPropSyncer() != nil {
		k.RemoveScheduler(player.GetPropSyncer().(*PropSync))
	}
	player.SetPropSyncer(nil)
	recs := player.GetRecNames()
	for _, r := range recs {
		rec := player.GetRec(r)
		rec.SetSyncer(nil)
	}
}

//获取玩家管理器
func (k *Kernel) GetPlayers() *PlayerManager {
	return Players
}
