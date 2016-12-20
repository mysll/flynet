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
	STATUS_VERSION_ERROR
)

var (
	ERR_GLOBALDATA_NOT_CREATE = fmt.Errorf("global data not found")
	ASYNCOP                   = -2
)

type globalClient struct {
	appid       int32
	status      int32
	dataversion int64
	errcount    int32
	disable     bool
}

type GlobalDataHelper struct {
	Callee
	Dispatch
	dataset       datatype.Entity
	dataCenter    string
	isServer      bool
	globalclients map[string]*globalClient
	dataversion   int64
	wait          chan struct{}
	ready         bool
	isnew         bool
}

func NewGlobalDataHelper() *GlobalDataHelper {
	h := &GlobalDataHelper{}
	return h
}

func (gd *GlobalDataHelper) SetServer() {
	gd.globalclients = make(map[string]*globalClient)
	gd.dataCenter = core.Name
	gd.isServer = true
	RegisterCallee("GlobalSet", gd)
}

func (gd *GlobalDataHelper) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("CreateGlobalData", gd.CreateGlobalData)
	//数据操作
	s.RegisterCallback("GlobalDataAdd", gd.GlobalDataAdd)
	s.RegisterCallback("GlobalDataSet", gd.GlobalDataSet)
	s.RegisterCallback("GlobalDataSetGrid", gd.GlobalDataSetGrid)
	s.RegisterCallback("GlobalDataSetRow", gd.GlobalDataSetRow)
	s.RegisterCallback("GlobalDataAddRow", gd.GlobalDataAddRow)
	s.RegisterCallback("GlobalDataAddRowValues", gd.GlobalDataAddRowValues)
	s.RegisterCallback("GlobalDataDelRow", gd.GlobalDataDelRow)
	s.RegisterCallback("GlobalDataClearRecord", gd.GlobalDataClearRecord)

	//更新
	s.RegisterCallback("GlobalDataAddData", gd.GlobalDataAddData)
	s.RegisterCallback("GlobalDataUpdate", gd.GlobalDataUpdate)
	s.RegisterCallback("GlobalDataRecAppend", gd.GlobalDataRecAppend)
	s.RegisterCallback("GlobalDataRecDelete", gd.GlobalDataRecDelete)
	s.RegisterCallback("GlobalDataRecClear", gd.GlobalDataRecClear)
	s.RegisterCallback("GlobalDataRecModify", gd.GlobalDataRecModify)
	s.RegisterCallback("GlobalDataRecSetRow", gd.GlobalDataRecSetRow)
}

//rpc function
func (gd *GlobalDataHelper) GlobalDataAdd(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var name, datatype string
	if err := ParseArgs(msg, &name, &datatype); err != nil {
		log.LogError(err)
		return share.ERR_ARGS_ERROR, nil
	}
	gd.addData(name, datatype)
	return 0, nil
}

func (gd *GlobalDataHelper) GlobalDataSet(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var index int
	var name string
	var val datatype.Any
	if err := ParseArgs(msg, &index, &name, &val); err != nil {
		log.LogError(err)
		return share.ERR_ARGS_ERROR, nil
	}

	err := gd.Set(index, name, val.Val)
	if err != nil {
		log.LogError(err)
	}

	return 0, nil
}

func (gd *GlobalDataHelper) GlobalDataSetGrid(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var index int
	var name string
	var row, col int
	var val datatype.Any
	if err := ParseArgs(msg, &index, &name, &row, &col, &val); err != nil {
		log.LogError(err)
		return share.ERR_ARGS_ERROR, nil
	}

	err := gd.SetGrid(index, name, row, col, val.Val)
	if err != nil {
		log.LogError(err)
	}

	return 0, nil
}

func (gd *GlobalDataHelper) GlobalDataSetRow(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var index int
	var name string
	var row int
	var val []datatype.Any
	if err := ParseArgs(msg, &index, &name, &row, &val); err != nil {
		log.LogError(err)
		return share.ERR_ARGS_ERROR, nil
	}

	args := make([]interface{}, len(val))
	for k, v := range val {
		args[k] = v.Val
	}

	err := gd.SetRow(index, name, row, args...)
	if err != nil {
		log.LogError(err)
	}

	return 0, nil
}

