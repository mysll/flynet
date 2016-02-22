package base

import (
	"libs/log"
	"libs/rpc"
	"pb/c2s"
	"pb/s2c"
	"server"
	"share"

	"github.com/golang/protobuf/proto"
)

type Account struct {
	SendBuf []byte
}

func (a *Account) SelectUser(mailbox rpc.Mailbox, args c2s.Selectuser) error {
	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil {
		//角色没有找到
		return nil
	}
	if player.State != STATE_LOGGED {
		log.LogWarning("player state not logged")
		return nil
	}

	player.ChooseRole = args.GetRolename()
	return App.DbBridge.selectUser(mailbox, args.GetRolename(), int(args.GetRoleindex()))
}

func (a *Account) CreatePlayer(mailbox rpc.Mailbox, args c2s.Create) error {
	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil {
		//角色没有找到
		return nil
	}
	if player.State != STATE_LOGGED {
		log.LogWarning("player state not logged")
		return nil
	}
	obj, err := App.CreateRole("Player", args)
	if err != nil {
		return err
	}

	save := share.GetSaveData(obj)
	err = App.DbBridge.createRole(mailbox, obj, player.Account, args.GetName(), int(args.GetIndex()), save)
	App.Destroy(obj.GetObjId())
	return err
}

func (a *Account) Login(mailbox rpc.Mailbox, args c2s.Enterbase) error {
	if App.Login.checkClient(args.GetUser(), args.GetKey()) {
		if pl := App.Players.AddPlayer(mailbox.Id); pl != nil {
			log.LogMessage("add player:", mailbox)
			pl.Account = args.GetUser()
			pl.State = STATE_LOGGED
			if args.GetRolename() != "" {
				pl.ChooseRole = args.GetRolename()
				return App.DbBridge.selectUser(mailbox, args.GetRolename(), int(args.GetRoleindex()))
			}
			return App.DbBridge.getUserInfo(mailbox, args.GetUser())
		}
		log.LogError("player add failed", mailbox)
		return nil
	} else {
		log.LogDebug(args.GetUser(), args.GetKey())
		err := &s2c.Error{}
		err.ErrorNo = proto.Int32(share.ERROR_LOGIN_FAILED)
		return server.MailTo(nil, &mailbox, "Login.Error", err)
	}
}

func NewAccount() *Account {
	a := &Account{}
	a.SendBuf = make([]byte, 0, 64*1024)
	return a
}
