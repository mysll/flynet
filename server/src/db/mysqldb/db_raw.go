package mysqldb

import (
	"bytes"
	"database/sql"
	"fmt"
	"server"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
	"server/util"
	"strings"
	"time"
)

type Database struct {
	*rpc.Thread
}

func NewRaw(pool int) *Database {
	idb := &Database{}
	idb.Thread = rpc.NewThread("raw", pool, 32)
	return idb
}

// 发送信件（收信人名称，信件类型(0-99为普通信件)，时间，内容，附件）
func (this *Database) sendLetter(source uint64, source_name string, recv uint64, recv_acc, recv_name string, typ int32, title, content, appendix string) (uint64, error) {
	sqlconn := db.sql
	uid, err := sqlconn.GetUid("serial_no")
	if err != nil {
		return 0, err
	}
	_, err = sqlconn.Exec("INSERT INTO `letter`(`serial_no`,`send_time`,`msg_acc`, `msg_name`,`msg_uid`,`source`,`source_name`,`msg_type`,`msg_title`,`msg_content`,`msg_appendix`) VALUES(?,?,?,?,?,?,?,?,?,?,?)",
		uid, time.Now().Format(util.TIME_LAYOUT), recv_acc, recv_name, recv, source, source_name, typ, title, content, appendix)

	return uid, err
}

// 发送系统信件（收件人为空表示发给所有玩家）
func (this *Database) systemLetter(source_name string, recv uint64, recv_acc, recv_name string, typ int32, title, content, appendix string) (uint64, error) {
	if recv != 0 {
		return this.sendLetter(0, source_name, recv, recv_acc, recv_name, typ, title, content, appendix)
	}

	//群发
	stmt, err := db.sql.GetDB().Prepare("INSERT INTO `letter`(`serial_no`,`send_time`,`msg_acc`, `msg_name`,`msg_uid`,`source`,`source_name`,`msg_type`,`msg_title`,`msg_content`,`msg_appendix`) VALUES(?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	r, err := db.sql.Query("SELECT `uid`,`account`,`rolename` FROM `role_info`")
	if err != nil {
		return 0, err
	}
	defer r.Close()
	for r.Next() {
		serial_no, _ := db.sql.GetUid("serial_no")
		var uid uint64
		var account string
		var rolename string
		r.Scan(&uid, &account, &rolename)
		stmt.Exec(serial_no, time.Now().Format(util.TIME_LAYOUT), account, rolename, uid, 0, source_name, typ, title, content, appendix)
	}
	return 0, nil
}

// 收信人接收并删除信件
func (this *Database) recvLetter(uid uint64) (*share.LetterInfo, error) {
	return this.recvLetterBySerial(uid, 0)
}

// 收信人接收并删除指定流水号的信件
func (this *Database) recvLetterBySerial(uid uint64, serial_no uint64) (*share.LetterInfo, error) {
	var sql string
	var args []interface{}
	if serial_no == 0 {
		sql = "SELECT `serial_no`,`send_time`,`msg_acc`, `msg_name`,`source`,`source_name`,`msg_type`,`msg_title`,`msg_content`,`msg_appendix` FROM `letter` WHERE `msg_uid`=? ORDER BY `send_time` LIMIT 0,1"
		args = []interface{}{uid}
	} else {
		sql = "SELECT `serial_no`,`send_time`,`msg_acc`, `msg_name`,`source`,`source_name`,`msg_type`,`msg_title`,`msg_content`,`msg_appendix` FROM `letter` WHERE `msg_uid`=? AND `serial_no` = ? ORDER BY `send_time` LIMIT 0,1"
		args = []interface{}{uid, serial_no}
	}

	row, err := db.sql.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	defer row.Close()

	if !row.Next() {
		return nil, fmt.Errorf("letter not found")
	}

	info := &share.LetterInfo{}

	if err := row.Scan(&info.Serial_no, &info.Send_time, &info.Recv_acc, &info.Recv_name, &info.Source, &info.Source_name, &info.Msg_type, &info.Title, &info.Content, &info.Appendix); err != nil {
		return nil, err
	}

	db.sql.Exec("DELETE FROM `letter` WHERE `serial_no`=?", info.Serial_no)

	return info, nil
}

// 收信人查看信件（不删除）
func (this *Database) lookLetter(uid uint64) ([]*share.LetterInfo, error) {
	row, err := db.sql.Query("SELECT `serial_no`,`send_time`,`msg_acc`, `msg_name`,`source`,`source_name`,`msg_type`,`msg_title`,`msg_content`,`msg_appendix` FROM `letter` WHERE `msg_uid`=? ORDER BY `send_time`", uid)
	if err != nil {
		return nil, err
	}

	defer row.Close()

	letters := make([]*share.LetterInfo, 0, 100)
	for row.Next() {
		info := &share.LetterInfo{}

		if err := row.Scan(&info.Serial_no, &info.Send_time, &info.Recv_acc, &info.Recv_name, &info.Source, &info.Source_name, &info.Msg_type, &info.Title, &info.Content, &info.Appendix); err != nil {
			return nil, err
		}
		letters = append(letters, info)
	}
	return letters, nil
}

// 收信人查询信件数量
func (this *Database) queryLetter(uid uint64) (int, error) {
	row, err := db.sql.Query("SELECT COUNT(`serial_no`) FROM `letter` WHERE `msg_uid`=?", uid)
	if err != nil {
		return 0, err
	}

	defer row.Close()
	if !row.Next() {
		return 0, nil
	}

	var count int
	row.Scan(&count)
	return count, nil
}

// 发信人清理已发信件,send_back是否退信
func (this *Database) cleanLetter(uid uint64, days int, typ int32, new_type int32, send_back bool) error {
	return nil
}

// 发信人清理指定流水号的已发信件
func (this *Database) cleanLetterBySerial(uid uint64, serial_no uint64, new_type int32, send_back bool) error {
	return nil
}

// 收信人退回指定流水号的信件
func (this *Database) backLetterBySerial(uid uint64, serial_no uint64, new_type int32) error {
	return nil
}

func (this *Database) RecvLetter(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var uid uint64
	var serial_no uint64
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &uid, &serial_no, &callback, &callbackparams)) {
		return 0, nil
	}

	letter, err := this.recvLetterBySerial(uid, serial_no)
	if err != nil {
		log.LogError(err)
		return share.ERR_REPLY_FAILED, nil
	}

	if callback == "_" {
		return 0, nil
	}
	callbackparams["result"] = false
	if letter != nil {
		callbackparams["result"] = true
		callbackparams["letter"] = letter
	}

	server.Check(server.MailTo(nil, &mailbox, callback, callbackparams))
	return 0, nil
}

