package base

import (
	"container/list"
	"libs/log"
	"libs/rpc"
	"pb/s2c"
	"server"
	"share"

	"github.com/golang/protobuf/proto"
)

type AreaBridge struct {
	pending map[string]*list.List
	buf     []byte
}

func (a *AreaBridge) getArea(mailbox rpc.Mailbox, id string) error {
	app := server.GetAppByType("areamgr")
	if app == nil {
		return server.ErrAppNotFound
	}

	err := app.Call(&mailbox, "AreaMgr.GetArea", id)
	if err != nil {
		log.LogError(err)
	}
	return err
}

func (a *AreaBridge) GetAreaBak(mailbox rpc.Mailbox, areaid string) error {
	if areaid == "" {
		log.LogError("enter area failed")
		return nil
	}
	player := App.Players.GetPlayer(mailbox.Id)
	if player != nil {
		player.AreaId = areaid
		player.State = STATE_ENTERAREA
		//player.EnterScene(areaid)
		a.enterArea(player, areaid)
		log.LogMessage("enter area:", areaid)
	}

	return nil
}

func (a *AreaBridge) enterArea(player *BasePlayer, areaid string) error {
	ap := server.GetApp(areaid)
	if ap == nil {
		if _, ok := a.pending[areaid]; ok {
			a.pending[areaid].PushBack(player.Mailbox)
			return nil
		} else {
			l := list.New()
			l.PushBack(player.Mailbox)
			a.pending[areaid] = l
			return nil
		}
	}
	if err := a.areaAddPlayer(ap, player); err != nil {
		log.LogError(err)
		return err
	}

	return nil
}

func (a *AreaBridge) checkPending(appid string) {
	if l, ok := a.pending[appid]; ok {
		ap := server.GetApp(appid)
		var next *list.Element
		for e := l.Front(); e != nil; e = next {
			next = e.Next()
			mb := e.Value.(rpc.Mailbox)
			l.Remove(e)
			player := App.Players.GetPlayer(mb.Id)
			if player == nil {
				continue
			}

			err := a.areaAddPlayer(ap, player)
			if err != nil {
				log.LogError(err)
			}
		}
	}
}

func (a *AreaBridge) areaAddPlayer(ap *server.RemoteApp, player *BasePlayer) error {
	addplayer, err := share.GetPlayerInfo(player.Account, player.ChooseRole, player.trans.Scene, player.trans.Pos.X, player.trans.Pos.Y, player.trans.Pos.Z, player.trans.Dir, player.Entity)
	if err != nil {
		return err
	}
	err = ap.Call(&player.Mailbox, "BaseProxy.AddPlayer", addplayer)
	if err != nil {
		return err
	}

	player.State = STATE_ENTERAREA
	return nil
}

func (a *AreaBridge) AddPlayerBak(mailbox rpc.Mailbox, res string) error {
	if res == "ok" {
		player := App.Players.GetPlayer(mailbox.Id)
		if player == nil {
			log.LogFatalf("can not be nil")
			return nil
		}

		if player.State != STATE_ENTERAREA { //可能客户端已经断开了，则让玩家下线
			player.Leave()
			return nil
		}

		player.EnterScene()
		return nil
	} else {
		err := &s2c.Error{}
		err.ErrorNo = proto.Int32(share.ERROR_ROLE_ENTERAREA_ERROR)
		return server.MailTo(nil, &mailbox, "error", err)
	}

}

func (a *AreaBridge) areaRemovePlayer(player *BasePlayer, typ int) {
	area := server.GetApp(player.AreaId)
	if area == nil {
		player.SaveToDb(true)
		return
	}

	err := area.Call(&player.Mailbox, "BaseProxy.RemovePlayer", typ)
	if err != nil {
		player.SaveToDb(true)
		return
	}

}

func (a *AreaBridge) RemovePlayerBak(mailbox rpc.Mailbox, res string) error {
	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil {
		log.LogFatalf("can not be nil", mailbox)
		return nil
	}
	player.LeaveArea()
	return nil
}

func NewAreaBridge() *AreaBridge {
	a := &AreaBridge{}
	a.pending = make(map[string]*list.List, 32)
	a.buf = make([]byte, 0, 512*1024)
	return a
}
