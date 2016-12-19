package task

import (
	"logicdata/entity"
	"pb/c2s"
	"server"
	"server/libs/log"
	"server/libs/rpc"
)

type TaskLogic struct {
}

func NewTaskLogic() *TaskLogic {
	logic := &TaskLogic{}
	return logic
}

func (t *TaskLogic) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("Submit", t.Submit)
}

func (t *TaskLogic) Submit(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	args := &c2s.Reqreceivetask{}
	if server.Check(server.ParseProto(msg, args)) {
		return 0, nil
	}
	p := Module.GetCore().Players.FindPlayer(mailbox.Uid)
	if p == nil || p.GetEntity() == nil {
		log.LogError("player not found")
		return 0, nil
	}
	player := p.GetEntity().(*entity.Player)

	taskid := args.GetTaskid()

	server.Check(Module.TaskSystem.Submit(player, taskid))
	return 0, nil
}
