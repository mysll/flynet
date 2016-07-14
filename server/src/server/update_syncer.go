package server

import (
	. "data/datatype"
	"libs/log"
	"libs/rpc"
	"pb/s2c"
	"util"

	"github.com/golang/protobuf/proto"
)

type PropSync struct {
	SchedulerBase
	mailbox rpc.Mailbox
	objid   ObjectID
}

func NewPropSync(mb rpc.Mailbox, objid ObjectID) *PropSync {
	ts := &PropSync{}
	ts.mailbox = mb
	ts.objid = objid
	return ts
}

func (ps *PropSync) OnUpdate() {
	player := core.GetEntity(ps.objid)
	if player == nil {
		core.RemoveScheduler(ps)
		return
	}
	ps.UpdateAll(player)
}

func (ps *PropSync) UpdateAll(player Entityer) error {
	data, _ := player.SerialModify()
	if data == nil {
		return nil
	}

	update := &s2c.UpdateProperty{}
	update.Self = proto.Bool(true)
	update.Index = proto.Int32(0)
	update.Serial = proto.Int32(0)
	update.Propinfo = data
	err := MailTo(nil, &ps.mailbox, "Entity.Update", update)
	if err != nil {
		log.LogError(err)
		return err
	}
	player.ClearModify()
	return nil
}

func (ps *PropSync) Update(index int16, value interface{}) {
	update := &s2c.UpdateProperty{}
	update.Self = proto.Bool(true)
	update.Index = proto.Int32(0)
	update.Serial = proto.Int32(0)
	ar := util.NewStoreArchiver(nil)
	ar.Write(index)
	ar.Write(value)
	update.Propinfo = ar.Data()
	err := MailTo(nil, &ps.mailbox, "Entity.Update", update)
	if err != nil {
		log.LogError(err)
	}
}
