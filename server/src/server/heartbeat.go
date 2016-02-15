package server

import (
	"libs/log"
	"time"
	"util/hash"
)

type Tick struct {
	Name     string
	NameHash int32
	Callback HeartBearter
	Params   interface{}
	Last     time.Time
	Interval time.Duration
	Count    int32
	next     *Tick
}

type HeartBearter func(t time.Duration, count int32, args interface{})

type Heartbeat struct {
	freeHeart  *Tick
	heartbeats *Tick
	checktime  time.Time
}

//心跳循环处理
func (h *Heartbeat) Pump() {

	now := time.Now()
	it := h.heartbeats
	for it != nil {
		if dur := now.Sub(it.Last); it.Count != 0 && dur >= it.Interval {
			if it.Count > 0 {
				it.Count--
			}
			h.checktime = time.Now()
			it.Callback(dur, it.Count, it.Params)
			if delay := time.Now().Sub(h.checktime); delay > warninglvl {
				log.LogWarning("heartbeat call ", it.NameHash, " delay:", delay.Seconds())
			}
			it.Last = now
		}
		it = it.next
	}

	it = h.heartbeats
	var parent *Tick
	for it != nil {
		if it.Count == 0 {
			next := it.next
			h.delTick(parent, it)
			it = next
			continue
		}

		parent = it
		it = it.next
	}
}

//查找心跳是否存在
func (h *Heartbeat) Find(heartbeat string) bool {
	namehash := hash.DJBHash(heartbeat)
	it := h.heartbeats
	for it != nil {
		if it.NameHash == namehash && it.Name == heartbeat {
			return true
		}
		it = it.next
	}

	return false
}

//重设一个心跳的次数
func (h *Heartbeat) SetHeartbeatCount(heartbeat string, count int32) bool {
	namehash := hash.DJBHash(heartbeat)
	it := h.heartbeats
	for it != nil {
		if it.NameHash == namehash && it.Name == heartbeat {
			it.Count = count
			return true
		}
		it = it.next
	}
	return false
}

//重设一个心跳的间隔
func (h *Heartbeat) ResetHeartbeatInterval(heartbeat string, t time.Duration) bool {
	if t < time.Millisecond {
		log.LogError("heartbeat duration must above 1 millisecond,", t)
		return false
	}

	namehash := hash.DJBHash(heartbeat)
	it := h.heartbeats
	for it != nil {
		if it.NameHash == namehash && it.Name == heartbeat {
			it.Interval = t
			return true
		}
		it = it.next
	}

	return false
}

//增加一个心跳，heartbeat是心跳的名称，t时间间隔，count为回调的次数，-1则表示无限，callback回调的函数，params是回调函数的参数
func (h *Heartbeat) AddHeartbeat(heartbeat string, t time.Duration, count int32, callback HeartBearter, params interface{}) bool {
	if t < time.Millisecond {
		log.LogError("heartbeat duration must above 1 millisecond,", t)
		return false
	}
	if h.Find(heartbeat) {
		log.LogError("heartbeat already add", heartbeat)
		return false
	}
	tick := h.getTick()
	tick.Name = heartbeat
	tick.NameHash = hash.DJBHash(heartbeat)
	tick.Interval = t
	tick.Count = count
	tick.Last = time.Now()
	tick.next = h.heartbeats
	tick.Callback = callback
	tick.Params = params
	h.heartbeats = tick

	return true
}

//移除一个名为heartbeat的心跳函数
func (h *Heartbeat) RemoveHeartbeat(heartbeat string) bool {
	namehash := hash.DJBHash(heartbeat)
	it := h.heartbeats
	for it != nil {
		if it.NameHash == namehash && it.Name == heartbeat {
			it.Count = 0
			return true
		}
		it = it.next
	}

	return false
}

//移除所有的心跳
func (h *Heartbeat) RemoveAllHeartbeat() {
	it := h.heartbeats
	for it != nil {
		it.Count = 0
		it = it.next
	}
}

func (h *Heartbeat) getTick() *Tick {
	tick := h.freeHeart
	if tick == nil {
		tick = new(Tick)
	} else {
		h.freeHeart = tick.next
		*tick = Tick{}
	}

	return tick
}

func (h *Heartbeat) delTick(parent *Tick, t *Tick) {
	if parent == nil {
		h.heartbeats = t.next
	} else {
		parent.next = t.next
	}
	log.LogDebug("heartbeat is deled:", t.Name)
	t.Callback = nil
	t.Params = nil
	t.next = h.freeHeart
	h.freeHeart = t
}

//创建一个新的心跳模块
func NewHeartbeat() *Heartbeat {
	return &Heartbeat{}
}
