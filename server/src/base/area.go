package base

import (
	"container/list"
	"pb/s2c"
	"server"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"

	"github.com/golang/protobuf/proto"
)

type AreaBridge struct {
	pending map[string]*list.List
	buf     []byte
}

func (t *AreaBridge) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("GetAreaBak", t.GetAreaBak)
	s.RegisterCallback("AddPlayerBak", t.AddPlayerBak)
	s.RegisterCallback("RemovePlayerBak", t.RemovePlayerBak)
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

func (a *AreaBridge) GetAreaBak(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	r := server.NewMessageReader(msg)
	areaid, err := r.ReadString()
	if server.Check(err) {
		return 0, nil
	}
	if areaid == "" {
		log.LogError("enter area failed")
		return 0, nil
	}
	p := App.Players.FindPlayer(mailbox.Uid)
	if p != nil {
		player := p.(*BasePlayer)
		player.AreaId = areaid
		player.State = STATE_ENTERAREA
		//player.EnterScene(areaid)
		a.enterArea(player, areaid)
		log.LogMessage("enter area:", areaid)
	}

	return 0, nil
}

func (a *AreaBridge) enterArea(player *BasePlayer, areaid string) error {
	ap := server.GetAppByName(areaid)
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
		ap := server.GetAppByName(appid)
		var next *list.Element
		for e := l.Front(); e != nil; e = next {
			next = e.Next()
			mb := e.Value.(rpc.Mailbox)
			l.Remove(e)
			p := App.Players.FindPlayer(mb.Uid)
			if p == nil {
				continue
			}

			player := p.(*BasePlayer)
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

func (a *AreaBridge) AddPlayerBak(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	r := server.NewMessageReader(msg)
	res, err := r.ReadString()
	if server.Check(err) {
		return 0, nil
	}
	if res == "ok" {
		p := App.Players.FindPlayer(mailbox.Uid)
		if p == nil {
			log.LogFatalf("can not be nil", mailbox)
			return 0, nil
		}

		player := p.(*BasePlayer)
		if player.State != STATE_ENTERAREA { //可能客户端已经断开了，则让玩家下线
			player.Leave()
			return 0, nil
		}

		player.EnterScene()
		return 0, nil
	} else {
		err := &s2c.Error{}
		err.ErrorNo = proto.Int32(share.ERROR_ROLE_ENTERAREA_ERROR)
		server.Check(server.MailTo(nil, &mailbox, "error", err))
		return 0, nil
	}

}

func (a *AreaBridge) areaRemovePlayer(player *BasePlayer, typ int) {
	area := server.GetAppByName(player.AreaId)
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

func (a *AreaBridge) RemovePlayerBak(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	p := App.Players.FindPlayer(mailbox.Uid)
	if p == nil {
		log.LogFatalf("can not be nil", mailbox)
		return 0, nil
	}

	player := p.(*BasePlayer)
	player.LeaveArea()
	return 0, nil
}

func NewAreaBridge() *AreaBridge {
	a := &AreaBridge{}
	a.pending = make(map[string]*list.List, 32)
	a.buf = make([]byte, 0, 512*1024)
	return a
}
