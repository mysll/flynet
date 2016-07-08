package mysqldb

import (
	"errors"
	"libs/log"
	"server"
)

var (
	db *MysqlDB
)

type MysqlDB struct {
	pools      int
	sql        SqlWrapper
	Account    *Account
	DBRaw      *Database
	dbname     string
	ds         string
	nameunique bool
	//wg      util.WaitGroupWrapper
	limit int
}

func (self *MysqlDB) InitDB(db string, source string, threads int, entity string, role string, limit int, nameunique bool) error {
	var err error
	self.sql, err = NewConn("mysql", source)
	if err != nil {
		return err
	}

	self.pools = threads
	self.ds = source
	self.dbname = db
	self.limit = limit
	self.nameunique = nameunique
	if !checkDb(entity, role) {
		return errors.New("database need sync")
	}
	self.Account = NewAccount(self.pools)
	self.DBRaw = NewRaw(self.pools)
	server.RegisterRemote("Account", self.Account)
	server.RegisterRemote("Database", self.DBRaw)
	self.Account.Start()
	self.DBRaw.Start()
	log.LogMessage("connect to mysql:", source)
	return nil
}

func (self *MysqlDB) KeepAlive() {
	self.sql.Ping()
}

func (self *MysqlDB) Close() {
	self.Account.Quit = true
	self.DBRaw.Quit = true
	self.Account.Wait()
	self.DBRaw.Wait()
	self.sql.Close()
}

func NewMysqlDB() *MysqlDB {
	if db != nil {
		return db
	}
	db = &MysqlDB{}
	return db
}
