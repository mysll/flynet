//Package server Server是整个服务器的核心，所以扩展的服务都继承于Server,每个服务是一个独立的进程，这个服务可以是单独运行在一台服务器上，
// 也可以将多个服务启动在同一台电脑上，这主要取决于你的服务器配置。
// Server提供的基础功能包含：
//		对外提供连接服务(主要是面向客户端的连接服务)
//		服务之间的通讯和RPC，包含服务之间的相互发现，服务结点的自动管理
//		服务主循环处理
//		游戏性相关处理(包含对象管理，游戏回调，事件回调等，参见Kernel的接口)
// 如何开发一个新的服务：
//		type NewApp struct {
//			*Server
//		}
//		NewApp.Server = server.NewServer(NewApp, appid)
// 启动这个服务:
//		if NewApp.Start(...) {
//			NewApp.Wait()
//		}
package server

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"server/data/helper"
	"server/libs/common/event"
	"server/libs/log"
	"server/libs/rpc"
	"server/util"
	"syscall"
	"time"

	"github.com/bitly/go-simplejson"
)

const (
	SENDBUFLEN = 16 * 1024
	RPCBUFFER  = 1024
)

type Server struct {
	kernel           *Kernel
	StartArgs        *simplejson.Json
	Type             string
	Host             string
	Port             int
	ClientHost       string
	ClientPort       int
	AppId            int32
	Name             string
	Fronted          bool
	Debug            bool
	PProfPort        int
	MustAppReady     bool
	MailBox          rpc.Mailbox
	Emitter          *event.EventList
	Eventer          *EventListener
	WaitGroup        *util.WaitGroupWrapper
	IsReady          bool
	Time             TimeInfo
	AssetPath        string
	Closing          bool
	Players          *PlayerManager
	channel          map[string]*Channel
	noder            *peer
	clientListener   net.Listener
	exitChannel      chan struct{}
	shutdown         chan struct{}
	rpcListener      net.Listener //rpc监听
	rpcServer        *rpc.Server
	rpcCh            chan *rpc.RpcCall
	apper            Apper
	quit             bool
	clientList       *ClientList
	startTime        time.Time
	s2chelper        *S2CHelper
	c2shelper        *C2SHelper
	teleport         *TeleportHelper
	Sockettype       string
	rpcProto         ProtoCodec
	globalset        string
	globaldataserver bool
	enableglobaldata bool
	maxglobalentry   int
	globalHelper     *GlobalDataHelper
	modules          map[string]Moduler
}

type StartStoper interface {
	Start(args *simplejson.Json) bool
	Stop()
}

//获取kernel
func (svr *Server) Kernel() *Kernel {
	return svr.kernel
}

//获取程序启动到现在的毫秒数
func (svr *Server) GetTickCount() int32 {
	return int32(time.Now().Sub(svr.startTime) / 1000000)
}

