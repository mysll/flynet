package server

import (
	"libs/rpc"
	"share"
)

//数据库操作封装
type DBWarp struct {
	db *RemoteApp
}

func NewDBWarp(db *RemoteApp) DBWarp {
	if db == nil {
		panic("db is nil")
	}

	if db.Type != "database" {
		panic("is not database")
	}

	return DBWarp{db}
}

//删除信件
func (warp DBWarp) RecvLetter(mailbox *rpc.Mailbox, uid uint64, serial_no uint64, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.RecvLetter", uid, serial_no, callback, callbackparams)
}

//获取信件数量
func (warp DBWarp) QueryLetter(mailbox *rpc.Mailbox, uid uint64, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.QueryLetter", uid, callback, callbackparams)
}

//查看信件
func (warp DBWarp) LookLetter(mailbox *rpc.Mailbox, uid uint64, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.LookLetter", uid, callback, callbackparams)
}

//发送系统邮件，发件人名称， 收件人帐号，收件人角色名，信件类型，信件标题，信件内容，附件
func (warp DBWarp) SendSystemLetter(mailbox *rpc.Mailbox, source_name, recvacc, recvrole string, typ int, title, content, appendix string, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.SendSystemLetter", source_name, recvacc, recvrole, typ, title, content, appendix, callback, callbackparams)
}

//获取某个表的记录行数
//示例：warp.Count(nil, "test1", "`id`=1", "DBWarp.CountBack", share.DBParams{})
//      返回test1表中id=1的行个数
func (warp DBWarp) Count(mailbox *rpc.Mailbox, tbl string, condition string, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.Count", tbl, condition, callback, callbackparams)
}

//查询记录回调函数格式
func (warp DBWarp) CountBack(mailbox rpc.Mailbox, callbackparams share.DBParams, count int32) error {
	return nil
}

//插入多条记录
//示例：warp.InsertRows(nil, "test1", []string{"id", "value"}, []interface{}{1,"test1", 2, "test2"}, "DBWarp.InsertRowsBack", share.DBParams{})
//      向test1表中插入两条记录
func (warp DBWarp) InsertRows(mailbox *rpc.Mailbox, tbl string, keys []string, values []interface{}, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.InsertRows", tbl, keys, values, callback, callbackparams)
}

//插入多条记录的回调函数
func (warp DBWarp) InsertRowsBack(mailbox rpc.Mailbox, callbackparams share.DBParams, eff int, err string) error {
	return nil
}

//插入一条记录
//示例：warp.InsertRow(nil, "test1", map[string]interface{"id":1, "value":"test1"}, "DBWarp.InsertRowBack", share.DBParams{})
//		向test1表中插入一条记录
func (warp DBWarp) InsertRow(mailbox *rpc.Mailbox, tbl string, values map[string]interface{}, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.InsertRow", tbl, values, callback, callbackparams)
}

//插入一条记录的回调函数
func (warp DBWarp) InsertRowBack(mailbox rpc.Mailbox, callbackparams share.DBParams, eff int, err string) error {
	return nil
}

//删除记录
//示例：warp.DeleteRow(nil, "test1", "`id`=1", "DBWarp.DeleteRowBack", share.DBParams{})
//      删除test1表中，id=1的记录
func (warp DBWarp) DeleteRow(mailbox *rpc.Mailbox, tbl string, condition string, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.DeleteRow", tbl, condition, callback, callbackparams)
}

//删除回调的格式
func (warp DBWarp) DeleteRowBack(mailbox rpc.Mailbox, callbackparams share.DBParams, eff int, err string) error {
	return nil
}

//更新一条记录
//示例：warp.UpdateRow(nil, "test1", map[string]interface{}{"value":"test3"}, "`id`=1", "DBWarp.UpdateRowBack", share.DBParams{})
//      更新id=1的记录，设置value字段为test3
func (warp DBWarp) UpdateRow(mailbox *rpc.Mailbox, tbl string, values map[string]interface{}, condition string, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.UpdateRow", tbl, values, condition, callback, callbackparams)
}

//更新一条记录的回调函数
func (warp DBWarp) UpdateRowBack(mailbox rpc.Mailbox, callbackparams share.DBParams, eff int, err string) error {
	return nil
}

//查询一条记录
//示例：warp.QueryRow(nil, "test1", "`id`=1", "", "DBWarp.QueryRowBack", share.DBParams{})
//      查询id=1的一条记录
func (warp DBWarp) QueryRow(mailbox *rpc.Mailbox, tbl string, condition string, orderby string, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.QueryRow", tbl, condition, orderby, callback, callbackparams)
}

//查询一条记录的回调函数
func (warp DBWarp) QueryRowBack(mailbox rpc.Mailbox, callbackparams share.DBParams, result share.DBRow) error {
	return nil
}

//查询多条记录
//示例：warp.QueryRows(nil, "test1", "", "`id` DESC", 0, 10, "DBWarp.QueryRowsBack", share.DBParams{})
//      查询test1表中前10条记录，按id进行降序排列
func (warp DBWarp) QueryRows(mailbox *rpc.Mailbox, tbl string, condition string, orderby string, index int, count int, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.QueryRows", tbl, condition, orderby, index, count, callback, callbackparams)
}

//查询多条记录的回调函数
func (warp DBWarp) QueryRowsBack(mailbox rpc.Mailbox, callbackparams share.DBParams, result []share.DBRow) error {
	return nil
}

//查询sql语句
//示例：warp.QuerySql(nil, "SELECT * FROM `test1` ORDERBY `id` DESC LIMIT 0 10")
//      查询test1表中前10条记录，按id进行降序排列
func (warp DBWarp) QuerySql(mailbox *rpc.Mailbox, sqlstr string, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.QuerySql", sqlstr, callback, callbackparams)
}

//查询sql的回调函数
func (warp DBWarp) QuerySqlBack(mailbox rpc.Mailbox, callbackparams share.DBParams, result []share.DBRow) error {
	return nil
}

//执行sql语句
//示例：warp.ExecSql(nil, "DELETE FROM `test1` WHERE `id`=1")
//      删除test1中id=1的记录
func (warp DBWarp) ExecSql(mailbox *rpc.Mailbox, sqlstr string, callback string, callbackparams share.DBParams) error {
	return warp.db.Call(mailbox, "Database.ExecSql", sqlstr, callback, callbackparams)
}

//执行一句sql的回调函数
func (warp DBWarp) ExecSqlBack(mailbox rpc.Mailbox, callbackparams share.DBParams, eff int, err string) error {
	return nil
}
