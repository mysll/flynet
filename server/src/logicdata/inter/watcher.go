package inter

import (
	"bytes"
	"encoding/gob"
	. "server/data/datatype"
	"server/libs/log"
)

type Watcher interface {
	AddObject(id ObjectID, typ int) bool
	RemoveObject(id ObjectID, typ int) bool
	SetRange(r int)
	GetRange() int
	ClearAll()
}

type AOI struct {
	watchers map[int]map[int32]ObjectID
	Range    int
}

func (a *AOI) Init() {
	a.watchers = make(map[int]map[int32]ObjectID, 10)
	a.Range = 1
}

func (a *AOI) Clear() {
	for _, v := range a.watchers {
		for k := range v {
			delete(v, k)
		}
	}
}

func (a *AOI) ClearAll() {
	a.Clear()
}

func (a *AOI) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	var err error

	err = encoder.Encode(a.Range)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (a *AOI) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	var err error
	err = decoder.Decode(&a.Range)
	if err != nil {
		return err
	}
	return nil
}

func (a *AOI) AddObject(id ObjectID, typ int) bool {
	if v, ok := a.watchers[typ]; ok {
		if _, ok := v[id.Index]; ok {
			return false
		}
		v[id.Index] = id
		return true
	}

	w := make(map[int32]ObjectID, 100)

	w[id.Index] = id
	a.watchers[typ] = w

	log.LogDebug("aoi add object:", id)
	return true
}

func (a *AOI) RemoveObject(id ObjectID, typ int) bool {
	if v, ok := a.watchers[typ]; ok {
		if _, ok := v[id.Index]; !ok {
			return false
		}
		delete(v, id.Index)
		log.LogDebug("aoi remove object:", id)
		return true
	}
	return false
}

func (a *AOI) SetRange(r int) {
	a.Range = r
}

func (a *AOI) GetRange() int {
	return a.Range
}
