package base

import (
	"libs/log"
	"math/rand"
	"server"
	"sync"
	"time"
	"util/hash"
)

var (
	App *BaseApp
)

type BaseApp struct {
	*server.Server
	Players    *PlayerList
	startinit  sync.Once
	Proxy      *Proxy
	Account    *Account
	DbBridge   *DbBridge
	AreaBridge *AreaBridge
	Login      *Login
	Sync       *Sync
	config     *Config
	Letter     *LetterSystem
	tasksystem *TaskSystem
	taskLogic  *TaskLogic
}

func (b *BaseApp) OnPrepare() bool {
	rand.Seed(time.Now().UTC().UnixNano() + int64(hash.DJBHash(App.Id)))
	log.LogMessage(b.Id, " prepared")
	return true
}

func (b *BaseApp) OnShutdown() bool {
	c := b.GetClientList()
	c.CloseAll()
	log.LogInfo("kick all user")
	return false
}

//客户端断开连接
func (b *BaseApp) OnClientLost(id int64) {
	pl := b.Players.GetPlayer(id)
	if pl == nil || pl.Deleted {
		return
	}
	log.LogInfo("client offline,", id)
	pl.Disconnect()
}

func (b *BaseApp) StartInit() {
	db := server.GetAppByType("database")
	db.Call(nil, "Account.ClearStatus", b.Id)
	//server.NewDBWarp(db).SendSystemLetter(nil, "system", "", "", 1, "测试", "这是一封测试邮件", "", "_", share.DBParams{})
}

func (b *BaseApp) OnMustAppReady() {
	log.LogInfo("must app ready")
	b.startinit.Do(b.StartInit)
}

func (b *BaseApp) OnStart() {
	b.config.load()
	b.tasksystem.LoadTaskInfo()
}

func (b *BaseApp) OnFrame() {
	b.Players.ClearDeleted()
	if b.Players.Count() == 0 && b.Closing {
		log.LogInfo("shutdown")
		b.Shutdown()
	}
}

func (b *BaseApp) OnFlush() {
	b.Login.checkCached()
	b.Players.CheckNewDay()
}

func (b *BaseApp) OnReady(appid string) {
	App.AreaBridge.checkPending(appid)
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {

	App = &BaseApp{
		Players:    NewPlayerList(),
		Login:      NewLogin(),
		Proxy:      NewProxy(),
		Account:    NewAccount(),
		DbBridge:   NewDbBridge(),
		AreaBridge: NewAreaBridge(),
		Sync:       NewSync(),
		config:     NewConfig(),
		Letter:     NewLetterSystem(),
		tasksystem: NewTaskSystem(),
		taskLogic:  NewTaskLogic(),
	}

	server.RegisterCallee("role", &RoleCallee{})
	server.RegisterCallee("Player", &Player{})

	server.RegisterHandler("Account", App.Account)
	server.RegisterHandler("MailBox", App.Letter)
	server.RegisterHandler("Task", App.taskLogic)

	server.RegisterRemote("Login", App.Login)
	server.RegisterRemote("DbBridge", App.DbBridge)
	server.RegisterRemote("AreaBridge", App.AreaBridge)
	server.RegisterRemote("Sync", App.Sync)

}