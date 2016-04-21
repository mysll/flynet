package server

import (
	"errors"
	"fmt"
	"libs/log"
	"libs/rpc"
	"pb/s2c"
	"sync"

	"github.com/golang/protobuf/proto"
)

var (
	applock        sync.RWMutex
	RemoteApps     = make(map[int32]*RemoteApp)
	RemoteAppName  = make(map[string]int32)
	ErrNotFoundApp = errors.New("app not found")
	ErrUnreachable = errors.New("unreachable router")
)

//往DB写日志，当前app必须是database,log_name为写日志的对象名称，可以是一个玩家的名字，可以是其它的信息。
//log_source记录的大类别, log_type记录的小类别,log_content, log_comment则为记录的具体内容和注释
func Log(log_name string, log_source, log_type int32, log_content, log_comment string) error {
	db := GetAppByType("database")
	if db == nil {
		return ErrAppNotFound
	}

	return db.Log(log_name, log_source, log_type, log_content, log_comment)
}

//发送一个错误提示
func Error(src *rpc.Mailbox, dest *rpc.Mailbox, method string, errno int32) error {
	errmsg := &s2c.Error{}
	errmsg.ErrorNo = proto.Int32(errno)
	return MailTo(src, dest, method, errmsg)
}

func GetLocalApp() *RemoteApp {
	return &RemoteApp{Id: core.AppId, Name: core.Name}
}

//rpc调用，src==>dest, src 为空，则自动填充为当前app
func MailTo(src *rpc.Mailbox, dest *rpc.Mailbox, method string, args ...interface{}) error {
	applock.RLock()
	defer applock.RUnlock()
	log.LogMessage("mailto:", *dest, "/", method)
	var app *RemoteApp
	if dest.App == core.AppId {
		app = &RemoteApp{Id: core.AppId, Name: core.Name}
	} else {
		var exist bool
		if app, exist = RemoteApps[dest.App]; !exist {
			return ErrNotFoundApp
		}
	}

	if dest.Flag == 0 {
		return app.Call(src, method, args...)
	} else if dest.Flag == 1 {
		return app.ClientCall(src, dest.Id, method, args[0])
	} else {
		return ErrUnreachable
	}

	return nil
}

func MailToAndCallback(src *rpc.Mailbox, dest *rpc.Mailbox, method string, cb rpc.ReplyCB, args ...interface{}) error {
	applock.RLock()
	defer applock.RUnlock()
	log.LogMessage("mailto:", *dest, "/", method)
	var app *RemoteApp
	if dest.App == core.AppId {
		app = &RemoteApp{Name: core.Name}
	} else {
		var exist bool
		if app, exist = RemoteApps[dest.App]; !exist {
			return ErrNotFoundApp
		}
	}

	if dest.Flag == 0 {
		return app.CallBack(src, method, cb, args...)
	} else if dest.Flag == 1 {
		return fmt.Errorf("client not support callback")
	} else {
		return ErrUnreachable
	}

	return nil
}

//获取远程进程的个数
func GetAppCount() int {
	applock.RLock()
	defer applock.RUnlock()
	return len(RemoteApps)
}

//通过name获取远程进程
func GetAppByName(name string) *RemoteApp {
	applock.RLock()
	defer applock.RUnlock()
	if appid, exist := RemoteAppName[name]; exist {
		return RemoteApps[appid]
	}
	return nil
}

func GetAppById(id int32) *RemoteApp {
	applock.RLock()
	defer applock.RUnlock()
	if app, exist := RemoteApps[id]; exist {
		return app
	}
	return nil
}

//通过类型获取远程进程
func GetAppByType(typ string) *RemoteApp {
	applock.RLock()
	defer applock.RUnlock()
	for _, v := range RemoteApps {
		if v.Type == typ {
			return v
		}
	}
	return nil
}

//获取某个类型的所有远程进程的appid
func GetAppIdsByType(typ string) []string {
	applock.RLock()
	defer applock.RUnlock()
	ret := make([]string, 0, 10)
	for _, v := range RemoteApps {
		if v.Type == typ {
			ret = append(ret, v.Name)
		}
	}
	return ret
}

//增加一个远程进程
func AddApp(typ string, id int32, name string, host string, port int, clienthost string, clientport int, ready bool) {
	applock.Lock()
	defer applock.Unlock()
	if core.Name == name {
		return
	}
	if appid, ok := RemoteAppName[name]; ok {
		app := RemoteApps[appid]
		if app.Id == id &&
			app.Host == host &&
			app.Port == port &&
			app.ClientHost == clienthost &&
			app.ClientPort == clientport { //已经存在
			return
		}

		app.Close()
		delete(RemoteAppName, name)
		delete(RemoteApps, appid)
	}

	RemoteApps[id] = &RemoteApp{Id: id, Type: typ, Name: name, Host: host, Port: port, ClientHost: clienthost, ClientPort: clientport, Ready: ready}
	RemoteAppName[name] = id
	log.LogInfo(core.Name, "> add server:", *RemoteApps[id])
	if ready {
		RemoteApps[id].SetReady(ready)
		core.Eventer.DispatchEvent("ready", name)
	}
}

//移除一个远程进程
func RemoveAppById(id int32) {
	applock.Lock()
	defer applock.Unlock()
	if core.AppId == id {
		return
	}
	if app, ok := RemoteApps[id]; ok {
		app.Close()
		delete(RemoteApps, id)
		log.LogInfo(core.Name, "> remove server:", app.Name)
		return
	}

	log.LogError(core.Name, "> remove server failed, ", id)
}

func RemoveAppByName(name string) {
	applock.Lock()
	defer applock.Unlock()
	if core.Name == name {
		return
	}
	if appid, ex := RemoteAppName[name]; ex {
		app := RemoteApps[appid]
		app.Close()
		delete(RemoteApps, appid)
		delete(RemoteAppName, name)
		log.LogInfo(core.Name, "> remove server:", name)
		return
	}

	log.LogError(core.Name, "> remove server failed, ", name)
}
