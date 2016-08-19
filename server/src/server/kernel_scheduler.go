package server

import "server/libs/log"

var (
	scheduler   = make(map[int32]Scheduler, 1024)
	schedulerid int32
)

type Scheduler interface {
	SetSchedulerID(id int32)
	GetSchedulerID() int32
	OnUpdate()
}

type SchedulerBase struct {
	id int32
}

func (sb *SchedulerBase) SetSchedulerID(id int32) {
	sb.id = id
}

func (sb *SchedulerBase) GetSchedulerID() int32 {
	return sb.id
}

func (k *Kernel) AddScheduler(s Scheduler) {
	if s == nil {
		return
	}
	schedulerid++
	s.SetSchedulerID(schedulerid)
	scheduler[schedulerid] = s
	log.LogDebug("add scheduler:", schedulerid)
}

func (k *Kernel) GetScheduler(id int32) Scheduler {
	if s, exist := scheduler[id]; exist {
		return s
	}
	return nil
}

func (k *Kernel) RemoveScheduler(s Scheduler) {
	if s == nil {
		return
	}
	if _, exist := scheduler[s.GetSchedulerID()]; exist {
		delete(scheduler, s.GetSchedulerID())
		log.LogDebug("remove scheduler:", s.GetSchedulerID(), " total:", len(scheduler))
		s.SetSchedulerID(-1)
	}
}

func (k *Kernel) RemoveSchedulerById(id int32) {
	if _, exist := scheduler[id]; exist {
		delete(scheduler, id)
		log.LogDebug("remove scheduler:", id, " total:", len(scheduler))
	}
}

func (k *Kernel) OnUpdate() {
	//更新调度器
	for _, s := range scheduler {
		s.OnUpdate()
	}
}
