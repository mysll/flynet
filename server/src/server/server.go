package server

import (
	"data/helper"
	"fmt"
	"libs/common/event"
	"libs/log"
	"libs/rpc"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
	"util"

	"github.com/bitly/go-simplejson"
)

var (
	context = &Context{}
)

const (
	SENDBUFLEN = 16 * 1024
	RPCBUFFER  = 1024
)

type Server struct {
	*Timer
	StartArgs       *simplejson.Json
	Type            string
	Host            string
	Port            int
	ClientHost      string
	ClientPort      int
	AppId           int32
	Id              string
	Fronted         bool
	Debug           bool
	PProfPort       int
	ObjectFactory   *Factory
	MustAppReady    bool
	MailBox         rpc.Mailbox
	Emitter         *event.EventList
	Eventer         *EventListener
	WaitGroup       *util.WaitGroupWrapper
	IsReady         bool
	Kernel          *Kernel
	Time            Time
	AssetPath       string
	Closing         bool
	channel         map[string]*Channel
	noder           *peer
	clientListener  net.Listener
	exitChannel     chan struct{}
	shutdown        chan struct{}
	clientTcpServer *ClientHandler //客户端
	rpcListener     net.Listener   //rpc监听
	rpcServer       *rpc.Server
	rpcCh           chan *rpc.RpcCall
	apper           Apper
	quit            bool
	clientList      *ClientList
	sceneBeat       *SceneBeat
	startTime       time.Time
	s2chelper       *S2CHelper
	c2shelper       *C2SHelper
}

type StartStoper interface {
	Start(args *simplejson.Json) bool
	Stop()
}

func (svr *Server) GetTickCount() int32 {
	return int32(time.Now().Sub(svr.startTime) % math.MaxInt32)
}

func (svr *Server) Start(master string, localip string, outerip string, typ string, args *simplejson.Json) bool {
	svr.startTime = time.Now()
	defer func() {
		if e := recover(); e != nil {
			log.LogFatalf(e)
		}
	}()

	svr.Type = typ
	svr.StartArgs = args
	if id, ok := args.CheckGet("id"); ok {
		v, err := id.String()
		if err != nil {
			panic(err)
		}
		svr.Id = v
	} else {
		panic("app id not defined")
	}
	svr.MailBox = rpc.Mailbox{Address: svr.Id, Type: ""}

	//now := time.Now()
	//log.WriteToFile(fmt.Sprintf("log/%s_%d_%d_%d_%d_%d_%d.log", svr.Id, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()))

	log.LogMessage("id:", svr.AppId)

	svr.Host = localip
	if svr.Host == "" {
		if host, ok := args.CheckGet("host"); ok {
			v, err := host.String()
			if err != nil {
				panic(err)
			}
			svr.Host = v
		} else {
			panic(svr.Id + " host not defined")
		}
	}

	if port, ok := args.CheckGet("port"); ok {
		v, err := port.Int()
		if err != nil {
			panic(err)
		}
		svr.Port = v
	} else {
		panic(svr.Id + "port not defined")
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
		log.TraceInfo(svr.Id, "init link")
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", svr.ClientHost, svr.ClientPort))
		if err != nil {
			panic(err)
		}

		svr.clientListener = listener
		if svr.ClientPort == 0 {
			svr.ClientPort = listener.Addr().(*net.TCPAddr).Port
		}

		tcpserver := &ClientHandler{}
		svr.clientTcpServer = tcpserver
		svr.WaitGroup.Wrap(func() { util.TCPServer(svr.clientListener, tcpserver) })
		log.TraceInfo(svr.Id, "start link complete")
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
	svr.Kernel.setSerial(serial)
	log.LogMessage("start serial:", fmt.Sprintf("%X", serial))
	if !svr.apper.OnPrepare() {

		return false
	}

	if svr.AssetPath != "" {
		helper.LoadAllConfig(svr.AssetPath)
	}

	//内部rpc注册
	svr.rpcServer = createRpc()
	svr.rpcCh = make(chan *rpc.RpcCall, RPCBUFFER)
	svr.WaitGroup.Wrap(func() { rpc.CreateService(svr.rpcServer, svr.rpcListener, svr.rpcCh) })

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
	log.TraceInfo(svr.Id, "is shutdown")
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
	log.TraceInfo(svr.Id, " stopped")
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

func NewServer(app Apper, id int32) *Server {
	s := &Server{}
	context.Server = s
	s.AppId = id
	s.Timer = NewTimer()
	s.WaitGroup = &util.WaitGroupWrapper{}
	s.exitChannel = make(chan struct{})
	s.shutdown = make(chan struct{})
	s.Eventer = NewEvent()
	s.clientList = NewClientList()
	s.apper = app
	s.Emitter = event.NewEventList()
	s.ObjectFactory = NewFactory()
	s.Kernel = NewKernel()
	s.channel = make(map[string]*Channel, 32)
	s.sceneBeat = NewSceneBeat()

	s.s2chelper = NewS2CHelper()
	s.c2shelper = &C2SHelper{}

	RegisterRemote("S2CHelper", s.s2chelper)
	RegisterHandler("C2SHelper", s.c2shelper)
	return s
}
