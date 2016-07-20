package mysqldb

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	. "logicdata/parser"
)

type DBConn struct {
	Conn *sql.DB
	Type string ""
}

/*
数据库操作接口
*/
type SqlWrapper interface {
	Close()
	Exec(query string, args ...interface{}) (result driver.Result, err error)
	Ping() error
	Query(query string, args ...interface{}) (row *sql.Rows, err error)
	QueryTblInfo(tbl_schema, tbl_name string) (fields TblField, empty bool, err error)
	UseDB(db string) (err error)
	DropDB(db string) (err error)
	CreateTable(tbl_name string, fields TblField, incNum int) (err error)
	DropTable(tbl_name string) (err error)
	Begin() (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	RollBack(tx *sql.Tx) error
	GetUid(typ string) (uint64, error)
	GetDB() *sql.DB
}

/*
创建一个新的数据库连接，driverName是使用的数据库类型("mysql")
datasource为连接串，各个数据库都不一样
连接示例：
	NewConn("mysql", root:123456@tcp(127.0.0.1:3306)/nx_district?charset=utf8")
*/
func NewConn(driverName, datasource string) (SqlWrapper, error) {
	switch driverName {
	case "mysql":
		conn, err := NewMysqlConn(datasource)
		if err != nil {
			return nil, err
		}
		conn.DBConn.Conn.SetMaxOpenConns(32)
		conn.DBConn.Conn.SetMaxIdleConns(32)
		return conn, nil
	}

	return nil, errors.New("driver not found")
}
