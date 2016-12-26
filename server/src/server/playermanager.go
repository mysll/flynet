package server

import (
	"fmt"
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
	"server/util"
)

const (
	UniqueType_Session = iota
	UniqueType_UID
	UniqueType_Name
)

const (
	E_NEWDAY = "newday"
)

type PlayerInfo struct {
	Mailbox  rpc.Mailbox
	Account  string
	acchash  int32
	Name     string
	namehash int32
	Session  int64
	Entity   datatype.Entity
	State    int
	Deleted  bool
}

func (pi *PlayerInfo) GetMailbox() rpc.Mailbox {
	return pi.Mailbox
}

func (pi *PlayerInfo) SetMailbox(mb rpc.Mailbox) {
	pi.Mailbox = mb
}

func (pi *PlayerInfo) GetAccount() string {
	return pi.Account
}

func (pi *PlayerInfo) GetAccountHash() int32 {
	return pi.acchash
}

func (pi *PlayerInfo) SetAccount(acc string) {
	pi.Account = acc
}

func (pi *PlayerInfo) GetName() string {
	return pi.Name
}

func (pi *PlayerInfo) GetNameHash() int32 {
	return pi.namehash
}

func (pi *PlayerInfo) SetName(name string) {
	pi.Name = name
}

func (pi *PlayerInfo) GetSession() int64 {
	return pi.Session
}

func (pi *PlayerInfo) SetSession(s int64) {
	pi.Session = s
}

func (pi *PlayerInfo) GetEntity() datatype.Entity {
	return pi.Entity
}

func (pi *PlayerInfo) SetEntity(ent datatype.Entity) {
	pi.Entity = ent
}

func (pi *PlayerInfo) GetState() int {
	return pi.State
}

func (pi *PlayerInfo) SetState(s int) {
	pi.State = s
}

func (pi *PlayerInfo) GetDeleted() bool {
	return pi.Deleted
}

func (pi *PlayerInfo) SetDeleted(d bool) {
	pi.Deleted = d
}

func (pi *PlayerInfo) OnBeforeDelete() {

}

func (pi *PlayerInfo) OnDelete() {

}

func (pi *PlayerInfo) OnClear() {

}

func (pi *PlayerInfo) OnEvent(event string) {
	core.kernel.Emit(pi.Entity, event, nil)
}

func (pi *PlayerInfo) UpdateHash() {
	pi.acchash = util.DJBHash(pi.Account)
	pi.namehash = util.DJBHash(pi.Name)
}

type PlayerHandler interface {
	GetMailbox() rpc.Mailbox
	SetMailbox(mb rpc.Mailbox)
	GetAccount() string
	GetAccountHash() int32
	SetAccount(acc string)
	GetName() string
	GetNameHash() int32
	SetName(name string)
	GetSession() int64
	SetSession(s int64)
	GetEntity() datatype.Entity
	SetEntity(ent datatype.Entity)
	GetState() int
	SetState(s int)
	GetDeleted() bool
	SetDeleted(d bool)
	OnBeforeDelete()
	OnDelete()
	OnClear()
	OnEvent(event string)
	UpdateHash()
}

type InitFunc func() PlayerHandler

type PlayerManager struct {
	Dispatch
	UniqueType int
	Players    map[uint64]PlayerHandler
	serial     uint64
	initfunc   InitFunc
}

func NewPlayerManager(typ int, f InitFunc) *PlayerManager {
	if core.Players != nil {
		return core.Players
	}
	pm := &PlayerManager{}
	pm.Players = make(map[uint64]PlayerHandler, 512)
	pm.UniqueType = typ
	pm.initfunc = f
	core.kernel.AddDispatchNoName(pm, DP_FRAME)
	core.Players = pm
	return pm
}

func (pm *PlayerManager) GetNewSerial() uint64 {
	pm.serial++
	return pm.serial
}

func (pm *PlayerManager) AddNewPlayer(uid uint64) (PlayerHandler, error) {
	if _, dup := pm.Players[uid]; dup {
		return nil, fmt.Errorf("add player uid is exist")
	}

	player := pm.initfunc()
	player.UpdateHash()
	pm.Players[uid] = player
	return player, nil
}

func (pm *PlayerManager) AddPlayer(uid uint64, player PlayerHandler) error {

	if _, dup := pm.Players[uid]; dup {
		return fmt.Errorf("add player uid is exist")
	}
	player.UpdateHash()
	pm.Players[uid] = player
	return nil
}

func (pm *PlayerManager) FindPlayer(uid uint64) PlayerHandler {
	if player, dup := pm.Players[uid]; dup {
		return player
	}
	return nil
}

func (pm *PlayerManager) FindPlayerBySession(session int64) PlayerHandler {
	for _, v := range pm.Players {
		if v.GetSession() == session {
			return v
		}
	}

	return nil
}

func (pm *PlayerManager) FindPlayerByName(name string) PlayerHandler {
	hash := util.DJBHash(name)
	for _, v := range pm.Players {
		if hash == v.GetNameHash() && v.GetName() == name {
			return v
		}
	}

	return nil
}

func (pm *PlayerManager) FindPlayerByAccountAndName(acc string, name string, exclude uint64) PlayerHandler {
	hash1 := util.DJBHash(acc)
	hash2 := util.DJBHash(name)
	for k, v := range pm.Players {
		if k == exclude {
			continue
		}
		if hash1 == v.GetAccountHash() && hash2 == v.GetNameHash() && v.GetAccount() == acc && v.GetName() == name {
			return v
		}
	}
	return nil
}

func (pm *PlayerManager) SwitchPlayer(player PlayerHandler) error {
	oldsession := player.GetSession()
	//被顶替的玩家
	replace := pm.FindPlayerByAccountAndName(player.GetAccount(), player.GetName(), uint64(player.GetSession()))
	if replace == nil {
		log.LogError("old player not found")
		return nil
	}

	//交换连接
	core.SwitchConn(replace.GetSession(), player.GetSession())
	//把原来的玩家踢下线
	mb := player.GetMailbox()
	Error(nil, &mb, "Login.Error", share.ERROR_ROLE_REPLACE)
	core.DelayKickUser(oldsession, 5)
	//同步玩家数据
	core.kernel.AttachPlayer(replace.GetEntity(), replace.GetMailbox())

	player.SetAccount(player.GetAccount() + "*replace*")
	player.SetName(player.GetName() + "*replace*")
	log.LogInfo("switch player: old:", oldsession, " new:", replace.GetSession())
	return nil
}

func (pm *PlayerManager) Count() int {
	return len(pm.Players)
}

func (pm *PlayerManager) RemovePlayer(uid uint64) {
	if player, dup := pm.Players[uid]; dup && !player.GetDeleted() {
		player.OnBeforeDelete()
		player.SetDeleted(true)
	}
}

func (pm *PlayerManager) Emit(event string) {
	for _, pl := range pm.Players {
		if !pl.GetDeleted() {
			pl.OnEvent(event)
		}
	}
}

func (pm *PlayerManager) OnFrame() {
	//check delete
	for uid, pl := range pm.Players {
		if pl.GetDeleted() {
			pl.OnDelete()
			delete(pm.Players, uid)
			pl.OnClear()
			log.LogDebug("delete player:", uid)
			log.LogDebug("remain players:", pm.Count())
		}
	}
}
