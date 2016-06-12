package mysqldb

import (
	"fmt"
	l "libs/log"
	. "libs/parser"
	"strings"
)

var (

	//ID表
	idtbl = TblField{
		FieldList: []Field{
			Field{"id", "BIGINT(20) UNSIGNED", false, true},
			Field{"stub", "VARCHAR(10)", false, false},
		},
		IndexInfo: map[string][]string{
			"PRIMARY":             []string{"id"},
			"UNIQUE INDEX `stub`": []string{"stub"},
		},
	}
	//角色表
	roletbl = TblField{
		FieldList: []Field{
			Field{"uid", "BIGINT(20) UNSIGNED", false, false},
			Field{"account", "CHAR(100)", false, false},
			Field{"rolename", "CHAR(64)", false, false},
			Field{"createtime", "DATETIME", false, false},
			Field{"lastlogintime", "DATETIME", false, false},
			Field{"locktime", "DATETIME", false, false},
			Field{"roleindex", "TINYINT(1) UNSIGNED", false, false},
			Field{"roleinfo", "TEXT", true, false},
			Field{"entity", "CHAR(100)", false, false},
			Field{"deleted", "TINYINT(1) UNSIGNED", true, false},
			Field{"locked", "TINYINT(1) UNSIGNED", true, false},
			Field{"status", "TINYINT(1) UNSIGNED", false, false},
			Field{"serverid", "CHAR(64)", true, false},
			Field{"scene", "CHAR(64)", true, false},
			Field{"scene_x", "FLOAT", true, false},
			Field{"scene_y", "FLOAT", true, false},
			Field{"scene_z", "FLOAT", true, false},
			Field{"scene_dir", "FLOAT", true, false},
			Field{"landtimes", "INT(10) UNSIGNED", false, false},
		},
		IndexInfo: map[string][]string{
			"PRIMARY":      []string{"uid"},
			"INDEX `role`": []string{"account", "rolename"},
		},
	}
)

type Sync struct {
	conn       *MySql
	db         string
	datasrouce string
}

func (s *Sync) syncTable(tbl_name string, tblinfo TblField, incNum int) error {
	if oldfield, empty, _ := s.conn.QueryTblInfo(s.db, tbl_name); empty {
		err := s.conn.CreateTable(tbl_name, tblinfo, incNum)
		if err != nil {
			panic(err)
		}
		l.TraceInfo("sync", "create table ", tbl_name, " ok")
	} else {
		alert := make([]Field, 0, 10)
		add := make([]Field, 0, 10)
		for _, newf := range tblinfo.FieldList {
			find := false
			for _, oldf := range oldfield.FieldList {
				if strings.ToLower(newf.Name) == strings.ToLower(oldf.Name) {
					if strings.ToLower(oldf.Type) != strings.ToLower(newf.Type) {
						alert = append(alert, newf)
					}
					find = true
					break
				}
			}
			if !find {
				add = append(add, newf)
			}
		}

		del := make([]Field, 0, 10)
		for _, oldf := range oldfield.FieldList {
			find := false
			for _, newf := range tblinfo.FieldList {
				if strings.ToLower(newf.Name) == strings.ToLower(oldf.Name) {
					find = true
					break
				}
			}
			if !find {
				del = append(del, oldf)
			}
		}

		for _, f := range alert {
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

			sql := fmt.Sprintf("ALTER TABLE `%s` CHANGE `%s` `%s` %s %s %s;", tbl_name, f.Name, f.Name, f.Type, nulltype, autoinc)
			_, err := s.conn.Exec(sql)
			if err != nil {
				panic(err)
			}

			l.TraceInfo("sync", sql, " ok")
		}

		for _, f := range add {
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

			/*if len(del) > 0 {
				sql := fmt.Sprintf("ALTER TABLE `%s` CHANGE `%s` `%s` %s %s %s;", tbl_name, del[0].Name, f.Name, f.Type, nulltype, autoinc)
				_, err := s.conn.Exec(sql)
				if err != nil {
					panic(err)
				}

				l.TraceInfo("sync", sql, " ok")
				del = del[1:]
				break
			} */
			sql := fmt.Sprintf("ALTER TABLE `%s` ADD `%s` %s %s %s;", tbl_name, f.Name, f.Type, nulltype, autoinc)
			_, err := s.conn.Exec(sql)
			if err != nil {
				panic(err)
			}

			l.TraceInfo("sync", sql, "ok")

		}

		for _, f := range del {
			sql := fmt.Sprintf("ALTER TABLE `%s`  DROP `%s` ;", tbl_name, f.Name)
			_, err := s.conn.Exec(sql)
			if err != nil {
				panic(err)
			}

			l.TraceInfo("sync", sql, " ok")
		}
	}

	return nil
}

func (s *Sync) syncObject(obj *Object) {
	tbl_name := "tbl_" + strings.ToLower(obj.Name)
	tblinfo := obj.CreateTable()
	s.syncTable(tbl_name, tblinfo, 0)

	records := obj.CreateRecordTable()
	for k := range records {
		tbl_name = strings.ToLower(fmt.Sprintf("tbl_%s_%s", obj.Name, k))

		s.syncTable(tbl_name, records[k], 0)
	}
}

func (s *Sync) syncSysDb() {
	//取用户名
	row, _ := s.conn.Query("SELECT CURRENT_USER()")
	row.Next()
	var user string
	row.Scan(&user)
	row.Close()
	s.syncTable("sys_uid", idtbl, 1000)
	s.syncTable("role_info", roletbl, 0)
}

func (s *Sync) SyncDB(path string, drop bool, role string) {

	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
		s.conn.Close()
	}()
	if drop {
		s.conn.DropDB(s.db)
		l.TraceInfo("sync", "drop db ", s.db)
	}

	s.conn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", s.db))
	s.conn.UseDB(s.db)
	s.syncSysDb()

	for k, t := range userTables {
		s.syncTable(k, t, 0)
	}

	LoadAllDef(path)

	for _, obj := range Defs {
		if obj.Name == role || obj.Persistent == "true" {
			s.syncObject(obj)
			l.TraceInfo("sync", "process ", obj.Name, " complete")
		}

	}
	l.TraceInfo("sync", "all complete")
}

func CreateSyncDB(db string, datasource string) *Sync {
	s := &Sync{}
	c, err := NewMysqlConn(datasource)
	if err != nil {
		l.TraceInfo("dbmgr", err)
		panic(err)
	}

	s.conn = conn
	s.db = db
	s.datasrouce = datasource
	return s
}
