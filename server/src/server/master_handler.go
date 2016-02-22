package server

import (
	"libs/log"
	"share"
	"util"
)

type master_peer struct {
}

func (mp *master_peer) Handle(id uint16, msgbody []byte) error {
	switch id {
	case share.M_ADD_SERVER:
		var as share.AddApp
		if err := share.DecodeMsg(msgbody, &as); err != nil {
			return err
		}
		AddApp(as.Type, as.AppId, as.Host, as.Port, as.ClientHost, as.ClientPort, as.Ready)
	case share.M_REMOVE_SERVER:
		var rs share.RemoveApp
		if err := share.DecodeMsg(msgbody, &rs); err != nil {
			return err
		}
		RemoveApp(rs.AppId)
	case share.M_SERVER_LIST:
		var sl share.AppInfo
		if err := share.DecodeMsg(msgbody, &sl); err != nil {
			return err
		}
		for _, a := range sl.Apps {
			AddApp(a.Type, a.AppId, a.Host, a.Port, a.ClientHost, a.ClientPort, a.Ready)
		}
	case share.M_HEARTBEAT:
		data, err := util.CreateMsg(nil, []byte{}, share.M_HEARTBEAT)
		if err != nil {
			log.LogFatalf(err)
		}
		core.noder.Send(data)
	case share.M_READY:
		var ready share.AppReady
		if err := share.DecodeMsg(msgbody, &ready); err != nil {
			return err
		}
		app := GetApp(ready.AppId)
		if app != nil {
			app.SetReady(true)
		} else {
			log.LogFatalf("app not found")
		}
	case share.M_MUSTAPPREADY:
		{
			log.LogMessage("must app ready")
			core.MustReady()
		}
	case share.M_SHUTDOWN:
		core.Closing = true
		close(core.exitChannel)
		return nil
	}

	core.Emitter.Push(MASERTINMSG, map[string]interface{}{"msg": MasterMsg{id, msgbody}}, false)
	return nil

}
