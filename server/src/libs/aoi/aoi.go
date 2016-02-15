package aoi

import (
	. "data/datatype"
	"libs/common/event"
)

type AOIer interface {
	Clear()
	GetEvent() *event.EventList
	GetIdsByPos(pos Vector3, ranges int) []ObjectID
	GetIdsByRange(pos Vector3, ranges int, types []int) []ObjectID
	AddObject(pos Vector3, obj ObjectID, typ int) bool
	RemoveObject(pos Vector3, obj ObjectID, typ int) bool
	UpdateObject(obj ObjectID, typ int, oldpos Vector3, newpos Vector3) error
	GetWatchers(pos Vector3, types []int) []ObjectID
	AddWatcher(watcher ObjectID, typ int, pos Vector3, ranges int) bool
	RemoveWatcher(watcher ObjectID, typ int, pos Vector3, ranges int) bool
	UpdateWatcher(watcher ObjectID, typ int, oldPos Vector3, newPos Vector3, oldRange, newRange int) bool
}

type AOI struct {
	aoi AOIer
}

func (this *AOI) GetEvent() *event.EventList {
	return this.aoi.GetEvent()
}

func (this *AOI) GetIdsByPos(pos Vector3, ranges int) []ObjectID {
	return this.aoi.GetIdsByPos(pos, ranges)
}

func (this *AOI) GetIdsByRange(pos Vector3, ranges int, types []int) []ObjectID {
	return this.aoi.GetIdsByRange(pos, ranges, types)
}

func (this *AOI) AddObject(pos Vector3, obj ObjectID, typ int) bool {
	return this.aoi.AddObject(pos, obj, typ)
}

func (this *AOI) RemoveObject(pos Vector3, obj ObjectID, typ int) bool {
	return this.aoi.RemoveObject(pos, obj, typ)
}

func (this *AOI) UpdateObject(obj ObjectID, typ int, oldpos Vector3, newpos Vector3) error {
	return this.aoi.UpdateObject(obj, typ, oldpos, newpos)
}

func (this *AOI) GetWatchers(pos Vector3, types []int) []ObjectID {
	return this.aoi.GetWatchers(pos, types)
}

func (this *AOI) AddWatcher(watcher ObjectID, typ int, pos Vector3, ranges int) bool {
	return this.aoi.AddWatcher(watcher, typ, pos, ranges)
}

func (this *AOI) RemoveWatcher(watcher ObjectID, typ int, pos Vector3, ranges int) bool {
	return this.aoi.RemoveWatcher(watcher, typ, pos, ranges)
}

func (this *AOI) UpdateWatcher(watcher ObjectID, typ int, oldPos Vector3, newPos Vector3, oldRange, newRange int) bool {
	return this.aoi.UpdateWatcher(watcher, typ, oldPos, newPos, oldRange, newRange)
}

func (this *AOI) Clear() {

}

func NewAOI(inst AOIer) *AOI {
	aoi := &AOI{}
	aoi.aoi = inst
	return aoi
}
