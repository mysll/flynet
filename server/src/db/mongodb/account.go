package mongodb

import (
	"errors"
	"libs/log"
	"libs/rpc"
	"pb/c2s"
	"pb/s2c"
	"server"
	"share"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"gopkg.in/mgo.v2/bson"
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
	if !a.quit {
		mb := r.GetSrc()
		a.queue[int(mb.Uid)%a.pools] <- r // 队列满了，就会阻塞在这里
	} else {
		log.LogWarning("it's quit, drop the message")
	}

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
			log.LogMessage(caller.GetSrc(), " rpc call:", caller.GetMethod(), ", thread:", id)
			start_time = time.Now()
			err := caller.Call()
			if err != nil {
				log.LogError("rpc error:", err)
			}
			delay = time.Now().Sub(start_time)
			if delay > warninglvl {
				log.LogWarning("rpc call ", caller.GetMethod(), " delay:", delay.Nanoseconds()/1000000, "ms")
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
	start := time.Now()
	err := f()
	if err != nil {
		log.LogError("db process error:", err)
	}
	delay := time.Now().Sub(start)
	if delay > time.Millisecond*10 {
		log.LogWarning("db process ", fname, " cost:", delay.Nanoseconds()/1000000, "ms")
	}
	atomic.AddInt32(&a.numprocess, -1)

	return err
}

func (a *Account) CheckAccount(mailbox rpc.Mailbox, login c2s.Loginuser) error {
	return nil
}

func (a *Account) ClearStatus(mailbox rpc.Mailbox, serverid string) error {
	return a.process("ClearStatus", func() error {
		c := db.DB.C(ROLEINFO)
		c.Update(bson.M{"status": 1, "serverid": serverid}, bson.M{"$set": bson.M{"status": 0, "serverid": ""}})
		return nil
	})
}

func (a *Account) LoadUser(mailbox rpc.Mailbox, info share.LoadUser) error {
	return a.process("LoadUser", func() error {
		bak := share.LoadUserBak{}
		app := server.GetAppById(mailbox.App)
		if app == nil {
			return server.ErrAppNotFound
		}

		c := db.DB.C(ROLEINFO)
		roleinfo := RoleInfo{}
		if err := c.Find(bson.M{"rolename": info.RoleName, "roleindex": info.Index}).One(&roleinfo); err != nil {
			return app.Call(&mailbox, "DbBridge.SelectUserBak", bak)
		}

		if roleinfo.Status == 1 {
			err := &s2c.Error{}
			err.ErrorNo = proto.Int32(share.ERROR_ROLE_USED)
			log.LogError("role inuse", info.RoleName)
			return server.MailTo(nil, &mailbox, "Login.Error", err)
		}

		//Loaduser
		data, err := LoadUser(db.DB, roleinfo.Uid, roleinfo.Entity)
		if err != nil {
			log.LogError(err)
			return err
		}

		c.Update(bson.M{"rolename": roleinfo.Rolename}, bson.M{"$set": bson.M{"lastlogintime": time.Now(), "status": 1, "serverid": mailbox.App}})

		bak.Name = info.RoleName
		bak.Scene = roleinfo.Scene
		bak.X = roleinfo.Scene_x
		bak.Y = roleinfo.Scene_y
		bak.Z = roleinfo.Scene_z
		bak.Data = &data

		log.LogMessage("load role succeed:", info.RoleName)

		return app.Call(&mailbox, "DbBridge.SelectUserBak", bak)
	})
	return nil
}

func (a *Account) CreateUser(mailbox rpc.Mailbox, info share.CreateUser) error {
	if info.SaveData.Data == nil {
		return errors.New("save data is nil")
	}
	return a.process("CreateUser", func() error {
		app := server.GetAppById(mailbox.App)
		if app == nil {
			return server.ErrAppNotFound
		}

		c := db.DB.C(ROLEINFO)
		result := RoleInfo{}
		if count, err := c.Find(bson.M{"account": info.Account}).Count(); err == nil {
			if count >= db.limit {
				errmsg := &s2c.Error{}
				errmsg.ErrorNo = proto.Int32(share.ERROR_ROLE_LIMIT)
				server.MailTo(&mailbox, &mailbox, "Role.Error", errmsg)
				return nil
			}
		}

		if err := c.Find(bson.M{"rolename": info.Name}).One(&result); err == nil {
			log.LogError("name conflict")
			err := &s2c.Error{}
			err.ErrorNo = proto.Int32(share.ERROR_NAME_CONFLIT)
			return server.MailTo(nil, &mailbox, "Role.Error", err)
		}

		uid := db.getNextSequence("userid")
		if uid == 0 {
			err := errors.New("uid is zero")
			log.LogError(err)
			return err
		}

		user := RoleInfo{
			Uid:           uid,
			Account:       info.Account,
			Rolename:      info.Name,
			Createtime:    time.Now(),
			Lastlogintime: time.Time{},
			Locktime:      time.Time{},
			Roleindex:     int8(info.Index),
			Roleinfo:      "",
			Entity:        info.SaveData.Data.Typ,
			Deleted:       0,
			Locked:        0,
			Status:        0,
			Serverid:      "",
			Scene:         info.Scene,
			Scene_x:       info.X,
			Scene_y:       info.Y,
			Scene_z:       info.Z,
		}

		if err := c.Insert(user); err != nil {
			log.LogError(err)
			return app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error())
		}

		if err := SaveToDb(db.DB, uid, &info.SaveData); err != nil {
			c.Remove(bson.M{"_id": uid})
			log.LogError(err)
			return app.Call(&mailbox, "DbBridge.CreateRoleBack", err.Error())

		}

		log.LogMessage("player:", info.Name, " create succeed")
		return app.Call(&mailbox, "DbBridge.CreateRoleBack", "ok")
	})

}

