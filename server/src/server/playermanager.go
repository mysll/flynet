package server

import (
	"fmt"
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/util"
	"time"
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
	Mailbox rpc.Mailbox
	Account string
	Name    string
	Session int64
	Entity  datatype.Entityer
	State   int
	Deleted bool
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

func (pi *PlayerInfo) SetAccount(acc string) {
	pi.Account = acc
}

func (pi *PlayerInfo) GetName() string {
	return pi.Name
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

func (pi *PlayerInfo) GetEntityer() datatype.Entityer {
	return pi.Entity
}

func (pi *PlayerInfo) SetEntityer(ent datatype.Entityer) {
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

}

type PlayerInfoer interface {
	GetMailbox() rpc.Mailbox
	SetMailbox(mb rpc.Mailbox)
	GetAccount() string
	SetAccount(acc string)
	GetName() string
	SetName(name string)
	GetSession() int64
	SetSession(s int64)
	GetEntityer() datatype.Entityer
	SetEntityer(ent datatype.Entityer)
	GetState() int
	SetState(s int)
	GetDeleted() bool
	SetDeleted(d bool)
	OnBeforeDelete()
	OnDelete()
	OnClear()
	OnEvent(event string)
}

type InitFunc func() PlayerInfoer

type PlayerManager struct {
	Dispatch
	UniqueType int
	Players    map[uint64]PlayerInfoer
	namemap    map[string]uint64
	serial     uint64
	initfunc   InitFunc
	lasttime   time.Time
}

var Players *PlayerManager

func NewPlayerManager(typ int, f InitFunc) *PlayerManager {
	if Players != nil {
		return Players
	}
	pm := &PlayerManager{}
	pm.Players = make(map[uint64]PlayerInfoer, 512)
	pm.namemap = make(map[string]uint64, 512)
	pm.UniqueType = typ
	pm.initfunc = f
	pm.lasttime = time.Now()
	core.AddDispatchNoName(pm, DP_UPDATE|DP_FLUSH)
	Players = pm
	return Players
}

func (pm *PlayerManager) GetNewSerial() uint64 {
	pm.serial++
	return pm.serial
}

func (pm *PlayerManager) AddNewPlayer(uid uint64) (PlayerInfoer, error) {
	if _, dup := pm.Players[uid]; dup {
		return nil, fmt.Errorf("add player uid is exist")
	}

	player := pm.initfunc()
	pm.Players[uid] = player
	return player, nil
}

func (pm *PlayerManager) AddPlayer(uid uint64, player PlayerInfoer) error {

	if _, dup := pm.Players[uid]; dup {
		return fmt.Errorf("add player uid is exist")
	}

	pm.Players[uid] = player
	return nil
}

func (pm *PlayerManager) UpdateName(uid uint64) {
	if player, dup := pm.Players[uid]; dup && !player.GetDeleted() {
		name := player.GetName()
		if name != "" {
			if _, dup := pm.namemap[name]; dup {
				log.LogError("name dup")
				return
			}
			pm.namemap[name] = uid
		}
	}
}

func (pm *PlayerManager) FindPlayer(uid uint64) PlayerInfoer {
	if player, dup := pm.Players[uid]; dup && !player.GetDeleted() {
		return player
	}
	return nil
}

func (pm *PlayerManager) FindPlayerByName(name string) PlayerInfoer {
	if uid, dup := pm.namemap[name]; dup {
		player := pm.Players[uid]
		if !player.GetDeleted() {
			return player
		}
	}

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

func (pm *PlayerManager) OnFlush() {
	if util.IsSameDay(pm.lasttime, time.Now()) {
		return
	}

	for _, pl := range pm.Players {
		if !pl.GetDeleted() {
			pl.OnEvent(E_NEWDAY)
		}
	}

	pm.lasttime = time.Now()
}

func (pm *PlayerManager) OnUpdate() {
	//check delete
	for uid, pl := range pm.Players {
		if pl.GetDeleted() {
			pl.OnDelete()
			delete(pm.Players, uid)
			delete(pm.namemap, pl.GetName())
			pl.OnClear()
			log.LogDebug("delete player:", uid)
			log.LogDebug("remain players:", pm.Count())
		}
	}
}
