package master

import (
	"net"
	"server/libs/log"
	"server/share"
	"server/util"
)

type tcp_server struct {
}

func (tcp *tcp_server) Handle(clientconn net.Conn) {

	buff := make([]byte, 2048)
	id, msgbody, err := util.ReadPkg(clientconn, buff)
	if err != nil {
		log.LogError(err)
		clientconn.Close()
		return
	}

	switch id {
	case share.M_REGISTER_APP:
		var reg share.RegisterApp
		if err := share.DecodeMsg(msgbody, &reg); err != nil {
			clientconn.Close()
			log.LogFatalf(err)
		}
		app := &app{typ: reg.Type, id: reg.Id, name: reg.Name, conn: clientconn, host: reg.Host, port: reg.Port, clienthost: reg.ClientHost, clientport: reg.ClientPort}
		log.LogMessage(app.id, ":", app.conn.RemoteAddr().String())
		app.SendList()
		AddApp(app)
		app.Loop()
	case share.M_REGISTER_AGENT:
		var reg share.RegisterAgent
		if err := share.DecodeMsg(msgbody, &reg); err != nil {
			clientconn.Close()
			log.LogFatalf(err)
		}

		agent := &AgentNode{id: reg.AgentId, nobalance: reg.NoBalance}
		agent.Handle(clientconn)
		context.agentlist.AddAgent(agent)
		log.LogMessage("agent add:", reg.AgentId)
	default:
		log.LogError("first message must reg app")
		return
	}

}
