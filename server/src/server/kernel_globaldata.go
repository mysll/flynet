package server

import (
	"fmt"
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
)

const (
	STATUS_NONE = iota
	STATUS_CREATE
	STATUS_CREATED
)

type globalClient struct {
	appid       int32
	status      int32
	dataversion int64
	errcount    int32
	disable     bool
}

type GlobalDataHelper struct {
	Dispatch
	dataset       datatype.Entityer
	dataCenter    string
	isServer      bool
	globalclients map[string]*globalClient
	dataversion   int64
}

func NewGlobalDataHelper() *GlobalDataHelper {
	h := &GlobalDataHelper{}
	return h
}

func (gd *GlobalDataHelper) SetServer() {
	gd.globalclients = make(map[string]*globalClient)
	gd.dataCenter = core.Name
	gd.isServer = true
}

func (gd *GlobalDataHelper) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("CreateGlobalData", gd.CreateGlobalData)
	s.RegisterCallback("AddGlobalData", gd.AddGlobalData)
}

//client function
func (gd *GlobalDataHelper) CreateGlobalData(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	reply = CreateMessage(core.Name)
	var datainfo *datatype.EntityInfo
	if err := ParseArgs(msg, &gd.dataCenter, &datainfo); err != nil {
		return share.ERR_ARGS_ERROR, reply
	}

	if !core.enableglobaldata {
		return share.ERR_FUNC_BEGIN + 1, reply
	}

	ent, err := core.CreateFromArchive(datainfo, nil)
	if err != nil {
		return share.ERR_REPLY_FAILED, reply
	}

	gd.dataset = ent

	log.LogMessage("create global data succeed")
	return share.ERR_REPLY_SUCCEED, reply
}

func (gd *GlobalDataHelper) OnCreateGlobalData(msg *rpc.Message) {
	var appname string
	if err := ParseArgs(msg, &appname); err != nil {
		log.LogError(err)
		return
	}
	ret := GetReplyError(msg)

	client, exist := gd.globalclients[appname]
	if !exist {
		return
	}

	if ret == share.ERR_FUNC_BEGIN+1 {
		client.disable = false
		log.LogMessage(appname, " global data is disabled")
		return
	}

	if ret != share.ERR_REPLY_SUCCEED {

		client.errcount++
		client.status = STATUS_NONE
		return

	}

	client.status = STATUS_CREATED
	log.LogMessage("create global data to ", appname, " succeed")
}

//server function
func (gd *GlobalDataHelper) AddGlobalData(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var name, datatype string
	if err := ParseArgs(msg, &name, &datatype); err != nil {
		log.LogError(err)
		return share.ERR_ARGS_ERROR, nil
	}
	gd.AddData(name, datatype)
	return 0, nil
}

func (gd *GlobalDataHelper) OnLoadGlobalData(msg *rpc.Message) {
	ret := GetReplyError(msg)
	if ret == share.ERR_REPLY_FAILED {
		log.LogError("global data is empty")
		if err := gd.createDataSet(); err != nil {
			log.LogError("create global data error:", err)
		}
		return
	}

	if ret != share.ERR_REPLY_SUCCEED {
		log.LogError("global data load error, errcode:", ret)
		return
	}

	var callbackparams share.DBParams
	var savedata share.DbSave
	if err := ParseArgs(msg, &callbackparams, &savedata); err != nil {
		log.LogError("load global data error:", err)
		return
	}

	ent, err := core.CreateFromDb(&savedata)
	if err != nil {
		log.LogError("create global data set failed, err:", err)
	}
	ent.SetInBase(true)
	gd.dataset = ent

	log.LogMessage("load global succeed")
}

func (gd *GlobalDataHelper) OnCreateDataSet(msg *rpc.Message) {
	ret := GetReplyError(msg)
	if ret == share.ERR_REPLY_SUCCEED {
		log.LogMessage("create global data succeed")
		return
	}

	log.LogError("create global data failed:", ret)
}

