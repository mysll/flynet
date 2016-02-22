package base

import (
	"data/entity"
	"errors"
	"libs/log"
	"libs/rpc"
	"pb/s2c"
	"server"
	"share"
	"util"

	"github.com/golang/protobuf/proto"
)

type DbBridge struct {
}

func (d *DbBridge) LookLetterBack(mailbox rpc.Mailbox, params share.DBParams) error {
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
		return server.ErrAppNotFound
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

func (d *DbBridge) createRole(mailbox rpc.Mailbox, obj entity.Entityer, account string, name string, index int, save *share.DbSave) error {
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

func (d *DbBridge) CreateRoleBack(mailbox rpc.Mailbox, errstr string) error {
	if errstr != "ok" {
		err := &s2c.Error{}
		err.ErrorNo = proto.Int32(share.ERROR_CREATE_ROLE_ERROR)
		return server.MailTo(nil, &mailbox, "Role.Error", err)
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

	return d.getUserInfo(mailbox, player.Account)
}

func (d *DbBridge) getUserInfo(mailbox rpc.Mailbox, account string) error {
	db := server.GetAppByType("database")

	if db != nil {
		return db.Call(&mailbox, "Account.GetUserInfo", account)
	}

	return server.ErrAppNotFound
}

func (d *DbBridge) selectUser(mailbox rpc.Mailbox, rolename string, index int) error {
	db := server.GetAppByType("database")

	if db != nil {
		loaduser := share.LoadUser{}
		loaduser.RoleName = rolename
		loaduser.Index = index
		return db.Call(&mailbox, "Account.LoadUser", loaduser)
	}

	return server.ErrAppNotFound
}

func (d *DbBridge) RoleInUse(mailbox rpc.Mailbox, serverid string) error {
	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil {
		return nil
	}

	if serverid == App.Id {

		return App.Players.SwitchPlayer(player)
	}

	app := server.GetApp(serverid)
	if app == nil {
		return server.ErrAppNotFound
	}

	app.Call(&mailbox, "Login.SwitchPlayer", player.Account)
	return nil
}

func (d *DbBridge) SelectUserBak(mailbox rpc.Mailbox, save share.LoadUserBak) error {
	db := server.GetAppByType("database")
	info := share.ClearUser{save.Name}
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
		return server.MailTo(nil, &mailbox, "Login.Error", err)
	}

	err := player.LoadPlayer(save)
	if err != nil {
		db.Call(&mailbox, "Account.ClearPlayerStatus", info)
		return err
	}

	status := server.GetAppByType("status")
	if status != nil {
		status.Call(&mailbox, "PlayerList.UpdatePlayer", player.Account, player.ChooseRole, App.Id)
	}
	//App.AreaBridge.getArea(mailbox, save.Scene) //这里会自动加入场景
	return nil
}

func (d *DbBridge) SavePlayerBak(mailbox rpc.Mailbox, err string) error {
	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil || player.State != STATE_SAVING {
		return errors.New("player not found")
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

func (d *DbBridge) UpdateUserInfo(mailbox rpc.Mailbox, infos share.UpdateUserBak) error {
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
