package base

import (
	"pb/c2s"
	"pb/s2c"
	"server"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"

	"github.com/golang/protobuf/proto"
)

type Account struct {
	SendBuf []byte
}

func (t *Account) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("SelectUser", t.SelectUser)
	s.RegisterCallback("CreatePlayer", t.CreatePlayer)
	s.RegisterCallback("Login", t.Login)
}

func (a *Account) SelectUser(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	args := &c2s.Selectuser{}
	if server.Check(server.ParseProto(msg, args)) {
		return 0, nil
	}
	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil {
		//角色没有找到
		return 0, nil
	}
	if player.State != STATE_LOGGED {
		log.LogWarning("player state not logged")
		return 0, nil
	}

	player.ChooseRole = args.GetRolename()
	err := App.DbBridge.selectUser(mailbox, player.Account, args.GetRolename(), int(args.GetRoleindex()))
	if err != nil {
		log.LogError(err)
	}

	return 0, nil
}

func (a *Account) CreatePlayer(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	args := &c2s.Create{}
	if server.Check(server.ParseProto(msg, args)) {
		return 0, nil
	}

	player := App.Players.GetPlayer(mailbox.Id)
	if player == nil {
		//角色没有找到
		return 0, nil
	}
	if player.State != STATE_LOGGED {
		log.LogWarning("player state not logged")
		return 0, nil
	}
	obj, err := App.CreateRole("Player", args)
	if err != nil {
		log.LogError(err)
		return 0, nil
	}

	save := share.GetSaveData(obj)
	server.Check(App.DbBridge.createRole(mailbox, obj, player.Account, args.GetName(), int(args.GetIndex()), save))
	App.Destroy(obj.GetObjId())
	return 0, nil
}

func (a *Account) Login(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	args := &c2s.Enterbase{}
	if server.Check(server.ParseProto(msg, args)) {
		return 0, nil
	}
	if App.Login.checkClient(args.GetUser(), args.GetKey()) {
		if pl := App.Players.AddPlayer(mailbox.Id); pl != nil {
			log.LogMessage("add player:", mailbox)
			pl.Account = args.GetUser()
			pl.State = STATE_LOGGED
			if args.GetRolename() != "" {
				pl.ChooseRole = args.GetRolename()
				server.Check(App.DbBridge.selectUser(mailbox, pl.Account, args.GetRolename(), int(args.GetRoleindex())))
				return 0, nil
			}
			server.Check(App.DbBridge.getUserInfo(mailbox, args.GetUser()))
			return 0, nil
		}
		log.LogError("player add failed", mailbox)
		return 0, nil
	} else {
		log.LogDebug(args.GetUser(), args.GetKey())
		err := &s2c.Error{}
		err.ErrorNo = proto.Int32(share.ERROR_LOGIN_FAILED)
		server.Check(server.MailTo(nil, &mailbox, "Login.Error", err))
		return 0, nil
	}
}

func NewAccount() *Account {
	a := &Account{}
	a.SendBuf = make([]byte, 0, 64*1024)
	return a
}
