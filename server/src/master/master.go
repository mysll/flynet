package master

import (
	"fmt"
	"libs/log"
	"net"
	"util"
)

type App struct {
	Apps     map[string]string `json:apps`
	MustApps []string          `json:mustapps`
}

var (
	Context context
)

type Master struct {
	Host        string
	Port        int
	AppDef      App
	ConsolePort int
	Template    string
	AppArgs     map[string][]byte
	tcpListener net.Listener
	waitGroup   *util.WaitGroupWrapper
	app         map[string]*app
	quit        chan int
}

func (m *Master) Start() {
	log.TraceInfo("master", "start")
	Context = context{m}
	tcpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", m.Host, m.Port))
	if err != nil {
		log.LogError(err)
		log.LogFatalf(err)
	}
	m.tcpListener = tcpListener
	context := &context{m}

	tcpserver := &tcp_server{context}
	m.waitGroup.Wrap(func() { util.TCPServer(m.tcpListener, tcpserver) })

	StartApp(m)
	if m.ConsolePort != 0 {
		StartConsoleServer(m)
		log.LogMessage("console start at:", m.ConsolePort)
	}
}

func (m *Master) Wait(ch chan int) {
	m.quit = ch
	<-m.quit
}

func (m *Master) Stop() {
	m.quit <- 1
}

func (m *Master) Exit() {
	log.TraceInfo("master", "stop")

	for _, a := range m.app {
		a.Close()
	}

	if m.tcpListener != nil {
		m.tcpListener.Close()
	}

	m.waitGroup.Wait()
}

func NewMaster() *Master {
	m := &Master{}
	m.AppArgs = make(map[string][]byte)
	m.waitGroup = &util.WaitGroupWrapper{}
	m.app = make(map[string]*app)
	return m
}
