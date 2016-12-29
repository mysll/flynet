package base

import (
	"logicdata/entity"
	"math/rand"
	_ "pb"
	"server"
	"server/libs/log"
	"server/libs/rpc"
	"server/util"
	"sync"
	"time"
)

var (
	App *BaseApp
)

type BaseApp struct {
	*server.Server
	startinit  sync.Once
	Account    *Account
	DbBridge   *DbBridge
	AreaBridge *AreaBridge
	Login      *Login
	Sync       *Sync
	config     *Config
}

func (b *BaseApp) IsBase() bool {
	return true
}

func (b *BaseApp) OnPrepare() bool {
	rand.Seed(time.Now().UTC().UnixNano() + int64(util.DJBHash(App.Name)))
	log.LogMessage(b.Name, " prepared")
	server.NewPlayerManager(entity.GetType("Player"), NewBasePlayer)
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
	p := b.Players.FindPlayerBySession(id)
	if p == nil {
		return
	}

	pl := p.(*BasePlayer)
	if pl.GetDeleted() {
		return
	}
	log.LogInfo("client offline,", id)
	pl.Disconnect()
}

func (b *BaseApp) StartInit() {
	db := server.GetAppByType("database")
	db.Call(nil, "Account.ClearStatus", b.Name)
}

func (b *BaseApp) OnMustAppReady() {
	log.LogInfo("must app ready")
	b.startinit.Do(b.StartInit)
}

func (b *BaseApp) OnStart() {
	b.config.load()
}

func (b *BaseApp) OnFrame() {
	if b.Closing && b.Players.Count() == 0 {
		log.LogInfo("shutdown")
		b.Shutdown()
	}
}

func (b *BaseApp) OnFlush() {
	b.Login.checkCached()
}

func (b *BaseApp) OnReady(appid string) {
	App.AreaBridge.checkPending(appid)
}

func (b *BaseApp) OnSceneTeleported(mailbox rpc.Mailbox, result bool) {
	log.LogDebug("teleport to scene, result:", result)
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {

	App = &BaseApp{
		Login:      NewLogin(),
		Account:    NewAccount(),
		DbBridge:   NewDbBridge(),
		AreaBridge: NewAreaBridge(),
		Sync:       NewSync(),
		config:     NewConfig(),
	}

	server.RegisterCallee("role", &RoleCallee{})
	server.RegisterCallee("Player", &Player{})

	server.RegisterHandler("Account", App.Account)

	server.RegisterRemote("Login", App.Login)
	server.RegisterRemote("DbBridge", App.DbBridge)
	server.RegisterRemote("AreaBridge", App.AreaBridge)
	server.RegisterRemote("Sync", App.Sync)

	//server.RegisterModule("taskmodule", task.Module)
	//server.RegisterModule("lettermodule", letter.Module)
}
