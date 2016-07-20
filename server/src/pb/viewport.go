package pb

import (
	"pb/s2c"
	. "server/data/datatype"

	"github.com/golang/protobuf/proto"
)

type PBViewportCodec struct {
}

func (vt *PBViewportCodec) GetCodecInfo() string {
	return "use protobuf"
}

//容器创建
func (vt *PBViewportCodec) ViewportCreate(id int32, container Entityer) interface{} {
	msg := &s2c.CreateView{}
	msg.Entity = proto.String(container.ObjTypeName())
	msg.ViewId = proto.Int32(id)
	msg.Capacity = proto.Int32(container.GetCapacity())
	return msg
}

//删除容器
func (vp *PBViewportCodec) ViewportDelete(id int32) interface{} {
	msg := &s2c.DeleteView{}
	msg.ViewId = proto.Int32(id)
	return msg
}

//容器里增加对象
func (vp *PBViewportCodec) ViewportNotifyAdd(id int32, index int32, object Entityer) interface{} {
	msg := &s2c.ViewAdd{}
	msg.ViewId = proto.Int32(id)
	msg.Entity = proto.String(object.ObjTypeName())
	msg.Index = proto.Int32(index)
	msg.Props, _ = object.Serial()
	return msg
}

//容器里移除对象
func (vp *PBViewportCodec) ViewportNotifyRemove(id int32, index int32) interface{} {
	msg := &s2c.ViewRemove{}
	msg.ViewId = proto.Int32(id)
	msg.Index = proto.Int32(index)
	return msg
}

//容器交换位置
func (vp *PBViewportCodec) ViewportNotifyExchange(srcid int32, src int32, destid int32, dest int32) interface{} {
	msg := &s2c.ViewExchange{}
	msg.ViewId1 = proto.Int32(srcid)
	msg.Index1 = proto.Int32(src)
	msg.ViewId2 = proto.Int32(destid)
	msg.Index2 = proto.Int32(dest)
	return msg
}

func (vp *PBViewportCodec) OnUpdate(id int32, child Entityer) interface{} {

	data, _ := child.SerialModify()
	if data == nil {
		return nil
	}

	msg := &s2c.ViewobjProperty{}
	msg.ViewId = proto.Int32(id)
	msg.Index = proto.Int32(int32(child.GetIndex()))
	msg.Props = data
	return msg
}
