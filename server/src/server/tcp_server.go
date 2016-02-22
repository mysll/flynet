package server

import (
	"errors"
	"io"
	"libs/log"
	"libs/rpc"
	"net"
	"share"
	"util"

	pb "github.com/golang/protobuf/proto"
)

type ClientCodec struct {
	rwc      io.ReadWriteCloser
	cachebuf []byte
	node     *ClientNode
}

func (c *ClientCodec) ReadRequestHeader(req *rpc.Request) error {
	req.ServiceMethod = "C2SC2SHelper.Call"
	req.Count = 1
	return nil
}

func (c *ClientCodec) ReadRequestBody(body interface{}) error {
	for {
		id, data, err := util.ReadPkg(c.rwc, c.cachebuf)
		if err != nil {
			return err
		}
		switch id {
		case share.C2S_PING:
			c.node.Ping()
			break
		case share.C2S_RPC:
			if pmsg, ok := body.(pb.Message); ok {
				return pb.Unmarshal(data, pmsg)
			}
			return errors.New("args not protobuf")
		}
	}
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
	id := core.clientList.Add(clientconn)
	mailbox := rpc.NewMailBox(core.Id, "session", id, core.AppId)
	core.Emitter.Push(NEWUSERCONN, map[string]interface{}{"id": id}, true)
	cl := core.clientList.FindNode(id)
	cl.MailBox = mailbox
	cl.Run()
	codec := &ClientCodec{}
	codec.rwc = clientconn
	codec.cachebuf = make([]byte, SENDBUFLEN)
	codec.node = cl
	log.LogInfo("new client:", mailbox, ",", clientconn.RemoteAddr())
	core.rpcServer.ServeCodec(codec, core.rpcCh)
	core.Emitter.Push(LOSTUSERCONN, map[string]interface{}{"id": cl.Session}, true)
	log.LogMessage("client handle quit")
}
