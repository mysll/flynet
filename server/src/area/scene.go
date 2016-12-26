package area

import (
	"server"
	. "server/data/datatype"
	"server/libs/log"
)

type Scene struct {
	server.Callee
}

func (s *Scene) OnAfterAdd(self Entity, sender Entity, index int) int {
	log.LogMessage("scene add obj", sender.ObjectId(), index)
	self.FindExtraData("cell").(*cell).AddObject(sender)
	return 1
}

func (s *Scene) OnRemove(self Entity, sender Entity, index int) int {
	log.LogMessage("scene remove obj", sender.ObjectId())
	//解除所有的心跳
	App.Kernel().DeatchBeat(sender)
	self.FindExtraData("cell").(*cell).RemoveObject(sender)

	return 1
}

func (s *Scene) OnDestroy(self Entity, sender Entity) int {
	log.LogMessage("scene destroy", sender)
	return 1
}
