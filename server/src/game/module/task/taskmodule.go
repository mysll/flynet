package task

import (
	"errors"
	"server"
)

var Module *TaskModule

type TaskModule struct {
	server.BaseModule
	TaskSystem *TaskSystem
}

func NewTaskModule() *TaskModule {
	tm := &TaskModule{}
	return tm
}

func (m *TaskModule) Init() {
	m.TaskSystem = NewTaskSystem()
	server.RegisterCalleePriority("Player", m.TaskSystem, server.PRIORITY_LOWEST)
	server.RegisterCalleePriority("Container", m.TaskSystem, server.PRIORITY_LOWEST)
	server.RegisterHandler("Task", NewTaskLogic())

	server.RegisterCallee("Player", &PlayerTask{})
}

func (m *TaskModule) Load() error {
	if m.TaskSystem.LoadTaskInfo() {
		return nil
	}

	return errors.New("load task config failed")
}

func init() {
	Module = NewTaskModule()
}
