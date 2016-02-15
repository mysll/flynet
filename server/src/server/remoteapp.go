package server

import (
	"errors"
	"fmt"
	"libs/log"
	"libs/rpc"
	"net"
	"pb/s2c"
	"share"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

var (
	applock        sync.RWMutex
	RemoteApps     = make(map[string]*RemoteApp)
	ErrNotFoundApp = errors.New("app not found")
	ErrUnreachable = errors.New("unreachable router")
)

//远程进程
type RemoteApp struct {
	sync.Mutex
	Type       string
	AppId      string
	Host       string
	Port       int
	ClientHost string
	ClientPort int
	Ready      bool
	Conn       net.Conn
	RpcClient  *rpc.Client
}

//进程就绪
func (app *RemoteApp) SetReady(ready bool) {
	app.Ready = ready
	log.LogInfo(app.AppId, " is ready")
	context.Server.Emitter.Push(NEWAPPREADY, map[string]interface{}{"id": app.AppId}, true)
}

//向客户端发起一个远程调用
func (app *RemoteApp) ClientCall(src *rpc.Mailbox, session int64, method string, args interface{}) error {

	var err error
	var pb share.S2CMsg
	pb.Sender = app.AppId
	pb.To = session
	pb.Method = method
	r := &s2c.Rpc{}
	r.Sender = proto.String(app.AppId)
	r.Servicemethod = proto.String(method)
	if val, ok := args.(proto.Message); ok {
		if r.Data, err = proto.Marshal(val); err != nil {
			return err
		}
	} else {
		return errors.New("args must be proto.Message")
	}

	if pb.Data, err = proto.Marshal(r); err != nil {
		return err
	}

	if src == nil {
		src = &context.Server.MailBox
	}

	app.Lock()
	defer app.Unlock()

	if app.AppId == context.Server.Id {
		return context.Server.s2chelper.Call(*src, pb)
	}

	if app.Conn == nil {
		if app.Conn == nil {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", app.Host, app.Port), time.Second)
			if err != nil {
				return err
			}
			app.Conn = conn
			app.RpcClient = rpc.NewClient(conn)
			log.LogMessage("rpc connected:", app.AppId, ",", fmt.Sprintf("%s:%d", app.Host, app.Port))
		}
	}

	err = app.RpcClient.Call("S2SS2CHelper.Call", *src, pb)
	if err == rpc.ErrShutdown {
		app.Close()
	}

	return err
}

//向客户端广播调用
func (app *RemoteApp) ClientBroadcast(src *rpc.Mailbox, session []int64, method string, args interface{}) error {

	var err error
	var pb share.S2CBrocast
	pb.Sender = app.AppId
	pb.To = session
	pb.Method = method
	r := &s2c.Rpc{}
	r.Sender = proto.String(app.AppId)
	r.Servicemethod = proto.String(method)
	if val, ok := args.(proto.Message); ok {
		if r.Data, err = proto.Marshal(val); err != nil {
			return err
		}
	} else {
		return errors.New("args must be proto.Message")
	}

	if pb.Data, err = proto.Marshal(r); err != nil {
		return err
	}

	if src == nil {
		src = &context.Server.MailBox
	}
	app.Lock()
	defer app.Unlock()

	if app.AppId == context.Server.Id {
		return context.Server.s2chelper.Broadcast(*src, pb)
	}

	if app.Conn == nil {

		if app.Conn == nil {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", app.Host, app.Port), time.Second)
			if err != nil {
				return err
			}
			app.Conn = conn
			app.RpcClient = rpc.NewClient(conn)
			log.LogMessage("rpc connected:", app.AppId, ",", fmt.Sprintf("%s:%d", app.Host, app.Port))
		}
	}

	err = app.RpcClient.Call("S2SS2CHelper.Broadcast", *src, pb)
	if err == rpc.ErrShutdown {
		app.Close()
	}

	return err
}

//往DB写日志，当前app必须是database,log_name为写日志的对象名称，可以是一个玩家的名字，可以是其它的信息。
//log_source记录的大类别, log_type记录的小类别,log_content, log_comment则为记录的具体内容和注释
func (app *RemoteApp) Log(log_name string, log_source, log_type int32, log_content, log_comment string) error {
	if app.Type == "database" {
		return app.Call(nil, "Database.Log", log_name, log_source, log_type, log_content, log_comment)
	}

	return Log(log_name, log_source, log_type, log_content, log_comment)
}

//发起一次远程调用。
//src为调用方,为nil则会自动填充为当前app的mailbox,
//method远程方法为，一般格式为:模块.方法
//args则为调用参数
func (app *RemoteApp) Call(src *rpc.Mailbox, method string, args ...interface{}) error {

	if src == nil {
		src = &context.Server.MailBox
	}

	app.Lock()
	defer app.Unlock()

	if app.AppId == context.Server.Id {
		log.LogMessage("rpc inner call:", method)
		return context.Server.rpcServer.DirectCall("S2S"+method, *src, args...)
	}

	if app.Conn == nil {

		if app.Conn == nil {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", app.Host, app.Port), time.Second)
			if err != nil {
				log.LogError(err)
				return err
			}
			app.Conn = conn
			app.RpcClient = rpc.NewClient(conn)
			log.LogMessage("rpc connected:", app.AppId, ",", fmt.Sprintf("%s:%d", app.Host, app.Port))
		}
	}
	log.LogMessage("remote call:", app.AppId, "/", method)
	err := app.RpcClient.Call("S2S"+method, *src, args...)

	if err != nil {
		app.Close()
		log.LogError(err)
	}

	return err
}

//处理由客户端的发起远程调用
func (app *RemoteApp) Handle(src rpc.Mailbox, method string, args interface{}) error {
	app.Lock()
	defer app.Unlock()

	if app.AppId == context.Server.Id {
		log.LogMessage("rpc inner handle:", method)
		return context.Server.rpcServer.DirectCall("C2S"+method, src, args)
	}

	if app.Conn == nil {

		if app.Conn == nil {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", app.Host, app.Port), time.Second)
			if err != nil {
				return err
			}
			app.Conn = conn
			app.RpcClient = rpc.NewClient(conn)
			log.LogMessage("rpc connected:", app.AppId, ",", fmt.Sprintf("%s:%d", app.Host, app.Port))
		}
	}
	log.LogMessage("remote handle:", app.AppId, "/", method)
	err := app.RpcClient.Call("C2S"+method, src, args)
	if err == rpc.ErrShutdown {
		log.LogError(err)
		app.Close()
	}
	return err
}

//关闭
func (app *RemoteApp) Close() {
	if app.Conn != nil {
		app.RpcClient.Close()
		app.RpcClient = nil
		app.Conn = nil
	}

	context.Server.Emitter.Push(APPLOST, map[string]interface{}{"id": app.AppId}, true)
}

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

//rpc调用，src==>dest, src 为空，则自动填充为当前app
func MailTo(src *rpc.Mailbox, dest *rpc.Mailbox, method string, args ...interface{}) error {
	applock.RLock()
	defer applock.RUnlock()
	log.LogMessage("mailto:", *dest, "/", method)
	var app *RemoteApp
	if dest.Address == context.Server.Id {
		app = &RemoteApp{AppId: context.Server.Id}
	} else {
		var exist bool
		if app, exist = RemoteApps[dest.Address]; !exist {
			return ErrNotFoundApp
		}
	}

	if dest.Type == "" {
		return app.Call(src, method, args...)
	} else if dest.Type == "session" {
		return app.ClientCall(src, dest.Id, method, args[0])
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

//通过appid获取远程进程
func GetApp(id string) *RemoteApp {
	applock.RLock()
	defer applock.RUnlock()
	if k, exist := RemoteApps[id]; exist {
		return k
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
			ret = append(ret, v.AppId)
		}
	}
	return ret
}

//增加一个远程进程
func AddApp(typ string, id string, host string, port int, clienthost string, clientport int, ready bool) {
	applock.Lock()
	defer applock.Unlock()
	if context.Server.Id == id {
		return
	}
	if app, ok := RemoteApps[id]; ok {
		if app.Host == host ||
			app.Port == port ||
			app.ClientHost == clienthost ||
			app.ClientPort == clientport { //可能是master重启了
			return
		}
	}

	RemoteApps[id] = &RemoteApp{Type: typ, AppId: id, Host: host, Port: port, ClientHost: clienthost, ClientPort: clientport, Ready: ready}
	log.LogInfo(context.Server.Id, "> add server:", *RemoteApps[id])
	if ready {
		RemoteApps[id].SetReady(ready)
		context.Server.Eventer.DispatchEvent("ready", id)
	}
}

//移除一个远程进程
func RemoveApp(id string) {
	applock.Lock()
	defer applock.Unlock()
	if context.Server.Id == id {
		return
	}
	if _, ok := RemoteApps[id]; ok {
		RemoteApps[id].Close()
		delete(RemoteApps, id)
		log.LogInfo(context.Server.Id, "> remove server:", id)
		return
	}

	log.LogError(context.Server.Id, "> remove server failed, ", id)

}