func (gd *GlobalDataHelper) GlobalDataAddRow(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var index int
	var name string
	var row int
	if err := ParseArgs(msg, &index, &name, &row); err != nil {
		log.LogError(err)
		return share.ERR_ARGS_ERROR, nil
	}

	index, err := gd.AddRow(index, name, row)
	if err != nil {
		log.LogError(err)
	}

	if index == -1 {
		log.LogError("add record row failed")
	}

	return 0, nil
}

func (gd *GlobalDataHelper) GlobalDataAddRowValues(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var index int
	var name string
	var row int
	var val []datatype.Any
	if err := ParseArgs(msg, &index, &name, &row, &val); err != nil {
		log.LogError(err)
		return share.ERR_ARGS_ERROR, nil
	}

	args := make([]interface{}, len(val))
	for k, v := range val {
		args[k] = v.Val
	}
	index, err := gd.AddRowValues(index, name, row, args...)
	if err != nil {
		log.LogError(err)
	}

	if index == -1 {
		log.LogError("add record row failed")
	}

	return 0, nil
}

func (gd *GlobalDataHelper) GlobalDataDelRow(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var index int
	var name string
	var row int
	if err := ParseArgs(msg, &index, &name, &row); err != nil {
		log.LogError(err)
		return share.ERR_ARGS_ERROR, nil
	}

	err := gd.DelRow(index, name, row)
	if err != nil {
		log.LogError(err)
	}
	return 0, nil
}

func (gd *GlobalDataHelper) GlobalDataClearRecord(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var index int
	var name string
	if err := ParseArgs(msg, &index, &name); err != nil {
		log.LogError(err)
		return share.ERR_ARGS_ERROR, nil
	}

	err := gd.ClearRecord(index, name)
	if err != nil {
		log.LogError(err)
	}

	return 0, nil
}

//callback
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

//数据库加载回调
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
	if gd.dataset != nil {
		core.Destroy(gd.dataset.ObjectId())
	}
	gd.dataset = ent
	gd.ready = true
	gd.OnDataReady()
	log.LogMessage("load global succeed")
}

func (gd *GlobalDataHelper) OnDataChanged(msg *rpc.Message) {
	var appname string
	var version int64

	ret := GetReplyError(msg)
	if ret == share.ERR_REPLY_SUCCEED {
		if err := ParseArgs(msg, &appname, &version); err != nil {
			log.LogError("parse args failed: ", err)
			return
		}

		client, exist := gd.globalclients[appname]
		if !exist {
			return
		}

		client.dataversion = version
		return
	}

	//失败
	if err := ParseArgs(msg, &appname); err != nil {
		log.LogError("parse args failed: ", err)
		return
	}

	if ret == ERR_GLOBALDATA_NOT_FOUND { //版本错误
		client, exist := gd.globalclients[appname]
		if !exist {
			return
		}

		client.status = STATUS_VERSION_ERROR
	}

	log.LogError(appname, " global data sync failed, errcode:", ret)
}

func (gd *GlobalDataHelper) OnSaveDataSet(msg *rpc.Message) {
	if gd.wait != nil {
		close(gd.wait)
		gd.wait = nil
	}
	ret := GetReplyError(msg)
	if ret == share.ERR_REPLY_SUCCEED {
		gd.isnew = false
		log.LogMessage("save global data succeed")
		if !gd.ready {
			gd.ready = true
			gd.OnDataReady()
		}
		return
	}

	log.LogError("save global data failed:", ret)
}

//operater function
//创建数据集合
func (gd *GlobalDataHelper) createDataSet() error {
	if gd.dataset == nil {
		ent, err := core.CreateContainer(core.globalset, core.maxglobalentry)
		if err != nil {
			return fmt.Errorf("create global data failed, %s", err.Error())
		}

		ent.SetDBId(core.GetUid())
		ent.Set("Name", "GlobalData")
		gd.dataset = ent
		gd.isnew = true
	}

	db := GetAppByType("database")
	if db == nil {
		return ErrAppNotFound
	}

	return db.CallBack(nil, "Database.SaveObject", gd.OnSaveDataSet, share.GetSaveData(gd.dataset), share.DBParams{})
}

func (gd *GlobalDataHelper) SaveData(wait chan struct{}, quit bool) error {
	if gd.dataset == nil {
		return nil
	}
	db := GetAppByType("database")
	if db == nil {
		return ErrAppNotFound
	}

	if wait != nil {
		gd.wait = wait
	}

	core.apper.OnPerSaveGlobalData(quit)
	return db.CallBack(nil, "Database.UpdateObject", gd.OnSaveDataSet, share.GetSaveData(gd.dataset), share.DBParams{})
}