func (this *Database) LookLetter(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var uid uint64
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &uid, &callback, &callbackparams)) {
		return 0, nil
	}

	letters, err := this.lookLetter(uid)
	if err != nil {
		log.LogError(err)
		return 0, nil
	}

	callbackparams["result"] = false
	if letters != nil {
		callbackparams["result"] = true
		callbackparams["letters"] = letters
	}

	server.Check(server.MailTo(nil, &mailbox, callback, callbackparams))
	return 0, nil
}

func (this *Database) QueryLetter(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var uid uint64
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &uid, &callback, &callbackparams)) {
		return 0, nil
	}

	count, _ := this.queryLetter(uid)
	callbackparams["count"] = count
	server.Check(server.MailTo(nil, &mailbox, callback, callbackparams))
	return 0, nil
}

func (this *Database) SendSystemLetter(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var source_name string
	var recvacc, recvrole string
	var typ int32
	var title, content, appendix string
	var callback string
	var callbackparams share.DBParams
	if server.Check(server.ParseArgs(msg, &source_name, &recvacc, &recvrole, &typ, &title, &content, &appendix, &callback, &callbackparams)) {
		return 0, nil
	}

	var roleid uint64
	if recvacc == "" && recvrole == "" {
		roleid = 0
	} else {
		var err error
		roleid, err = GetRoleUid(recvacc, recvrole)
		if err != nil {
			callbackparams["result"] = false
			callbackparams["err"] = "role not found"
			if callback != "_" {
				server.Check(server.MailTo(nil, &mailbox, callback, callbackparams))
			}
			log.LogError(err)
			return 0, nil
		}
	}

	serial_no, err := this.systemLetter(source_name, roleid, recvacc, recvrole, typ, title, content, appendix)
	if err != nil {
		if callback != "_" {
			callbackparams["result"] = false
			callbackparams["err"] = err.Error()
			server.Check(server.MailTo(nil, &mailbox, callback, callbackparams))
		}
		log.LogError(err)
		return 0, nil
	}

	if callback != "_" {
		callbackparams["result"] = true
		callbackparams["serial_no"] = serial_no
		server.Check(server.MailTo(nil, &mailbox, callback, callbackparams))
	}
	return 0, nil

}

