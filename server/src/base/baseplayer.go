package base

import (
	"fmt"
	"logicdata/entity"
	"pb/s2c"
	"server"
	. "server/data/datatype"
	"server/libs/log"
	"server/share"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	STATE_UNKNOWN   = iota
	STATE_LOGGED    //已登录
	STATE_READY     //已就绪
	STATE_ENTERAREA //进入场景服务器(过程中)
	STATE_SWITCH    //切换场景
	STATE_GAMING    //进入场景服务器完成
	STATE_LEAVE     //退出游戏
	STATE_SAVING    //保存中
	STATE_DELETING  //正在删除
)

type BasePlayer struct {
	server.PlayerInfo

	ChooseRole  string
	trans       server.Transform
	RoleInfo    string
	Offline     bool
	AreaId      string
	LandTimes   int32
	propsyncer  *server.PropSync
	tablesyncer *server.TableTrans
	lastupdate  time.Time
	saveid      server.TimerID
}

func NewBasePlayer() server.PlayerHandler {
	bp := &BasePlayer{}
	return bp
}

//客户端断开连接
func (p *BasePlayer) Disconnect() {
	if p.Offline {
		return
	}

	p.Offline = true
	App.Kernel().Disconnect(p.Entity)
	p.Leave()
	log.LogInfo("player disconnect:", p.ChooseRole, " session:", p.Session)
}

func TimeToDel(intervalid server.TimerID, count int32, args interface{}) {
	App.Players.RemovePlayer(args.(uint64))
}

func (p *BasePlayer) TimeToSave(intervalid server.TimerID, count int32, args interface{}) {
	p.SaveToDb(false)
}

func (p *BasePlayer) Leave() {

	if p.State < STATE_READY { //还没有创建角色
		App.Players.RemovePlayer(p.Mailbox.Uid)
		return
	}

	if p.State == STATE_ENTERAREA {
		p.State = STATE_LEAVE
		return
	}

	if p.State == STATE_GAMING {
		p.State = STATE_LEAVE
		App.AreaBridge.areaRemovePlayer(p, share.REMOVE_OFFLINE)
		return
	}

	if p.State == STATE_SWITCH {
		p.State = STATE_LEAVE
		return
	}

	p.SaveToDb(true)
}

func (p *BasePlayer) LeaveArea() {
	if p.State == STATE_LEAVE {
		p.SaveToDb(true)
	}
}

func (p *BasePlayer) SaveToDb(offline bool) {
	typ := share.SAVETYPE_TIMER
	if offline {
		p.State = STATE_SAVING
		typ = share.SAVETYPE_OFFLINE
	}

	if typ == share.SAVETYPE_TIMER {
		if !p.Entity.NeedSave() {
			return
		}
	}

	log.LogInfo("save player,", p.ChooseRole, ", type ", typ)
	//写数据到数据库
	if err := App.DbBridge.savePlayer(p, typ); err != nil {
		log.LogError(err)
		if p.Entity != nil {
			p.SaveFailed()
		}
	}
}

func (p *BasePlayer) SaveFailed() {
	if p.State == STATE_SAVING {
		now := time.Now()
		f := fmt.Sprintf("dump/%s_%d_%d_%d_%d_%d_%d.log", p.Account, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

		log.LogError("save player failed, dump info into:", f)
		go App.Kernel().DumpInfo(*p, f)
		App.Kernel().Timeout(time.Second*5, TimeToDel, p.Mailbox.Uid)
	}
}

func (p *BasePlayer) LoadPlayer(data share.LoadUserBak) error {
	p.trans = server.Transform{data.Scene, Vector3{data.X, data.Y, data.Z}, data.Dir}
	p.RoleInfo = data.Data.RoleInfo
	p.LandTimes = data.LandTimes
	var err error
	p.Entity, err = App.Kernel().CreateFromDb(data.Data)
	if err != nil {
		log.LogError(err)
		return err
	}
	p.State = STATE_READY
	player := p.Entity.(*entity.Player)
	App.Kernel().SetRoleInfo(p.Entity, p.RoleInfo)
	App.Kernel().SetLandpos(p.Entity, p.trans)
	p.Entity.SetExtraData("account", p.Account)
	p.Entity.SetUID(p.Mailbox.Uid)
	log.LogInfo("load player succeed,", player.GetName())
	p.saveid = App.Kernel().AddTimer(time.Minute*5, -1, p.TimeToSave, nil)

	if player.GetLastUpdateTime() == 0 {
		player.SetLastUpdateTime(time.Now().Unix())
		p.lastupdate = time.Now()
	} else {
		p.lastupdate = time.Unix(player.GetLastUpdateTime(), 0)
	}

	//同步玩家
	App.Kernel().AttachPlayer(p.Entity, p.Mailbox)

	return err
}

func (p *BasePlayer) DeletePlayer() {
	if p.Entity != nil {
		App.Kernel().DetachPlayer(p.Entity)
		p.Entity.SetQuiting()
		App.Kernel().Destroy(p.Entity.ObjectId())
		log.LogInfo("player destroy:", p.ChooseRole, " session:", p.Session)
	}
	App.Kernel().CancelTimer(p.saveid)
}

func (p *BasePlayer) PlayerReady() {

	if p.LandTimes == 0 {
		App.Kernel().Command(p.Entity.ObjectId(), p.Entity.ObjectId(), share.PLAYER_FIRST_LAND, nil)
	}

	server.MailTo(nil, &p.Mailbox, "Role.Ready", &s2c.Void{})
}

func (p *BasePlayer) EnterScene() error {
	p.State = STATE_GAMING
	enter := &s2c.EnterScene{}
	enter.Name = proto.String(p.ChooseRole)
	err := server.MailTo(nil, &p.Mailbox, "Login.EnterScene", enter)
	if err != nil {
		log.LogError(err)
	}

	return err
}
