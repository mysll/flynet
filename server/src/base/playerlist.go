package base

import (
	"libs/log"
	"libs/rpc"
	"pb/s2c"
	"server"
	"share"
	"time"
	"util"
)

type PlayerList struct {
	players  map[int64]*BasePlayer
	lasttime time.Time
}

func (p *PlayerList) SwitchPlayer(player *BasePlayer) error {
	oldsession := player.Session
	//被顶替的玩家
	replace := p.FindPlayer(player.Account, player.ChooseRole, player.Session)
	if replace == nil {
		log.LogError("old player not found")
		return nil
	}

	//交换连接
	App.SwitchConn(replace.Session, player.Session)
	//把原来的玩家踢下线
	server.Error(nil, &player.Mailbox, "Login.Error", share.ERROR_ROLE_REPLACE)
	App.DelayKickUser(oldsession, 5)
	//同步玩家数据
	App.AttachPlayer(replace.Entity, replace.Mailbox)

	player.Account += "*replace*"
	player.ChooseRole += "*replace*"
	log.LogInfo("switch player: old:", oldsession, " new:", replace.Session)
	return nil
}

func (p *PlayerList) FindPlayer(acc, role string, exclude int64) *BasePlayer {
	for k, v := range p.players {
		if k == exclude {
			continue
		}

		if v.Account == acc && v.ChooseRole == role {
			return v
		}
	}

	return nil
}

func (p *PlayerList) Count() int {
	return len(p.players)
}

func (p *PlayerList) AddPlayer(session int64) *BasePlayer {

	if _, dup := p.players[session]; dup {
		return nil
	}

	pl := &BasePlayer{}
	pl.Mailbox = rpc.NewMailBox(1, session, App.AppId)
	pl.Session = session
	p.players[session] = pl
	return pl
}

func (p *PlayerList) GetPlayer(session int64) *BasePlayer {
	if pl, exist := p.players[session]; exist && !pl.Deleted {
		return pl
	}
	return nil
}

func (p *PlayerList) RemovePlayer(session int64) bool {
	if pl, exist := p.players[session]; exist && !pl.Deleted {
		status := server.GetAppByType("status")
		if status != nil {
			status.Call(nil, "PlayerList.UpdatePlayer", pl.Account, pl.ChooseRole, "")
		}

		pl.Deleted = true
		pl.State = STATE_DELETING
		log.LogDebug("Remove player:", pl.ChooseRole, " session:", session)
		return true
	}

	return false
}

func (p *PlayerList) ClearDeleted() {
	for session, pl := range p.players {
		if pl.Deleted {
			pl.DeletePlayer()
			delete(p.players, session)
			log.LogDebug("delete player:", session)
			log.LogDebug("remain players:", p.Count())
		}
	}
}

func (p *PlayerList) CheckNewDay() {
	if util.IsSameDay(p.lasttime, time.Now()) {
		return
	}

	newday := &s2c.Respnewday{}
	for _, pl := range p.players {
		if !pl.Deleted {
			server.MailTo(nil, &pl.Mailbox, "SyncTime.RespNewDay", newday)
		}
	}

	p.lasttime = time.Now()
}

func NewPlayerList() *PlayerList {
	pl := &PlayerList{}
	pl.players = make(map[int64]*BasePlayer)
	pl.lasttime = time.Now()
	return pl
}
