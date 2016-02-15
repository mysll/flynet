package mongodb

import (
	"time"
)

type Counter struct {
	Id_ string `bson:"_id"`
	Seq uint64
}

type RoleInfo struct {
	Uid           uint64 `bson:"_id"`
	Account       string
	Rolename      string
	Createtime    time.Time
	Lastlogintime time.Time
	Locktime      time.Time
	Roleindex     int8
	Roleinfo      string
	Entity        string
	Deleted       int8
	Locked        int8
	Status        int8
	Serverid      string
	Scene         string
	Scene_x       float32
	Scene_y       float32
	Scene_z       float32
}

type Childs struct {
	Parent_Id uint64
	Child_Id  uint64
	Type      string
	Index     int
}
