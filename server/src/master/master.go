package master

import (
	"fmt"
	"net"
	"server/libs/log"
	"server/share"
	"server/util"
	"time"
)

type App struct {
	Apps     map[string]string `json:apps`
	MustApps []string          `json:mustapps`
}

var (
	context *Master
)

type Master struct {
	Agent       bool
	AgentId     string
	Host        string
	Port        int
	LocalIP     string
	OuterIP     string
	AppDef      App
	ConsolePort int
	Template    string
	AppArgs     map[string][]byte
	tcpListener net.Listener
	waitGroup   *util.WaitGroupWrapper
	app         map[int32]*app
	quit        chan int
	agent       *Agent
	agentlist   *AgentList
	WaitAgents  int
	waitfor     bool
	NoBalance   bool
}

func (m *Master) Start() {
	context = m

	if !m.Agent {
		log.TraceInfo("master", "start")
		tcpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", m.Host, m.Port))
		if err != nil {
			log.LogError(err)
			log.LogFatalf(err)
		}
		m.tcpListener = tcpListener

		tcpserver := &tcp_server{}
		m.waitGroup.Wrap(func() { util.TCPServer(m.tcpListener, tcpserver) })
	}

	if !m.Agent && m.ConsolePort != 0 {
		StartConsoleServer(m)
		log.LogMessage("console start at:", m.ConsolePort)
	}

	if m.Agent {
		log.TraceInfo("master agent", "start")
		m.agent = &Agent{}
		m.agent.Connect(m.Host, m.Port, true, m.ConnectToMaster)
	} else {
		m.agentlist = NewAgentList()
		m.waitfor = true
		if m.WaitAgents == 0 {
			m.waitfor = false
			StartApp(m)
		}
	}
}

func (m *Master) AgentsDown() {
	if m.waitfor && m.agentlist.Count() >= m.WaitAgents {
		m.waitfor = false
		StartAppBlance(m)
	}
}

func (m *Master) ConnectToMaster() {
	m.agent.Register(m.AgentId, m.NoBalance)
	StartApp(m)
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

	if m.Agent {
		m.agent.Close()
	} else {
		m.agentlist.CloseAll()
	}

	for _, a := range m.app {
		a.Close()
	}

	if m.tcpListener != nil {
		m.tcpListener.Close()
	}

	log.LogInfo("wait all app quit")
	for len(m.app) != 0 {
		time.Sleep(time.Second)
	}
	m.waitGroup.Wait()
}

func (m *Master) CreateApp(reqid string, appid string, appuid int32, typ string, startargs string, callbackapp int32) {
	if appuid == 0 {
		appuid = GetAppUid()
	}

	if m.agentlist != nil {
		agent := m.agentlist.GetMinLoadAgent()
		if agent != nil {
			if agent.load < Load {
				//远程创建
				err := agent.CreateApp(reqid, appid, appuid, typ, startargs, callbackapp)
				if err != nil && callbackapp != 0 {
					data, err := share.CreateAppBakMsg(reqid, appuid, err.Error())
					if err != nil {
						log.LogFatalf(err)
					}

					m.SendToApp(callbackapp, data)
				}
				return
			}
		}
	}

	//本地创建
	err := CreateApp(appid, appuid, typ, startargs)
	res := "ok"
	if err != nil {
		res = err.Error()
	}

	if callbackapp != 0 {
		data, err := share.CreateAppBakMsg(reqid, appuid, res)
		if err != nil {
			log.LogFatalf(err)
		}

		m.SendToApp(callbackapp, data)
	}
}

func (m *Master) SendToApp(app int32, data []byte) error {
	if app, exist := m.app[app]; exist {
		return app.Send(data)
	}

	return fmt.Errorf("app not found")
}

func NewMaster() *Master {
	m := &Master{}
	m.AppArgs = make(map[string][]byte)
	m.waitGroup = &util.WaitGroupWrapper{}
	m.app = make(map[int32]*app)
	return m
}
