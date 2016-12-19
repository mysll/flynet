package base

import (
	"logicdata/entity"
	"server"
	"server/data/datatype"
	"server/share"
)

type Player struct {
	server.Callee
}

func (p *Player) OnLoad(self datatype.Entity, typ int) int {
	//player := self.(*entity.Player)
	if typ == share.LOAD_DB {

	}

	return 1
}

func (c *Player) OnPropertyChange(self datatype.Entity, prop string, old interface{}) int {
	//player := self.(*entity.Player)
	switch prop {
	}
	return 1
}

func (c *Player) OnStore(self datatype.Entity, typ int) int {
	return 1
}

func (c *Player) OnDisconnect(self datatype.Entity) int {
	return 1
}

func (c *Player) OnCommand(self datatype.Entity, sender datatype.Entity, msgid int, msg interface{}) int {
	player := self.(*entity.Player)
	switch msgid {
	case share.PLAYER_FIRST_LAND:
		c.FirstLand(player)
	}
	return 1
}

func (c *Player) FirstLand(player *entity.Player) {

}

func (c *Player) OnReady(self datatype.Entity, first bool) int {
	return 1
}