func (a *Account) GetUserInfo(mailbox rpc.Mailbox, account string) error {
	return a.process("GetUserInfo", func() error {

		c := db.DB.C(ROLEINFO)
		roleinfos := []RoleInfo{}
		c.Find(bson.M{"account": account, "deleted": 0}).All(&roleinfos)

		info := &s2c.RoleInfo{}
		info.UserInfo = make([]*s2c.Role, 0, 4)

		for _, r := range roleinfos {
			if r.Locked == 0 {
				role := &s2c.Role{}
				role.Name = proto.String(r.Rolename)
				role.Index = proto.Int32(int32(r.Roleindex))
				role.Roleinfo = proto.String(r.Roleinfo)
				info.UserInfo = append(info.UserInfo, role)
			}
		}

		return server.MailTo(nil, &mailbox, "roleinfo", info)
	})
}

func (a *Account) ClearPlayerStatus(mailbox rpc.Mailbox, info share.ClearUser) error {
	return a.process("ClearPlayerStatus", func() error {
		c := db.DB.C(ROLEINFO)
		c.Update(bson.M{"rolename": info.Name}, bson.M{"$set": bson.M{"status": 0, "serverid": ""}})

		c.Update(bson.M{"rolename": info.Name}, bson.M{"$set": bson.M{"status": 0, "serverid": ""}})

		log.LogMessage("player clear status,", info.Name)
		return nil
	})
}

func (a *Account) SavePlayer(mailbox rpc.Mailbox, data share.UpdateUser) error {
	return a.process("SavePlayer", func() error {
		base := rpc.NewMailBox(0, 0, mailbox.App)

		//save
		if err := SaveToDb(db.DB, 0, &data.SaveData); err != nil {
			log.LogError(err)
			return server.MailTo(&mailbox, &base, "DbBridge.SavePlayerBak", err.Error())
		}

		if data.Type == share.SAVETYPE_OFFLINE {
			if err := a.ClearPlayerStatus(mailbox, share.ClearUser{Account: data.Account, Name: data.Name}); err != nil {
				log.LogError(err)
				return server.MailTo(&mailbox, &base, "DbBridge.SavePlayerBak", err.Error())
			}
		}
		log.LogMessage("player:", data.Name, " save succeed")
		return server.MailTo(&mailbox, &base, "DbBridge.SavePlayerBak", "ok")
	})
}