func (gd *GlobalDataHelper) createDataSet() error {
	if gd.dataset == nil {
		ent, err := core.Create(core.globalset)
		if err != nil {
			return fmt.Errorf("create globat data failed, %s", err.Error())
		}

		ent.SetInBase(true)
		ent.SetDbId(core.GetUid())
		ent.Set("Name", "GlobalData")
		gd.dataset = ent
	}

	db := GetAppByType("database")
	if db == nil {
		return ErrAppNotFound
	}

	return db.CallBack(nil, "Database.SaveObject", gd.OnCreateDataSet, share.GetSaveData(gd.dataset), share.DBParams{})
}

func (gd *GlobalDataHelper) AddData(name string, datatype string) error {
	if gd.dataset == nil {
		return fmt.Errorf("dataset is nil")
	}

	if gd.dataset.GetChildByName(name) != nil {
		return fmt.Errorf("global data(%s) exist", name)
	}

	data, err := core.Create(datatype)
	if err != nil {
		return fmt.Errorf("create global data(%s) failed", datatype)
	}

	data.Set("Name", name)
	_, err = gd.dataset.AddChild(-1, data)
	return err
}

func (gd *GlobalDataHelper) GetData(name string) datatype.Entityer {
	if gd.dataset == nil {
		return nil
	}

	return gd.dataset.GetChildByName(name)
}

func (gd *GlobalDataHelper) OnAppReady(appname string) {
	if gd.isServer {
		app := GetAppByName(appname)
		if app == nil {
			panic("app is nil")
		}

		if _, dup := gd.globalclients[appname]; dup {
			panic("already register")
		}
		gc := &globalClient{}
		gc.appid = app.Id
		gd.globalclients[appname] = gc
	}
}

func (gd *GlobalDataHelper) OnAppLost(appname string) {
	if gd.isServer {
		if _, exist := gd.globalclients[appname]; !exist {
			panic("not have")
		}
		delete(gd.globalclients, appname)
	}
}

func (gd *GlobalDataHelper) OnFrame() {
	if gd.dataset == nil || !gd.isServer {
		return
	}

	for _, client := range gd.globalclients {
		if client.disable { //没有开启
			continue
		}

		if client.status == STATUS_NONE && client.errcount < 3 {
			app := GetAppById(client.appid)
			if app == nil {
				continue
			}

			entityinfo, err := share.GetItemInfo(gd.dataset, true)
			if err != nil {
				panic(err)
			}

			client.dataversion = gd.dataversion
			if app.CallBack(nil, "GlobalHelper.CreateGlobalData", gd.OnCreateGlobalData, gd.dataCenter, entityinfo) != nil {
				client.errcount++
				continue
			}

			client.status = STATUS_CREATE
		}
	}
}

//加载全局数据
func (gd *GlobalDataHelper) LoadGlobalData() error {
	core.AddDispatchNoName(gd, DP_FRAME)

	log.LogMessage("begin load global data")
	db := GetAppByType("database")
	if db == nil {
		return ErrAppNotFound
	}

	return db.CallBack(nil, "Database.LoadObjectByName", gd.OnLoadGlobalData, core.globalset, "GlobalData", share.DBParams{})
}

//增加全局数据
func (kernel *Kernel) AddGlobalData(name string, datatype string) error {
	if core.globalHelper.isServer {
		return core.globalHelper.AddData(name, datatype)
	} else {
		if !core.enableglobaldata {
			return fmt.Errorf("global data is disable, please enable `enableglobaldata`")
		}
		app := GetAppByName(core.globalHelper.dataCenter)
		if app == nil {
			return ErrAppNotFound
		}
		return app.Call(nil, "GlobalHelper.AddGlobalData", name, datatype)
	}
}

//获取全局数据
func (kernel *Kernel) GetGlobalData(name string) datatype.Entityer {
	if !core.globalHelper.isServer && !core.enableglobaldata {
		log.LogError("global data is disable, please enable `enableglobaldata`")
		return nil
	}
	return core.globalHelper.GetData(name)
}
