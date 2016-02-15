package server

import (
	"errors"
	"libs/log"
	"libs/rpc"
	"pb/c2s"
	"share"
	"util"
)

var (
	ErrClientNotFound = errors.New("client not found")
	ErrAppNotFound    = errors.New("rpc call app not found")
	packagesize       = 1400 //MTU大小
)

//客户端向服务器的远程调用辅助工具
type C2SHelper struct {
	sendbuf []byte
}

//处理客户端的调用
func (ch *C2SHelper) Call(sender rpc.Mailbox, request c2s.Rpc) error {
	node := request.GetNode()

	var app *RemoteApp
	if node == "." {
		app = &RemoteApp{AppId: context.Server.Id}
	} else {
		if app = GetApp(node); app == nil {
			return ErrAppNotFound
		}
	}

	log.LogMessage("client call: ", node, "/", request.GetServicemethod())
	return app.Handle(sender, request.GetServicemethod(), request.GetData())
}

//服务器向客户端的远程调用
type S2CHelper struct {
	sendbuf   []byte
	cachedata map[int64]*Message
}

func NewS2CHelper() *S2CHelper {
	sc := &S2CHelper{}
	sc.sendbuf = make([]byte, 0, SENDBUFLEN)
	sc.cachedata = make(map[int64]*Message)
	return sc
}

//发送缓存数据
func (s *S2CHelper) flush() {
	for k, v := range s.cachedata {
		if v != nil {
			c := context.Server.clientList.FindNode(k)
			if c != nil {
				if err := c.SendMessage(v); err != nil {
					log.LogError(k, err)
				}
			}
			v.Free()
		}
		delete(s.cachedata, k)
	}
}

//处理服务器向客户端的调用，对消息进行封装转成客户端的协议
func (s *S2CHelper) Call(src rpc.Mailbox, request share.S2CMsg) error {
	out, err := util.CreateMsg(s.sendbuf, request.Data, share.S2C_RPC)
	if err != nil {
		return err
	}

	return s.call(src, request.To, request.Method, out)
}

func (s *S2CHelper) call(src rpc.Mailbox, session int64, method string, out []byte) error {
	c := context.Server.clientList.FindNode(session)
	if c == nil {
		return ErrClientNotFound
	}

	cachedata, exist := s.cachedata[session]

	if exist && cachedata != nil { //优先写入缓存
		if len(cachedata.Body)+len(out) <= packagesize { //可以写入
			cachedata.Body = append(cachedata.Body, out...)
			return nil
		}

		//超出长度，则先发送当前消息，保证消息顺序
		if err := c.SendMessage(cachedata); err != nil {
			log.LogError(session, err)
		}
		cachedata.Free()
		cachedata = nil
		s.cachedata[session] = nil
	}

	if len(out) > packagesize { //超出包体大小,直接发送
		//直接发送
		if err := c.Send(out); err != nil {
			log.LogError(session, err)
		}
		return nil
	}

	//写入缓冲区
	cachedata = NewMessage(packagesize)
	s.cachedata[session] = cachedata
	cachedata.Body = append(cachedata.Body, out...)
	return nil
}

//处理服务器向客户端的广播
func (s *S2CHelper) Broadcast(src rpc.Mailbox, request share.S2CBrocast) error {
	for _, to := range request.To {
		out, err := util.CreateMsg(s.sendbuf, request.Data, share.S2C_RPC)
		if err != nil {
			return err
		}

		if err = s.call(src, to, request.Method, out); err != nil {
			log.LogError(to, err)
		}
	}
	return nil
}
