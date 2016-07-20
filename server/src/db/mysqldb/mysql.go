package mysqldb

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	. "logicdata/parser"
	"server/libs/log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mysll/go-uuid/uuid"
)

type MySql struct {
	DBConn
}

/*
创建一个新的mysql连接
*/
func NewMysqlConn(datasource string) (conn *MySql, err error) {
	db, e := sql.Open("mysql", datasource)
	if e != nil {
		err = e
		return
	}

	if e := db.Ping(); e != nil {
		err = e
		return
	}

	conn = &MySql{}
	conn.Conn = db

	conn.Type = "mysql"
	log.TraceInfo("mysql", "connected")
	return
}

func (conn *MySql) GetDB() *sql.DB {
	return conn.DBConn.Conn
}

/*
执行一个sql操作，返回执行的结果
*/
func (conn *MySql) Exec(query string, args ...interface{}) (result driver.Result, err error) {
	if len(args) == 0 {
		result, err = conn.Conn.Exec(query)
	} else {
		stmt, e := conn.Conn.Prepare(query)
		if e != nil {
			err = e
			return
		}
		defer stmt.Close()
		result, err = stmt.Exec(args...)
	}

	return
}

/*
执行一个sql查询，返回查询的结果
*/
func (conn *MySql) Query(query string, args ...interface{}) (row *sql.Rows, err error) {
	if len(args) == 0 {
		row, err = conn.Conn.Query(query)
	} else {
		stmt, e := conn.Conn.Prepare(query)
		if e != nil {
			err = e
			return
		}
		defer stmt.Close()
		row, err = stmt.Query(args...)
	}

	return
}

/*
关闭数据库连接
*/
func (conn *MySql) Close() {
	conn.Conn.Close()
	log.TraceInfo("mysql", "closed")
}

func (conn *MySql) queryTblIndex(tbl_schema, tbl_name string) (index map[string][]string, err error) {
	row, e := conn.Query(fmt.Sprintf("show index from `%s`.`%s`", tbl_schema, tbl_name))
	if e != nil {
		err = e
		return
	}

	defer row.Close()

	index = make(map[string][]string)

	for row.Next() {
		var tbl string
		var unique int
		var key string
		var seqindex int
		var col_name string
		var collaction string
		var cardinality sql.NullString
		var subpart sql.NullString
		var pkg sql.NullString
		var isnull string
		var idxtype string
		var comment string

		e := row.Scan(&tbl, &unique, &key, &seqindex, &col_name, &collaction, &cardinality, &subpart, &pkg, &isnull, &idxtype, &comment)
		if e != nil {
			err = e
			return
		}

		var indexkey string
		if key == "PRIMARY" {
			indexkey = "PRIMARY"
		} else if unique == 0 {
			indexkey = fmt.Sprintf("UNIQUE INDEX `%s`", key)
		} else {
			indexkey = fmt.Sprintf("INDEX `%s`", key)
		}

		if _, dup := index[indexkey]; dup {
			index[indexkey] = append(index[key], col_name)
		} else {
			index[indexkey] = make([]string, 0, 2)
			index[indexkey] = append(index[key], col_name)
		}

	}

	return
}

/*
查询表格的字段信息
*/
func (conn *MySql) QueryTblInfo(tbl_schema, tbl_name string) (fields TblField, empty bool, err error) {

	rowcount, e := conn.Query("select count(*) from information_schema.columns where  table_schema=? and  table_name=? limit 0, 1", tbl_schema, tbl_name)
	if e != nil {
		log.LogError(err.Error())
		err = e
		return
	}

	defer rowcount.Close()
	if !rowcount.Next() {
		return
	}

	var count int
	rowcount.Scan(&count)
	if count <= 0 {
		empty = true
		return
	}

	fields.FieldList = make([]Field, count)

	inforows, e2 := conn.Query("select column_name, column_type, is_nullable, extra from information_schema.columns where  table_schema=? and  table_name=?", tbl_schema, tbl_name)
	if e2 != nil {
		log.LogError(err.Error())
		err = e2
		return
	}

	defer inforows.Close()

	var colname, coltype, isnull string
	var extra sql.NullString
	index := 0
	for inforows.Next() {
		e := inforows.Scan(&colname, &coltype, &isnull, &extra)
		if e != nil {
			log.LogError(e.Error())
			err = e
			return
		}

		fields.FieldList[index] = Field{colname, coltype, isnull == "YES", extra.Valid && (extra.String == "auto_increment")}
		index++

	}

	indexinfo, e := conn.queryTblIndex(tbl_schema, tbl_name)
	if e != nil {
		err = e
		return
	}

	fields.IndexInfo = indexinfo
	return
}

func (conn *MySql) Ping() error {
	return conn.Conn.Ping()
}

/*
选择数据库
*/
func (conn *MySql) UseDB(db string) (err error) {
	if _, err = conn.Exec(fmt.Sprintf("use `%s`", db)); err != nil {
		return err
	}
	log.LogMessage("use db:", db)
	return
}

/*
删除数据库
*/
func (conn *MySql) DropDB(db string) (err error) {
	_, err = conn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", db))
	return
}

/*
删除一个表
*/
func (conn *MySql) DropTable(tbl_name string) (err error) {
	_, err = conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tbl_name))
	return
}

/*
创建一个表
*/
func (conn *MySql) CreateTable(tbl_name string, fields TblField, incNum int) (err error) {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (", tbl_name))

	for _, f := range fields.FieldList {
		var nulltype string
		if !f.IsNull {
			nulltype = "NOT NULL"
		} else {
			nulltype = "DEFAULT NULL"
		}
		autoinc := ""
		if f.AutoInc {
			autoinc = " AUTO_INCREMENT"
		}
		buffer.WriteString(fmt.Sprintf("`%s` %s %s%s,", f.Name, f.Type, nulltype, autoinc))

	}

	if _, exist := fields.IndexInfo["PRIMARY"]; exist {
		buffer.WriteString(fmt.Sprintf("PRIMARY KEY (`%s`),", fields.IndexInfo["PRIMARY"][0]))
	}

	for key := range fields.IndexInfo {
		if key == "PRIMARY" {
			continue
		}
		cols := ""

		for _, col := range fields.IndexInfo[key] {
			cols += fmt.Sprintf("`%s`,", col)
		}
		c := []byte(cols)
		buffer.WriteString(fmt.Sprintf("%s (%s),", key, string(c[:len(c)-1])))
	}

	b := buffer.Bytes()
	sqlstr := string(b[:len(b)-1])
	sqlstr += fmt.Sprintf(")ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=%d", incNum)

	_, err = conn.Exec(sqlstr)
	return
}

func (conn *MySql) Begin() (*sql.Tx, error) {
	return conn.DBConn.Conn.Begin()
}

func (conn *MySql) Commit(tx *sql.Tx) error {
	return tx.Commit()
}

func (conn *MySql) RollBack(tx *sql.Tx) error {
	return tx.Rollback()
}

func (conn *MySql) GetUid(typ string) (id uint64, err error) {
	uuid.SetNodeID([]byte(typ))
	uid := uuid.NewRandom()
	r := bytes.NewBuffer(uid)
	err = binary.Read(r, binary.BigEndian, &id)
	id = id & 0x7FFFFFFFFFFFFFFF
	return
}