//增加一个全局数据
func (gd *GlobalDataHelper) addData(name string, datatype string) error {
	if gd.dataset == nil {
		return fmt.Errorf("dataset is nil")
	}

	if gd.dataset.FindChildByName(name) != nil {
		return fmt.Errorf("global data(%s) exist", name)
	}

	data, err := core.Create(datatype)
	if err != nil {
		return fmt.Errorf("create global data(%s) failed", datatype)
	}

	//消除ERROR
	data.SetInBase(true)
	data.Set("Name", name)

	index, err := core.AddChild(gd.dataset.ObjectId(), data.ObjectId(), -1)
	if err == nil {

		entityinfo, err := share.GetItemInfo(data, false)
		if err != nil {
			return err
		}

		v := gd.dataChange()
		for _, client := range gd.globalclients {
			if client.disable || client.status == STATUS_NONE { //没有开启
				continue
			}

			app := GetAppById(client.appid)
			if app == nil {
				continue
			}

			app.CallBack(nil, "GlobalHelper.GlobalDataAddData", gd.OnDataChanged, index, v, entityinfo)
		}
	}
	return err
}

//获取全局数据
func (gd *GlobalDataHelper) getData(name string) datatype.Entity {
	if gd.dataset == nil {
		return nil
	}

	return gd.dataset.FindChildByName(name)
}

