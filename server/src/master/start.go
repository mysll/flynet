package master

import (
	"errors"
	"github.com/bitly/go-simplejson"
	"libs/log"
)

var (
	AppId int32
)

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
			Start(startapp, id, k, string(appargs))
			idx++
		}
	}
}

func GetAppName(typ string) string {
	if v, exist := Context.master.AppDef.Apps[typ]; exist {
		return v
	}

	return ""
}
