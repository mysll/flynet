package mysqldb

import (
	"bytes"
	"database/sql"
	"fmt"
	"libs/log"
	"libs/rpc"
	"server"
	"share"
	"sync/atomic"
	"time"
	"util"
)

type Database struct {
	numprocess int32
	queue      []chan *rpc.RpcCall
	quit       bool
	pools      int
}

func NewRaw(pool int) *Database {
	idb := &Database{}
	idb.queue = make([]chan *rpc.RpcCall, pool)
	idb.pools = pool
	for i := 0; i < pool; i++ {
		idb.queue[i] = make(chan *rpc.RpcCall, 128)
	}

	return idb
}

func (this *Database) Push(r *rpc.RpcCall) bool {

	mb := r.Src.Interface().(rpc.Mailbox)
	this.queue[int(mb.Uid)%this.pools] <- r // 队列满了，就会阻塞在这里

	return true
}

func (this *Database) work(id int) {
	log.LogMessage("db work, id:", id)
	var start_time time.Time
	var delay time.Duration
	warninglvl := 50 * time.Millisecond
	for {
		select {
		case caller := <-this.queue[id]:
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
			if this.quit {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func (this *Database) Do() {
	log.LogMessage("start db interface thread, total:", this.pools)
	for i := 0; i < this.pools; i++ {
		id := i
		db.wg.Wrap(func() { this.work(id) })
	}
}

func (this *Database) process(fname string, f func() error) error {
	atomic.AddInt32(&this.numprocess, 1)
	err := f()
	if err != nil {
		log.LogError("db process error:", err)
	}
	atomic.AddInt32(&this.numprocess, -1)
	return err
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

func (this *Database) RecvLetter(mailbox rpc.Mailbox, uid uint64, serial_no uint64, callback string, callbackparams share.DBParams) error {
	letter, err := this.recvLetterBySerial(uid, serial_no)
	if err != nil {
		log.LogError(err)
	}

	if callback == "_" {
		return nil
	}
	callbackparams["result"] = false
	if letter != nil {
		callbackparams["result"] = true
		callbackparams["letter"] = letter
	}

	return server.MailTo(nil, &mailbox, callback, callbackparams)
}

func (this *Database) LookLetter(mailbox rpc.Mailbox, uid uint64, callback string, callbackparams share.DBParams) error {
	letters, err := this.lookLetter(uid)
	if err != nil {
		log.LogError(err)
	}

	callbackparams["result"] = false
	if letters != nil {
		callbackparams["result"] = true
		callbackparams["letters"] = letters
	}

	return server.MailTo(nil, &mailbox, callback, callbackparams)
}

func (this *Database) QueryLetter(mailbox rpc.Mailbox, uid uint64, callback string, callbackparams share.DBParams) error {
	count, _ := this.queryLetter(uid)
	callbackparams["count"] = count
	return server.MailTo(nil, &mailbox, callback, callbackparams)
}

func (this *Database) SendSystemLetter(mailbox rpc.Mailbox, source_name string, recvacc, recvrole string, typ int32, title, content, appendix string, callback string, callbackparams share.DBParams) error {
	return this.process("SendLetter", func() error {

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
					return server.MailTo(nil, &mailbox, callback, callbackparams)
				}
				return err
			}
		}

		serial_no, err := this.systemLetter(source_name, roleid, recvacc, recvrole, typ, title, content, appendix)
		if err != nil {
			if callback != "_" {
				callbackparams["result"] = false
				callbackparams["err"] = err.Error()
				return server.MailTo(nil, &mailbox, callback, callbackparams)
			}
			return err
		}

		if callback != "_" {
			callbackparams["result"] = true
			callbackparams["serial_no"] = serial_no
			return server.MailTo(nil, &mailbox, callback, callbackparams)
		}
		return nil
	})
}

func (this *Database) Log(mailbox rpc.Mailbox, log_name string, log_source, log_type int32, log_content, log_comment string) error {
	return this.process("Log", func() error {
		sqlconn := db.sql
		uid, err := sqlconn.GetUid("serial_no")
		if err != nil {
			log.LogError(err)
			return err
		}
		sql := fmt.Sprintf("INSERT INTO `log_data`(`serial_no`, `log_time`,`log_name`, `log_source`, `log_type`, `log_content`, `log_comment`) VALUES(?,?,?,?,?,?,?)")
		_, err = sqlconn.Exec(sql, uid, time.Now().Format(util.TIME_LAYOUT), log_name, log_source, log_type, log_content, log_comment)
		if err != nil {
			log.LogError(err, sql)
		}
		return err
	})
}

func (this *Database) Count(mailbox rpc.Mailbox, tbl string, condition string, callback string, callbackparams share.DBParams) error {
	return this.process("Count", func() error {
		sqlconn := db.sql
		var r *sql.Rows
		var err error

		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
		}

		if condition != "" {
			condition = fmt.Sprintf("WHERE %s", condition)
		}
		sqlstr := fmt.Sprintf("SELECT COUNT(*) FROM `%s` %s LIMIT 1", tbl, condition)
		if r, err = sqlconn.Query(sqlstr); err != nil {
			log.LogError("sql:", sqlstr)
			return err
		}
		defer r.Close()
		if !r.Next() {
			return app.Call(nil, callback, -1)
		}

		var count int32
		if err = r.Scan(&count); err != nil {
			log.LogError(err)
			return app.Call(nil, callback, -1)
		}

		return app.Call(nil, callback, callbackparams, count)
	})
}

func (this *Database) InsertRows(mailbox rpc.Mailbox, tbl string, keys []string, values []interface{}, callback string, callbackparams share.DBParams) error {
	return this.process("InsertRows", func() error {
		sqlconn := db.sql
		var r sql.Result
		var err error
		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
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
			return fmt.Errorf("length is not match")
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
				return err
			}
			return app.Call(nil, callback, callbackparams, 0, err.Error())
		}

		if callback == "_" {
			return nil
		}
		eff, _ := r.RowsAffected()
		return app.Call(nil, callback, callbackparams, eff, "")
	})
}

