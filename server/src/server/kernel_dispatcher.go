package server

import (
	"container/list"
)

var dispatcherList = list.New()

type Dispatcher interface {
	OnBeatRun()
	OnBeginUpdate()
	OnUpdate()
	OnLastUpdate()
	OnFrame()
	OnFlush()
}

//调度组件
type Dispatch struct {
}

func (d *Dispatch) OnBeatRun() {
}

func (d *Dispatch) OnBeginUpdate() {
}

func (d *Dispatch) OnUpdate() {
}

func (d *Dispatch) OnLastUpdate() {
}

func (d *Dispatch) OnFrame() {
}

func (d *Dispatch) OnFlush() {
}

//挂载组件
func (k *Kernel) AddDispatch(d Dispatcher) bool {
	for ele := dispatcherList.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(Dispatcher) == d {
			return false
		}
	}

	dispatcherList.PushBack(d)
	return true
}

func (k *Kernel) RemoveDispatch(d Dispatcher) bool {
	for ele := dispatcherList.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(Dispatcher) == d {
			dispatcherList.Remove(ele)
			return true
		}
	}
	return false
}
