package status

import (
	"libs/log"
	"libs/rpc"
	"server"
	"share"
)

type DBUtil struct {
}

func NewDBUtil() *DBUtil {
	util := &DBUtil{}
	return util
}

func (this *DBUtil) CountTest(mailbox rpc.Mailbox, count int32) error {
	log.LogDebug("count,", count)
	return nil
}

func (this *DBUtil) testcall() {
	db := server.GetAppByType("database")
	if db != nil {
		db.Call(nil, "Database.Count", "role_info", "", "DBUtil.CountTest")
	} else {
		log.LogError("db not found")
	}

}

func (this *DBUtil) QueryTest(mailbox rpc.Mailbox, result share.DBRow) error {

	for k, _ := range result {
		v, _ := result.GetString(k)
		log.LogDebug(k, ":", v)
	}
	return nil
}

func (this *DBUtil) queryTest() {
	db := server.GetAppByType("database")
	if db != nil {
		db.Call(nil, "Database.QueryRow", "role_info", "", "", "DBUtil.QueryTest")
	} else {
		log.LogError("db not found")
	}
}

func (this *DBUtil) QueryTest2(mailbox rpc.Mailbox, result []share.DBRow, index, count int) error {
	log.LogDebug(result)
	return nil
}

func (this *DBUtil) queryTest2() {
	db := server.GetAppByType("database")
	if db != nil {
		db.Call(nil, "Database.QueryRows", "role_info", "", "account ASC", 0, 3, "DBUtil.QueryTest2")
	} else {
		log.LogError("db not found")
	}
}

func (this *DBUtil) InsertTest2(mailbox rpc.Mailbox, args share.DBParams, eff int, err string) error {
	log.LogDebug(eff, err)
	return nil
}

func (this *DBUtil) insertTest2() {
	db := server.GetAppByType("database")
	if db != nil {
		db.Call(nil, "Database.InsertRows", "survive_ranking", []string{"id", "name", "score", "userdata"}, []interface{}{uint64(1231213), "test1", 100, []byte{},
			uint64(1231214), "test2", 101, []byte{},
			uint64(1231215), "test3", 102, []byte{},
		}, "DBUtil.InsertTest2", share.DBParams{})
	} else {
		log.LogError("db not found")
	}
}
