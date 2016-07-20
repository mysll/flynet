package server

import (
	"errors"
	"io"
	"net"
	"server/libs/log"
	"server/share"
	"server/util"
	"strings"
)

type peerhandle interface {
	Handle(id uint16, data []byte) error
}

type peer struct {
	conn net.Conn
	addr string
	rwc  io.ReadWriteCloser
	h    peerhandle
}

func (p *peer) Connect() error {
	conn, err := net.Dial("tcp", p.addr)
	if err != nil {
		return err
	}

	p.conn = conn
	p.rwc = p.conn
	go p.readloop()
	return nil
}

func (p *peer) Reg(server *Server) error {

	out, err := share.CreateRegisterAppMsg(server.Type, server.AppId, server.Name, server.Host, server.Port, server.ClientHost, server.ClientPort)
	if err != nil {
		log.LogFatalf(err)
	}

	if _, err := p.rwc.Write(out); err != nil {
		return err
	}

	return nil
}

func (p *peer) Ready() error {
	out, err := util.CreateMsg(nil, []byte{}, share.M_READY)
	if err != nil {
		log.LogFatalf(err)
	}

	if _, err := p.rwc.Write(out); err != nil {
		return err
	}

	return nil
}

func (p *peer) Send(data []byte) error {
	if p.conn == nil {
		return errors.New("socket not create")
	}

	_, err := p.rwc.Write(data)
	return err
}

func (p *peer) Close() {
	if p.conn != nil {
		p.rwc.Close()
	}
}

func (p *peer) readloop() {
	buffer := make([]byte, 2048)
	for !core.quit {
		id, msg, err := util.ReadPkg(p.rwc, buffer)
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
				log.LogError(err)
			}
			break
		}

		if p.h != nil {
			if err := p.h.Handle(id, msg); err != nil {
				log.LogError(err)
				break
			}
		}
	}

	log.LogInfo("peer node closed")
	p.Close()
}
