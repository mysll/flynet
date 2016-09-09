package server

import (
	"fmt"
	"net"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
	"sync"
	"time"
)

//远程进程
type RemoteApp struct {
	sync.Mutex
	Id               int32
	Type             string
	Name             string
	Host             string
	Port             int
	ClientHost       string
	ClientPort       int
	Ready            bool
	EnableGlobalData bool
	Conn             net.Conn
	RpcClient        *rpc.Client
}

//进程就绪
func (app *RemoteApp) SetReady(ready bool) {
	app.Ready = ready
	log.LogInfo(app.Name, " is ready")
	core.Emitter.Push(NEWAPPREADY, map[string]interface{}{"id": app.Name}, true)
}

//向客户端发起一个远程调用
func (app *RemoteApp) ClientCall(src *rpc.Mailbox, session int64, method string, args interface{}) error {

	var err error
	var pb share.S2CMsg
	pb.Sender = app.Name
	pb.To = session
	pb.Method = method

	if pb.Data, err = core.rpcProto.CreateRpcMessage(app.Name, method, args); err != nil {
		log.LogError(err)
		return err
	}

	if src == nil {
		src = &core.MailBox
	}

	app.Lock()
	defer app.Unlock()

	if app.Id == core.AppId { //进程内调用
		msg := NewMessage()
		msg.Write(pb)
		msg.Flush()
		core.s2chelper.Call(*src, msg.GetMessage())
		msg.Free()
		return nil
	}

	if app.Conn == nil {
		if app.Conn == nil {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", app.Host, app.Port), time.Second)
			if err != nil {
				return err
			}
			app.Conn = conn
			app.RpcClient = rpc.NewClient(conn)
			log.LogMessage("rpc connected:", app.Name, ",", fmt.Sprintf("%s:%d", app.Host, app.Port))
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
	pb.Sender = app.Name
	pb.To = session
	pb.Method = method
	if pb.Data, err = core.rpcProto.CreateRpcMessage(app.Name, method, args); err != nil {
		log.LogError(err)
		return err
	}

	if src == nil {
		src = &core.MailBox
	}
	app.Lock()
	defer app.Unlock()

	if app.Id == core.AppId { //进程内调用
		msg := NewMessage()
		msg.Write(pb)
		msg.Flush()
		core.s2chelper.Broadcast(*src, msg.GetMessage())
		msg.Free()
		return nil
	}

	if app.Conn == nil {

		if app.Conn == nil {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", app.Host, app.Port), time.Second)
			if err != nil {
				return err
			}
			app.Conn = conn
			app.RpcClient = rpc.NewClient(conn)
			log.LogMessage("rpc connected:", app.Name, ",", fmt.Sprintf("%s:%d", app.Host, app.Port))
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
		src = &core.MailBox
	}

	app.Lock()
	defer app.Unlock()

	if app.Id == core.AppId { //进程内调用
		log.LogMessage("rpc inner call:", method)
		return core.rpcServer.Call("S2S"+method, *src, args...)
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
			log.LogMessage("rpc connected:", app.Name, ",", fmt.Sprintf("%s:%d", app.Host, app.Port))
		}
	}
	log.LogMessage("remote call:", app.Name, "/", method)
	err := app.RpcClient.Call("S2S"+method, *src, args...)

	if err != nil {
		app.Close()
		log.LogError(err)
	}

	return err
}

func (app *RemoteApp) CallBack(src *rpc.Mailbox, method string, cb rpc.ReplyCB, args ...interface{}) error {

	if src == nil {
		src = &core.MailBox
	}

	app.Lock()
	defer app.Unlock()

	if app.Id == core.AppId { //进程内调用
		log.LogMessage("rpc inner call:", method)
		return core.rpcServer.CallBack("S2S"+method, *src, cb, args...)
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
			log.LogMessage("rpc connected:", app.Name, ",", fmt.Sprintf("%s:%d", app.Host, app.Port))
		}
	}
	log.LogMessage("remote call:", app.Name, "/", method)
	err := app.RpcClient.CallBack("S2S"+method, *src, cb, args...)

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

	if app.Id == core.AppId {
		log.LogMessage("rpc inner handle:", method)
		return core.rpcServer.Call("C2S"+method, src, args)
	}

	if app.Conn == nil {

		if app.Conn == nil {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", app.Host, app.Port), time.Second)
			if err != nil {
				return err
			}
			app.Conn = conn
			app.RpcClient = rpc.NewClient(conn)
			log.LogMessage("rpc connected:", app.Name, ",", fmt.Sprintf("%s:%d", app.Host, app.Port))
		}
	}
	log.LogMessage("remote handle:", app.Name, "/", method)
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

	core.Emitter.Push(APPLOST, map[string]interface{}{"id": app.Name}, true)
}