func (this *Database) Log(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var log_name string
	var log_source, log_type int32
	var log_content, log_comment string

	if server.Check(server.ParseArgs(msg, &log_name, &log_source, &log_type, &log_content, &log_comment)) {
		return 0, nil
	}

	sqlconn := db.sql
	uid, err := sqlconn.GetUid("serial_no")
	if err != nil {
		log.LogError(err)
		return 0, nil
	}
	sql := fmt.Sprintf("INSERT INTO `log_data`(`serial_no`, `log_time`,`log_name`, `log_source`, `log_type`, `log_content`, `log_comment`) VALUES(?,?,?,?,?,?,?)")
	server.Check2(sqlconn.Exec(sql, uid, time.Now().Format(util.TIME_LAYOUT), log_name, log_source, log_type, log_content, log_comment))
	return 0, nil
}

func (this *Database) Count(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var tbl string
	var condition string
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &tbl, &condition, &callback, &callbackparams)) {
		return 0, nil
	}

	sqlconn := db.sql
	var r *sql.Rows
	var err error
	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
	}

	if condition != "" {
		condition = fmt.Sprintf("WHERE %s", condition)
	}
	sqlstr := fmt.Sprintf("SELECT COUNT(*) FROM `%s` %s LIMIT 1", tbl, condition)
	if r, err = sqlconn.Query(sqlstr); err != nil {
		log.LogError("sql:", sqlstr)
		return 0, nil
	}
	defer r.Close()
	if !r.Next() {
		server.Check(app.Call(nil, callback, -1))
		return 0, nil
	}

	var count int32
	if err = r.Scan(&count); err != nil {
		log.LogError(err)
		server.Check(app.Call(nil, callback, -1))
		return 0, nil
	}

	server.Check(app.Call(nil, callback, callbackparams, count))
	return 0, nil
}

func (this *Database) InsertRows(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var tbl string
	var keys []string
	var values []interface{}
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &tbl, &keys, &values, &callback, &callbackparams)) {
		return 0, nil
	}

	sqlconn := db.sql
	var r sql.Result
	var err error
	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	var sql bytes.Buffer
	sql.WriteString("INSERT INTO `")
	sql.WriteString(tbl)
	sql.WriteString("`(")
	split := " "
	for _, v := range keys {
		sql.WriteString(split)
		split = ", "
		sql.WriteString("`")
		sql.WriteString(v)
		sql.WriteString("`")
	}
	sql.WriteString(") VALUES")
	split = " "
	if len(values)%len(keys) != 0 {
		log.LogError("length is not match")
		return 0, nil
	}
	rows := len(values) / len(keys)
	for j := 0; j < rows; j++ {
		sql.WriteString(split)
		sql.WriteString("(")
		split = " "
		for i := 0; i < len(keys); i++ {
			sql.WriteString(split)
			sql.WriteString("?")
			split = ", "
		}
		split = ", "
		sql.WriteString(")")

	}

	if r, err = sqlconn.Exec(sql.String(), values...); err != nil {
		log.LogError("sql:", sql.String())
		if callback == "_" {
			log.LogError(err)
			return 0, nil
		}
		server.Check(app.Call(nil, callback, callbackparams, 0, err.Error()))
		return 0, nil
	}

	if callback == "_" {
		return 0, nil
	}
	eff, _ := r.RowsAffected()
	server.Check(app.Call(nil, callback, callbackparams, eff, ""))
	return 0, nil
}

func (this *Database) InsertRow(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var tbl string
	var values map[string]interface{}
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &tbl, &values, &callback, &callbackparams)) {
		return 0, nil
	}

	sqlconn := db.sql
	var r sql.Result
	var err error

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	args := make([]interface{}, 0, len(values))
	var sql bytes.Buffer
	sql.WriteString("INSERT INTO `")
	sql.WriteString(tbl)
	sql.WriteString("` SET")
	split := " "
	for k, v := range values {
		sql.WriteString(split)
		split = ", "
		sql.WriteString("`")
		sql.WriteString(k)
		sql.WriteString("`")
		sql.WriteString("=?")
		args = append(args, v)
	}

	if r, err = sqlconn.Exec(sql.String(), args...); err != nil {
		log.LogError(sql.String())
		if callback == "_" {
			log.LogError(err)
			return 0, nil
		}
		server.Check(app.Call(nil, callback, callbackparams, 0, err.Error()))
		return 0, nil
	}
	if callback == "_" {
		return 0, nil
	}
	eff, _ := r.RowsAffected()
	server.Check(app.Call(nil, callback, callbackparams, eff, ""))
	return 0, nil

}

