package server

import (
	. "server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
)

var pt PropCodec

type PropCodec interface {
	GetCodecInfo() string
	UpdateAll(object Entity, self bool) interface{}
	Update(index int16, value interface{}, self bool, objid ObjectID) interface{}
}

type PropSync struct {
	SchedulerBase
	mailbox rpc.Mailbox
	objid   ObjectID
}

func NewPropSync(mb rpc.Mailbox, objid ObjectID) *PropSync {
	if pt == nil {
		panic("prop transport not set")
	}

	log.LogMessage("prop sync proto:", pt.GetCodecInfo())
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

func (ps *PropSync) UpdateAll(player Entity) error {
	data, _ := player.SerialModify()
	if data == nil {
		return nil
	}

	update := pt.UpdateAll(player, true)
	if update == nil {
		return nil
	}
	err := MailTo(nil, &ps.mailbox, "Entity.Update", update)
	if err != nil {
		log.LogError(err)
		return err
	}
	player.ClearModify()
	return nil
}

func (ps *PropSync) Update(self Entity, index int16, value interface{}) {
	update := pt.Update(index, value, true, ps.objid)
	if update == nil {
		return
	}
	err := MailTo(nil, &ps.mailbox, "Entity.Update", update)
	if err != nil {
		log.LogError(err)
	}
}

func RegisterPropCodec(t PropCodec) {
	pt = t
}
