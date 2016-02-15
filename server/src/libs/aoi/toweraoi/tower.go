package toweraoi

import (
	. "data/datatype"
)

type Tower struct {
	Ids      map[int32]ObjectID
	Watchers map[int]map[int32]ObjectID
	TypeMap  map[int]map[int32]ObjectID
}

func (t *Tower) clear() {
	for k := range t.Ids {
		delete(t.Ids, k)
	}

	for _, v := range t.Watchers {
		for k := range v {
			delete(v, k)
		}
	}

	for _, v := range t.TypeMap {
		for k := range v {
			delete(v, k)
		}
	}
}

func (t *Tower) add(obj ObjectID, typ int) bool {
	if _, ok := t.Ids[obj.Index]; ok {
		return false
	}
	t.Ids[obj.Index] = obj

	if _, ok := t.TypeMap[typ]; !ok {
		t.TypeMap[typ] = make(map[int32]ObjectID, 200)
	}

	t.TypeMap[typ][obj.Index] = obj
	return true
}

func (t *Tower) remove(obj ObjectID, typ int) bool {
	if _, ok := t.Ids[obj.Index]; !ok {
		return false
	}
	delete(t.Ids, obj.Index)
	if _, ok := t.TypeMap[typ]; ok {
		delete(t.TypeMap[typ], obj.Index)
	}
	return true
}

func (t *Tower) getIds() []ObjectID {
	if len(t.Ids) == 0 {
		return nil
	}

	objs := make([]ObjectID, 0, len(t.Ids))
	for _, o := range t.Ids {
		objs = append(objs, o)
	}
	return objs
}

func (t *Tower) getIdsByTypes(types []int) []ObjectID {
	count := 0
	for _, typ := range types {
		if _, ok := t.TypeMap[typ]; ok {
			count += len(t.TypeMap[typ])
		}
	}
	if count == 0 {
		return nil
	}

	objs := make([]ObjectID, 0, count)
	for _, typ := range types {
		if ts, ok := t.TypeMap[typ]; ok {
			for _, o := range ts {
				objs = append(objs, o)
			}
		}
	}
	return objs

}

func (t *Tower) addWatcher(watcher ObjectID, typ int) bool {

	if _, ok := t.Watchers[typ]; !ok {
		t.Watchers[typ] = make(map[int32]ObjectID)
	}

	t.Watchers[typ][watcher.Index] = watcher

	return true
}

func (t *Tower) removeWatcher(watcher ObjectID, typ int) {
	if _, ok := t.Watchers[typ]; !ok {
		return
	}

	delete(t.Watchers[typ], watcher.Index)
}

func (t *Tower) getAllWatchers() []ObjectID {
	count := 0
	for _, ws := range t.Watchers {
		count += len(ws)
	}

	if count == 0 {
		return nil
	}

	result := make([]ObjectID, 0, count)

	for _, ws := range t.Watchers {
		for _, w := range ws {
			result = append(result, w)
		}
	}

	return result
}

func (t *Tower) getWatchers(types []int) []ObjectID {
	if len(types) == 0 {
		return nil
	}

	count := 0
	for _, ts := range types {
		if _, ok := t.Watchers[ts]; ok {
			count += len(t.Watchers[ts])
		}
	}

	if count == 0 {
		return nil
	}

	result := make([]ObjectID, 0, count)
	for _, ts := range types {
		if _, ok := t.Watchers[ts]; ok {
			for _, ws := range t.Watchers[ts] {
				result = append(result, ws)
			}
		}
	}

	return result
}

func NewTower() *Tower {
	t := &Tower{}
	t.Ids = make(map[int32]ObjectID, 1024)
	t.Watchers = make(map[int]map[int32]ObjectID, 1024)
	t.TypeMap = make(map[int]map[int32]ObjectID, 1024)
	return t
}
