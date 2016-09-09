package server

import (
	"errors"
	"fmt"
	"server/libs/log"
	"server/libs/rpc"
	"sync"
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
	errmsg := core.rpcProto.ErrorMsg(errno)
	return MailTo(src, dest, method, errmsg)
}

func GetLocalApp() *RemoteApp {
	return &RemoteApp{Id: core.AppId, Name: core.Name}
}

//rpc调用，src==>dest, src 为空，则自动填充为当前app
func MailTo(src *rpc.Mailbox, dest *rpc.Mailbox, method string, args ...interface{}) error {
	applock.RLock()
	log.LogMessage("mailto:", *dest, "/", method)
	var app *RemoteApp
	if dest.App == core.AppId {
		app = &RemoteApp{Id: core.AppId, Name: core.Name}
	} else {
		var exist bool
		if app, exist = RemoteApps[dest.App]; !exist {
			applock.RUnlock()
			return ErrNotFoundApp
		}
	}
	applock.RUnlock()

	if dest.Flag == 0 {
		return app.Call(src, method, args...)
	} else if dest.Flag == 1 {
		return app.ClientCall(src, dest.Id, method, args[0])
	} else {
		return ErrUnreachable
	}
}

func MailToAndCallback(src *rpc.Mailbox, dest *rpc.Mailbox, method string, cb rpc.ReplyCB, args ...interface{}) error {
	applock.RLock()
	log.LogMessage("mailto:", *dest, "/", method)
	var app *RemoteApp
	if dest.App == core.AppId {
		app = &RemoteApp{Name: core.Name}
	} else {
		var exist bool
		if app, exist = RemoteApps[dest.App]; !exist {
			applock.RUnlock()
			return ErrNotFoundApp
		}
	}
	applock.RUnlock()

	if dest.Flag == 0 {
		return app.CallBack(src, method, cb, args...)
	} else if dest.Flag == 1 {
		return fmt.Errorf("client not support callback")
	} else {
		return ErrUnreachable
	}
}

//获取远程进程的个数
func GetAppCount() int {
	applock.RLock()
	l := len(RemoteApps)
	applock.RUnlock()
	return l
}

//获取appid
func GetAppIdByName(name string) int32 {
	applock.RLock()
	if appid, exist := RemoteAppName[name]; exist {
		applock.RUnlock()
		return appid
	}
	applock.RUnlock()
	return -1
}

//通过name获取远程进程
func GetAppByName(name string) *RemoteApp {
	applock.RLock()
	if appid, exist := RemoteAppName[name]; exist {
		applock.RUnlock()
		return RemoteApps[appid]
	}
	applock.RUnlock()
	return nil
}

func GetAppById(id int32) *RemoteApp {
	applock.RLock()
	if app, exist := RemoteApps[id]; exist {
		applock.RUnlock()
		return app
	}
	applock.RUnlock()
	return nil
}

//通过类型获取远程进程
func GetAppByType(typ string) *RemoteApp {
	applock.RLock()
	for _, v := range RemoteApps {
		if v.Type == typ {
			applock.RUnlock()
			return v
		}
	}
	applock.RUnlock()
	return nil
}

//获取某个类型的所有远程进程的appid
func GetAppIdsByType(typ string) []string {
	applock.RLock()
	ret := make([]string, 0, 10)
	for _, v := range RemoteApps {
		if v.Type == typ {
			ret = append(ret, v.Name)
		}
	}
	applock.RUnlock()
	return ret
}

//增加一个远程进程
func AddApp(typ string, id int32, name string, host string, port int, clienthost string, clientport int, ready bool, enableglobaldata bool) {
	if core.Name == name {
		return
	}
	applock.Lock()
	if appid, ok := RemoteAppName[name]; ok {
		app := RemoteApps[appid]
		if app.Id == id &&
			app.Host == host &&
			app.Port == port &&
			app.ClientHost == clienthost &&
			app.ClientPort == clientport { //已经存在
			applock.Unlock()
			return
		}

		app.Close()
		delete(RemoteAppName, name)
		delete(RemoteApps, appid)
	}

	RemoteApps[id] = &RemoteApp{Id: id, Type: typ, Name: name, Host: host, Port: port, ClientHost: clienthost, ClientPort: clientport, Ready: ready, EnableGlobalData: enableglobaldata}
	RemoteAppName[name] = id
	log.LogInfo(core.Name, "> add server:", *RemoteApps[id])
	if ready {
		RemoteApps[id].SetReady(ready)
		core.Eventer.DispatchEvent("ready", name)
	}
	applock.Unlock()
}

//移除一个远程进程
func RemoveAppById(id int32) {
	if core.AppId == id {
		return
	}

	applock.Lock()
	if app, ok := RemoteApps[id]; ok {
		app.Close()
		delete(RemoteApps, id)
		log.LogInfo(core.Name, "> remove server:", app.Name)
		applock.Unlock()
		return
	}

	log.LogError(core.Name, "> remove server failed, ", id)
	applock.Unlock()
}

func RemoveAppByName(name string) {
	if core.Name == name {
		return
	}

	applock.Lock()
	if appid, ex := RemoteAppName[name]; ex {
		app := RemoteApps[appid]
		app.Close()
		delete(RemoteApps, appid)
		delete(RemoteAppName, name)
		log.LogInfo(core.Name, "> remove server:", name)
		applock.Unlock()
		return
	}

	log.LogError(core.Name, "> remove server failed, ", name)
	applock.Unlock()
}
