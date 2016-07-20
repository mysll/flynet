package areamgr

import (
	"container/list"
	"fmt"
	"server"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
	"sync"
)

const (
	AREA_CREATING = iota
	AREA_CREATED
)

type area struct {
	AppId  string
	Status int
}

type Areas struct {
	l          sync.RWMutex
	lastareaid int
	areas      map[string]*area
	pending    map[string]*list.List
}

func (a *Areas) GetArea(mailbox rpc.Mailbox, id string) error {
	a.l.Lock()
	defer a.l.Unlock()

	log.LogMessage("GetArea")
	app := server.GetAppByName(mailbox.App)
	if app == nil {
		return server.ErrAppNotFound
	}

	if areainfo, exist := a.areas[id]; exist {
		if areainfo.Status == AREA_CREATED {
			return app.Call(&mailbox, "AreaBridge.GetAreaBak", areainfo.AppId)
		} else {
			a.pending[id].PushBack(mailbox)
			return nil
		}
	}

	a.lastareaid++
	appid := fmt.Sprintf("area_%d", a.lastareaid)
	data, err := share.CreateAppMsg("area",
		id,
		appid,
		fmt.Sprintf(`{ "id":"%s", "host":"127.0.0.1", "port":0, "areaid":"%s"}`,
			appid,
			id),
		App.Name,
	)
	if err != nil {
		log.LogError(err)
		return app.Call(&mailbox, "AreaBridge.GetAreaBak", "")
	}

	err = App.SendToMaster(data)
	if err != nil {
		log.LogError(err)
		return app.Call(&mailbox, "AreaBridge.GetAreaBak", "")
	}

	ar := &area{}
	ar.AppId = appid
	ar.Status = AREA_CREATING
	a.areas[id] = ar
	l := list.New()
	l.PushBack(mailbox)
	a.pending[id] = l

	log.LogMessage(a)
	return nil
}

func (a *Areas) createAppBak(bak share.CreateAppBak) {
	a.l.Lock()
	defer a.l.Unlock()
	appid := ""
	if bak.Res == "ok" {
		a.areas[bak.Id].Status = AREA_CREATED
		appid = bak.AppId
	} else {
		delete(a.areas, bak.Id)
	}

	log.LogMessage(bak, a)

	p := a.pending[bak.Id]
	var next *list.Element
	for e := p.Front(); e != nil; e = next {
		next = e.Next()
		mailbox := e.Value.(rpc.Mailbox)
		p.Remove(e)
		app := server.GetAppByName(mailbox.App)
		if app == nil {
			continue
		}
		app.Call(&mailbox, "AreaBridge.GetAreaBak", appid)
	}

	delete(a.pending, bak.Id)
}

func NewAreas() *Areas {
	a := &Areas{}
	a.lastareaid = 1000
	a.pending = make(map[string]*list.List, 8)
	a.areas = make(map[string]*area, 32)
	return a
}
