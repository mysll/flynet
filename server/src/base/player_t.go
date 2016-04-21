package base

import (
	"data/entity"
	"server"
	"share"
)

type Player struct {
	server.Callee
}

func (p *Player) OnLoad(self entity.Entityer, typ int) int {
	//player := self.(*entity.Player)
	if typ == share.LOAD_DB {

	}

	return 1
}

func (c *Player) OnPropertyChange(self entity.Entityer, prop string, old interface{}) int {
	//player := self.(*entity.Player)
	switch prop {
	}
	return 1
}

func (c *Player) OnStore(self entity.Entityer, typ int) int {
	return 1
}

func (c *Player) OnDisconnect(self entity.Entityer) int {
	return 1
}

func (c *Player) OnCommand(self entity.Entityer, sender entity.Entityer, msgid int, msg interface{}) int {
	player := self.(*entity.Player)
	switch msgid {
	case share.PLAYER_FIRST_LAND:
		c.FirstLand(player)
	}
	return 1
}

func (c *Player) FirstLand(player *entity.Player) {

}

func (c *Player) OnReady(self entity.Entityer, first bool) int {
	return 1
}
