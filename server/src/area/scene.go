package area

import (
	"server"
	. "server/data/datatype"
	"server/libs/log"
)

type Scene struct {
	server.Callee
}

func (s *Scene) OnAfterAdd(self Entityer, sender Entityer, index int) int {
	log.LogMessage("scene add obj", sender.GetObjId(), index)
	self.GetExtraData("cell").(*cell).AddObject(sender)
	return 1
}

func (s *Scene) OnRemove(self Entityer, sender Entityer, index int) int {
	log.LogMessage("scene remove obj", sender.GetObjId())
	//解除所有的心跳
	App.DeatchBeat(sender)
	self.GetExtraData("cell").(*cell).RemoveObject(sender)

	return 1
}

func (s *Scene) OnDestroy(self Entityer, sender Entityer) int {
	log.LogMessage("scene destroy", sender)
	return 1
}
