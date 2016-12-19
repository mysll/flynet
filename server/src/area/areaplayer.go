package area

import (
	"server"
	. "server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
)

const (
	STATE_GAMING = iota
	STATE_DELETING
	STATE_SAVING
)

type AreaPlayer struct {
	server.PlayerInfo

	Base  rpc.Mailbox
	Trans server.Transform
	Cell  *cell
	Quit  bool
}

func NewAreaPlayer() server.PlayerHandler {
	ap := &AreaPlayer{}
	return ap
}

func (a *AreaPlayer) LoadPlayer(mailbox rpc.Mailbox, data *EntityInfo) error {
	ent, err := App.CreateFromArchive(data,
		map[string]interface{}{
			"mailbox": mailbox,
			"base":    rpc.Mailbox{App: mailbox.App},
			"sync":    true})
	if err != nil {
		log.LogError(err)
		return err
	}

	nameinter, err := ent.Get("Name")
	if err != nil {
		log.LogError(err)
		App.Destroy(ent.ObjectId())
		return err
	}
	name := nameinter.(string)

	a.Mailbox = mailbox
	a.Base = rpc.Mailbox{App: mailbox.App}
	a.Name = name
	a.Entity = ent
	a.Entity.SetUID(mailbox.Uid)
	a.Deleted = false
	a.Quit = false
	a.UpdateHash()
	return nil
	/*
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
	*/
}

func (a *AreaPlayer) RemovePlayer(mailbox rpc.Mailbox) bool {
	/*if pl, exist := p.players[mailbox.Uid]; exist && !pl.Deleted {
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
	}*/

	return false
}

func (a *AreaPlayer) DeletePlayer() {
	if a.Entity != nil {
		App.Destroy(a.Entity.ObjectId())
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
