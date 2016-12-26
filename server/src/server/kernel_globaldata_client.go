package server

import (
	"fmt"
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
)

var (
	ERR_GLOBALDATA_DISABLED          = int32(share.ERR_FUNC_BEGIN + 1)
	ERR_GLOBALDATA_VERSION_NOT_MATCH = int32(share.ERR_FUNC_BEGIN + 2)
	ERR_GLOBALDATA_NOT_FOUND         = int32(share.ERR_FUNC_BEGIN + 3)
)

func (gd *GlobalDataHelper) CreateGlobalData(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	reply = CreateMessage(core.Name)
	var datainfo *datatype.EntityInfo
	if err := ParseArgs(msg, &gd.dataCenter, &datainfo, &gd.dataversion); err != nil {
		return share.ERR_ARGS_ERROR, reply
	}

	if !core.enableglobaldata {
		return ERR_GLOBALDATA_DISABLED, reply
	}

	ent, err := core.kernel.CreateFromArchive(datainfo, nil)
	if err != nil {
		return share.ERR_REPLY_FAILED, reply
	}

	if gd.dataset != nil {
		core.kernel.Destroy(gd.dataset.ObjectId())
	}
	gd.dataset = ent
	log.LogMessage("create global data succeed, version:", gd.dataversion)
	core.apper.OnGlobalDataCreated()
	return share.ERR_REPLY_SUCCEED, reply
}

func (gd *GlobalDataHelper) GlobalDataAddData(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	reply = CreateMessage(core.Name)
	var cindex int
	var version int64
	var datainfo *datatype.EntityInfo
	if err := ParseArgs(msg, &cindex, &version, &datainfo); err != nil {
		return share.ERR_ARGS_ERROR, reply
	}

	if !core.enableglobaldata {
		return ERR_GLOBALDATA_DISABLED, reply
	}

	if gd.dataversion+1 != version {
		return ERR_GLOBALDATA_VERSION_NOT_MATCH, reply
	}

	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_FOUND, reply
	}

	ent, err := core.kernel.CreateFromArchive(datainfo, nil)
	if err != nil {
		return share.ERR_FUNC_BEGIN + 4, reply
	}

	_, err = gd.dataset.AddChild(cindex, ent)
	if err != nil {
		return share.ERR_FUNC_BEGIN + 5, reply
	}

	AppendArgs(reply, version)
	gd.dataversion = version
	log.LogMessage("update globaldata,current version:", gd.dataversion)
	return share.ERR_REPLY_SUCCEED, reply

}

func (gd *GlobalDataHelper) GlobalDataUpdate(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	reply = CreateMessage(core.Name)
	var cindex int
	var version int64
	var pindex int16
	var val datatype.Any
	if err := ParseArgs(msg, &cindex, &version, &pindex, &val); err != nil {
		return share.ERR_ARGS_ERROR, reply
	}

	if !core.enableglobaldata {
		return ERR_GLOBALDATA_DISABLED, reply
	}

	if gd.dataversion+1 != version {
		return ERR_GLOBALDATA_VERSION_NOT_MATCH, reply
	}

	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_FOUND, reply
	}

	cld := gd.dataset.GetChild(cindex)
	if cld == nil {
		return share.ERR_FUNC_BEGIN + 4, reply
	}

	if err := cld.SetByIndex(pindex, val.Val); err != nil {
		return share.ERR_FUNC_BEGIN + 5, reply
	}

	AppendArgs(reply, version)
	gd.dataversion = version
	log.LogMessage("update globaldata,current version:", gd.dataversion)
	return share.ERR_REPLY_SUCCEED, reply

}

func (gd *GlobalDataHelper) GlobalDataRecAppend(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	reply = CreateMessage(core.Name)
	var cindex int
	var version int64
	var recname string
	var row int
	var rowdata []byte
	if err := ParseArgs(msg, &cindex, &version, &recname, &row, &rowdata); err != nil {
		return share.ERR_ARGS_ERROR, reply
	}

	if !core.enableglobaldata {
		return ERR_GLOBALDATA_DISABLED, reply
	}

	if gd.dataversion+1 != version {
		return ERR_GLOBALDATA_VERSION_NOT_MATCH, reply
	}

	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_FOUND, reply
	}

	cld := gd.dataset.GetChild(cindex)
	if cld == nil {
		return share.ERR_FUNC_BEGIN + 4, reply
	}

	rec := cld.FindRec(recname)
	if rec == nil {
		return share.ERR_FUNC_BEGIN + 5, reply
	}

	if -1 == rec.AddByBytes(row, rowdata) {
		return share.ERR_FUNC_BEGIN + 6, reply
	}

	AppendArgs(reply, version)
	gd.dataversion = version
	log.LogMessage("update globaldata,current version:", gd.dataversion)
	return share.ERR_REPLY_SUCCEED, reply

}

