package db

import (
	"db/mysqldb"
)

type Sync interface {
	SyncDB(path string, drop bool, role string)
}

func CreateSyncDB(db string, datasource string) Sync {
	return mysqldb.CreateSyncDB(db, datasource)
}