//启动服务,master是master的监听地址,localip是服务的rpc启动地址,outerip是对外的服务地址,typ是服务类型,argstr启动参数
func (svr *Server) Start(master string, localip string, outerip string, typ string, argstr string) bool {
	svr.startTime = time.Now()
	defer func() {
		if e := recover(); e != nil {
			log.LogFatalf(e)
		}
	}()

	args, err := simplejson.NewJson([]byte(argstr))
	if err != nil {
		panic(err)
	}
	svr.Type = typ
	svr.StartArgs = args
	if name, ok := args.CheckGet("name"); ok {
		v, err := name.String()
		if err != nil {
			panic(err)
		}
		svr.Name = v
	} else {
		panic("app id not defined")
	}
	svr.MailBox = rpc.NewMailBox(0, 0, svr.AppId)

	log.LogMessage("id:", svr.AppId, ", name:", svr.Name)

	svr.Host = localip
	if svr.Host == "" {
		if host, ok := args.CheckGet("host"); ok {
			v, err := host.String()
			if err != nil {
				panic(err)
			}
			svr.Host = v
		} else {
			panic(svr.Name + " host not defined")
		}
	}

	if port, ok := args.CheckGet("port"); ok {
		v, err := port.Int()
		if err != nil {
			panic(err)
		}
		svr.Port = v
	} else {
		panic(svr.Name + "port not defined")
	}

	svr.ClientHost = outerip
	if svr.ClientHost == "" {
		if clienthost, ok := args.CheckGet("clienthost"); ok {
			v, err := clienthost.String()
			if err != nil {
				panic(err)
			}
			svr.ClientHost = v
		}
	}

	if clientport, ok := args.CheckGet("clientport"); ok {
		v, err := clientport.Int()
		if err != nil {
			panic(err)
		}
		svr.ClientPort = v
	}

	if fronted, ok := args.CheckGet("fronted"); ok {
		v, err := fronted.Bool()
		if err != nil {
			panic(err)
		}

		svr.Fronted = v
	}

	svr.Sockettype = "native"
	if Sockettype, ok := args.CheckGet("sockettype"); ok {
		v, err := Sockettype.String()
		if err != nil {
			panic(err)
		}

		svr.Sockettype = v
	}

	//rpc端口
	r, err := net.Listen("tcp", fmt.Sprintf("%s:%d", svr.Host, svr.Port))
	if err != nil {
		panic(err)
	}
	svr.rpcListener = r
	if svr.Port == 0 {
		svr.Port = r.Addr().(*net.TCPAddr).Port
	}

	if svr.Fronted && !svr.apper.RawSock() {
		log.TraceInfo(svr.Name, "init link")
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", svr.ClientHost, svr.ClientPort))
		if err != nil {
			panic(err)
		}

		svr.clientListener = listener
		if svr.ClientPort == 0 {
			svr.ClientPort = listener.Addr().(*net.TCPAddr).Port
		}

		switch svr.Sockettype {
		case "websocket":
			svr.WaitGroup.Wrap(func() { util.WSServer(svr.clientListener, &WSClientHandler{}) })
		default:
			svr.WaitGroup.Wrap(func() { util.TCPServer(svr.clientListener, &ClientHandler{}) })
		}

		log.TraceInfo(svr.Name, "start link complete")
	}

	mp := &masterPeer{}
	peer := &peer{addr: master, h: mp}
	if err := peer.Connect(); err != nil {
		panic(err)
	}

	svr.noder = peer

	signalChan := make(chan os.Signal, 1)
	go func() {
		for {
			if _, ok := <-signalChan; !ok {
				return
			}
		}

	}()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	if debug, ok := args.CheckGet("debug"); ok {
		v, err := debug.Bool()
		if err == nil {
			svr.Debug = v
		}
	}

	if pfport, ok := args.CheckGet("pprof"); ok {
		v, err := pfport.Int()
		if err == nil {
			svr.PProfPort = v
		}
	}

	if assets, ok := args.CheckGet("assets"); ok {
		v, err := assets.String()
		if err == nil {
			svr.AssetPath = v
		}
	}

	if loglevel, ok := args.CheckGet("loglevel"); ok {
		v, err := loglevel.Int()
		if err == nil {
			log.SetLogLevel("stdout", v)
		}
	}

	svr.globalset = "GlobalSet"
	if gs, ok := args.CheckGet("globalset"); ok {
		v, err := gs.String()
		if err == nil {
			svr.globalset = v
		}
	}

	svr.globaldataserver = false
	if enable, ok := args.CheckGet("globaldataserver"); ok {
		v, err := enable.Bool()
		if err == nil {
			svr.globaldataserver = v
		}
	}

	svr.enableglobaldata = false
	if enable, ok := args.CheckGet("enableglobaldata"); ok {
		v, err := enable.Bool()
		if err == nil {
			svr.enableglobaldata = v
		}
	}

	svr.maxglobalentry = 10
	if max, ok := args.CheckGet("maxglobalentry"); ok {
		v, err := max.Int()
		if err == nil {
			svr.maxglobalentry = v
		}
	}

	log.LogMessage("global data status is ", svr.enableglobaldata)

	if svr.globaldataserver {
		svr.globalHelper.SetServer()
		log.LogMessage("promote to global data server")
	}

	serial := (uint64(time.Now().Unix()%0x80000000) << 32) | (uint64(svr.AppId) << 24)
	svr.kernel.setSerial(serial)
	log.LogMessage("start serial:", fmt.Sprintf("%X", serial))
	if !svr.apper.OnPrepare() {
		return false
	}
	svr.kernel.CurrentInBase(svr.apper.IsBase())
	if svr.AssetPath != "" {
		helper.LoadAllConfig(svr.AssetPath)
	}

	initAllModules()

	//内部rpc注册
	svr.rpcCh = make(chan *rpc.RpcCall, RPCBUFFER)
	svr.rpcServer = createRpc(svr.rpcCh)
	svr.WaitGroup.Wrap(func() { rpc.CreateService(svr.rpcServer, svr.rpcListener) })

	svr.rpcProto = codec
	if svr.rpcProto == nil {
		panic("proto not set")
	}
	log.LogMessage("client proto:", svr.rpcProto.GetCodecInfo())

	loadAllModules()

	svr.Ready()

	return true
}

