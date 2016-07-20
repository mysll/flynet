package mysqldb

import (
	"fmt"
	. "logicdata/parser"
	l "server/libs/log"
	"strings"
)

func checkTable(tbl_name string, tblinfo TblField) bool {
	if oldfield, empty, _ := db.sql.QueryTblInfo(db.dbname, tbl_name); empty {
		return false
	} else {
		for _, newf := range tblinfo.FieldList {
			find := false
			for _, oldf := range oldfield.FieldList {
				if strings.ToLower(newf.Name) == strings.ToLower(oldf.Name) {
					if strings.ToLower(oldf.Type) != strings.ToLower(newf.Type) {
						return false
					}
					find = true
					break
				}
			}
			if !find {
				return false
			}
		}
	}
	return true
}

func checkObject(obj *Object) bool {
	tbl_name := "tbl_" + strings.ToLower(obj.Name)
	tblinfo := obj.CreateTable()
	if !checkTable(tbl_name, tblinfo) {
		return false
	}

	records := obj.CreateRecordTable()
	for k := range records {
		tbl_name = strings.ToLower(fmt.Sprintf("tbl_%s_%s", obj.Name, k))

		if !checkTable(tbl_name, records[k]) {
			return false
		}
	}

	return true
}

func checkSysDb() bool {
	//检查系统表
	if !checkTable("sys_uid", idtbl) {
		return false
	}

	//用户帐号信息
	if !checkTable("role_info", roletbl) {
		return false
	}

	for k, t := range userTables {
		if !checkTable(k, t) {
			return false
		}
	}
	return true
}

func checkDb(path string, role string) bool {

	if !checkSysDb() {
		l.TraceInfo("dbmgr", ` system table need sync, you need use "-sync" to sync the database`)
		return false
	}

	LoadAllDef(path)

	for k, obj := range Defs {
		if obj.Name == role || obj.Persistent == "true" {
			if !checkObject(obj) {
				l.TraceInfo("dbmgr", k, ` is changed, you need use "-sync" to sync the database `)
				return false
			}
		}
	}
	/*
		dir, _ := os.Open(path)
		files, _ := dir.Readdir(0)

		for _, f := range files {
			if !f.IsDir() {
				obj := ParseEntity(path + "/" + f.Name())
				if obj.Name == role {
					if !checkObject(obj) {
						l.TraceInfo("dbmgr", f.Name(), ` is changed, you need use "-sync" to sync the database `)
						return false
					}
				}
			}
		}
	*/

	r, _ := db.sql.Query("SELECT * FROM `sys_uid` LIMIT 1 ")
	if !r.Next() {
		db.sql.GetUid("userid")
	}
	r.Close()
	l.TraceInfo("dbmgr", "check db complete")

	return true
}