func (this *Database) InsertRow(mailbox rpc.Mailbox, tbl string, values map[string]interface{}, callback string, callbackparams share.DBParams) error {
	return this.process("InsertRow", func() error {
		sqlconn := db.sql
		var r sql.Result
		var err error
		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
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
				return err
			}
			return app.Call(nil, callback, callbackparams, 0, err.Error())
		}
		if callback == "_" {
			return nil
		}
		eff, _ := r.RowsAffected()
		return app.Call(nil, callback, callbackparams, eff, "")
	})
}

func (this *Database) DeleteRow(mailbox rpc.Mailbox, tbl string, condition string, callback string, callbackparams share.DBParams) error {
	return this.process("DeleteRow", func() error {
		sqlconn := db.sql
		var r sql.Result
		var err error
		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
		}

		if condition != "" {
			condition = fmt.Sprintf(" WHERE %s", condition)
		}

		sqlstr := fmt.Sprintf("DELETE FROM `%s`%s", tbl, condition)
		if r, err = sqlconn.Exec(sqlstr); err != nil {
			log.LogError("sql:", sqlstr)
			return app.Call(nil, callback, callbackparams, 0, err.Error())
		}

		if callback == "_" {
			return nil
		}

		eff, _ := r.RowsAffected()
		return app.Call(nil, callback, callbackparams, eff, "")
	})
}

func (this *Database) UpdateRow(mailbox rpc.Mailbox, tbl string, values map[string]interface{}, condition string, callback string, callbackparams share.DBParams) error {
	return this.process("UpdateRow", func() error {
		sqlconn := db.sql
		var r sql.Result
		var err error

		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
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
				return err
			}
			return app.Call(nil, callback, callbackparams, 0, err.Error())
		}
		if callback == "_" {
			return nil
		}
		eff, _ := r.RowsAffected()
		return app.Call(nil, callback, callbackparams, eff, "")
	})
}

func (this *Database) QueryRow(mailbox rpc.Mailbox, tbl string, condition string, orderby string, callback string, callbackparams share.DBParams) error {
	return this.process("QueryRow", func() error {
		sqlconn := db.sql
		var r *sql.Rows
		var err error

		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
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
			return err
		}
		defer r.Close()
		var cols []string
		cols, err = r.Columns()
		if err != nil {
			return app.Call(nil, callback, callbackparams, share.DBRow{})
		}

		if !r.Next() {
			return app.Call(nil, callback, callbackparams, share.DBRow{})
		}

		result := make([]interface{}, len(cols))
		for k := range result {
			result[k] = new([]byte)
		}

		err = r.Scan(result...)
		if err != nil {
			log.LogError("sql:", sqlstr)
			return app.Call(nil, callback, callbackparams, share.DBRow{"error": []byte(err.Error())})
		}

		mapresult := make(share.DBRow, len(cols))
		for k, v := range cols {
			mapresult[v] = *result[k].(*[]byte)
		}

		return app.Call(nil, callback, callbackparams, mapresult)
	})
}