func (gd *GlobalDataHelper) GlobalDataRecDelete(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	reply = CreateMessage(core.Name)
	var cindex int
	var version int64
	var recname string
	var row int

	if err := ParseArgs(msg, &cindex, &version, &recname, &row); err != nil {
		return share.ERR_ARGS_ERROR, reply
	}

	if !core.enableglobaldata {
		return ERR_GLOBALDATA_DISABLED, reply
	}

	if gd.dataversion+1 != version {
		return ERR_GLOBALDATA_VERSION_NOT_MATCH, reply
	}

	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_FOUND, reply
	}

	cld := gd.dataset.GetChild(cindex)
	if cld == nil {
		return share.ERR_FUNC_BEGIN + 4, reply
	}

	rec := cld.FindRec(recname)
	if rec == nil {
		return share.ERR_FUNC_BEGIN + 5, reply
	}

	rec.Del(row)
	AppendArgs(reply, version)
	gd.dataversion = version
	log.LogMessage("update globaldata,current version:", gd.dataversion)
	return share.ERR_REPLY_SUCCEED, reply

}

func (gd *GlobalDataHelper) GlobalDataRecClear(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	reply = CreateMessage(core.Name)
	var cindex int
	var version int64
	var recname string

	if err := ParseArgs(msg, &cindex, &version, &recname); err != nil {
		return share.ERR_ARGS_ERROR, reply
	}

	if !core.enableglobaldata {
		return ERR_GLOBALDATA_DISABLED, reply
	}

	if gd.dataversion+1 != version {
		return ERR_GLOBALDATA_VERSION_NOT_MATCH, reply
	}

	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_FOUND, reply
	}

	cld := gd.dataset.GetChild(cindex)
	if cld == nil {
		return share.ERR_FUNC_BEGIN + 4, reply
	}

	rec := cld.FindRec(recname)
	if rec == nil {
		return share.ERR_FUNC_BEGIN + 5, reply
	}

	rec.Clear()
	AppendArgs(reply, version)
	gd.dataversion = version
	log.LogMessage("update globaldata,current version:", gd.dataversion)
	return share.ERR_REPLY_SUCCEED, reply
}

func (gd *GlobalDataHelper) GlobalDataRecModify(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	reply = CreateMessage(core.Name)
	var cindex int
	var version int64
	var recname string
	var row, col int
	var val datatype.Any
	if err := ParseArgs(msg, &cindex, &version, &recname, &row, &col, &val); err != nil {
		return share.ERR_ARGS_ERROR, reply
	}

	if !core.enableglobaldata {
		return ERR_GLOBALDATA_DISABLED, reply
	}

	if gd.dataversion+1 != version {
		return ERR_GLOBALDATA_VERSION_NOT_MATCH, reply
	}

	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_FOUND, reply
	}

	cld := gd.dataset.GetChild(cindex)
	if cld == nil {
		return share.ERR_FUNC_BEGIN + 4, reply
	}

	rec := cld.FindRec(recname)
	if rec == nil {
		return share.ERR_FUNC_BEGIN + 5, reply
	}

	if err := rec.Set(row, col, val.Val); err != nil {
		return share.ERR_FUNC_BEGIN + 6, reply
	}
	AppendArgs(reply, version)
	gd.dataversion = version
	log.LogMessage("update globaldata,current version:", gd.dataversion)
	return share.ERR_REPLY_SUCCEED, reply
}

