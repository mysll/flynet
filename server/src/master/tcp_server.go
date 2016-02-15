package master

import (
	"libs/log"
	"net"
	"share"
	"util"
)

type tcp_server struct {
	context *context
}

func (tcp *tcp_server) Handle(clientconn net.Conn) {

	buff := make([]byte, 2048)
	id, msgbody, err := util.ReadPkg(clientconn, buff)
	if err != nil {
		log.LogError(err)
		clientconn.Close()
		return
	}

	var reg share.RegisterApp
	switch id {
	case share.M_REGISTER_APP:
		if err := share.DecodeMsg(msgbody, &reg); err != nil {
			clientconn.Close()
			log.LogFatalf(err)
		}
	default:
		log.LogError("first message must reg app")
		return
	}

	app := &app{typ: reg.Type, id: reg.AppId, conn: clientconn, host: reg.Host, port: reg.Port, clienthost: reg.ClientHost, clientport: reg.ClientPort}

	log.LogMessage(app.id, ":", app.conn.RemoteAddr().String())

	app.SendList()
	AddApp(app)
	app.Loop()
}
