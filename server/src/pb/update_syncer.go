package pb

import (
	"pb/s2c"
	. "server/data/datatype"
	"server/util"

	"github.com/golang/protobuf/proto"
)

type PBPropCodec struct {
}

func (ps *PBPropCodec) GetCodecInfo() string {
	return "use protobuf"
}

func (ps *PBPropCodec) UpdateAll(object Entity, self bool) interface{} {
	data, _ := object.SerialModify()
	if data == nil {
		return nil
	}

	objid := object.ObjectId()
	update := &s2c.UpdateProperty{}
	update.Self = proto.Bool(self)
	update.Index = proto.Int32(objid.Index)
	update.Serial = proto.Int32(objid.Serial)
	update.Propinfo = data
	return update
}

func (ps *PBPropCodec) Update(index int16, value interface{}, self bool, objid ObjectID) interface{} {
	update := &s2c.UpdateProperty{}
	update.Self = proto.Bool(self)
	update.Index = proto.Int32(objid.Index)
	update.Serial = proto.Int32(objid.Serial)
	ar := util.NewStoreArchiver(nil)
	ar.Write(index)
	ar.Write(value)
	update.Propinfo = ar.Data()
	return update
}
