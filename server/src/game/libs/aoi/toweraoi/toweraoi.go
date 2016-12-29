package toweraoi

import (
	"errors"
	"fmt"
	"math"
	. "server/data/datatype"
	"server/libs/common/event"
)

type TowerPos struct {
	X, Y int
}

type TowerAOI struct {
	width       float32
	height      float32
	towerWidth  float32
	towerHeight float32
	rangeLimit  int
	max         TowerPos
	towers      [][]*Tower
	Emitter     *event.EventList
}

func (this *TowerAOI) Init() {
	iloop := int(math.Ceil(float64(this.width / this.towerWidth)))
	jloop := int(math.Ceil(float64(this.height / this.towerHeight)))
	this.max.X = iloop - 1
	this.max.Y = jloop - 1
	this.towers = make([][]*Tower, iloop)
	for i := 0; i < iloop; i++ {
		this.towers[i] = make([]*Tower, jloop)
		for j := 0; j < jloop; j++ {
			this.towers[i][j] = NewTower()
		}
	}
}

func (this *TowerAOI) Clear() {
	for i := 0; i <= this.max.X; i++ {
		for j := 0; j <= this.max.Y; j++ {
			this.towers[i][j].clear()
		}
	}
}

func (this *TowerAOI) GetEvent() *event.EventList {
	return this.Emitter
}

func (this *TowerAOI) GetIdsByPos(pos Vector3, ranges int) []ObjectID {
	if !this.checkPos(pos) || ranges < 0 {
		return nil
	}

	result := make([]ObjectID, 0, 100)
	if ranges > this.rangeLimit {
		ranges = this.rangeLimit
	}
	p := this.transPos(pos)
	start, end := getPosLimit(p, ranges, this.max)

	for i := start.X; i <= end.X; i++ {
		for j := start.Y; j <= end.Y; j++ {
			result = append(result, this.towers[i][j].getIds()...)
		}
	}
	return result
}

func (this *TowerAOI) GetIdsByRange(pos Vector3, ranges int, types []int) []ObjectID {
	if !this.checkPos(pos) || ranges < 0 {
		return nil
	}

	result := make([]ObjectID, 0, 100)
	if ranges > this.rangeLimit {
		ranges = this.rangeLimit
	}
	p := this.transPos(pos)
	start, end := getPosLimit(p, ranges, this.max)

	for i := start.X; i <= end.X; i++ {
		for j := start.Y; j <= end.Y; j++ {
			result = append(result, this.towers[i][j].getIdsByTypes(types)...)
		}
	}

	return result
}

func (this *TowerAOI) AddObject(pos Vector3, obj ObjectID, typ int) bool {
	if this.checkPos(pos) {
		p := this.transPos(pos)
		this.towers[p.X][p.Y].add(obj, typ)
		this.Emitter.Push(
			"add", map[string]interface{}{
				"id":       obj,
				"type":     typ,
				"watchers": this.towers[p.X][p.Y].getAllWatchers()},
			false)
		return true
	}
	return false
}

func (this *TowerAOI) RemoveObject(pos Vector3, obj ObjectID, typ int) bool {
	if this.checkPos(pos) {
		p := this.transPos(pos)
		this.towers[p.X][p.Y].remove(obj, typ)
		this.Emitter.Push(
			"remove", map[string]interface{}{
				"id":       obj,
				"type":     typ,
				"watchers": this.towers[p.X][p.Y].getAllWatchers()},
			false)
		return true
	}
	return false
}

func (this *TowerAOI) UpdateObject(obj ObjectID, typ int, oldpos Vector3, newpos Vector3) error {
	if !this.checkPos(oldpos) || !this.checkPos(newpos) {
		return nil
	}
	p1 := this.transPos(oldpos)
	p2 := this.transPos(newpos)

	if p1.X == p2.X && p1.Y == p2.Y {
		return nil
	} else {
		if this.towers[p1.X] == nil || this.towers[p2.X] == nil {
			return errors.New(fmt.Sprintf("AOI pos error ! oldPos : %v, newPos : %v, p1 : %v, p2 : %v", oldpos, newpos, p1, p2))
		}

		oldtower := this.towers[p1.X][p1.Y]
		newtower := this.towers[p2.X][p2.Y]
		oldtower.remove(obj, typ)
		newtower.add(obj, typ)
		this.Emitter.Push(
			"update", map[string]interface{}{
				"id":          obj,
				"type":        typ,
				"oldWatchers": oldtower.getAllWatchers(),
				"newWatchers": newtower.getAllWatchers()},
			false)
	}

	return nil
}

func (this *TowerAOI) GetWatchers(pos Vector3, types []int) []ObjectID {
	if this.checkPos(pos) {
		p := this.transPos(pos)
		return this.towers[p.X][p.Y].getWatchers(types)
	}
	return nil
}

func (this *TowerAOI) AddWatcher(watcher ObjectID, typ int, pos Vector3, ranges int) bool {
	if ranges < 0 {
		return false
	}
	if ranges > this.rangeLimit {
		ranges = this.rangeLimit
	}
	p := this.transPos(pos)
	start, end := getPosLimit(p, ranges, this.max)
	for i := start.X; i <= end.X; i++ {
		for j := start.Y; j <= end.Y; j++ {
			this.towers[i][j].addWatcher(watcher, typ)
		}
	}

	return true
}

