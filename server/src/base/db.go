package base

import (
	"errors"
	l "game/module/letter"
	"logicdata/entity"
	"server"
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
	"server/util"
)

type DbBridge struct {
}

func (t *DbBridge) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("LookLetterBack", t.LookLetterBack)
	s.RegisterCallback("CreateRoleBack", t.CreateRoleBack)
	s.RegisterCallback("RoleInUse", t.RoleInUse)
	s.RegisterCallback("SelectUserBak", t.SelectUserBak)
	s.RegisterCallback("SavePlayerBak", t.SavePlayerBak)
	s.RegisterCallback("UpdateUserInfo", t.UpdateUserInfo)
}

func (d *DbBridge) LookLetterBack(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	params := share.DBParams{}
	r := server.NewMessageReader(msg)
	if server.Check(r.ReadObject(&params)) {
		return 0, nil
	}

	if !params["result"].(bool) {
		return 0, nil
	}

	mailbox = params["mailbox"].(rpc.Mailbox)
	p := App.Players.FindPlayer(mailbox.Uid)
	if p == nil {
		log.LogError("player not found, id:", mailbox.Id)
		//角色没有找到
		return 0, nil
	}
	player := p.(*BasePlayer).Entity.(*entity.Player)

	db := server.GetAppByType("database")
	if db == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	warp := server.NewDBWarp(db)
	letters := params["letters"].([]*share.LetterInfo)
	for _, letter := range letters {
		idx := player.MailBox_r.AddRowValue(-1, letter.Source, letter.Source_name, util.UTC2Loc(letter.Send_time.Time.UTC()).Unix(), letter.Title, letter.Content, letter.Appendix, 0, letter.Serial_no, letter.Msg_type)
		if idx == -1 {
			//邮箱满了
			server.Error(nil, &mailbox, "Letter.Error", l.ERR_MAILBOX_FULL)
			break
		}
		//删信
		warp.RecvLetter(nil, player.DBId(), letter.Serial_no, "_", share.DBParams{})
	}

	return 0, nil
}

func (d *DbBridge) createRole(mailbox rpc.Mailbox, obj datatype.Entity, account string, name string, index int, save *share.DbSave) error {
	db := server.GetAppByType("database")
	if db != nil {
		trans := App.Kernel().GetLandpos(obj)
		cu := share.CreateUser{}
		cu.Account = account
		cu.Name = name
		cu.Index = index
		cu.Scene, cu.X, cu.Y, cu.Z, cu.Dir = trans.Scene, trans.Pos.X, trans.Pos.Y, trans.Pos.Z, trans.Dir
		cu.SaveData = *save
		return db.Call(&mailbox, "Account.CreateUser", cu)
	}

	return server.ErrAppNotFound
}

func (d *DbBridge) CreateRoleBack(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	r := server.NewMessageReader(msg)
	errstr, err := r.ReadString()
	if server.Check(err) {
		return 0, nil
	}
	if errstr != "ok" {
		server.Check(server.Error(nil, &mailbox, "Role.Error", share.ERROR_CREATE_ROLE_ERROR))
		return 0, nil
	}

	p := App.Players.FindPlayer(mailbox.Uid)
	if p == nil {
		log.LogError("player not found, id:", mailbox.Id)
		//角色没有找到
		return 0, nil
	}
	player := p.(*BasePlayer)

	if player.State != STATE_LOGGED {
		log.LogError("player state not logged")
		return 0, nil
	}

	server.Check(d.getUserInfo(mailbox, player.Account))
	return 0, nil
}

func (d *DbBridge) getUserInfo(mailbox rpc.Mailbox, account string) error {
	db := server.GetAppByType("database")

	if db != nil {
		return db.Call(&mailbox, "Account.GetUserInfo", account)
	}

	return server.ErrAppNotFound
}

func (d *DbBridge) selectUser(mailbox rpc.Mailbox, account string, rolename string, index int) error {
	db := server.GetAppByType("database")

	if db != nil {
		loaduser := share.LoadUser{}
		loaduser.Account = account
		loaduser.RoleName = rolename
		loaduser.Index = index
		return db.Call(&mailbox, "Account.LoadUser", loaduser)
	}

	return server.ErrAppNotFound
}

