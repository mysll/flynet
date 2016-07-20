package server

import (
	"container/list"
	. "server/data/datatype"
	"server/libs/log"
	"time"
)

type objectTick struct {
	serial   int64
	Name     string
	Oid      ObjectID
	ent      Entityer
	Params   interface{}
	Last     time.Time
	Interval time.Duration
	Count    int32
	deleted  bool
	next     *objectTick
}

type BeatTickSave struct {
	Name     string
	Interval time.Duration
	Last     time.Time
	Count    int32
	Args     interface{}
}

type beatSlot struct {
	info *list.List
}

func (slot *beatSlot) Add(tick *objectTick) {
	slot.info.PushBack(tick)
}

func (slot *beatSlot) Run(t time.Time) {
	count := slot.info.Len()
	if count == 0 {
		return
	}

	//存储当前队尾，用作循环结束的条件。因为执行过程中，执行过的会重新放到队尾。
	last := slot.info.Back().Value.(*objectTick).serial
	var next *list.Element
	for e := slot.info.Front(); e != nil; e = next {
		next = e.Next()
		tick := e.Value.(*objectTick)
		if !tick.deleted {
			if dur := t.Sub(tick.Last); dur >= tick.Interval {
				tick.Last = t
				if tick.Count > 0 {
					tick.Count--
					if tick.Count == 0 {
						tick.deleted = true
					}
				}

				if tick.ent == nil || tick.ent.GetDeleted() {
					tick.deleted = true
				} else {
					callee := GetCallee(tick.ent.ObjTypeName())
					for _, cl := range callee {
						res := cl.OnTimer(tick.ent, tick.Count, tick.Params)
						if res == 0 {
							break
						}
					}
				}

				slot.info.Remove(e)
				slot.info.PushBack(tick)

			} else { //如果前面时间不满足，后面的就不用处理了
				break
			}
		}
		if last == tick.serial {
			break
		}
	}
}

type beatIndex struct {
	t    time.Duration
	elem *objectTick
}

type SceneBeat struct {
	freeHeart  *objectTick
	heartbeats map[int64]*beatSlot
	beatHash   map[int32]map[string]*beatIndex
	serial     int64
}

func (this *SceneBeat) getTick() *objectTick {
	var tick *objectTick
	if this.freeHeart != nil {
		tick = this.freeHeart
		this.freeHeart = this.freeHeart.next
	} else {
		tick = new(objectTick)
	}
	*tick = objectTick{}
	return tick
}

func (this *SceneBeat) freeTick(tick *objectTick) {
	tick.Params = nil
	tick.ent = nil
	tick.next = this.freeHeart
	this.freeHeart = tick
}

func (this *SceneBeat) find(oid ObjectID, beat string) *beatIndex {
	if mbi, ok := this.beatHash[oid.Index]; ok {
		if mb, ok := mbi[beat]; ok {
			return mb
		}
	}

	return nil
}

func (this *SceneBeat) Pump() {
	t := time.Now()
	for _, bs := range this.heartbeats {
		bs.Run(t)
	}

	//清理需要删除的心跳
	for t, bs := range this.heartbeats {
		var next *list.Element
		for e := bs.info.Front(); e != nil; e = next {
			next = e.Next()
			tick := e.Value.(*objectTick)
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

func (this *SceneBeat) deleteTick(tick *objectTick) {
	delete(this.beatHash[tick.Oid.Index], tick.Name)
	if len(this.beatHash[tick.Oid.Index]) == 0 {
		delete(this.beatHash, tick.Oid.Index)
	}
	this.freeTick(tick)
}

func (this *SceneBeat) ResetCount(oid ObjectID, beat string, count int32) bool {
	l := this.find(oid, beat)
	if l == nil || l.elem.deleted {
		return false
	}

	l.elem.Count = count
	return true
}

func (this *SceneBeat) Find(oid ObjectID, beat string) bool {
	if this.find(oid, beat) != nil {
		return true
	}

	return false
}

func (this *SceneBeat) Add(obj Entityer, beat string, t time.Duration, count int32, param interface{}) bool {
	oid := obj.GetObjId()
	if this.find(oid, beat) != nil {
		log.LogError("heartbeat already add", beat)
		return false
	}

	if t < time.Millisecond {
		log.LogError("heartbeat duration must above 1 millisecond,", t)
		return false
	}

	if count == 0 {
		log.LogError("heartbeat count must above 0 or -1")
		return false
	}

	tick := this.getTick()
	this.serial++
	tick.serial = this.serial
	tick.Name = beat
	tick.Oid = oid
	tick.ent = obj
	tick.Params = param
	tick.Last = time.Now()
	tick.Interval = t
	tick.Count = count
	tick.deleted = false

	if hb, ok := this.heartbeats[int64(t)]; ok {
		hb.Add(tick)
	} else {
		hb = &beatSlot{}
		hb.info = list.New()
		hb.Add(tick)
		this.heartbeats[int64(t)] = hb
	}

	if beats, ok := this.beatHash[oid.Index]; ok {
		beats[beat] = &beatIndex{t, tick}
	} else {
		b := make(map[string]*beatIndex, 10)
		b[beat] = &beatIndex{t, tick}
		this.beatHash[oid.Index] = b
	}

	return true

}

func (this *SceneBeat) Remove(oid ObjectID, beat string) bool {
	bi := this.find(oid, beat)
	if bi != nil {
		bi.elem.deleted = true
		return true
	}

	return false
}

func (this *SceneBeat) RemoveObjectBeat(oid ObjectID) {
	if beats, ok := this.beatHash[oid.Index]; ok {
		if len(beats) == 0 {
			return
		}
		for _, v := range beats {
			if v.elem.deleted {
				continue
			}

			v.elem.deleted = true
		}
	}
}

func (this *SceneBeat) Deatch(object Entityer) bool {
	if object == nil {
		return false
	}

	id := object.GetObjId()
	if beats, ok := this.beatHash[id.Index]; ok {
		if len(beats) == 0 {
			return true
		}
		savebeat := make([]BeatTickSave, 0, len(beats))
		for _, v := range beats {
			if v.elem.deleted {
				continue
			}
			beat := BeatTickSave{}
			beat.Name = v.elem.Name
			beat.Last = v.elem.Last
			beat.Interval = v.elem.Interval
			beat.Count = v.elem.Count
			beat.Args = v.elem.Params
			v.elem.deleted = true
			savebeat = append(savebeat, beat)
		}
		object.SetExtraData("saveBeats", savebeat)
		return true
	}

	object.SetExtraData("saveBeats", nil)
	return true
}

func NewSceneBeat() *SceneBeat {
	beat := &SceneBeat{}
	beat.heartbeats = make(map[int64]*beatSlot, 100)
	beat.beatHash = make(map[int32]map[string]*beatIndex, 1000)
	beat.serial = 0
	return beat
}