func (this *TowerAOI) RemoveWatcher(watcher ObjectID, typ int, pos Vector3, ranges int) bool {
	if ranges < 0 {
		return false
	}

	if ranges > this.rangeLimit {
		ranges = this.rangeLimit
	}

	p := this.transPos(pos)

	start, end := getPosLimit(p, ranges, this.max)

	for i := start.X; i <= end.X; i++ {
		for j := start.Y; j <= end.Y; j++ {
			this.towers[i][j].removeWatcher(watcher, typ)
		}
	}

	return true
}

func (this *TowerAOI) UpdateWatcher(watcher ObjectID, typ int, oldPos Vector3, newPos Vector3, oldRange int, newRange int) bool {
	if !this.checkPos(oldPos) || !this.checkPos(newPos) {
		return false
	}
	p1 := this.transPos(oldPos)
	p2 := this.transPos(newPos)

	if p1.X == p2.X && p1.Y == p2.Y {
		return true
	} else {
		if oldRange < 0 || newRange < 0 {
			return false
		}
		if oldRange > this.rangeLimit {
			oldRange = this.rangeLimit
		}
		if newRange > this.rangeLimit {
			newRange = this.rangeLimit
		}
		addTowers, removeTowers := this.getChangedTowers(p1, p2, oldRange, newRange)
		addObjs := make([]ObjectID, 0, 100)
		removeObjs := make([]ObjectID, 0, 100)
		for _, t := range addTowers {
			t.addWatcher(watcher, typ)
			addObjs = append(addObjs, t.getIds()...)
		}

		for _, t := range removeTowers {
			t.removeWatcher(watcher, typ)
			removeObjs = append(removeObjs, t.getIds()...)
		}

		this.Emitter.Push(
			"updateWatcher", map[string]interface{}{
				"id":         watcher,
				"type":       typ,
				"addObjs":    addObjs,
				"removeObjs": removeObjs},
			false)
	}
	return true
}

/**
 * Get changed towers for girven pos
 * @param p1 {Object} The origin position
 * @param p2 {Object} The now position
 * @param r1 {Number} The old range
 * @param r2 {Number} The new range
 */
func (this *TowerAOI) getChangedTowers(p1 TowerPos, p2 TowerPos, r1 int, r2 int) ([]*Tower, []*Tower) {
	var start1, end1 = getPosLimit(p1, r1, this.max)
	var start2, end2 = getPosLimit(p2, r2, this.max)

	removeTowers := make([]*Tower, 0, 10)
	addTowers := make([]*Tower, 0, 10)

	for i := start1.X; i <= end1.X; i++ {
		for j := start1.Y; j <= end1.Y; j++ {
			if !isInRect(TowerPos{i, j}, start2, end2) {
				removeTowers = append(removeTowers, this.towers[i][j])
			}
		}
	}

	for i := start2.X; i <= end2.X; i++ {
		for j := start2.Y; j <= end2.Y; j++ {
			if !isInRect(TowerPos{i, j}, start1, end1) {
				addTowers = append(addTowers, this.towers[i][j])
			}
		}
	}

	return addTowers, removeTowers

}

/**
 * Check if the pos is valid;
 * @return {Boolean} Test result
 */
func (this *TowerAOI) checkPos(pos Vector3) bool {
	if pos.X < 0 || pos.Z < 0 || pos.X >= this.width || pos.Z >= this.height {
		return false
	}
	return true
}

/**
 * Trans the absolut pos to tower pos. For example : (210, 110} -> (1, 0), for tower width 200, height 200
 *
 */
func (this *TowerAOI) transPos(pos Vector3) TowerPos {
	return TowerPos{
		X: int(math.Floor(float64(pos.X / this.towerWidth))),
		Y: int(math.Floor(float64(pos.Z / this.towerHeight))),
	}
}

/**
 * Get the postion limit of given range
 * @param pos {Object} The center position
 * @param range {Number} The range
 * @param max {max} The limit, the result will not exceed the limit
 * @return The pos limitition
 */
func getPosLimit(pos TowerPos, ranges int, max TowerPos) (start TowerPos, end TowerPos) {

	if pos.X-ranges < 0 {
		start.X = 0
		end.X = 2 * ranges
	} else if pos.X+ranges > max.X {
		end.X = max.X
		start.X = max.X - 2*ranges
	} else {
		start.X = pos.X - ranges
		end.X = pos.X + ranges
	}

	if pos.Y-ranges < 0 {
		start.Y = 0
		end.Y = 2 * ranges
	} else if pos.Y+ranges > max.Y {
		end.Y = max.Y
		start.Y = max.Y - 2*ranges
	} else {
		start.Y = pos.Y - ranges
		end.Y = pos.Y + ranges
	}
	if start.X < 0 {
		start.X = 0
	}
	if end.X > max.X {
		end.X = max.X
	}
	if start.Y < 0 {
		start.Y = 0
	}
	if end.Y > max.Y {
		end.Y = max.Y
	}

	return
}

/**
 * Check if the pos is in the rect
 */
func isInRect(pos TowerPos, start TowerPos, end TowerPos) bool {
	return (pos.X >= start.X && pos.X <= end.X && pos.Y >= start.Y && pos.Y <= end.Y)
}

func NewTowerAOI(w float32, h float32, tw float32, th float32, limit int) *TowerAOI {
	aoi := &TowerAOI{}
	aoi.width = w
	aoi.height = h
	aoi.towerWidth = tw
	aoi.towerHeight = th
	aoi.rangeLimit = limit
	aoi.Emitter = event.NewEventList()
	aoi.Init()
	return aoi
}
