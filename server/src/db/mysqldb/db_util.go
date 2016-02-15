package mysqldb

import (
	"data/entity"
	"encoding/json"
	"fmt"
	"share"
	"time"
)

func GetRoleUid(account, role string) (uint64, error) {
	r, err := db.sql.Query("SELECT `uid` FROM `role_info` WHERE `account`=? AND `rolename`=?", account, role)
	if err != nil {
		return 0, err
	}

	defer r.Close()
	if !r.Next() {
		return 0, fmt.Errorf("role not found")
	}

	var roleid uint64
	r.Scan(&roleid)
	return roleid, nil
}

func CreateUser(sqlconn SqlWrapper, id uint64, acc string, name string, index int, scene string, x, y, z, dir float32, typ string, data *share.DbSave) error {
	_, err := sqlconn.Exec("INSERT INTO `role_info`(`uid`,`account`,`rolename`,`createtime`,`lastlogintime`,`locktime`,`roleindex`,`roleinfo`,`entity`,`deleted`,`locked`,`status`,`serverid`,`scene`,`scene_x`,`scene_y`,`scene_z`,`scene_dir`,`landtimes`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		id, acc, name, time.Now().Format("2006-01-02 15:04:05"), time.Time{}, time.Time{}, index, data.RoleInfo, typ, 0, 0, 0, "", scene, x, y, z, dir, 0)
	if err != nil {
		return err
	}

	err = SaveItem(sqlconn, true, id, data.Data)
	if err != nil {
		sqlconn.Exec("DELETE FROM `role_info` WHERE `uid`=?", id)
		return err
	}
	return nil
}

func UpdateItem(sqlconn SqlWrapper, id uint64, data *share.SaveEntity) error {
	return SaveItem(sqlconn, false, id, data)
}

func SaveItem(sqlconn SqlWrapper, insert bool, id uint64, data *share.SaveEntity) error {
	var err error
	if data == nil {
		return nil
	}

	childjson, err := json.Marshal(data.Childs)
	if err != nil {
		return err
	}
	//写入自身数据
	if s, ok := data.Obj.(entity.DBSaveLoader); ok {
		if insert {
			err = s.Insert(sqlconn, id, ",`childinfo`", ",?", childjson)
			if err != nil {
				return err
			}
		} else {
			err = s.Update(sqlconn, id, "`childinfo`=?,", "", childjson)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func LoadUser(sqlconn SqlWrapper, uid uint64, ent string) (loaddata share.DbSave, err error) {
	loaddata.Data, err = LoadEntity(sqlconn, uid, ent, 0)
	return
}

func parseEntity(data *share.SaveEntity) error {

	if data == nil {
		return nil
	}

	bdata := data.Obj.(map[string]interface{})
	e := entity.CreateSaveLoader(data.Typ)

	err := e.Unmarshal(bdata)
	if err != nil {
		return err
	}

	data.Obj = e
	if data.Base != nil {
		err = parseEntity(data.Base)
		if err != nil {
			return err
		}
	}

	for _, c := range data.Childs {
		err = parseEntity(c)
		if err != nil {
			return err
		}
	}

	return nil

}

func LoadEntity(sqlconn SqlWrapper, uid uint64, ent string, index int) (*share.SaveEntity, error) {
	si := &share.SaveEntity{}
	si.Typ = ent
	si.DBId = uid
	si.Index = index
	e := entity.CreateSaveLoader(ent)
	var childjson []byte
	err := e.Load(sqlconn, uid, ",`childinfo`", &childjson)
	if err != nil {
		return nil, err
	}
	si.Obj = e

	var childinfo []*share.SaveEntity
	json.Unmarshal(childjson, &childinfo)
	for _, c := range childinfo {
		parseEntity(c)
	}
	si.Childs = childinfo
	return si, nil
}
