package mysqldb

import (
	"database/sql"
	"errors"
	"libs/log"
	"libs/rpc"
	"pb/c2s"
	"pb/s2c"
	"server"
	"share"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
)

type Account struct {
	numprocess int32
	queue      []chan *rpc.RpcCall
	quit       bool
	pools      int
}

func NewAccount(pool int) *Account {
	a := &Account{}
	a.queue = make([]chan *rpc.RpcCall, pool)
	a.pools = pool
	for i := 0; i < pool; i++ {
		a.queue[i] = make(chan *rpc.RpcCall, 128)
	}
	return a
}

func (a *Account) Push(r *rpc.RpcCall) bool {

	mb := r.Src.Interface().(rpc.Mailbox)
	a.queue[int(mb.Uid)%a.pools] <- r // 队列满了，就会阻塞在这里

	return true
}

func (a *Account) work(id int) {
	log.LogMessage("db work, id:", id)
	var start_time time.Time
	var delay time.Duration
	warninglvl := 50 * time.Millisecond
	for {
		select {
		case caller := <-a.queue[id]:
			log.LogMessage(caller.Src.Interface(), " rpc call:", caller.Req.ServiceMethod, ", thread:", id)
			start_time = time.Now()
			err := caller.Call()
			if err != nil {
				log.LogError("rpc error:", err)
			}
			delay = time.Now().Sub(start_time)
			if delay > warninglvl {
				log.LogWarning("rpc call ", caller.Req.ServiceMethod, " delay:", delay.Nanoseconds()/1000000, "ms")
			}
			caller.Free()

			break
		default:
			if a.quit {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func (a *Account) Do() {
	log.LogMessage("start db thread, total:", a.pools)
	for i := 0; i < a.pools; i++ {
		id := i
		db.wg.Wrap(func() { a.work(id) })
	}
}

func (a *Account) process(fname string, f func() error) error {
	atomic.AddInt32(&a.numprocess, 1)
	err := f()
	if err != nil {
		log.LogError("db process error:", err)
	}
	atomic.AddInt32(&a.numprocess, -1)
	return err
}

func (a *Account) CheckAccount(mailbox rpc.Mailbox, login c2s.Loginuser) error {
	return nil
}

func (a *Account) ClearStatus(mailbox rpc.Mailbox, serverid string) error {
	return a.process("ClearStatus", func() error {
		_, err := db.sql.Exec("UPDATE `role_info` SET `status`=?, `serverid`=? WHERE `status`=? and `serverid`=?", 0, "", 1, serverid)
		return err
	})
}

func (a *Account) LoadUser(mailbox rpc.Mailbox, info share.LoadUser) error {
	return a.process("LoadUser", func() error {
		sqlconn := db.sql
		var r *sql.Rows
		var err error
		bak := share.LoadUserBak{}
		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
		}

		if r, err = sqlconn.Query("SELECT `uid`,`locktime`,`locked`,`entity`, `status`, `serverid`, `scene`, `scene_x`, `scene_y`, `scene_z`, `scene_dir`, `roleinfo`, `landtimes` FROM `role_info` WHERE `rolename`=? and `roleindex`=? LIMIT 1", info.RoleName, info.Index); err != nil {
			log.LogError(err)
			return err
		}
		if !r.Next() {
			log.LogError("user not found")
			r.Close()
			return app.Call(&mailbox, "DbBridge.SelectUserBak", bak)
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
			return app.Call(&mailbox, "DbBridge.RoleInUse", serverid)
		}

		data, err := LoadUser(sqlconn, uid, ent)
		if err != nil {
			log.LogError(err)
			return err
		}

		if _, err = sqlconn.Exec("UPDATE `role_info` set `lastlogintime`=?,`status`=?,`serverid`=?,`landtimes`=`landtimes`+1 WHERE `rolename`=? LIMIT 1", time.Now().Format("2006-01-02 15:04:05"), 1, mailbox.Address, info.RoleName); err != nil {
			log.LogError(err)
			return err
		}
		data.RoleInfo = roleinfo
		bak.Name = info.RoleName
		bak.Scene = scene
		bak.X = x
		bak.Y = y
		bak.Z = z
		bak.Dir = dir
		bak.LandTimes = landtimes
		bak.Data = &data
		return app.Call(&mailbox, "DbBridge.SelectUserBak", bak)
	})
}

func (a *Account) CreateUser(mailbox rpc.Mailbox, info share.CreateUser) error {
	if info.SaveData.Data == nil {
		return errors.New("save data is nil")
	}
	return a.process("CreateUser", func() error {
		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
		}
		sqlconn := db.sql
		var r *sql.Rows
		var err error

		if r, err = sqlconn.Query("SELECT count(*) FROM `role_info` WHERE `account`=?", info.Account); err != nil {
			log.LogError(err)
			return app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error())
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

		if r, err = sqlconn.Query("SELECT `uid` FROM `role_info` WHERE `rolename`=? LIMIT 1", info.Name); err != nil {
			log.LogError(err)
			return app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error())
		}
		if r.Next() {
			log.LogError("name conflict")
			r.Close()
			errmsg := &s2c.Error{}
			errmsg.ErrorNo = proto.Int32(share.ERROR_NAME_CONFLIT)
			server.MailTo(&mailbox, &mailbox, "Role.Error", errmsg)
			return nil
		}
		r.Close()

		uid, err := sqlconn.GetUid("userid")
		if err != nil {
			log.LogError(err)
			return app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error())
		}

		if err = CreateUser(sqlconn, uid, info.Account, info.Name, info.Index, info.Scene, info.X, info.Y, info.Z, info.Dir, info.SaveData.Data.Typ, &info.SaveData); err != nil {
			log.LogError(err)
			return app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error())
		}
		return a.GetUserInfo(mailbox, info.Account)
		//return app.Call(&mailbox, "DbBridge.CreateRoleBack", "ok")

	})
}

