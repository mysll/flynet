package areamgr

import (
	_ "pb"
	"server"
	"server/libs/log"
	"server/share"
)

var (
	App *AreaMgr
)

type AreaMgr struct {
	*server.Server
	quit chan int
	Area *Areas
}

func (app *AreaMgr) OnPrepare() bool {
	log.LogMessage(app.AppId, " prepared")
	return true
}

func (app *AreaMgr) OnEvent(e string, args map[string]interface{}) {
	switch e {
	case server.MASERTINMSG:
		msg := args["msg"].(server.MasterMsg)
		if msg.Id == share.M_CREATEAPP_BAK {
			var cab share.CreateAppBak
			if err := share.DecodeMsg(msg.Body, &cab); err != nil {
				log.LogError(err)
				return
			}
			App.Area.createAppBak(cab)
		}
	}
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {
	App = &AreaMgr{
		Area: NewAreas(),
	}
	server.RegisterRemote("AreaMgr", App.Area)
}
