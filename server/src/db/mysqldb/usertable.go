package mysqldb

import (
	. "logicdata/parser"
)

var (
	userTables = map[string]TblField{
		"main_ranking": TblField{
			FieldList: []Field{
				Field{"id", "BIGINT(20) UNSIGNED", false, false},
				Field{"score", "BIGINT(20) UNSIGNED", false, false},
				Field{"name", "VARCHAR(32)", false, false},
				Field{"userdata", "TEXT", true, false},
			},
			IndexInfo: map[string][]string{
				"PRIMARY":           []string{"id"},
				"INDEX `nameindex`": []string{"name"},
			},
		},
		"survive_ranking": TblField{
			FieldList: []Field{
				Field{"id", "BIGINT(20) UNSIGNED", false, false},
				Field{"score", "BIGINT(20) UNSIGNED", false, false},
				Field{"name", "VARCHAR(32)", false, false},
				Field{"userdata", "TEXT", true, false},
			},
			IndexInfo: map[string][]string{
				"PRIMARY":           []string{"id"},
				"INDEX `nameindex`": []string{"name"},
			},
		},
		"pub_data": TblField{
			FieldList: []Field{
				Field{"key", "VARCHAR(64)", false, false},
				Field{"value", "VARCHAR(255)", false, false},
			},
			IndexInfo: map[string][]string{
				"PRIMARY": []string{"key"},
			},
		},
		"letter": TblField{
			FieldList: []Field{
				Field{"serial_no", "BIGINT(20) UNSIGNED", false, true},
				Field{"send_time", "DATETIME", true, false},
				Field{"msg_acc", "VARCHAR(32)", true, false},
				Field{"msg_name", "VARCHAR(32)", true, false},
				Field{"msg_uid", "BIGINT(20) UNSIGNED", true, false},
				Field{"source", "BIGINT(20) UNSIGNED", true, false},
				Field{"source_name", "VARCHAR(32)", true, false},
				Field{"msg_type", "INT(10)", true, false},
				Field{"msg_title", "VARCHAR(255)", true, false},
				Field{"msg_content", "VARCHAR(255)", true, false},
				Field{"msg_appendix", "VARCHAR(255)", true, false},
			},
			IndexInfo: map[string][]string{
				"PRIMARY":           []string{"serial_no"},
				"INDEX `roleindex`": []string{"msg_uid"},
			},
		},
		"log_data": TblField{
			FieldList: []Field{
				Field{"serial_no", "BIGINT(20) UNSIGNED", false, true},
				Field{"log_time", "DATETIME", true, false},
				Field{"log_name", "VARCHAR(32)", true, false},
				Field{"log_source", "INT(10) UNSIGNED", true, false},
				Field{"log_type", "INT(10) UNSIGNED", true, false},
				Field{"log_content", "VARCHAR(255)", true, false},
				Field{"log_comment", "VARCHAR(255)", true, false},
			},
			IndexInfo: map[string][]string{
				"PRIMARY": []string{"serial_no"},
			},
		},
	}
)
