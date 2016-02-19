package master

import (
	"io"
	"libs/log"
	"share"
	"strings"
	"sync"
	"util"
)

type AgentNode struct {
	id        string
	load      int32
	nobalance bool
	rwc       io.ReadWriteCloser
	quit      bool
}

func (a *AgentNode) Handle(conn io.ReadWriteCloser) {
	a.rwc = conn
	go a.readloop()
}

func (a *AgentNode) HandleMsg(id uint16, msg []byte) error {
	switch id {
	case share.M_SHUTDOWN:
		a.Close()
	case share.M_FORWARD_MSG:
		forward := &share.SendAppMsg{}
		err := share.DecodeMsg(msg, forward)
		if err != nil {
			log.LogError(err)
			return err
		}
		err = context.SendToApp(forward.AppId, forward.Data)
		if err != nil {
			log.LogError(err)
			return err
		}
	}
	return nil
}

func (a *AgentNode) Close() {
	if a.rwc != nil {
		a.rwc.Close()
	}
	a.quit = true
}

func (a *AgentNode) CreateApp(reqid string, appid string, appuid int32, typ string, args string, callapp string) error {
	data, err := share.CreateAppMsg(typ, reqid, appid, appuid, args, callapp)
	if err != nil {
		log.LogError(err)
		return err
	}

	_, err = a.rwc.Write(data)
	if err != nil {
		return err
	}

	a.load++
	return nil
}

func (a *AgentNode) readloop() {
	buffer := make([]byte, 2048)
	for !a.quit {
		id, msg, err := util.ReadPkg(a.rwc, buffer)
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
				log.LogError(err)
			}
			break
		}

		if err := a.HandleMsg(id, msg); err != nil {
			log.LogError(err)
			break
		}
	}

	log.LogInfo("agent node closed")
	a.Close()
	context.agentlist.RemoveAgent(a.id)
}

type AgentList struct {
	sync.RWMutex
	agents map[string]*AgentNode
}

func NewAgentList() *AgentList {
	al := &AgentList{}
	al.agents = make(map[string]*AgentNode)
	return al
}

func (al *AgentList) Count() int {
	return len(al.agents)
}

func (al *AgentList) AddAgent(an *AgentNode) {
	al.Lock()
	if a, exist := al.agents[an.id]; exist {
		a.Close()
		delete(al.agents, an.id)
	}
	al.agents[an.id] = an
	al.Unlock()
	context.AgentsDown()
}

func (al *AgentList) RemoveAgent(id string) {
	al.Lock()
	defer al.Unlock()
	if a, exist := al.agents[id]; exist {
		a.Close()
		delete(al.agents, id)
	}
}

func (al *AgentList) GetMinLoadAgent() *AgentNode {
	al.Lock()
	defer al.Unlock()
	var ret *AgentNode
	for _, agent := range al.agents {
		if agent.nobalance {
			continue
		}
		if ret == nil {
			ret = agent
			continue
		}
		if ret.load > agent.load {
			ret = agent
		}
	}

	return ret
}

func (al *AgentList) CloseAll() {
	al.Lock()
	defer al.Unlock()

	data, err := util.CreateMsg(nil, []byte{}, share.M_SHUTDOWN)
	if err != nil {
		log.LogFatalf(err)
	}

	for k, agent := range al.agents {
		_, err = agent.rwc.Write(data)
		if err != nil {
			log.LogInfo(k, " closed")
		} else {
			log.LogInfo(k, " send shutdown")
		}
	}
}
