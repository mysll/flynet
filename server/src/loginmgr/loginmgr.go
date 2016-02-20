package loginmgr

import (
	"fmt"
	"libs/log"
	"net"
	"server"
	"sync"
	"util"
)

var (
	App = &LoginMgr{}
)

type LoginMgr struct {
	*server.Server
	l        sync.Mutex
	serial   int
	listener net.Listener
}

func (l *LoginMgr) OnPrepare() bool {
	log.TraceInfo(l.Id, "init link")
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", l.ClientHost, l.ClientPort))
	if err != nil {
		panic(err)
	}
	l.listener = listener
	tcpserver := &handler{}
	l.WaitGroup.Wrap(func() { util.TCPServer(listener, tcpserver) })
	log.TraceInfo(l.Id, "start link complete")
	return true
}

func (l *LoginMgr) OnShutdown() bool {
	l.listener.Close()
	return true
}

func (l *LoginMgr) RawSock() bool {
	return true
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {
}