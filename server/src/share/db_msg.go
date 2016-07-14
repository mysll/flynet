package share

import (
	"data/datatype"
	"fmt"
	"strconv"

	"github.com/go-sql-driver/mysql"
)

type DBParams map[string]interface{}
type DBRow map[string][]byte

type LetterInfo struct {
	Serial_no                uint64         //序列号
	Send_time                mysql.NullTime //发送时间
	Source                   uint64         //发信人UID
	Source_name              string         //发信人姓名
	Recv_acc, Recv_name      string         //接收人帐号和角色名
	Msg_type                 int32          //消息类型
	Title, Content, Appendix string         //消息标题，内容，附件
}

func (row DBRow) Count() int {
	return len(row)
}

func (row DBRow) GetInt32(key string) (int32, error) {
	if val, ok := row[key]; ok {
		i64, err := strconv.ParseInt(string(val), 10, 32)
		if err != nil {
			return 0, err
		}

		return int32(i64), err

	}

	return 0, fmt.Errorf("%s not found", key)

}

func (row DBRow) GetUint32(key string) (uint32, error) {
	if val, ok := row[key]; ok {
		ui64, err := strconv.ParseUint(string(val), 10, 32)
		if err != nil {
			return 0, err
		}

		return uint32(ui64), err

	}

	return 0, fmt.Errorf("%s not found", key)

}

func (row DBRow) GetInt64(key string) (int64, error) {
	if val, ok := row[key]; ok {
		return strconv.ParseInt(string(val), 10, 64)
	}

	return 0, fmt.Errorf("%s not found", key)
}

func (row DBRow) GetUint64(key string) (uint64, error) {
	if val, ok := row[key]; ok {
		return strconv.ParseUint(string(val), 10, 64)
	}

	return 0, fmt.Errorf("%s not found", key)
}

func (row DBRow) GetFloat32(key string) (float32, error) {
	if val, ok := row[key]; ok {
		i64, err := strconv.ParseFloat(string(val), 32)
		if err != nil {
			return 0, err
		}
		return float32(i64), err
	}

	return 0, fmt.Errorf("%s not found", key)
}

func (row DBRow) GetFloat64(key string) (float64, error) {
	if val, ok := row[key]; ok {
		return strconv.ParseFloat(string(val), 64)
	}

	return 0, fmt.Errorf("%s not found", key)
}

func (row DBRow) GetString(key string) (string, error) {
	if val, ok := row[key]; ok {
		return string(val), nil
	}

	return "", fmt.Errorf("%s not found", key)
}

func (row DBRow) GetBytes(key string) ([]byte, error) {
	if val, ok := row[key]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("%s not found", key)
}

type SaveEntity struct {
	DBId   uint64        `bson:"_id" json:"d"`
	Typ    string        `json:"t"`
	Index  int           `json:"i"`
	Obj    interface{}   `json:"o"`
	Base   *SaveEntity   `json:"b"`
	Childs []*SaveEntity `json:"c"`
}

type DbSave struct {
	Data     *SaveEntity
	RoleInfo string
}

type CreateUser struct {
	Account  string
	Name     string
	Index    int
	Scene    string
	X        float32
	Y        float32
	Z        float32
	Dir      float32
	SaveData DbSave
}

type UpdateUser struct {
	Type     int
	Account  string
	Name     string
	Scene    string
	X        float32
	Y        float32
	Z        float32
	Dir      float32
	SaveData DbSave
}

type ClearUser struct {
	Account string
	Name    string
}

type ObjectInfo struct {
	ObjId datatype.ObjectID
	DBId  uint64
}

type UpdateUserBak struct {
	Infos []ObjectInfo
}

type LoadUser struct {
	Account  string
	RoleName string
	Index    int
}

type LoadUserBak struct {
	Account   string
	Name      string
	Scene     string
	X         float32
	Y         float32
	Z         float32
	Dir       float32
	LandTimes int32
	Data      *DbSave
}

func GetSaveData(ent datatype.Entityer) *DbSave {
	data := &DbSave{}
	data.Data = GetEntityData(ent, false, 0)
	roleinfo := ent.GetExtraData("roleinfo")
	if roleinfo != nil {
		data.RoleInfo = roleinfo.(string)
	}
	return data
}

func GetEntityData(ent datatype.Entityer, base bool, depth int) *SaveEntity {
	if !base && !ent.IsSave() {
		return nil
	}
	s := &SaveEntity{}
	ent.SyncToDb()
	s.Typ = ent.ObjTypeName()
	s.DBId = ent.GetDbId()
	s.Obj = ent.GetSaveLoader()
	s.Index = ent.GetIndex()
	if ent.Base() != nil {
		s.Base = GetEntityData(ent.Base(), true, depth)
	}
	if base {
		return s
	}
	if depth > 1 {
		return s
	}

	clds := ent.GetChilds()
	l := len(clds)
	if l > 0 {
		s.Childs = make([]*SaveEntity, 0, l)
		for _, e := range clds {
			if e != nil {
				if child := GetEntityData(e, false, depth+1); child != nil {
					s.Childs = append(s.Childs, child)
				}
			}
		}
	}

	return s
}
