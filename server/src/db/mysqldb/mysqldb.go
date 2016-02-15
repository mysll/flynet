package mysqldb

import (
	"errors"
	"libs/log"
	"server"
	"util"
)

var (
	db *MysqlDB
)

type MysqlDB struct {
	pools   int
	sql     SqlWrapper
	Account *Account
	DBRaw   *Database
	dbname  string
	ds      string
	wg      util.WaitGroupWrapper
	limit   int
}

func (self *MysqlDB) InitDB(db string, source string, threads int, entity string, role string, limit int) error {
	var err error
	self.sql, err = NewConn("mysql", source)
	if err != nil {
		return err
	}

	self.pools = threads
	self.ds = source
	self.dbname = db
	self.limit = limit
	if !checkDb(entity, role) {
		return errors.New("database need sync")
	}

	self.Account = NewAccount(self.pools)
	self.DBRaw = NewRaw(self.pools)
	self.Account.Do()
	self.DBRaw.Do()
	server.RegisterRemote("Account", self.Account)
	server.RegisterRemote("Database", self.DBRaw)
	log.LogMessage("connect to mysql:", source)
	return nil
}

func (self *MysqlDB) KeepAlive() {
	self.sql.Ping()
}

func (self *MysqlDB) Close() {
	self.Account.quit = true
	self.DBRaw.quit = true
	self.wg.Wait()
	self.sql.Close()
}

func NewMysqlDB() *MysqlDB {
	if db != nil {
		return db
	}
	db = &MysqlDB{}
	return db
}
