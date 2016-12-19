package server

import "server/libs/log"

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
	k.schedulerid++
	s.SetSchedulerID(k.schedulerid)
	k.scheduler[k.schedulerid] = s
	log.LogDebug("add scheduler:", k.schedulerid)
}

func (k *Kernel) GetScheduler(id int32) Scheduler {
	if s, exist := k.scheduler[id]; exist {
		return s
	}
	return nil
}

func (k *Kernel) RemoveScheduler(s Scheduler) {
	if s == nil {
		return
	}
	if _, exist := k.scheduler[s.GetSchedulerID()]; exist {
		delete(k.scheduler, s.GetSchedulerID())
		log.LogDebug("remove scheduler:", s.GetSchedulerID(), " total:", len(k.scheduler))
		s.SetSchedulerID(-1)
	}
}

func (k *Kernel) RemoveSchedulerById(id int32) {
	if _, exist := k.scheduler[id]; exist {
		delete(k.scheduler, id)
		log.LogDebug("remove scheduler:", id, " total:", len(k.scheduler))
	}
}

func (k *Kernel) OnUpdate() {
	//更新调度器
	for _, s := range k.scheduler {
		s.OnUpdate()
	}
}
