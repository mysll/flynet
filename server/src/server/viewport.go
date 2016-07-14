package server

import (
	. "data/datatype"
	"fmt"
	"libs/rpc"
	"pb/s2c"

	"github.com/golang/protobuf/proto"
)

type ViewportData struct {
	Id        int32
	Container Entityer
}

type Viewport struct {
	SchedulerBase
	Views   map[int32]*ViewportData
	Owner   Entityer
	mailbox rpc.Mailbox
}

func NewViewport(p Entityer, mailbox rpc.Mailbox) *Viewport {
	vp := &Viewport{}
	vp.Views = make(map[int32]*ViewportData)
	vp.Owner = p
	vp.mailbox = mailbox
	return vp
}

func (vp *Viewport) ClearAll() {
	for id, _ := range vp.Views {
		vp.RemoveViewport(id)
	}
}

func (vp *Viewport) AddViewport(id int32, container Entityer) error {
	if _, exist := vp.Views[id]; exist {
		return fmt.Errorf("viewport is already open")
	}

	if container == nil {
		return fmt.Errorf("container is nil")
	}

	if container.GetCapacity() == -1 {
		return fmt.Errorf("container capacity not set")
	}

	container.SetExtraData("viewportid", id)
	vp.Views[id] = &ViewportData{id, container}
	vp.ViewportCreate(id)
	return nil
}

func (vp *Viewport) RemoveViewport(id int32) {
	if vd, exist := vp.Views[id]; exist {
		vp.ViewportDelete(id)
		vd.Container.RemoveExtraData("viewportid")
		delete(vp.Views, id)
	}
}

//容器创建
func (vp *Viewport) ViewportCreate(id int32) {
	vd := vp.Views[id]
	if vd == nil {
		return
	}
	msg := &s2c.CreateView{}
	msg.Entity = proto.String(vd.Container.ObjTypeName())
	msg.ViewId = proto.Int32(id)
	msg.Capacity = proto.Int32(vd.Container.GetCapacity())
	MailTo(nil, &vp.mailbox, "Viewport.Create", msg)
	childs := vd.Container.GetCapacity()
	for index := int32(0); index < childs; index++ {
		vp.ViewportNotifyAdd(id, index)
	}
}

//删除容器
func (vp *Viewport) ViewportDelete(id int32) {
	vd := vp.Views[id]
	if vd == nil {
		return
	}
	msg := &s2c.DeleteView{}
	msg.ViewId = proto.Int32(id)
	MailTo(nil, &vp.mailbox, "Viewport.Delete", msg)
}

//容器里增加对象
func (vp *Viewport) ViewportNotifyAdd(id int32, index int32) {
	vd := vp.Views[id]
	if vd == nil {
		return
	}

	child := vd.Container.GetChild(int(index))
	if child == nil {
		return
	}
	msg := &s2c.ViewAdd{}
	msg.ViewId = proto.Int32(id)
	msg.Entity = proto.String(child.ObjTypeName())
	msg.Index = proto.Int32(index)
	msg.Props, _ = child.Serial()
	MailTo(nil, &vp.mailbox, "Viewport.Add", msg)
}

//容器里移除对象
func (vp *Viewport) ViewportNotifyRemove(id int32, index int32) {
	vd := vp.Views[id]
	if vd == nil {
		return
	}
	msg := &s2c.ViewRemove{}
	msg.ViewId = proto.Int32(id)
	msg.Index = proto.Int32(index)
	MailTo(nil, &vp.mailbox, "Viewport.Remove", msg)
}

//容器交换位置
func (vp *Viewport) ViewportNotifyExchange(srcid int32, src int32, destid int32, dest int32) {
	vd1 := vp.Views[srcid]
	if vd1 == nil {
		return
	}

	vd2 := vp.Views[destid]
	if vd2 == nil {
		return
	}

	msg := &s2c.ViewExchange{}
	msg.ViewId1 = proto.Int32(srcid)
	msg.Index1 = proto.Int32(src)
	msg.ViewId2 = proto.Int32(destid)
	msg.Index2 = proto.Int32(dest)
	MailTo(nil, &vp.mailbox, "Viewport.Exchange", msg)
}

func (vp *Viewport) OnUpdate() {
	for _, vd := range vp.Views {
		childs := vd.Container.GetChilds()
		for _, child := range childs {
			if child == nil {
				continue
			}
			data, _ := child.SerialModify()
			if data == nil {
				continue
			}

			msg := &s2c.ViewobjProperty{}
			msg.ViewId = proto.Int32(vd.Id)
			msg.Index = proto.Int32(int32(child.GetIndex()))
			msg.Props = data
			MailTo(nil, &vp.mailbox, "Viewport.ObjUpdate", msg)
			child.ClearModify()
		}
	}
}
