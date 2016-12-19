package task

import (
	"logicdata/entity"
	"server"
	"server/data/datatype"
	"time"
)

type PlayerTask struct {
	server.Callee
}

func (pl *PlayerTask) OnReady(self datatype.Entity, first bool) int {
	if first {
		Module.GetCore().AddHeartbeat(self, "CheckTask", time.Minute, -1, nil)
		Module.TaskSystem.CheckTaskInfo(self.(*entity.Player))
	}

	return 1
}

func (pl *PlayerTask) OnTimer(self datatype.Entity, beat string, count int32, args interface{}) int {
	switch beat {
	case "CheckTask":
		//清理过期的邮件
		Module.TaskSystem.OnUpdate(self.(*entity.Player))
		return 0
	}
	return 1
}

func (pl *PlayerTask) OnEvent(self datatype.Entity, event string, args interface{}) int {
	switch event {
	case "newday":
		Module.TaskSystem.NewDay(self.(*entity.Player)) //新一天
	}
	return 1
}
