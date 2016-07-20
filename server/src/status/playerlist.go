package status

import (
	"server"
	"server/libs/log"
	"server/libs/rpc"
)

const (
	MAX_ROLES = 1
)

type PlayerInfo struct {
	RoleName string
	BaseId   string
}

type AccountInfo struct {
	roles map[string]*PlayerInfo
}

func (ai *AccountInfo) GetPlayInfo(role string) *PlayerInfo {
	if pi, ok := ai.roles[role]; ok {
		return pi
	}
	return nil
}

type PlayerList struct {
	accounts map[string]*AccountInfo
}

func (t *PlayerList) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("UpdatePlayer", t.UpdatePlayer)
	s.RegisterCallback("GetPlayerBase", t.GetPlayerBase)
}

func NewPlayerList() *PlayerList {
	pl := &PlayerList{}
	pl.accounts = make(map[string]*AccountInfo, 1024)
	return pl
}

func (pl *PlayerList) UpdatePlayer(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	account, err := r.ReadString()
	if server.Check(err) {
		return nil
	}
	role_name, err := r.ReadString()
	if server.Check(err) {
		return nil
	}
	base_id, err := r.ReadString()
	if server.Check(err) {
		return nil
	}

	if base_id == "" {
		if acc, ok := pl.accounts[account]; ok {
			pl := acc.GetPlayInfo(role_name)
			if pl != nil {
				delete(acc.roles, account)
				log.LogMessage("remove ", account, ",", role_name)
			}
		}

		return nil
	}

	if acc, ok := pl.accounts[account]; ok {
		p := acc.GetPlayInfo(role_name)
		if p != nil {
			log.LogMessage("update ", account, ",", role_name, " base:", base_id)
			p.BaseId = base_id
			return nil
		}

		p = &PlayerInfo{role_name, base_id}
		acc.roles[role_name] = p
		log.LogMessage("add ", account, ",", role_name, " base:", base_id)
		return nil
	}

	p := &PlayerInfo{role_name, base_id}
	acc := &AccountInfo{}
	acc.roles = make(map[string]*PlayerInfo, MAX_ROLES)
	acc.roles[role_name] = p
	pl.accounts[account] = acc
	log.LogMessage("add ", account, ",", role_name, " base:", base_id)
	return nil
}

func (pl *PlayerList) GetPlayerBase(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	account, err := r.ReadString()
	if server.Check(err) {
		return nil
	}
	role_name, err := r.ReadString()
	if server.Check(err) {
		return nil
	}
	callback, err := r.ReadString()
	if server.Check(err) {
		return nil
	}

	app := server.GetAppById(mailbox.App)
	if app == nil {
		server.Check(server.ErrAppNotFound)
		return nil
	}

	if acc, ok := pl.accounts[account]; ok {
		pl := acc.GetPlayInfo(role_name)
		if pl != nil {
			server.Check(app.Call(&mailbox, callback, account, role_name, pl.BaseId))
			return nil
		}
	}

	server.Check(app.Call(&mailbox, callback, account, role_name, ""))
	return nil
}
