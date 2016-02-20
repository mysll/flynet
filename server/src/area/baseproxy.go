package area

import (
	. "data/datatype"
	"data/entity"
	"libs/log"
	"libs/rpc"
	"server"
	"share"
)

type BaseProxy struct {
}

func (b *BaseProxy) AddPlayer(mailbox rpc.Mailbox, player share.PlayerInfo) error {

	base := server.GetApp(mailbox.Address)
	if base == nil {
		return server.ErrAppNotFound
	}

	ap := App.players.AddPlayer(mailbox, player.Entity)
	if ap == nil {
		return base.Call(&mailbox, "AreaBridge.AddPlayerBak", "create player failed")
	}
	ap.State = STATE_GAMING
	ap.Trans = server.Transform{player.Scene, Vector3{player.X, player.Y, player.Z}, player.Dir}

	App.Kernel.EntryScene(ap.Entity)
	ap.State = STATE_GAMING
	return base.Call(&mailbox, "AreaBridge.AddPlayerBak", "ok")
}

func (b *BaseProxy) RemovePlayer(mailbox rpc.Mailbox, reason int) error {
	//同步数据
	player := App.players.GetPlayer(mailbox)
	if player == nil {
		log.LogError("player not found")
	}
	player.Save(true)

	var err error
	if player.Entity.GetExtraData("saveData") == nil {
		log.LogError("player save data is nil")
		return err
	}

	err = server.MailTo(&App.MailBox,
		&player.Base,
		"Sync.SyncPlayer",
		map[string]interface{}{
			"mailbox": player.Mailbox,
			"data":    player.Entity.GetExtraData("saveData")},
	)

	player.Entity.RemoveExtraData("saveData")

	return err
}

func (b *BaseProxy) SyncPlayerBak(mailbox rpc.Mailbox, info map[ObjectID]ObjectID) error {
	//同步数据
	player := App.players.GetPlayer(mailbox)
	if player == nil {
		log.LogError("player not found")
	}

	for k, v := range info {
		ent := App.Kernel.GetEntity(k)
		if ent == nil {
			log.LogError("object not found")
			continue
		}
		ent.SetExtraData("linkObj", v)
	}

	var err error
	if player.Quit {
		App.players.RemovePlayer(mailbox)
		err = server.MailTo(&mailbox,
			&player.Base,
			"AreaBridge.RemovePlayerBak",
			"ok",
		)
	}

	return err
}

func (b *BaseProxy) entryScene(scene entity.Entityer, player entity.Entityer) {

}

func NewBaseProxy() *BaseProxy {
	bp := &BaseProxy{}
	return bp
}