func (this *Database) DeleteRow(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var tbl string
	var condition string
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &tbl, &condition, &callback, &callbackparams)) {
		return 0, nil
	}

	sqlconn := db.sql
	var r sql.Result
	var err error
	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	if condition != "" {
		condition = fmt.Sprintf(" WHERE %s", condition)
	}

	sqlstr := fmt.Sprintf("DELETE FROM `%s`%s", tbl, condition)
	if r, err = sqlconn.Exec(sqlstr); err != nil {
		log.LogError("sql:", sqlstr)
		if callback != "_" {
			server.Check(app.Call(nil, callback, callbackparams, 0, err.Error()))
			return 0, nil
		}
	}

	if callback == "_" {
		return 0, nil
	}

	eff, _ := r.RowsAffected()
	server.Check(app.Call(nil, callback, callbackparams, eff, ""))
	return 0, nil
}

func (this *Database) UpdateRow(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var tbl string
	var values map[string]interface{}
	var condition string
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &tbl, &values, &condition, &callback, &callbackparams)) {
		return 0, nil
	}

	sqlconn := db.sql
	var r sql.Result
	var err error

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	if condition != "" {
		condition = fmt.Sprintf("WHERE %s", condition)
	}

	args := make([]interface{}, 0, len(values))
	var sql bytes.Buffer
	sql.WriteString("UPDATE `")
	sql.WriteString(tbl)
	sql.WriteString("` SET")
	split := " "
	for k, v := range values {
		sql.WriteString(split)
		split = ", "
		sql.WriteString("`")
		sql.WriteString(k)
		sql.WriteString("`")
		sql.WriteString("=?")
		args = append(args, v)
	}
	sqlstr := fmt.Sprintf("%s %s", sql.String(), condition)
	if r, err = sqlconn.Exec(sqlstr, args...); err != nil {
		log.LogError("sql:", sqlstr)
		if callback == "_" {
			log.LogError(err)
			return 0, nil
		}
		server.Check(app.Call(nil, callback, callbackparams, 0, err.Error()))
		return 0, nil
	}
	if callback == "_" {
		return 0, nil
	}
	eff, _ := r.RowsAffected()
	server.Check(app.Call(nil, callback, callbackparams, eff, ""))
	return 0, nil
}

func (this *Database) QueryRow(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var tbl string
	var condition string
	var orderby string
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &tbl, &condition, &orderby, &callback, &callbackparams)) {
		return 0, nil
	}

	sqlconn := db.sql
	var r *sql.Rows
	var err error

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	if condition != "" {
		condition = fmt.Sprintf(" WHERE %s", condition)
	}
	if orderby != "" {
		orderby = fmt.Sprintf(" ORDER BY %s", orderby)
	}
	sqlstr := fmt.Sprintf("SELECT * FROM `%s`%s%s LIMIT 1", tbl, condition, orderby)
	if r, err = sqlconn.Query(sqlstr); err != nil {
		log.LogError(sqlstr, " ", err)
		return 0, nil
	}
	defer r.Close()
	var cols []string
	cols, err = r.Columns()
	if err != nil {
		server.Check(app.Call(nil, callback, callbackparams, share.DBRow{}))
		return 0, nil
	}

	if !r.Next() {
		server.Check(app.Call(nil, callback, callbackparams, share.DBRow{}))
		return 0, nil
	}

	result := make([]interface{}, len(cols))
	for k := range result {
		result[k] = new([]byte)
	}

	err = r.Scan(result...)
	if err != nil {
		log.LogError("sql:", sqlstr)
		server.Check(app.Call(nil, callback, callbackparams, share.DBRow{"error": []byte(err.Error())}))
		return 0, nil
	}

	mapresult := make(share.DBRow, len(cols))
	for k, v := range cols {
		mapresult[v] = *result[k].(*[]byte)
	}

	server.Check(app.Call(nil, callback, callbackparams, mapresult))
	return 0, nil
}

