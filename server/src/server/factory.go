package server

import (
	"container/list"
	. "data/datatype"
	"errors"
	"libs/log"
	"time"
)

//对象工厂
type Factory struct {
	index   int32
	pool    *Pool
	objects map[int32]Entityer
	deletes *list.List
}

//创建一个对象
func (f *Factory) Create(typ string) (ent Entityer, err error) {
	id := ObjectID{}
	for i := 0; i < 100; i++ {
		f.index++
		if _, ok := f.objects[f.index]; ok {
			continue
		}
		id.Index = f.index
		id.Serial = int32(time.Now().UTC().UnixNano())
		break
	}
	if id.Index == 0 {
		err = errors.New("Create entity id error")
		return
	}
	ent = f.pool.Create(typ)
	if ent == nil {
		err = errors.New("create entity error:" + typ)
		return
	}
	ent.SetObjId(id)
	ent.SetDeleted(false)
	f.objects[id.Index] = ent
	return
}

//通过id获取一个对象
func (f *Factory) Find(id ObjectID) Entityer {
	if e, ok := f.objects[id.Index]; ok {
		if e.GetDeleted() || !e.GetObjId().Equal(id) {
			return nil
		}
		return e
	}

	return nil
}

//销毁一个对象
func (f *Factory) Destroy(id ObjectID) {
	if e, ok := f.objects[id.Index]; ok {
		if e.GetDeleted() || !e.GetObjId().Equal(id) {
			return
		}
		f.destroyObj(e)
	}
}

func (f *Factory) destroyObj(obj Entityer) {
	//查找是否有子对象，一并删除
	chs := obj.GetChilds()
	for _, c := range chs {
		if c != nil {
			f.destroyObj(c)
		}
	}
	parent := obj.GetParent()
	if parent != nil {
		parent.RemoveChild(obj)
	}
	obj.ClearChilds()
	obj.SetDeleted(true)
	f.deletes.PushBack(obj.GetObjId().Index)
}

func (f *Factory) destroySelf(obj Entityer) {
	obj.SetDeleted(true)
	f.deletes.PushBack(obj.GetObjId().Index)
}

func (f *Factory) realDestroy(id int32) {
	if k, ok := f.objects[id]; ok {
		f.pool.Free(k)
		delete(f.objects, id)
	}
}

//清理删除的对象
func (f *Factory) ClearDelete() {
	delcount := 0
	var next *list.Element
	for ele := f.deletes.Front(); ele != nil; ele = next {
		next = ele.Next()
		f.realDestroy(ele.Value.(int32))
		f.deletes.Remove(ele)
		delcount++
	}
	if delcount > 0 {
		log.LogDebug("deleted objects:", delcount, ",remain objects:", len(f.objects))
	}
}

//创建一个新的工厂
func NewFactory() *Factory {
	f := &Factory{}
	f.pool = NewEntityPool()
	f.objects = make(map[int32]Entityer)
	f.deletes = list.New()
	return f
}
