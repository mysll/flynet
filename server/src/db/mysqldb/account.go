package mysqldb

import (
	"database/sql"
	"libs/log"
	"libs/rpc"
	"pb/s2c"
	"server"
	"share"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
)

type Account struct {
	*rpc.Thread
}

func NewAccount(pool int) *Account {
	a := &Account{}
	a.Thread = rpc.NewThread("Account", pool, 128)
	return a
}

func (a *Account) CheckAccount(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	return nil
}

func (a *Account) ClearStatus(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	serverid, err := r.ReadString()
	if server.Check(err) {
		return nil
	}

	server.Check2(db.sql.Exec("UPDATE `role_info` SET `status`=?, `serverid`=? WHERE `status`=? and `serverid`=?", 0, "", 1, serverid))
	log.LogMessage("clear server:", serverid)
	return nil
}

func (a *Account) LoadUser(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	reader := server.NewMessageReader(msg)
	var info share.LoadUser
	if server.Check(reader.ReadObject(&info)) {
		return nil
	}

	sqlconn := db.sql
	var r *sql.Rows
	var err error
	bak := share.LoadUserBak{}
	bak.Account = info.Account
	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return nil
	}

	if r, err = sqlconn.Query("SELECT `uid`,`locktime`,`locked`,`entity`, `status`, `serverid`, `scene`, `scene_x`, `scene_y`, `scene_z`, `scene_dir`, `roleinfo`, `landtimes` FROM `role_info` WHERE `account`=? and `rolename`=? and `roleindex`=? LIMIT 1", info.Account, info.RoleName, info.Index); err != nil {
		server.Check(err)
		return nil
	}
	if !r.Next() {
		log.LogError("user not found")
		r.Close()
		server.Check(app.Call(&mailbox, "DbBridge.SelectUserBak", bak))
		return nil
	}
	var uid uint64
	var ltime mysql.NullTime
	var locked int
	var ent string
	var status int
	var serverid string
	var scene string
	var x, y, z, dir float32
	var roleinfo string
	var landtimes int32
	err = r.Scan(&uid, &ltime, &locked, &ent, &status, &serverid, &scene, &x, &y, &z, &dir, &roleinfo, &landtimes)
	if err != nil {
		log.LogError("scan user failed")
		r.Close()
		return nil
	}
	r.Close()

	if status == 1 {
		server.Check(app.Call(&mailbox, "DbBridge.RoleInUse", serverid))
		return nil
	}

	data, err := LoadUser(sqlconn, uid, ent)
	if server.Check(err) {
		server.Check(err)
		return nil
	}

	if _, err = sqlconn.Exec("UPDATE `role_info` set `lastlogintime`=?,`status`=?,`serverid`=?,`landtimes`=`landtimes`+1 WHERE `account`=? and `rolename`=? LIMIT 1", time.Now().Format("2006-01-02 15:04:05"), 1, app.Name, info.Account, info.RoleName); err != nil {
		log.LogError(err)
		return nil
	}
	data.RoleInfo = roleinfo
	bak.Account = info.Account
	bak.Name = info.RoleName
	bak.Scene = scene
	bak.X = x
	bak.Y = y
	bak.Z = z
	bak.Dir = dir
	bak.LandTimes = landtimes
	bak.Data = &data
	server.Check(app.Call(&mailbox, "DbBridge.SelectUserBak", bak))
	return nil
}

func (a *Account) CreateUser(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	reader := server.NewMessageReader(msg)
	var info share.CreateUser
	if server.Check(reader.ReadObject(&info)) {
		return nil
	}
	if info.SaveData.Data == nil {
		log.LogError("save data is nil")
		return nil
	}

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return nil
	}
	sqlconn := db.sql
	var r *sql.Rows
	var err error

	if r, err = sqlconn.Query("SELECT count(*) FROM `role_info` WHERE `account`=?", info.Account); err != nil {
		log.LogError(err)
		server.Check(app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error()))
		return nil
	}
	var count int
	if r.Next() {
		r.Scan(&count)
	}
	r.Close()

	if count >= db.limit {
		errmsg := &s2c.Error{}
		errmsg.ErrorNo = proto.Int32(share.ERROR_ROLE_LIMIT)
		server.MailTo(&mailbox, &mailbox, "Role.Error", errmsg)
		return nil
	}

	if db.nameunique { //名称是否唯一
		if r, err = sqlconn.Query("SELECT `uid` FROM `role_info` WHERE `account`=? and `rolename`=? LIMIT 1", info.Account, info.Name); err != nil {
			log.LogError(err)
			server.Check(app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error()))
			return nil
		}
		if r.Next() {
			log.LogError("name conflict")
			r.Close()
			errmsg := &s2c.Error{}
			errmsg.ErrorNo = proto.Int32(share.ERROR_NAME_CONFLIT)
			server.Check(server.MailTo(&mailbox, &mailbox, "Role.Error", errmsg))
			return nil
		}
		r.Close()
	}

	uid, err := sqlconn.GetUid("userid")
	if err != nil {
		log.LogError(err)
		server.Check(app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error()))
		return nil
	}

	if err = CreateUser(sqlconn, uid, info.Account, info.Name, info.Index, info.Scene, info.X, info.Y, info.Z, info.Dir, info.SaveData.Data.Typ, &info.SaveData); err != nil {
		log.LogError(err)
		server.Check(app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error()))
		return nil
	}

	server.Check(server.GetLocalApp().Call(&mailbox, "Account.GetUserInfo", info.Account))

	return nil
	//return app.Call(&mailbox, "DbBridge.CreateRoleBack", "ok")
}

