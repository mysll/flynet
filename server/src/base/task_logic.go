package base

import (
	"data/entity"
	"libs/log"
	"libs/rpc"
	"pb/c2s"
)

type TaskLogic struct {
}

func NewTaskLogic() *TaskLogic {
	logic := &TaskLogic{}
	return logic
}

func (t *TaskLogic) Submit(mailbox rpc.Mailbox, args c2s.Reqreceivetask) error {

	p := App.Players.GetPlayer(mailbox.Id)
	if p == nil || p.Entity == nil {
		log.LogError("player not found")
		return nil
	}
	player := p.Entity.(*entity.Player)

	taskid := args.GetTaskid()

	err := App.tasksystem.Submit(player, taskid)

	return err
}