//获取频道
func (svr *Server) GetChannel(typ string) *Channel {
	if c, ok := svr.channel[typ]; ok {
		return c
	}
	svr.channel[typ] = NewChannel()
	return svr.channel[typ]
}

//Wait 主循环
func (svr *Server) Wait() {
	go run(svr)
	log.LogMessage("server debug:", svr.Debug)
	//启动调试
	if svr.Debug {
		go func() {
			log.LogMessage("pprof start at:", svr.PProfPort)
			if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", svr.PProfPort), nil); err != nil {
				log.LogMessage("pprof start failed:", err)
			}
		}()
	}

	<-svr.exitChannel
	log.TraceInfo(svr.Name, "is shutdown")
	if svr.globaldataserver {
		//保存global data
		log.TraceInfo(svr.Name, "save global data")
		if err := svr.kernel.SaveGlobalData(true, true); err != nil {
			log.LogError(err)
		}
		log.TraceInfo(svr.Name, "save global data, ok")
	}

	//通知app进程即将退出
	if svr.apper.OnShutdown() {
		svr.Shutdown()
	}
	<-svr.shutdown
	if svr.noder != nil {
		svr.noder.Close()
	}
	if svr.clientListener != nil {
		svr.clientListener.Close()
	}

	if svr.rpcListener != nil {
		svr.rpcListener.Close()
	}
	svr.WaitGroup.Wait()
	log.TraceInfo(svr.Name, " stopped")

	unloadAllModules()
	//等待日志写入完毕
	<-time.After(time.Second)
	log.CloseLogger()
}

//关闭服务
func (svr *Server) Shutdown() {
	svr.quit = true
	close(svr.shutdown)
}

//服务已经就绪
func (svr *Server) Ready() {
	svr.noder.Reg(svr)
	svr.noder.Ready()
	svr.IsReady = true
}

//踢人
func (svr *Server) KickUser(session int64) {
	node := svr.clientList.FindNode(session)
	if node != nil {
		node.Close()
	}
}

//延迟踢人
func (svr *Server) DelayKickUser(session int64, sec int) {
	node := svr.clientList.FindNode(session)
	if node != nil {
		node.DelCount(sec)
	}
}

//顶号处理
func (svr *Server) SwitchConn(oldsession, newsession int64) bool {
	if svr.clientList.Switch(oldsession, newsession) {
		return true
	}
	return false
}

//获取客户端列表
func (svr *Server) GetClientList() *ClientList {
	return svr.clientList
}

//发送消息到master
func (svr *Server) SendToMaster(data []byte) error {
	return svr.noder.Send(data)
}

//关键应用已经启动完成
func (svr *Server) MustReady() {
	if !svr.MustAppReady {
		svr.MustAppReady = true
		svr.apper.OnMustAppReady()
	}

	if svr.globaldataserver {
		log.LogMessage("start global data service")
		err := svr.globalHelper.LoadGlobalData()
		if err != nil {
			log.LogError(err)
		}
	}
}

//某个app就绪通知
func (svr *Server) OnAppReady(app string) {
	svr.globalHelper.OnAppReady(app)
}

//某个app关闭通知
func (svr *Server) OnAppLost(app string) {
	svr.globalHelper.OnAppLost(app)
}
