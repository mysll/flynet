package area

import (
	"logicdata/inter"
	"server"
	. "server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
	"server/util"
)

const (
	STATE_GAMING = iota
	STATE_DELETING
	STATE_SAVING
)

type AreaPlayer struct {
	Mailbox rpc.Mailbox
	Base    rpc.Mailbox
	Account string
	Name    string
	Entity  Entityer
	State   int
	Deleted bool
	Trans   server.Transform
	Cell    *cell
	Quit    bool
}

func (a *AreaPlayer) DeletePlayer() {
	if a.Entity != nil {
		App.Destroy(a.Entity.GetObjId())
	}
}

func (a *AreaPlayer) Save(remove bool) {
	var typ int
	if remove {
		typ = share.SAVETYPE_OFFLINE
		a.Quit = true
	} else {
		typ = share.SAVETYPE_TIMER
	}

	App.Save(a.Entity, typ)
}

func (a *AreaPlayer) LevelScene() {
	a.Cell = nil
}

type PlayerList struct {
	server.Dispatch
	players  map[uint64]*AreaPlayer
	namelist map[string]rpc.Mailbox
}

func (p *PlayerList) Count() int {
	return len(p.players)
}

func (p *PlayerList) AddPlayer(mailbox rpc.Mailbox, data *EntityInfo) *AreaPlayer {
	if _, dup := p.players[mailbox.Uid]; dup {
		log.LogError("player already added,", mailbox)
		return nil
	}

	ent, err := App.CreateFromArchive(data,
		map[string]interface{}{
			"mailbox": mailbox,
			"base":    rpc.Mailbox{App: mailbox.App},
			"sync":    true})
	if err != nil {
		log.LogError(err)
		return nil
	}

	nameinter, err := ent.Get("Name")
	if err != nil {
		log.LogError(err)
		App.Destroy(ent.GetObjId())
		return nil
	}
	name := nameinter.(string)
	if _, dup := p.namelist[name]; dup {
		log.LogError("player name conflict")
		App.Destroy(ent.GetObjId())
		return nil
	}

	pl := &AreaPlayer{}
	pl.Mailbox = mailbox
	pl.Base = rpc.Mailbox{App: mailbox.App}
	pl.Name = name
	pl.Entity = ent
	pl.Deleted = false
	pl.Quit = false
	p.players[mailbox.Uid] = pl
	p.namelist[name] = mailbox
	cell := App.GetCell(1)
	if w, ok := ent.(inter.Watcher); ok {
		w.SetRange(2)
	}

	pl.Cell = cell
	App.PlaceObj(
		cell.scene,
		ent,
		Vector3{
			util.RandRangef(0, cell.width),
			0,
			util.RandRangef(0, cell.height)},
		0)

	log.LogDebug("Add player:", mailbox)
	return pl
}

func (p *PlayerList) GetPlayerByName(name string) *AreaPlayer {
	if mailbox, exist := p.namelist[name]; exist {
		return p.GetPlayer(mailbox)
	}
	return nil
}

func (p *PlayerList) GetPlayer(mailbox rpc.Mailbox) *AreaPlayer {
	if pl, exist := p.players[mailbox.Uid]; exist && !pl.Deleted {
		return pl
	}
	return nil
}

func (p *PlayerList) RemovePlayer(mailbox rpc.Mailbox) bool {
	if pl, exist := p.players[mailbox.Uid]; exist && !pl.Deleted {
		if pl.Cell != nil {
			err := App.RemoveChild(pl.Cell.scene, pl.Entity)
			if err != nil {
				log.LogError(err)
			}
		}
		pl.Deleted = true
		pl.State = STATE_DELETING
		log.LogDebug("Remove player:", mailbox)
		return true
	}

	return false
}

func (p *PlayerList) ClearDeleted() {
	for session, pl := range p.players {
		if pl.Deleted {
			pl.DeletePlayer()
			delete(p.players, session)
			delete(p.namelist, pl.Name)
			log.LogDebug("delete player:", pl.Mailbox)
			log.LogDebug("remain players:", p.Count())
		}
	}
}

func NewPlayerList() *PlayerList {
	pl := &PlayerList{}
	pl.players = make(map[uint64]*AreaPlayer, 512)
	pl.namelist = make(map[string]rpc.Mailbox, 512)
	return pl
}
