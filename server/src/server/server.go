package server

import (
	"fmt"
	"math"
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
	*Kernel
	StartArgs      *simplejson.Json
	Type           string
	Host           string
	Port           int
	ClientHost     string
	ClientPort     int
	AppId          int32
	Name           string
	Fronted        bool
	Debug          bool
	PProfPort      int
	ObjectFactory  *Factory
	MustAppReady   bool
	MailBox        rpc.Mailbox
	Emitter        *event.EventList
	Eventer        *EventListener
	WaitGroup      *util.WaitGroupWrapper
	IsReady        bool
	Time           Time
	AssetPath      string
	Closing        bool
	channel        map[string]*Channel
	noder          *peer
	clientListener net.Listener
	exitChannel    chan struct{}
	shutdown       chan struct{}
	rpcListener    net.Listener //rpc监听
	rpcServer      *rpc.Server
	rpcCh          chan *rpc.RpcCall
	apper          Apper
	quit           bool
	clientList     *ClientList
	startTime      time.Time
	s2chelper      *S2CHelper
	c2shelper      *C2SHelper
	teleport       *TeleportHelper
	Sockettype     string
	rpcProto       ProtoCodec
}

type StartStoper interface {
	Start(args *simplejson.Json) bool
	Stop()
}

func (svr *Server) GetTickCount() int32 {
	return int32(time.Now().Sub(svr.startTime) % math.MaxInt32)
}

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

	//now := time.Now()
	//log.WriteToFile(fmt.Sprintf("log/%s_%d_%d_%d_%d_%d_%d.log", svr.Name, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()))

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

	master_peer := &master_peer{}
	peer := &peer{addr: master, h: master_peer}
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

	serial := (uint64(time.Now().Unix()%0x80000000) << 32) | (uint64(svr.AppId) << 24)
	svr.setSerial(serial)
	log.LogMessage("start serial:", fmt.Sprintf("%X", serial))
	if !svr.apper.OnPrepare() {
		return false
	}
	svr.CurrentInBase(svr.apper.IsBase())
	if svr.AssetPath != "" {
		helper.LoadAllConfig(svr.AssetPath)
	}

	//内部rpc注册
	svr.rpcCh = make(chan *rpc.RpcCall, RPCBUFFER)
	svr.rpcServer = createRpc(svr.rpcCh)
	svr.WaitGroup.Wrap(func() { rpc.CreateService(svr.rpcServer, svr.rpcListener) })

	svr.rpcProto = codec
	if svr.rpcProto == nil {
		panic("proto not set")
	}
	log.LogMessage("client proto:", svr.rpcProto.GetCodecInfo())
	svr.Ready()

	return true
}

func (svr *Server) GetChannel(typ string) *Channel {
	if c, ok := svr.channel[typ]; ok {
		return c
	}
	svr.channel[typ] = NewChannel()
	return svr.channel[typ]
}

func (svr *Server) Wait() {
	go Run(svr)
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
	//等待日志写入完毕
	<-time.After(time.Second)
	log.CloseLogger()
}

func (svr *Server) Shutdown() {
	svr.quit = true
	close(svr.shutdown)
}

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

func (svr *Server) DelayKickUser(session int64, sec int) {
	node := svr.clientList.FindNode(session)
	if node != nil {
		node.DelCount(sec)
	}
}

//頂號處理
func (svr *Server) SwitchConn(oldsession, newsession int64) bool {
	if svr.clientList.Switch(oldsession, newsession) {
		return true
	}
	return false
}

func (svr *Server) GetClientList() *ClientList {
	return svr.clientList
}

func (svr *Server) SendToMaster(data []byte) error {
	return svr.noder.Send(data)
}

func (svr *Server) MustReady() {
	if !svr.MustAppReady {
		svr.MustAppReady = true
		svr.apper.OnMustAppReady()
	}
}
