package server

import (
	"fmt"
	. "server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
)

var (
	vt ViewportCodec
)

type ViewportCodec interface {
	GetCodecInfo() string
	ViewportCreate(id int32, container Entity) interface{}
	ViewportDelete(id int32) interface{}
	ViewportNotifyAdd(id int32, index int32, object Entity) interface{}
	ViewportNotifyRemove(id int32, index int32) interface{}
	ViewportNotifyExchange(srcid int32, src int32, destid int32, dest int32) interface{}
	OnUpdate(id int32, child Entity) interface{}
}

type ViewportData struct {
	Id        int32
	Container Entity
}

type Viewport struct {
	SchedulerBase
	Views   map[int32]*ViewportData
	Owner   Entity
	mailbox rpc.Mailbox
}

func NewViewport(p Entity, mailbox rpc.Mailbox) *Viewport {
	if vt == nil {
		panic("viewport transport not set")
	}
	log.LogMessage("viewport proto:", vt.GetCodecInfo())
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

func (vp *Viewport) AddViewport(id int32, container Entity) error {
	if _, exist := vp.Views[id]; exist {
		return fmt.Errorf("viewport is already open")
	}

	if container == nil {
		return fmt.Errorf("container is nil")
	}

	if container.Caps() == -1 {
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
	msg := vt.ViewportCreate(id, vd.Container)
	if msg == nil {
		return
	}
	MailTo(nil, &vp.mailbox, "Viewport.Create", msg)
	childs := vd.Container.Caps()
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
	msg := vt.ViewportDelete(id)
	if msg == nil {
		return
	}
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
	msg := vt.ViewportNotifyAdd(id, index, child)
	if msg == nil {
		return
	}
	MailTo(nil, &vp.mailbox, "Viewport.Add", msg)
}

//容器里移除对象
func (vp *Viewport) ViewportNotifyRemove(id int32, index int32) {
	vd := vp.Views[id]
	if vd == nil {
		return
	}
	msg := vt.ViewportNotifyRemove(id, index)
	if msg == nil {
		return
	}
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

	msg := vt.ViewportNotifyExchange(srcid, src, destid, dest)
	if msg == nil {
		return
	}
	MailTo(nil, &vp.mailbox, "Viewport.Exchange", msg)
}

func (vp *Viewport) OnUpdate() {
	for _, vd := range vp.Views {
		childs := vd.Container.AllChilds()
		for _, child := range childs {
			if child == nil {
				continue
			}
			data, _ := child.SerialModify()
			if data == nil {
				continue
			}

			msg := vt.OnUpdate(vd.Id, child)
			if msg == nil {
				continue
			}
			MailTo(nil, &vp.mailbox, "Viewport.ObjUpdate", msg)
			child.ClearModify()
		}
	}
}

func RegisterViewportCodec(t ViewportCodec) {
	vt = t
}
