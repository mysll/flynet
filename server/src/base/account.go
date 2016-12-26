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
	p := App.Players.FindPlayer(mailbox.Uid)
	if p == nil {
		//角色没有找到
		return 0, nil
	}
	player := p.(*BasePlayer)
	if player.State != STATE_LOGGED {
		log.LogWarning("player state not logged")
		return 0, nil
	}

	player.ChooseRole = args.GetRolename()
	player.Name = player.ChooseRole
	player.UpdateHash()
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

	p := App.Players.FindPlayer(mailbox.Uid)
	if p == nil {
		//角色没有找到
		return 0, nil
	}
	player := p.(*BasePlayer)
	if player.State != STATE_LOGGED {
		log.LogWarning("player state not logged")
		return 0, nil
	}
	obj, err := App.Kernel().CreateRole("Player", args)
	if err != nil {
		log.LogError(err)
		return 0, nil
	}

	save := share.GetSaveData(obj)
	server.Check(App.DbBridge.createRole(mailbox, obj, player.Account, args.GetName(), int(args.GetIndex()), save))
	App.Kernel().Destroy(obj.ObjectId())
	return 0, nil
}

func (a *Account) Login(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	args := &c2s.Enterbase{}
	if server.Check(server.ParseProto(msg, args)) {
		return 0, nil
	}
	if App.Login.checkClient(args.GetUser(), args.GetKey()) {
		if p, err := App.Players.AddNewPlayer(mailbox.Uid); err == nil {
			pl := p.(*BasePlayer)
			log.LogMessage("add player:", mailbox)
			pl.Account = args.GetUser()
			pl.State = STATE_LOGGED
			pl.UpdateHash()
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