func (d *DbBridge) RoleInUse(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	r := server.NewMessageReader(msg)
	serverid, err := r.ReadString()
	if server.Check(err) {
		return 0, nil
	}
	p := App.Players.FindPlayer(mailbox.Uid)
	if p == nil {
		return 0, nil
	}

	player := p.(*BasePlayer)

	if serverid == App.Name {
		server.Check(App.Players.SwitchPlayer(player))
		return 0, nil
	}

	app := server.GetAppByName(serverid)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	app.Call(&mailbox, "Login.SwitchPlayer", player.Account)
	return 0, nil
}

func (d *DbBridge) SelectUserBak(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	r := server.NewMessageReader(msg)
	save := share.LoadUserBak{}
	if server.Check(r.ReadObject(&save)) {
		return 0, nil
	}
	db := server.GetAppByType("database")
	info := share.ClearUser{save.Account, save.Name}
	p := App.Players.FindPlayer(mailbox.Uid)
	if p == nil {
		log.LogError("player not found, id:", mailbox.Id)
		db.Call(&mailbox, "Account.ClearPlayerStatus", info)
		//角色没有找到
		return 0, nil
	}

	player := p.(*BasePlayer)
	if player.State != STATE_LOGGED {
		log.LogError("player state not logged")
		db.Call(&mailbox, "Account.ClearPlayerStatus", info)
		player.Leave()
		return 0, nil
	}

	if save.Data == nil {
		db.Call(&mailbox, "Account.ClearPlayerStatus", info)
		server.Check(server.Error(nil, &mailbox, "Login.Error", share.ERROR_SELECT_ROLE_ERROR))
		return 0, nil
	}

	err := player.LoadPlayer(save)
	if server.Check(err) {
		db.Call(&mailbox, "Account.ClearPlayerStatus", info)
		return 0, nil
	}

	status := server.GetAppByType("status")
	if status != nil {
		status.Call(&mailbox, "PlayerList.UpdatePlayer", player.Account, player.ChooseRole, App.Name)
	}
	//App.AreaBridge.getArea(mailbox, save.Scene) //这里会自动加入场景
	return 0, nil
}

func (d *DbBridge) SavePlayerBak(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	r := server.NewMessageReader(msg)
	err, e := r.ReadString()
	if server.Check(e) {
		return 0, nil
	}
	p := App.Players.FindPlayer(mailbox.Uid)
	if p == nil {
		log.LogError(errors.New("player not found"))
		return 0, nil
	}

	player := p.(*BasePlayer)
	if player.State != STATE_SAVING {
		log.LogError(errors.New("player state not match"))
		return 0, nil
	}

	if err == "ok" {
		App.Players.RemovePlayer(mailbox.Uid)
	} else {
		player.Entity.SetSaveFlag()
		player.SaveFailed()
	}

	return 0, nil
}

func (d *DbBridge) savePlayer(p *BasePlayer, typ int) error {
	if p == nil || p.Entity == nil {
		return errors.New("player not created")
	}

	App.Kernel().Save(p.Entity, typ)
	trans := App.Kernel().GetLandpos(p.Entity)
	p.trans = trans
	save := share.UpdateUser{}
	save.Account = p.Account
	save.Name = p.ChooseRole
	save.Type = typ
	save.Scene, save.X, save.Y, save.Z, save.Dir = p.trans.Scene, p.trans.Pos.X, p.trans.Pos.Y, p.trans.Pos.Z, p.trans.Dir

	save.SaveData = *share.GetSaveData(p.Entity)
	db := server.GetAppByType("database")
	if db != nil {
		err := db.Call(&p.Mailbox, "Account.SavePlayer", save)
		if err == nil {
			p.Entity.ClearSaveFlag()
		}
		return err
	}

	return server.ErrAppNotFound

}

func (d *DbBridge) UpdateUserInfo(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	r := server.NewMessageReader(msg)
	infos := share.UpdateUserBak{}
	if server.Check(r.ReadObject(&infos)) {
		return 0, nil
	}
	for _, info := range infos.Infos {
		if e := App.Kernel().GetEntity(info.ObjId); e != nil {
			e.SetDBId(info.DBId)
			continue
		}

		log.LogWarning("update user info failed, object not found,", info.ObjId)
	}
	return 0, nil
}

func NewDbBridge() *DbBridge {
	db := &DbBridge{}
	return db
}
