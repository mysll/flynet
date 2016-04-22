package base

import (
	. "data/datatype"
	"data/entity"
	"fmt"
	"libs/log"
	"pb/c2s"
	"server"
)

const (
	DATA_VERSION = 1
)

type RoleCallee struct {
	server.Callee
}

func (r *RoleCallee) OnCreateRole(self entity.Entityer, args interface{}) int {
	createinfo := args.(*c2s.Create)
	player := self.(*entity.Player)
	err := App.LoadFromConfig(self, fmt.Sprintf("%d", createinfo.GetRoleid()))
	player.SetConfig("")
	if err != nil {
		log.LogError(err)
		return -1
	}

	player.SetDataVer(DATA_VERSION)
	player.SetName(createinfo.GetName())
	player.SetSex(int8(createinfo.GetSex()))

	App.SetLandpos(self, server.Transform{"hall", Vector3{0, 0, 0}, 0})
	App.SetRoleInfo(self, fmt.Sprintf("%d", player.GetSex()))
	return 1
}