func (this *Database) QueryRows(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var tbl string
	var condition string
	var orderby string
	var index, count int32
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &tbl, &condition, &orderby, &index, &count, &callback, &callbackparams)) {
		return 0, nil
	}

	sqlconn := db.sql
	var r *sql.Rows
	var err error
	if count <= 0 {
		log.LogError("count must above zero")
		return 0, nil
	}

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	if condition != "" {
		condition = fmt.Sprintf(" WHERE %s", condition)
	}
	if orderby != "" {
		orderby = fmt.Sprintf(" ORDER BY %s", orderby)
	}

	sqlstr := fmt.Sprintf("SELECT * FROM `%s`%s%s LIMIT %d, %d", tbl, condition, orderby, index, count)
	if r, err = sqlconn.Query(sqlstr); err != nil {
		log.LogError(err)
		return 0, nil
	}
	defer r.Close()
	var cols []string
	cols, err = r.Columns()
	if err != nil {
		server.Check(app.Call(nil, callback, callbackparams, []share.DBRow{share.DBRow{"error": []byte(err.Error())}}, index, count))
		return 0, nil
	}

	result := make([]interface{}, len(cols))
	for k := range result {
		result[k] = new([]byte)
	}

	arrresult := make([]share.DBRow, 0, count)
	for r.Next() {

		err = r.Scan(result...)
		if err != nil {
			log.LogError("sql:", sqlstr)
			server.Check(app.Call(nil, callback, callbackparams, []share.DBRow{share.DBRow{"error": []byte(err.Error())}}, index, count))
			return 0, nil
		}

		mapresult := make(share.DBRow, len(cols))
		for k, v := range cols {
			mapresult[v] = *result[k].(*[]byte)
		}
		arrresult = append(arrresult, mapresult)
	}

	server.Check(app.Call(nil, callback, callbackparams, arrresult, index, count))
	return 0, nil
}

func (this *Database) QuerySql(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var sqlstr string
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &sqlstr, &callback, &callbackparams)) {
		return 0, nil
	}

	sqlconn := db.sql
	var r *sql.Rows
	var err error

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	if r, err = sqlconn.Query(sqlstr); err != nil {
		log.LogError(err)
		return 0, nil
	}
	defer r.Close()
	var cols []string
	cols, err = r.Columns()
	if err != nil {
		server.Check(app.Call(nil, callback, callbackparams, []share.DBRow{share.DBRow{"error": []byte(err.Error())}}))
		return 0, nil
	}

	result := make([]interface{}, len(cols))
	for k := range result {
		result[k] = new([]byte)
	}

	arrresult := make([]share.DBRow, 0, 100)
	for r.Next() {

		err = r.Scan(result...)
		if err != nil {
			log.LogError("sql:", sqlstr)
			server.Check(app.Call(nil, callback, callbackparams, []share.DBRow{share.DBRow{"error": []byte(err.Error())}}))
			return 0, nil
		}

		mapresult := make(share.DBRow, len(cols))
		for k, v := range cols {
			mapresult[v] = *result[k].(*[]byte)
		}
		arrresult = append(arrresult, mapresult)
		if len(arrresult) == 100 {
			break
		}
	}

	server.Check(app.Call(nil, callback, callbackparams, arrresult))
	return 0, nil
}

func (this *Database) ExecSql(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var sqlstr string
	var callback string
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &sqlstr, &callback, &callbackparams)) {
		return 0, nil
	}

	sqlconn := db.sql
	var r sql.Result
	var err error

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return 0, nil
	}

	if r, err = sqlconn.Exec(sqlstr); err != nil {
		log.LogError("sql:", sqlstr)
		server.Check(app.Call(nil, callback, callbackparams, 0, err.Error()))
		return 0, nil
	}

	eff, _ := r.RowsAffected()
	server.Check(app.Call(nil, callback, callbackparams, eff, ""))
	return 0, nil
}

func (this *Database) SaveObject(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var object share.DbSave
	var callbackparams share.DBParams
	if server.Check(server.ParseArgs(msg, &object, &callbackparams)) {
		return share.ERR_ARGS_ERROR, nil
	}

	reply = server.CreateMessage(callbackparams)
	sqlconn := db.sql
	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return share.ERR_SYSTEM_ERROR, reply
	}

	err := SaveItem(sqlconn, true, object.Data.DBId, object.Data)

	if err != nil {
		return share.ERR_REPLY_FAILED, reply
	}

	return share.ERR_REPLY_SUCCEED, reply
}

