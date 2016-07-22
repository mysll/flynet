package server

import (
	"fmt"
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
)

type TeleportHelper struct {
}

func (t *TeleportHelper) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("AddPlayer", t.AddPlayer)
	s.RegisterCallback("SyncBaseFromSceneData", t.SyncBaseFromSceneData)
}

//当前服务器增加entity
func (t *TeleportHelper) AddPlayer(sender rpc.Mailbox, msg *rpc.Message) *rpc.Message {

	var playerinfo *datatype.EntityInfo
	var args []interface{}
	if !Check(ParseArgs(msg, &args)) || len(args) < 1 {
		log.LogError("args parse error")
		reply, _ := CreateMessage(sender, false)
		return reply
	}

	var ok bool
	if playerinfo, ok = args[0].(*datatype.EntityInfo); !ok {
		log.LogError("args parse error")
		reply, _ := CreateMessage(sender, false)
		return reply
	}

	pl, err := core.CreateFromArchive(playerinfo, nil)
	if err != nil {
		log.LogError(err)
		reply, _ := CreateMessage(sender, false)
		return reply
	}

	if !core.apper.OnTeleportFromBase(args[1:], pl) {
		core.Destroy(pl.GetObjId())
		reply, _ := CreateMessage(sender, false)
		return reply
	}

	reply, _ := CreateMessage(sender, true)
	return reply
}

//回调函数
func (t *TeleportHelper) AddPlayerBack(msg *rpc.Message) {
	var mailbox rpc.Mailbox
	var result bool
	if msg != nil {
		ParseArgs(msg, &mailbox, &result)
	}

	core.apper.OnSceneTeleported(mailbox, result)
}

func (t *TeleportHelper) SyncBaseFromSceneData(sender rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	var args []interface{}
	if !Check(ParseArgs(msg, &args)) || len(args) < 1 {
		reply, _ := CreateMessage(sender, false)
		return reply
	}

	result := core.apper.OnTeleportFromScene(args[0], args[1:])
	reply, _ := CreateMessage(sender, result)
	return reply
}

func (t *TeleportHelper) SyncBaseFromSceneDataBack(msg *rpc.Message) {
	var mailbox rpc.Mailbox
	var result bool
	if msg != nil {
		ParseArgs(msg, &mailbox, &result)
	}

	core.apper.OnBaseTeleported(mailbox, result)
}

//传送到场景
func (t *TeleportHelper) teleport(app *RemoteApp, player datatype.Entityer, mailbox rpc.Mailbox, args ...interface{}) error {
	playerinfo, err := share.GetItemInfo(player, false)
	if err != nil {
		return err
	}
	infos := make([]interface{}, 0, len(args)+1)
	infos = append(infos, playerinfo)
	infos = append(infos, args)
	app.CallBack(&mailbox, "Teleport.AddPlayer", t.AddPlayerBack, infos)
	return nil
}

//传送回base
func (t *TeleportHelper) teleportToBase(app *RemoteApp, object datatype.Entityer, mailbox rpc.Mailbox, args ...interface{}) error {
	sd := object.GetSceneData()
	infos := make([]interface{}, 0, len(args)+1)
	infos = append(infos, sd)
	infos = append(infos, args)
	return app.CallBack(&mailbox, "Teleport.SyncBaseFromSceneData", t.SyncBaseFromSceneDataBack, infos)
}

//从base传送到场景
func (k *Kernel) TeleportToAppByName(appname string, player datatype.Entityer, mailbox rpc.Mailbox, args ...interface{}) error {
	if player == nil {
		return fmt.Errorf("player is nil")
	}
	app := GetAppByName(appname)
	if app == nil {
		return ErrAppNotFound
	}

	return core.teleport.teleport(app, player, mailbox, args...)
}

//从base传送到场景
func (k *Kernel) TeleportToApp(appid int32, player datatype.Entityer, mailbox rpc.Mailbox, args ...interface{}) error {
	if player == nil {
		return fmt.Errorf("player is nil")
	}
	app := GetAppById(appid)
	if app == nil {
		return ErrAppNotFound
	}

	return core.teleport.teleport(app, player, mailbox, args...)
}

//从场景返回base
func (k *Kernel) TeleportToBaseByName(basename string, object datatype.Entityer, mailbox rpc.Mailbox, args ...interface{}) error {
	if object == nil {
		return fmt.Errorf("object is nil")
	}
	app := GetAppByName(basename)
	if app == nil {
		return ErrAppNotFound
	}

	return core.teleport.teleportToBase(app, object, mailbox, args...)
}

//从场景返回base
func (k *Kernel) TeleportToBaseById(baseid int32, object datatype.Entityer, mailbox rpc.Mailbox, args ...interface{}) error {
	if object == nil {
		return fmt.Errorf("object is nil")
	}
	app := GetAppById(baseid)
	if app == nil {
		return ErrAppNotFound
	}

	return core.teleport.teleportToBase(app, object, mailbox, args...)
}