func (gd *GlobalDataHelper) GlobalDataRecSetRow(sender rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	reply = CreateMessage(core.Name)
	var cindex int
	var version int64
	var recname string
	var row int
	var rowdata []byte
	if err := ParseArgs(msg, &cindex, &version, &recname, &row, &rowdata); err != nil {
		return share.ERR_ARGS_ERROR, reply
	}

	if !core.enableglobaldata {
		return ERR_GLOBALDATA_DISABLED, reply
	}

	if gd.dataversion+1 != version {
		return ERR_GLOBALDATA_VERSION_NOT_MATCH, reply
	}

	if gd.dataset == nil {
		return ERR_GLOBALDATA_NOT_FOUND, reply
	}

	cld := gd.dataset.GetChild(cindex)
	if cld == nil {
		return share.ERR_FUNC_BEGIN + 4, reply
	}

	rec := cld.FindRec(recname)
	if rec == nil {
		return share.ERR_FUNC_BEGIN + 5, reply
	}

	if err := rec.SetRowByBytes(row, rowdata); err != nil {
		log.LogError(err)
		return share.ERR_FUNC_BEGIN + 6, reply
	}

	AppendArgs(reply, version)
	gd.dataversion = version
	log.LogMessage("update globaldata,current version:", gd.dataversion)
	return share.ERR_REPLY_SUCCEED, reply

}

func sendGlobalDataAdd(name string, datatype string) error {
	if !core.enableglobaldata {
		return fmt.Errorf("global data is disable, please enable `enableglobaldata`")
	}
	app := GetAppByName(core.globalHelper.dataCenter)
	if app == nil {
		return ErrAppNotFound
	}
	return app.Call(nil, "GlobalHelper.GlobalDataAdd", name, datatype)
}

func sendGlobalDataSet(index int, name string, val interface{}) error {
	if !core.enableglobaldata {
		return fmt.Errorf("global data is disable, please enable `enableglobaldata`")
	}

	app := GetAppByName(core.globalHelper.dataCenter)
	if app == nil {
		return ErrAppNotFound
	}

	return app.Call(nil, "GlobalHelper.GlobalDataSet", index, name, datatype.NewAny(0, val))
}

func sendGlobalDataSetGrid(index int, rec string, row, col int, val interface{}) error {
	if !core.enableglobaldata {
		return fmt.Errorf("global data is disable, please enable `enableglobaldata`")
	}

	app := GetAppByName(core.globalHelper.dataCenter)
	if app == nil {
		return ErrAppNotFound
	}

	return app.Call(nil, "GlobalHelper.GlobalDataSetGrid", index, rec, row, col, datatype.NewAny(0, val))
}

func sendGlobalDataSetRow(index int, rec string, row int, args ...interface{}) error {
	if !core.enableglobaldata {
		return fmt.Errorf("global data is disable, please enable `enableglobaldata`")
	}

	app := GetAppByName(core.globalHelper.dataCenter)
	if app == nil {
		return ErrAppNotFound
	}

	sendargs := make([]datatype.Any, len(args))
	for k, v := range args {
		sendargs[k].Val = v
	}
	return app.Call(nil, "GlobalHelper.GlobalDataSetRow", index, rec, row, sendargs)
}

func sendGlobalDataAddRow(index int, rec string, row int) error {
	if !core.enableglobaldata {
		return fmt.Errorf("global data is disable, please enable `enableglobaldata`")
	}

	app := GetAppByName(core.globalHelper.dataCenter)
	if app == nil {
		return ErrAppNotFound
	}

	return app.Call(nil, "GlobalHelper.GlobalDataAddRow", index, rec, row)
}

func sendGlobalDataAddRowValues(index int, rec string, row int, args ...interface{}) error {
	if !core.enableglobaldata {
		return fmt.Errorf("global data is disable, please enable `enableglobaldata`")
	}

	app := GetAppByName(core.globalHelper.dataCenter)
	if app == nil {
		return ErrAppNotFound
	}

	sendargs := make([]datatype.Any, len(args))
	for k, v := range args {
		sendargs[k].Val = v
	}
	return app.Call(nil, "GlobalHelper.GlobalDataAddRowValues", index, rec, row, sendargs)
}

func sendGlobalDataDelRow(index int, rec string, row int) error {
	if !core.enableglobaldata {
		return fmt.Errorf("global data is disable, please enable `enableglobaldata`")
	}

	app := GetAppByName(core.globalHelper.dataCenter)
	if app == nil {
		return ErrAppNotFound
	}

	return app.Call(nil, "GlobalHelper.GlobalDataDelRow", index, rec, row)
}

func sendGlobalDataClearRecord(index int, rec string) error {
	if !core.enableglobaldata {
		return fmt.Errorf("global data is disable, please enable `enableglobaldata`")
	}

	app := GetAppByName(core.globalHelper.dataCenter)
	if app == nil {
		return ErrAppNotFound
	}

	return app.Call(nil, "GlobalHelper.GlobalDataClearRecord", index, rec)
}