func (this *Database) UpdateObject(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var object share.DbSave
	var callbackparams share.DBParams

	if server.Check(server.ParseArgs(msg, &object, &callbackparams)) {
		return share.ERR_ARGS_ERROR, nil
	}

	reply = server.CreateMessage(callbackparams)
	sqlconn := db.sql

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return share.ERR_SYSTEM_ERROR, reply
	}

	err := SaveItem(sqlconn, false, object.Data.DBId, object.Data)

	if err != nil {
		return share.ERR_REPLY_FAILED, reply
	}

	return share.ERR_REPLY_SUCCEED, reply
}

func (this *Database) LoadObject(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var ent string
	var dbid uint64
	var callbackparams share.DBParams
	var err error
	if server.Check(server.ParseArgs(msg, &ent, &dbid, &callbackparams)) {
		return share.ERR_ARGS_ERROR, nil
	}

	reply = server.CreateMessage(callbackparams)
	sqlconn := db.sql

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return share.ERR_SYSTEM_ERROR, reply
	}

	savedata := share.DbSave{}
	savedata.Data, err = LoadEntity(sqlconn, dbid, ent, 0)

	if err != nil {
		return share.ERR_REPLY_FAILED, reply
	}

	server.AppendArgs(reply, savedata)
	return share.ERR_REPLY_SUCCEED, reply
}

func (this *Database) LoadObjectByName(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var ent string
	var name string
	var callbackparams share.DBParams
	var err error
	if server.Check(server.ParseArgs(msg, &ent, &name, &callbackparams)) {
		return share.ERR_ARGS_ERROR, nil
	}

	reply = server.CreateMessage(callbackparams)
	sqlconn := db.sql

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return share.ERR_SYSTEM_ERROR, reply
	}

	savedata := share.DbSave{}
	savedata.Data, err = LoadEntityByName(sqlconn, name, ent, 0)

	if err != nil {
		return share.ERR_REPLY_FAILED, reply
	}

	server.AppendArgs(reply, savedata)
	return share.ERR_REPLY_SUCCEED, reply
}

func (this *Database) DeleteObject(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	var ent string
	var dbid uint64
	var callbackparams share.DBParams
	var err error

	if server.Check(server.ParseArgs(msg, &ent, &dbid, &callbackparams)) {
		return share.ERR_ARGS_ERROR, nil
	}

	reply = server.CreateMessage(callbackparams)
	sqlconn := db.sql

	app := server.GetAppById(mailbox.App)
	if app == nil {
		log.LogError(server.ErrAppNotFound)
		return share.ERR_SYSTEM_ERROR, reply
	}

	sqlstr := fmt.Sprintf("DELETE FROM `tbl_%s` WHERE `id`=?", strings.ToLower(ent))
	if _, err = sqlconn.Exec(sqlstr, dbid); err != nil {
		log.LogError("sql:", sqlstr)
		return share.ERR_REPLY_FAILED, reply
	}

	return share.ERR_REPLY_SUCCEED, reply
}

func (t *Database) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("RecvLetter", t.RecvLetter)
	s.RegisterCallback("LookLetter", t.LookLetter)
	s.RegisterCallback("QueryLetter", t.QueryLetter)
	s.RegisterCallback("SendSystemLetter", t.SendSystemLetter)
	s.RegisterCallback("Log", t.Log)
	s.RegisterCallback("Count", t.Count)
	s.RegisterCallback("InsertRows", t.InsertRows)
	s.RegisterCallback("InsertRow", t.InsertRow)
	s.RegisterCallback("DeleteRow", t.DeleteRow)
	s.RegisterCallback("UpdateRow", t.UpdateRow)
	s.RegisterCallback("QueryRow", t.QueryRow)
	s.RegisterCallback("QueryRows", t.QueryRows)
	s.RegisterCallback("QuerySql", t.QuerySql)
	s.RegisterCallback("ExecSql", t.ExecSql)
	s.RegisterCallback("SaveObject", t.SaveObject)
	s.RegisterCallback("UpdateObject", t.UpdateObject)
	s.RegisterCallback("LoadObject", t.LoadObject)
	s.RegisterCallback("LoadObjectByName", t.LoadObjectByName)
	s.RegisterCallback("DeleteObject", t.DeleteObject)
}