func (a *Account) GetUserInfo(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	reader := server.NewMessageReader(msg)
	account, err := reader.ReadString()
	if server.Check(err) {
		return nil
	}

	s := db.sql
	r, err := s.Query(`SELECT rolename, roleindex, roleinfo, locktime, locked 
FROM role_info
WHERE account=? and deleted =0`, account)
	if err != nil {
		log.LogError(err)
		return nil
	}
	defer r.Close()

	var rolename string
	var roleindex int32
	var roleinfo string
	var locktime mysql.NullTime
	var locked int

	info := &s2c.RoleInfo{}
	info.UserInfo = make([]*s2c.Role, 0, 4)
	for r.Next() {
		err = r.Scan(&rolename, &roleindex, &roleinfo, &locktime, &locked)
		if err != nil {
			log.LogError(err)
			return nil
		}
		if locked == 0 {
			role := &s2c.Role{}
			role.Name = proto.String(rolename)
			role.Index = proto.Int32(roleindex)
			role.Roleinfo = proto.String(roleinfo)

			info.UserInfo = append(info.UserInfo, role)
		}
	}
	server.Check(server.MailTo(nil, &mailbox, "Role.RoleInfo", info))
	return nil
}

func (a *Account) ClearPlayerStatus(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	var info share.ClearUser
	if server.Check(r.ReadObject(&info)) {
		return nil
	}

	sqlconn := db.sql
	_, err := sqlconn.Exec("UPDATE `role_info` SET `status`=?, `serverid`=? WHERE `account`=? and `rolename`=?", 0, "", info.Account, info.Name)
	if err != nil {
		log.LogError(err)
		return nil
	}
	log.LogMessage("player clear status,", info.Name)
	return nil
}

func (a *Account) SavePlayer(mailbox rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	r := server.NewMessageReader(msg)
	var data share.UpdateUser
	if server.Check(r.ReadObject(&data)) {
		return nil
	}
	if data.SaveData.Data == nil {
		log.LogError("save data is nil")
		return nil
	}

	sqlconn := db.sql
	base := rpc.NewMailBox(0, 0, mailbox.App)
	infos := make([]share.ObjectInfo, 0, 128)
	if err := UpdateItem(sqlconn, data.SaveData.Data.DBId, data.SaveData.Data); err != nil {
		log.LogError(err)
		server.Check(server.MailTo(&mailbox, &base, "DbBridge.SavePlayerBak", err.Error()))
		return nil
	}

	if data.Type == share.SAVETYPE_OFFLINE {
		_, err := db.sql.Exec("UPDATE `role_info` SET `status`=?, `serverid`=?, `scene`=?, `scene_x`=?, `scene_y`=?, `scene_z`=?, `scene_dir`=?, `roleinfo`=? WHERE `account`=? and `rolename`=?",
			0,
			"",
			data.Scene, data.X, data.Y, data.Z, data.Dir,
			data.SaveData.RoleInfo,
			data.Account,
			data.Name,
		)
		if err != nil {
			log.LogError(err)
			server.Check(server.MailTo(&mailbox, &base, "DbBridge.SavePlayerBak", err.Error()))
			return nil
		}

		server.Check(server.MailTo(&mailbox, &base, "DbBridge.SavePlayerBak", "ok"))
		return nil
	}

	bakinfo := share.UpdateUserBak{}
	bakinfo.Infos = infos
	server.Check(server.MailTo(&mailbox, &base, "DbBridge.UpdateUserInfo", bakinfo))
	return nil
}

func (t *Account) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("CheckAccount", t.CheckAccount)
	s.RegisterCallback("ClearStatus", t.ClearStatus)
	s.RegisterCallback("LoadUser", t.LoadUser)
	s.RegisterCallback("CreateUser", t.CreateUser)
	s.RegisterCallback("GetUserInfo", t.GetUserInfo)
	s.RegisterCallback("ClearPlayerStatus", t.ClearPlayerStatus)
	s.RegisterCallback("SavePlayer", t.SavePlayer)
}
