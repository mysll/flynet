package base

import (
	"data/entity"
	"libs/rpc"
	"pb/s2c"
	"server"
	"share"

	"github.com/golang/protobuf/proto"
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
	player := self.(*entity.Player)
	roleinfo := &s2c.RoleProperty{}
	roleinfo.Name = proto.String(player.GetName())
	roleinfo.Hp = proto.Int32(player.GetHP())
	roleinfo.Attack = proto.Int32(player.GetAttack())
	roleinfo.Defend = proto.Int32(player.GetDefence())
	roleinfo.Critical = proto.Int32(player.GetCrit())
	roleinfo.Strength = proto.Int32(player.GetStrength())
	roleinfo.MaxStrength = proto.Int32(player.GetMaxStrength())

	bi := make([]*s2c.BagItem, 0, 4)
	if player.GetItem0() != "" {
		item := &s2c.BagItem{}
		item.Slot = proto.Int32(1)
		item.Itemid = proto.String(player.GetItem0())
		item.Itemnum = proto.Int32(player.GetItem0Num())
		bi = append(bi, item)
	}

	if player.GetItem1() != "" {
		item := &s2c.BagItem{}
		item.Slot = proto.Int32(2)
		item.Itemid = proto.String(player.GetItem1())
		item.Itemnum = proto.Int32(player.GetItem1Num())
		bi = append(bi, item)
	}

	if player.GetWeapon() != "" {
		item := &s2c.BagItem{}
		item.Slot = proto.Int32(3)
		item.Itemid = proto.String(player.GetWeapon())
		item.Itemnum = proto.Int32(player.GetWeaponNum())
		bi = append(bi, item)
	}

	if player.GetEquip() != "" {
		item := &s2c.BagItem{}
		item.Slot = proto.Int32(4)
		item.Itemid = proto.String(player.GetEquip())
		item.Itemnum = proto.Int32(player.GetEquipNum())
		bi = append(bi, item)
	}

	roleinfo.Items = bi
	mb := player.GetExtraData("mailbox").(rpc.Mailbox)
	server.MailTo(nil, &mb, "Role.Info", roleinfo)

	return 1
}
