package server

import (
	"errors"
	"io"
	"libs/log"
	"libs/rpc"
	"net"
	"share"
	"util"

	"golang.org/x/net/websocket"
)

var (
	ERRNOTSUPPORT = errors.New("not support")
)

type ClientCodec struct {
	rwc      io.ReadWriteCloser
	cachebuf []byte
	node     *ClientNode
}

func (c *ClientCodec) ReadRequest(maxrc uint16) (*rpc.Message, error) {
	for {
		id, data, err := util.ReadPkg(c.rwc, c.cachebuf)
		if err != nil {
			return nil, err
		}
		switch id {
		case share.C2S_PING:
			c.node.Ping()
			break
		case share.C2S_RPC:
			msg := rpc.NewMessage(len(data))
			ar := NewHeadWriter(msg)
			ar.Write(uint64(0))
			ar.Write(c.node.MailBox.Uid)
			ar.Write("C2SC2SHelper.Call")
			msg.Header = msg.Header[:ar.Len()]
			msg.Body = append(msg.Body, data...)
			return msg, nil
		}
	}

	return nil, nil
}

func (c *ClientCodec) WriteResponse(seq uint64, body *rpc.Message) (err error) {
	return ERRNOTSUPPORT
}

func (c *ClientCodec) Close() error {
	return c.rwc.Close()
}

func (c *ClientCodec) Mailbox() *rpc.Mailbox {
	return &c.node.MailBox
}

type ClientHandler struct {
}

func (c *ClientHandler) Handle(clientconn net.Conn) {
	if core.Closing {
		clientconn.Close()
		return
	}

	id := core.clientList.Add(clientconn, clientconn.RemoteAddr().String())
	mailbox := rpc.NewMailBox(1, id, core.AppId)
	core.Emitter.Push(NEWUSERCONN, map[string]interface{}{"id": id}, true)
	cl := core.clientList.FindNode(id)
	cl.MailBox = mailbox
	cl.Run()
	codec := &ClientCodec{}
	codec.rwc = clientconn
	codec.cachebuf = make([]byte, SENDBUFLEN)
	codec.node = cl
	log.LogInfo("new client:", mailbox, ",", clientconn.RemoteAddr())
	core.rpcServer.ServeCodec(codec, rpc.MAX_BUF_LEN, core.rpcCh)
	core.Emitter.Push(LOSTUSERCONN, map[string]interface{}{"id": cl.Session}, true)
	log.LogMessage("client handle quit")
}

type WSClientHandler struct {
}

func (c *WSClientHandler) Handle(ws *websocket.Conn) {
	if core.Closing {
		ws.Close()
		return
	}
	rwc := NewWSConn(ws)
	id := core.clientList.Add(rwc, ws.RemoteAddr().String())
	mailbox := rpc.NewMailBox(1, id, core.AppId)
	core.Emitter.Push(NEWUSERCONN, map[string]interface{}{"id": id}, true)
	cl := core.clientList.FindNode(id)
	cl.MailBox = mailbox
	cl.Run()
	codec := &ClientCodec{}
	codec.rwc = rwc
	codec.cachebuf = make([]byte, SENDBUFLEN)
	codec.node = cl
	log.LogInfo("new client:", mailbox, ",", ws.RemoteAddr())
	core.rpcServer.ServeCodec(codec, rpc.MAX_BUF_LEN, core.rpcCh)
	core.Emitter.Push(LOSTUSERCONN, map[string]interface{}{"id": cl.Session}, true)
	log.LogMessage("client handle quit")
}