func (this *Database) QueryRows(mailbox rpc.Mailbox, tbl string, condition string, orderby string, index int, count int, callback string, callbackparams share.DBParams) error {
	return this.process("QueryRows", func() error {
		sqlconn := db.sql
		var r *sql.Rows
		var err error

		if count <= 0 {
			return fmt.Errorf("count must above zero")
		}

		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
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
			return err
		}
		defer r.Close()
		var cols []string
		cols, err = r.Columns()
		if err != nil {
			return app.Call(nil, callback, callbackparams, []share.DBRow{share.DBRow{"error": []byte(err.Error())}}, index, count)
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
				return app.Call(nil, callback, callbackparams, []share.DBRow{share.DBRow{"error": []byte(err.Error())}}, index, count)
			}

			mapresult := make(share.DBRow, len(cols))
			for k, v := range cols {
				mapresult[v] = *result[k].(*[]byte)
			}
			arrresult = append(arrresult, mapresult)
		}

		return app.Call(nil, callback, callbackparams, arrresult, index, count)
	})
}

func (this *Database) QuerySql(mailbox rpc.Mailbox, sqlstr string, callback string, callbackparams share.DBParams) error {
	return this.process("QuerySql", func() error {
		sqlconn := db.sql
		var r *sql.Rows
		var err error

		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
		}

		if r, err = sqlconn.Query(sqlstr); err != nil {
			log.LogError(err)
			return err
		}
		defer r.Close()
		var cols []string
		cols, err = r.Columns()
		if err != nil {
			return app.Call(nil, callback, callbackparams, []share.DBRow{share.DBRow{"error": []byte(err.Error())}})
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
				return app.Call(nil, callback, callbackparams, []share.DBRow{share.DBRow{"error": []byte(err.Error())}})
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

		return app.Call(nil, callback, callbackparams, arrresult)

	})
}

func (this *Database) ExecSql(mailbox rpc.Mailbox, sqlstr string, callback string, callbackparams share.DBParams) error {
	return this.process("ExecSql", func() error {
		sqlconn := db.sql
		var r sql.Result
		var err error
		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
		}

		if r, err = sqlconn.Exec(sqlstr); err != nil {
			log.LogError("sql:", sqlstr)
			return app.Call(nil, callback, callbackparams, 0, err.Error())
		}

		eff, _ := r.RowsAffected()
		return app.Call(nil, callback, callbackparams, eff, "")
	})
}

func (this *Database) SaveObject(mailbox rpc.Mailbox, object *share.SaveEntity, callback string, callbackparams share.DBParams) error {
	return this.process("SaveObject", func() error {
		sqlconn := db.sql
		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
		}

		err := SaveItem(sqlconn, true, object.DBId, object)

		callbackparams["result"] = "ok"
		if err != nil {
			callbackparams["result"] = err.Error()
		}
		if callback != "_" {
			return app.Call(nil, callback, callbackparams)
		}

		return err
	})
}

func (this *Database) UpdateObject(mailbox rpc.Mailbox, object *share.SaveEntity, callback string, callbackparams share.DBParams) error {
	return this.process("UpdateObject", func() error {
		sqlconn := db.sql

		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
		}

		err := SaveItem(sqlconn, false, object.DBId, object)
		callbackparams["result"] = "ok"
		if err != nil {
			callbackparams["result"] = err.Error()
		}

		if callback != "_" {
			return app.Call(nil, callback, callbackparams)
		}

		return err
	})
}

func (this *Database) LoadObject(mailbox rpc.Mailbox, ent string, dbid uint64, callback string, callbackparams share.DBParams) error {
	return this.process("UpdateObject", func() error {
		sqlconn := db.sql

		app := server.GetApp(mailbox.Address)
		if app == nil {
			return server.ErrAppNotFound
		}

		savedata := share.DbSave{}
		var err error
		savedata.Data, err = LoadEntity(sqlconn, dbid, ent, 0)

		if err != nil {
			callbackparams["result"] = err.Error()
		} else {
			callbackparams["result"] = "ok"
			callbackparams["data"] = savedata
		}

		if callback == "_" || callback == "" {
			log.LogError("need callback")
			return nil
		}

		return app.Call(nil, callback, callbackparams)
	})
}