func (a *Account) GetUserInfo(mailbox rpc.Mailbox, account string) error {
	return a.process("GetUserInfo", func() error {
		s := db.sql
		r, err := s.Query(`SELECT rolename, roleindex, roleinfo, locktime, locked 
FROM role_info
WHERE account=? and deleted =0`, account)
		if err != nil {
			log.LogError(err)
			return err
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
				return err
			}
			if locked == 0 {
				role := &s2c.Role{}
				role.Name = proto.String(rolename)
				role.Index = proto.Int32(roleindex)
				role.Roleinfo = proto.String(roleinfo)

				info.UserInfo = append(info.UserInfo, role)
			}
		}
		return server.MailTo(nil, &mailbox, "Role.RoleInfo", info)
	})
}

func (a *Account) ClearPlayerStatus(mailbox rpc.Mailbox, info share.ClearUser) error {
	return a.process("ClearPlayerStatus", func() error {
		sqlconn := db.sql

		_, err := sqlconn.Exec("UPDATE `role_info` SET `status`=?, `serverid`=? WHERE `rolename`=?", 0, "", info.Name)
		if err != nil {
			log.LogError(err)
			return err
		}
		log.LogMessage("player clear status,", info.Name)
		return nil
	})
}

func (a *Account) SavePlayer(mailbox rpc.Mailbox, data share.UpdateUser) error {
	if data.SaveData.Data == nil {
		return errors.New("save data is nil")
	}
	return a.process("SavePlayer", func() error {
		sqlconn := db.sql
		base := rpc.Mailbox{}
		base.Address = mailbox.Address
		infos := make([]share.ObjectInfo, 0, 128)
		if err := UpdateItem(sqlconn, data.SaveData.Data.DBId, data.SaveData.Data); err != nil {
			log.LogError(err)
			return server.MailTo(&mailbox, &base, "DbBridge.SavePlayerBak", err.Error())
		}

		if data.Type == share.SAVETYPE_OFFLINE {
			_, err := db.sql.Exec("UPDATE `role_info` SET `status`=?, `serverid`=?, `scene`=?, `scene_x`=?, `scene_y`=?, `scene_z`=?, `scene_dir`=?, `roleinfo`=? WHERE `rolename`=?",
				0,
				"",
				data.Scene, data.X, data.Y, data.Z, data.Dir,
				data.SaveData.RoleInfo,
				data.Name,
			)
			if err != nil {
				log.LogError(err)
				return server.MailTo(&mailbox, &base, "DbBridge.SavePlayerBak", err.Error())
			}

			return server.MailTo(&mailbox, &base, "DbBridge.SavePlayerBak", "ok")
		}

		bakinfo := share.UpdateUserBak{}
		bakinfo.Infos = infos
		return server.MailTo(&mailbox, &base, "DbBridge.UpdateUserInfo", bakinfo)
	})
}