func (gd *GlobalDataHelper) OnDataReady() {
	core.apper.OnGlobalDataLoaded()
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
		gc.disable = !app.EnableGlobalData
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

		if (client.status == STATUS_NONE || client.status == STATUS_VERSION_ERROR) && client.errcount < 3 {
			app := GetAppById(client.appid)
			if app == nil {
				continue
			}

			entityinfo, err := share.GetItemInfo(gd.dataset, true)
			if err != nil {
				panic(err)
			}

			client.dataversion = gd.dataversion
			if app.CallBack(nil, "GlobalHelper.CreateGlobalData", gd.OnCreateGlobalData, gd.dataCenter, entityinfo, gd.dataversion) != nil {
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

func (gd *GlobalDataHelper) dataChange() int64 {
	gd.dataversion++
	return gd.dataversion
}

//数据属性变动同步
func (gd *GlobalDataHelper) Update(self datatype.Entity, index int16, value interface{}) {
	if !gd.isServer {
		return
	}
	v := gd.dataChange()

	for _, client := range gd.globalclients {
		if client.disable || client.status == STATUS_NONE { //没有开启
			continue
		}

		app := GetAppById(client.appid)
		if app == nil {
			continue
		}

		app.CallBack(nil, "GlobalHelper.GlobalDataUpdate", gd.OnDataChanged, self.ChildIndex(), v, index, datatype.NewAny(0, value))
	}
}

//表格变动同步
func (gd *GlobalDataHelper) RecAppend(self datatype.Entity, rec datatype.Record, row int) {
	if !gd.isServer {
		return
	}

	v := gd.dataChange()

	for _, client := range gd.globalclients {
		if client.disable || client.status == STATUS_NONE { //没有开启
			continue
		}

		app := GetAppById(client.appid)
		if app == nil {
			continue
		}

		rowvalues, _ := rec.SerialRow(row)
		app.CallBack(nil, "GlobalHelper.GlobalDataRecAppend", gd.OnDataChanged, self.ChildIndex(), v, rec.Name(), row, rowvalues)
	}
}

func (gd *GlobalDataHelper) RecDelete(self datatype.Entity, rec datatype.Record, row int) {
	if !gd.isServer {
		return
	}

	v := gd.dataChange()

	for _, client := range gd.globalclients {
		if client.disable || client.status == STATUS_NONE { //没有开启
			continue
		}

		app := GetAppById(client.appid)
		if app == nil {
			continue
		}

		app.CallBack(nil, "GlobalHelper.GlobalDataRecDelete", gd.OnDataChanged, self.ChildIndex(), v, rec.Name(), row)
	}
}

func (gd *GlobalDataHelper) RecClear(self datatype.Entity, rec datatype.Record) {
	if !gd.isServer {
		return
	}
	v := gd.dataChange()

	for _, client := range gd.globalclients {
		if client.disable || client.status == STATUS_NONE { //没有开启
			continue
		}

		app := GetAppById(client.appid)
		if app == nil {
			continue
		}

		app.CallBack(nil, "GlobalHelper.GlobalDataRecClear", gd.OnDataChanged, self.ChildIndex(), v, rec.Name())
	}
}

func (gd *GlobalDataHelper) RecModify(self datatype.Entity, rec datatype.Record, row, col int) {
	if !gd.isServer {
		return
	}
	v := gd.dataChange()

	for _, client := range gd.globalclients {
		if client.disable || client.status == STATUS_NONE { //没有开启
			continue
		}

		app := GetAppById(client.appid)
		if app == nil {
			continue
		}

		val, _ := rec.Get(row, col)
		app.CallBack(nil, "GlobalHelper.GlobalDataRecModify", gd.OnDataChanged, self.ChildIndex(), v, rec.Name(), row, col, datatype.NewAny(0, val))
	}
}

func (gd *GlobalDataHelper) RecSetRow(self datatype.Entity, rec datatype.Record, row int) {
	if !gd.isServer {
		return
	}

	v := gd.dataChange()

	for _, client := range gd.globalclients {
		if client.disable || client.status == STATUS_NONE { //没有开启
			continue
		}

		app := GetAppById(client.appid)
		if app == nil {
			continue
		}

		rowvalues, _ := rec.SerialRow(row)

		app.CallBack(nil, "GlobalHelper.GlobalDataRecSetRow", gd.OnDataChanged, self.ChildIndex(), v, rec.Name(), row, rowvalues)
	}

}

func (gd *GlobalDataHelper) OnAfterAdd(self datatype.Entity, sender datatype.Entity, index int) int {
	if !gd.isServer {
		return 1
	}

	sender.SetPropUpdate(gd)
	recs := sender.RecordNames()
	for _, v := range recs {
		rec := sender.FindRec(v)
		if rec.IsVisible() {
			rec.SetMonitor(gd)
		}
	}
	return 1
}

func (gd *GlobalDataHelper) OnRemove(self datatype.Entity, sender datatype.Entity, index int) int {
	if !gd.isServer {
		return 1
	}
	sender.SetPropUpdate(nil)
	recs := sender.RecordNames()
	for _, v := range recs {
		rec := sender.FindRec(v)
		if rec.IsVisible() {
			rec.SetMonitor(nil)
		}
	}
	return 1
}

// 数据操作API
//查找全局数据的索引
func (gd *GlobalDataHelper) FindGlobalData(name string) int {
	if gd.dataset == nil {
		return -1
	}

	c := gd.dataset.FindChildByName(name)
	if c == nil {
		return -1
	}

	return c.ChildIndex()
}

func (gd *GlobalDataHelper) Set(index int, name string, val interface{}) error {
	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_CREATE
	}

	data := gd.dataset.GetChild(index)
	if data == nil {
		return fmt.Errorf("index(%d) not found", index)
	}

	return data.Set(name, val)
}

//设置单元格值
func (gd *GlobalDataHelper) SetGrid(index int, rec string, row, col int, val interface{}) error {
	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_CREATE
	}

	data := gd.dataset.GetChild(index)
	if data == nil {
		return fmt.Errorf("index(%d) not found", index)
	}

	record := data.FindRec(rec)
	if record == nil {
		return fmt.Errorf("index(%d) record(%s) not found", index, rec)
	}

	return record.Set(row, col, val)
}

//设置一行的值
func (gd *GlobalDataHelper) SetRow(index int, rec string, row int, args ...interface{}) error {
	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_CREATE
	}

	data := gd.dataset.GetChild(index)
	if data == nil {
		return fmt.Errorf("index(%d) not found", index)
	}

	record := data.FindRec(rec)
	if record == nil {
		return fmt.Errorf("index(%d) record(%s) not found", index, rec)
	}

	return record.SetRow(row, args...)
}

//增加一行数据,row插入的位置，-1表示插入在最后
func (gd *GlobalDataHelper) AddRowValues(index int, rec string, row int, args ...interface{}) (int, error) {
	if gd.dataset == nil {
		return -1, ERR_GLOBALDATA_NOT_CREATE
	}

	data := gd.dataset.GetChild(index)
	if data == nil {
		return -1, fmt.Errorf("index(%d) not found", index)
	}

	record := data.FindRec(rec)
	if record == nil {
		return -1, fmt.Errorf("index(%d) record(%s) not found", index, rec)
	}

	return record.Add(row, args...), nil
}

