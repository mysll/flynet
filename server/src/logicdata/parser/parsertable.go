package parser

import (
	"fmt"
	"server/data/datatype"
	"strconv"
	"strings"
)

/*
字段类型
*/
type Field struct {
	Name    string
	Type    string
	IsNull  bool
	AutoInc bool
}

type TblField struct {
	FieldList []Field
	IndexInfo map[string][]string
}

func GetDbType(t string, size int) string {
	st := ""
	switch t {
	case "int8":
		st = "TINYINT(4)"
	case "int16":
		st = "SMALLINT(6)"
	case "int32":
		st = "INT(10)"
	case "int64":
		st = "BIGINT(20)"
	case "uint8":
		st = "TINYINT(4) UNSIGNED"
	case "uint16":
		st = "SMALLINT(6) UNSIGNED"
	case "uint32":
		st = "INT(10) UNSIGNED"
	case "uint64":
		st = "BIGINT(20) UNSIGNED"
	case "float32":
		st = "FLOAT"
	case "float64":
		st = "DOUBLE"
	case "string":
		if size == 0 {
			size = 50
		}
		if size < 256 {
			st = fmt.Sprintf("VARCHAR(%d)", size)
		} else if size < 65536 {
			st = "TEXT"
		} else {
			st = "MEDIUMTEXT"
		}
	default:
		return ""
	}

	return st
}

func (obj *Object) CreateTable() TblField {
	tbl := TblField{}
	tbl.FieldList = make([]Field, 0, len(obj.Propertys)+2)
	tbl.FieldList = append(tbl.FieldList, []Field{Field{"id", "BIGINT(20) UNSIGNED", false, false},
		Field{"capacity", "SMALLINT(6)", true, false},
		Field{"configid", "CHAR(64)", true, false},
	}...)
	for _, p := range obj.Propertys {
		if p.Save == "true" {
			st := GetDbType(p.Type, p.Len)
			if st != "" {
				tbl.FieldList = append(tbl.FieldList, Field{"p_" + strings.ToLower(p.Name), st, false, false})
			} else {
				usertype := datatype.GetUserType(p.Type, strings.ToLower(p.Name))
				if usertype != nil {
					for _, ut := range usertype {
						if len(ut) != 3 {
							panic("user type serial failed")
						}

						typelen, e := strconv.ParseInt(ut[2], 10, 32)
						if e != nil {
							panic("user type length error")
						}
						tbl.FieldList = append(tbl.FieldList, Field{ut[0], GetDbType(ut[1], int(typelen)), false, false})
					}
				}
			}
		}
	}

	for _, r := range obj.Records {
		if r.Save == "true" && r.Type != "" {
			tbl.FieldList = append(tbl.FieldList, Field{"r_" + strings.ToLower(r.Name), strings.ToUpper(r.Type), false, false})
		}
	}
	tbl.FieldList = append(tbl.FieldList, Field{"childinfo", "MEDIUMTEXT", false, false})
	tbl.IndexInfo = make(map[string][]string)
	tbl.IndexInfo["PRIMARY"] = []string{"id"}

	return tbl

}

func (obj *Object) CreateRecordTable() map[string]TblField {

	records := make(map[string]TblField)
	for _, r := range obj.Records {
		if r.Save == "true" && r.Type == "" {

			tbl := TblField{}
			tbl.FieldList = make([]Field, 0, len(r.Columns)+3)

			tbl.FieldList = append(tbl.FieldList, Field{"id", "BIGINT(20) UNSIGNED", false, false})
			tbl.FieldList = append(tbl.FieldList, Field{"index", "SMALLINT(6)", false, false})
			tbl.FieldList = append(tbl.FieldList, Field{"delete", "TINYINT(1) UNSIGNED", false, false})

			for _, p := range r.Columns {
				st := GetDbType(p.Type, p.Len)
				if st != "" {
					tbl.FieldList = append(tbl.FieldList, Field{"p_" + strings.ToLower(p.Name), st, false, false})
				} else {
					usertype := datatype.GetUserType(p.Type, strings.ToLower(p.Name))
					if usertype != nil {
						for _, ut := range usertype {
							if len(ut) != 3 {
								panic("user type serial failed")
							}

							typelen, e := strconv.ParseInt(ut[2], 10, 32)
							if e != nil {
								panic("user type length error")
							}
							tbl.FieldList = append(tbl.FieldList, Field{ut[0], GetDbType(ut[1], int(typelen)), false, false})
						}
					}
				}
			}

			tbl.IndexInfo = make(map[string][]string)
			tbl.IndexInfo["INDEX id"] = []string{"id"}

			records[r.Name] = tbl
		}
	}

	return records

}
