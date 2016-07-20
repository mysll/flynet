package area

import (
	"logicdata/entity"
	"server"
	"server/libs/log"
)

type Scene struct {
	server.Callee
}

func (s *Scene) OnAfterAdd(self entity.Entityer, sender entity.Entityer, index int) int {
	log.LogMessage("scene add obj", sender.GetObjId(), index)
	self.GetExtraData("cell").(*cell).AddObject(sender)
	return 1
}

func (s *Scene) OnRemove(self entity.Entityer, sender entity.Entityer) int {
	log.LogMessage("scene remove obj", sender.GetObjId())
	//解除所有的心跳
	App.DeatchBeat(sender)
	self.GetExtraData("cell").(*cell).RemoveObject(sender)

	return 1
}

func (s *Scene) OnDestroy(self entity.Entityer, sender entity.Entityer) int {
	log.LogMessage("scene destroy", sender)
	return 1
}
