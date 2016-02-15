package server

import (
	"container/list"
	"reflect"
	"sync"
)

type callbacker struct {
	ptr uintptr
	l   func(interface{})
}

//事件监听模块
type EventListener struct {
	sync.RWMutex
	listener map[string]*list.List
}

//增加一个事件，提供事件名和回调函数
func (el *EventListener) AddEventListener(event string, l func(interface{})) {
	el.Lock()
	defer el.Unlock()
	if ls, ok := el.listener[event]; ok {
		ls.PushBack(callbacker{reflect.ValueOf(l).Pointer(), l})
	} else {
		lst := list.New()
		lst.PushBack(callbacker{reflect.ValueOf(l).Pointer(), l})
		el.listener[event] = lst
	}
}

//移除一个事件
func (el *EventListener) RemoveEventListener(event string, l func(interface{})) {
	el.Lock()
	defer el.Unlock()

	del := reflect.ValueOf(l).Pointer()
	if ls, ok := el.listener[event]; ok {
		var next *list.Element
		for e := ls.Front(); e != nil; e = next {
			next = e.Next()
			if cb, ok := e.Value.(callbacker); ok && cb.ptr == del {
				ls.Remove(e)
			}
		}
	}
}

//分发事件
func (el *EventListener) DispatchEvent(event string, params interface{}) {
	if ls, ok := el.listener[event]; ok {
		for e := ls.Front(); e != nil; e = e.Next() {
			if cb, ok := e.Value.(callbacker); ok {
				cb.l(params)
			}
		}
	}
}

//创建一个新的事件监听
func NewEvent() *EventListener {
	el := &EventListener{}
	el.listener = make(map[string]*list.List)
	return el
}
