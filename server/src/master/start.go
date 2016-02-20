package master

import (
	"errors"
	"libs/log"
	"sync/atomic"

	"github.com/bitly/go-simplejson"
)

var (
	AppUid int32
	Load   int32
)

func GetAppUid() int32 {
	atomic.AddInt32(&AppUid, 1)
	return AppUid
}

func StartApp(m *Master) {

	for k, v := range m.AppArgs {
		json, err := simplejson.NewJson(v)
		if err != nil {
			log.LogFatalf(err)
		}
		idx := 0

		startapp, ok := m.AppDef.Apps[k]
		if !ok {
			log.LogFatalf(errors.New("app not found"))
		}
		for {
			app := json.GetIndex(idx)
			if app.Interface() == nil {
				break
			}

			appargs, _ := app.MarshalJSON()
			id, err := app.Get("id").String()
			if err != nil {
				log.LogFatalf("app id not set")
				return
			}

			Start(startapp, id, GetAppUid(), k, string(appargs))
			idx++
		}
	}
}

func StartAppBlance(m *Master) {
	for k, v := range m.AppArgs {
		json, err := simplejson.NewJson(v)
		if err != nil {
			log.LogFatalf(err)
		}
		idx := 0

		for {
			app := json.GetIndex(idx)
			if app.Interface() == nil {
				break
			}

			appargs, _ := app.MarshalJSON()
			appid, err := app.Get("id").String()
			if err != nil {
				log.LogFatalf("app id not set")
				return
			}

			context.CreateApp("", appid, GetAppUid(), k, string(appargs), "")

			idx++
		}
	}
}

func CreateApp(id string, uid int32, typ string, startargs string) error {
	startapp := GetAppName(typ)
	err := Start(startapp, id, uid, typ, startargs)
	return err
}

func GetAppName(typ string) string {
	if v, exist := context.AppDef.Apps[typ]; exist {
		return v
	}

	return ""
}