package base

import (
	"errors"
	"logicdata/entity"
	"pb/s2c"
	"server"
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
	"server/util"

	"github.com/golang/protobuf/proto"
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

func (d *DbBridge) LookLetterBack(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	params := share.DBParams{}
	r := server.NewMessageReader(msg)
	if server.Check(r.ReadObject(&params)) {
		return nil
	}

	if !params["result"].(bool) {
		return nil
	}

	mailbox = params["mailbox"].(rpc.Mailbox)
	p := App.Players.GetPlayer(mailbox.Id)
	if p == nil {
		log.LogError("player not found, id:", mailbox.Id)
		//角色没有找到
		return nil
	}
	player := p.Entity.(*entity.Player)

	db := server.GetAppByType("database")
	if db == nil {
		log.LogError(server.ErrAppNotFound)
		return nil
	}

	warp := server.NewDBWarp(db)
	letters := params["letters"].([]*share.LetterInfo)
	for _, letter := range letters {
		idx := player.MailBox_r.AddRowValue(-1, letter.Source, letter.Source_name, util.UTC2Loc(letter.Send_time.Time.UTC()).Unix(), letter.Title, letter.Content, letter.Appendix, 0, letter.Serial_no, letter.Msg_type)
		if idx == -1 {
			//邮箱满了
			server.Error(nil, &mailbox, "Letter.Error", ERR_MAILBOX_FULL)
			break
		}
		//删信
		warp.RecvLetter(nil, player.GetDbId(), letter.Serial_no, "_", share.DBParams{})
	}

	return nil
}

func (d *DbBridge) createRole(mailbox rpc.Mailbox, obj datatype.Entityer, account string, name string, index int, save *share.DbSave) error {
	db := server.GetAppByType("database")
	if db != nil {
		trans := App.GetLandpos(obj)
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

func (d *DbBridge) CreateRoleBack(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	errstr, err := r.ReadString()
	if server.Check(err) {
		return nil
	}
	if errstr != "ok" {
		err := &s2c.Error{}
		err.ErrorNo = proto.Int32(share.ERROR_CREATE_ROLE_ERROR)
		server.Check(server.MailTo(nil, &mailbox, "Role.Error", err))
		return nil
	}

	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil {
		log.LogError("player not found, id:", mailbox.Id)
		//角色没有找到
		return nil
	}

	if player.State != STATE_LOGGED {
		log.LogError("player state not logged")
		return nil
	}

	server.Check(d.getUserInfo(mailbox, player.Account))
	return nil
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

func (d *DbBridge) RoleInUse(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	serverid, err := r.ReadString()
	if server.Check(err) {
		return nil
	}
	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil {
		return nil
	}

	if serverid == App.Name {
		server.Check(App.Players.SwitchPlayer(player))
		return nil
	}

	app := server.GetAppByName(serverid)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return nil
	}

	app.Call(&mailbox, "Login.SwitchPlayer", player.Account)
	return nil
}

func (d *DbBridge) SelectUserBak(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	save := share.LoadUserBak{}
	if server.Check(r.ReadObject(&save)) {
		return nil
	}
	db := server.GetAppByType("database")
	info := share.ClearUser{save.Account, save.Name}
	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil {
		log.LogError("player not found, id:", mailbox.Id)
		db.Call(&mailbox, "Account.ClearPlayerStatus", info)
		//角色没有找到
		return nil
	}

	if player.State != STATE_LOGGED {
		log.LogError("player state not logged")
		db.Call(&mailbox, "Account.ClearPlayerStatus", info)
		player.Leave()
		return nil
	}

	if save.Data == nil {
		err := &s2c.Error{}
		err.ErrorNo = proto.Int32(share.ERROR_SELECT_ROLE_ERROR)
		db.Call(&mailbox, "Account.ClearPlayerStatus", info)
		server.Check(server.MailTo(nil, &mailbox, "Login.Error", err))
		return nil
	}

	err := player.LoadPlayer(save)
	if server.Check(err) {
		db.Call(&mailbox, "Account.ClearPlayerStatus", info)
		return nil
	}

	status := server.GetAppByType("status")
	if status != nil {
		status.Call(&mailbox, "PlayerList.UpdatePlayer", player.Account, player.ChooseRole, App.Name)
	}
	//App.AreaBridge.getArea(mailbox, save.Scene) //这里会自动加入场景
	return nil
}

func (d *DbBridge) SavePlayerBak(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	err, e := r.ReadString()
	if server.Check(e) {
		return nil
	}
	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil || player.State != STATE_SAVING {
		log.LogError(errors.New("player not found"))
		return nil
	}

	if err == "ok" {
		App.Players.RemovePlayer(mailbox.Id)
	} else {
		player.Entity.SetSaveFlag()
		player.SaveFailed()
	}

	return nil
}

func (d *DbBridge) savePlayer(p *BasePlayer, typ int) error {
	if p == nil || p.Entity == nil {
		return errors.New("player not created")
	}

	App.Save(p.Entity, typ)
	trans := App.GetLandpos(p.Entity)
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

func (d *DbBridge) UpdateUserInfo(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	infos := share.UpdateUserBak{}
	if server.Check(r.ReadObject(&infos)) {
		return nil
	}
	for _, info := range infos.Infos {
		if e := App.GetEntity(info.ObjId); e != nil {
			e.SetDbId(info.DBId)
			continue
		}

		log.LogWarning("update user info failed, object not found,", info.ObjId)
	}
	return nil
}

func NewDbBridge() *DbBridge {
	db := &DbBridge{}
	return db
}
