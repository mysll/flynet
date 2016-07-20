// 定时器组件,采用时间轮的方法
// 需要注意的地方:
//     如果一个定时器是一直重复的,在不需要的时候,需要把定时器取消掉,不然会造成"泄漏".
// 对于有次数限制的定时器,不需要做这个操作.但是可以中途取消.
package server

import (
	"container/list"
	"math"
	"server/libs/log"
	"time"
)

type TimerCB func(intervalid TimerID, count int32, args interface{})

type TimerID int64
type timerTick struct {
	IntervalId TimerID
	Params     interface{}
	Last       time.Time
	Interval   time.Duration
	Count      int32
	deleted    bool
	cb         TimerCB
	next       *timerTick
}

type tickSlot struct {
	info *list.List
}

func (slot *tickSlot) Add(tick *timerTick) {
	slot.info.PushBack(tick)
}

func (slot *tickSlot) Run(t time.Time) {
	count := slot.info.Len()
	if count == 0 {
		return
	}
	//存储当前队尾，用作循环结束的条件。因为执行过程中，执行过的会重新放到队尾。
	last := slot.info.Back().Value.(*timerTick).IntervalId
	var next *list.Element
	for e := slot.info.Front(); e != nil; e = next {
		next = e.Next()
		tick := e.Value.(*timerTick)
		if !tick.deleted {
			if dur := t.Sub(tick.Last); dur >= tick.Interval {
				tick.Last = t
				if tick.Count > 0 {
					tick.Count--
					if tick.Count == 0 {
						tick.deleted = true
					}
				}
				tick.cb(tick.IntervalId, tick.Count, tick.Params)
				slot.info.Remove(e)
				slot.info.PushBack(tick)

			} else { //如果前面时间不满足，后面的就不用处理了
				break
			}
		}
		if last == tick.IntervalId {
			break
		}
	}
}

type tickIndex struct {
	t    time.Duration
	elem *timerTick
}

type Timer struct {
	freeHeart  *timerTick
	heartbeats map[int64]*tickSlot
	beatHash   map[TimerID]*tickIndex
	serial     TimerID
}

func (this *Timer) getTick() *timerTick {
	var tick *timerTick
	if this.freeHeart != nil {
		tick = this.freeHeart
		this.freeHeart = this.freeHeart.next
	} else {
		tick = new(timerTick)
	}
	*tick = timerTick{}
	return tick
}

func (this *Timer) freeTick(tick *timerTick) {
	tick.cb = nil
	tick.Params = nil
	tick.next = this.freeHeart
	this.freeHeart = tick
}

func (this *Timer) find(intervalId TimerID) *tickIndex {
	if mbi, ok := this.beatHash[intervalId]; ok {
		return mbi
	}

	return nil
}

func (this *Timer) Pump() {
	t := time.Now()
	for _, bs := range this.heartbeats {
		bs.Run(t)
	}

	//清理需要删除的心跳
	for t, bs := range this.heartbeats {
		var next *list.Element
		for e := bs.info.Front(); e != nil; e = next {
			next = e.Next()
			tick := e.Value.(*timerTick)
			if tick.deleted {
				bs.info.Remove(e)
				this.deleteTick(tick)
			}
		}
		if bs.info.Len() == 0 {
			delete(this.heartbeats, t)
		}
	}
}

func (this *Timer) deleteTick(tick *timerTick) {
	delete(this.beatHash, tick.IntervalId)
	this.freeTick(tick)
}

func (this *Timer) ResetCount(intervalId TimerID, count int32) bool {
	l := this.find(intervalId)
	if l == nil || l.elem.deleted {
		return false
	}

	l.elem.Count = count
	return true
}

func (this *Timer) Find(intervalId TimerID) bool {
	if this.find(intervalId) != nil {
		return true
	}

	return false
}

func (this *Timer) Timeout(t time.Duration, cb TimerCB, param interface{}) TimerID {
	return this.AddTimer(t, 1, cb, param)
}

func (this *Timer) AddTimer(t time.Duration, count int32, cb TimerCB, param interface{}) TimerID {

	if t < time.Millisecond {
		log.LogError("heartbeat duration must >= 1 millisecond,", t)
		return -1
	}

	if count == 0 {
		log.LogError("heartbeat count must > 0 or -1")
		return -1
	}

	tick := this.getTick()
	this.serial = (this.serial + 1) % math.MaxInt64
	if this.serial < 0 {
		this.serial = 1
	}

	tick.IntervalId = this.serial
	tick.Params = param
	tick.Last = time.Now()
	tick.Interval = t
	tick.Count = count
	tick.deleted = false
	tick.cb = cb
	if hb, ok := this.heartbeats[int64(t)]; ok {
		hb.Add(tick)
	} else {
		hb = &tickSlot{}
		hb.info = list.New()
		hb.Add(tick)
		this.heartbeats[int64(t)] = hb
	}

	if _, ok := this.beatHash[tick.IntervalId]; ok {
		panic("tick uid repeated")
	} else {
		this.beatHash[tick.IntervalId] = &tickIndex{t, tick}
	}

	return tick.IntervalId
}

func (this *Timer) Cancel(intervalId TimerID) bool {
	bi := this.find(intervalId)
	if bi != nil {
		bi.elem.deleted = true
		return true
	}
	return false
}

func NewTimer() *Timer {
	beat := &Timer{}
	beat.heartbeats = make(map[int64]*tickSlot, 100)
	beat.beatHash = make(map[TimerID]*tickIndex, 1000)
	beat.serial = 0
	return beat
}
