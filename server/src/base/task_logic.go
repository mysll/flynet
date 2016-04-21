package base

import (
	"data/entity"
	"libs/log"
	"libs/rpc"
	"pb/c2s"
	"server"
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

func (t *TaskLogic) Submit(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	args := &c2s.Reqreceivetask{}
	if err := server.ProtoParse(msg, args); err != nil {
		log.LogError(err)
		return nil
	}
	p := App.Players.GetPlayer(mailbox.Id)
	if p == nil || p.Entity == nil {
		log.LogError("player not found")
		return nil
	}
	player := p.Entity.(*entity.Player)

	taskid := args.GetTaskid()

	server.Check(App.tasksystem.Submit(player, taskid))
	return nil
}