//增加一行
func (gd *GlobalDataHelper) AddRow(index int, rec string, row int) (int, error) {
	if gd.dataset == nil {
		return -1, ERR_GLOBALDATA_NOT_CREATE
	}

	data := gd.dataset.GetChild(index)
	if data == nil {
		return -1, fmt.Errorf("index(%d) not found", index)
	}

	record := data.FindRec(rec)
	if record == nil {
		return -1, fmt.Errorf("index(%d) record(%s) not found", index, rec)
	}

	return record.AddRow(row), nil
}

//删除一行
func (gd *GlobalDataHelper) DelRow(index int, rec string, row int) error {
	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_CREATE
	}

	data := gd.dataset.GetChild(index)
	if data == nil {
		return fmt.Errorf("index(%d) not found", index)
	}

	record := data.FindRec(rec)
	if record == nil {
		return fmt.Errorf("index(%d) record(%s) not found", index, rec)
	}

	record.Del(row)
	return nil
}

//清除表格内容
func (gd *GlobalDataHelper) ClearRecord(index int, rec string) error {
	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_CREATE
	}

	data := gd.dataset.GetChild(index)
	if data == nil {
		return fmt.Errorf("index(%d) not found", index)
	}

	record := data.FindRec(rec)
	if record == nil {
		return fmt.Errorf("index(%d) record(%s) not found", index, rec)
	}

	record.Clear()
	return nil
}

//kernel API
//增加全局数据
func (kernel *Kernel) AddGlobalData(name string, datatype string) error {
	if core.globalHelper.isServer {
		return core.globalHelper.addData(name, datatype)
	} else {
		return sendGlobalDataAdd(name, datatype)
	}
}

//获取全局数据
func (kernel *Kernel) GetGlobalData(name string) datatype.Entity {
	if !core.globalHelper.isServer && !core.enableglobaldata {
		log.LogError("global data is disable, please enable `enableglobaldata`")
		return nil
	}
	return core.globalHelper.getData(name)
}

//查找globaldata的索引
func (kernel *Kernel) FindGlobalData(name string) int {
	return core.globalHelper.FindGlobalData(name)
}

//保存全局数据
func (kernel *Kernel) SaveGlobalData(wait bool, quit bool) error {
	if core.globalHelper.isServer {
		if wait {
			ch := make(chan struct{})
			if err := core.globalHelper.SaveData(ch, quit); err != nil {
				return err
			}
			<-ch
			return nil
		}
		return core.globalHelper.SaveData(nil, quit)
	}
	return nil
}

//数据操作API
func (kernel *Kernel) GlobalDataSet(index int, name string, val interface{}) error {
	if core.globalHelper.isServer {
		return core.globalHelper.Set(index, name, val)
	}

	return sendGlobalDataSet(index, name, val)
}

func (kernel *Kernel) GlobalDataSetGrid(index int, rec string, row, col int, val interface{}) error {
	if core.globalHelper.isServer {
		return core.globalHelper.SetGrid(index, rec, row, col, val)
	}

	return sendGlobalDataSetGrid(index, rec, row, col, val)
}

func (kernel *Kernel) GlobalDataSetRow(index int, rec string, row int, args ...interface{}) error {
	if core.globalHelper.isServer {
		return core.globalHelper.SetRow(index, rec, row, args...)
	}
	return sendGlobalDataSetRow(index, rec, row, args...)
}

func (kernel *Kernel) GlobalDataAddRow(index int, rec string, row int) (int, error) {
	if core.globalHelper.isServer {
		return core.globalHelper.AddRow(index, rec, row)
	}
	return ASYNCOP, sendGlobalDataAddRow(index, rec, row)
}

func (kernel *Kernel) GlobalDataAddRowValues(index int, rec string, row int, args ...interface{}) (int, error) {
	if core.globalHelper.isServer {
		return core.globalHelper.AddRowValues(index, rec, row, args...)
	}

	return ASYNCOP, sendGlobalDataAddRowValues(index, rec, row, args...)
}

func (kernel *Kernel) GlobalDataDelRow(index int, rec string, row int) error {
	if core.globalHelper.isServer {
		return core.globalHelper.DelRow(index, rec, row)
	}

	return sendGlobalDataDelRow(index, rec, row)
}

func (kernel *Kernel) GlobalDataClearRecord(index int, rec string) error {
	if core.globalHelper.isServer {
		return core.globalHelper.ClearRecord(index, rec)
	}
	return sendGlobalDataClearRecord(index, rec)
}
