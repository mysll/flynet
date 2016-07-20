package rpc

import (
	"net"
	"runtime"
	"server/libs/log"
)

func CreateRpcService(service map[string]interface{}, handle map[string]interface{}, ch chan *RpcCall) (rpcsvr *Server, err error) {
	rpcsvr = NewServer(ch)
	for k, v := range service {
		err = rpcsvr.RegisterName("S2S"+k, v)
		if err != nil {
			return
		}
	}

	for k, v := range handle {
		err = rpcsvr.RegisterName("C2S"+k, v)
		if err != nil {
			return
		}
	}

	return
}

func CreateService(rs *Server, l net.Listener) {

	log.LogMessage("rpc start at:", l.Addr().String())
	for {
		conn, err := l.Accept()
		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			log.LogWarning("TCP", "temporary Accept() failure - ", err.Error())
			runtime.Gosched()
			continue
		}
		if err != nil {
			log.LogWarning("rpc accept quit")
			break
		}
		//启动服务
		log.LogInfo("new rpc client,", conn.RemoteAddr())
		go rs.ServeConn(conn, MAX_BUF_LEN)
	}
}
