package master

import (
	"errors"
	"fmt"
	"io"
	"net"
	"server/libs/log"
	"server/share"
	"server/util"
	"strings"
	"time"
)

type Agent struct {
	conn net.Conn
	addr string
	rwc  io.ReadWriteCloser
	quit bool
}

func (a *Agent) Connect(host string, port int, keepconnect bool, cb func()) error {
	a.addr = fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", a.addr)
	if err != nil {
		if keepconnect {
			go a.TryConnect(cb)
		}
		return err
	}

	a.conn = conn
	a.rwc = a.conn
	cb()
	go a.readloop()
	return nil
}

func (a *Agent) TryConnect(cb func()) {
	for !a.quit {
		time.Sleep(time.Second * 5)
		conn, err := net.Dial("tcp", a.addr)
		if err != nil {
			continue
		}

		a.conn = conn
		a.rwc = a.conn
		cb()
		go a.readloop()

		break
	}
}

func (a *Agent) Close() {
	if a.conn != nil {
		a.rwc.Close()
	}

	a.quit = true
}

func (a *Agent) Send(data []byte) error {
	if a.conn == nil {
		return errors.New("socket not create")
	}

	_, err := a.rwc.Write(data)
	return err
}

func (a *Agent) Register(agentid string, nobalance bool) error {
	out, err := share.CreateRegisterAgent(agentid, nobalance)
	if err != nil {
		log.LogFatalf(err)
	}

	if _, err := a.rwc.Write(out); err != nil {
		return err
	}

	return nil
}

func (a *Agent) CreateApp(create share.CreateApp) {
	err := CreateApp(create.AppName, create.AppUid, create.Type, create.Args)
	res := "ok"
	if err != nil {
		res = err.Error()
		log.LogError(err)
	}

	if create.CallApp != 0 {
		data, err := share.CreateAppBakMsg(create.ReqId, create.AppUid, res)
		if err != nil {
			log.LogError(err)
			return
		}

		out, err := share.CreateForwardMsg(create.CallApp, data)
		if err != nil {
			log.LogError(err)
			return
		}

		a.Send(out)
	}
}

func (a *Agent) Handle(id uint16, msgbody []byte) error {
	switch id {
	case share.M_SHUTDOWN:
		context.Stop()
	case share.M_CREATEAPP:
		create := share.CreateApp{}
		if err := share.DecodeMsg(msgbody, &create); err != nil {
			log.LogError(err)
		}

		go a.CreateApp(create)

	}
	return nil
}

func (a *Agent) readloop() {
	buffer := make([]byte, 2048)
	for !a.quit {
		id, msg, err := util.ReadPkg(a.rwc, buffer)
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
				log.LogError(err)
			}
			break
		}

		if err := a.Handle(id, msg); err != nil {
			log.LogError(err)
			break
		}
	}

	log.LogInfo("agent pipe quit")
	a.Close()
}
