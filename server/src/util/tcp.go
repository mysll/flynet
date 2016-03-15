package util

import (
	"libs/log"
	"net"
	"net/http"
	"runtime"
	"strings"

	"golang.org/x/net/websocket"
)

type TCPHandler interface {
	Handle(net.Conn)
}

func TCPServer(listener net.Listener, handler TCPHandler) {
	log.TraceInfo("TCP", "listening on ", listener.Addr().String())
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				log.LogWarning("TCP", "temporary Accept() failure - ", err.Error())
				runtime.Gosched()
				continue
			}
			// theres no direct way to detect this error because it is not exposed
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.LogError("listener.Accept() - ", err.Error())
			}
			break
		}
		go handler.Handle(clientConn)
	}

	log.TraceInfo("TCP", "closing ", listener.Addr().String())
}

type WSHandler interface {
	Handle(ws *websocket.Conn)
}

func WSServer(l net.Listener, handler WSHandler) {
	log.TraceInfo("TCP", "listening on ", l.Addr().String())
	http.Handle("/socket.io", websocket.Handler(handler.Handle))
	http.Serve(l, nil)
	log.TraceInfo("TCP", "closing ", l.Addr().String())
}
