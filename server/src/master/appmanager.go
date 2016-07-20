package master

import (
	"server/libs/log"
	"server/share"
	"sync"
)

var (
	applock  sync.RWMutex
	mustapps = map[string]string{}
)

func Ready(app *app) {
	applock.Lock()
	defer applock.Unlock()

	out, err := share.CreateReadyMsg(app.id)
	if err != nil {
		log.LogFatalf(err)
	}

	ismustapp := false
	for _, v := range context.AppDef.MustApps {
		if app.typ == v {
			mustapps[app.typ] = app.name
			ismustapp = true

			if len(mustapps) == len(context.AppDef.MustApps) {
				log.LogMessage("must app ready")
			}
			break
		}
	}

	var out1 []byte
	if len(mustapps) == len(context.AppDef.MustApps) {
		out1, err = share.CreateMustAppReadyMsg()
		if err != nil {
			log.LogFatalf(err)
		}
	}

	for _, v := range context.app {
		if v.id != app.id {
			v.conn.Write(out)
		}
		if len(out1) > 0 { //mustapp 已经都启动了
			if ismustapp { //当前是mustapp的ready，则给所有的app发送mustappready消息。
				v.conn.Write(out1)
			} else if v.id == app.id { //当前是其它应用的ready,而且mustapp都启动了，则只需要给当前的app发送mustappready就可以了
				v.conn.Write(out1)
			}

		}
	}

}

func AddApp(app *app) {
	if _, ok := context.app[app.id]; ok {
		RemoveApp(app.id)
	}

	applock.Lock()
	defer applock.Unlock()

	out, err := share.CreateAddServerMsg(app.typ, app.id, app.name, app.host, app.port, app.clienthost, app.clientport, app.ready)
	if err != nil {
		log.LogFatalf(err)
	}
	for _, v := range context.app {
		v.conn.Write(out)
	}

	context.app[app.id] = app
}

func RemoveApp(id int32) {
	applock.Lock()
	defer applock.Unlock()
	if _, ok := context.app[id]; !ok {
		return
	}

	app := context.app[id]
	app.Close()
	delete(context.app, id)

	out, err := share.CreateRemoveServerMsg(id)
	if err != nil {
		log.LogFatalf(err)
	}
	for _, v := range context.app {
		v.conn.Write(out)
	}

}
