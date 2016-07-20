package server

import (
	"errors"
	"server/libs/log"
	"server/libs/rpc"
	"server/share"
	"server/util"
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

func (t *C2SHelper) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("Call", t.Call)
}

//处理客户端的调用
func (ch *C2SHelper) Call(sender rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	node, sm, data, err := core.rpcProto.DecodeRpcMessage(msg)
	if err != nil {
		log.LogError(err)
		return nil
	}

	var app *RemoteApp
	if node == "." {
		app = &RemoteApp{Id: core.AppId, Name: core.Name}
	} else {
		if app = GetAppByName(node); app == nil {
			log.LogError(ErrAppNotFound)
			return nil
		}
	}

	log.LogMessage("client call: ", node, "/", sm)
	if err := app.Handle(sender, sm, data); err != nil {
		log.LogError(err)
	}
	return nil
}

//服务器向客户端的远程调用
type S2CHelper struct {
	sendbuf   []byte
	cachedata map[int64]*rpc.Message
}

func (t *S2CHelper) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("Call", t.Call)
}

func NewS2CHelper() *S2CHelper {
	sc := &S2CHelper{}
	sc.sendbuf = make([]byte, 0, SENDBUFLEN)
	sc.cachedata = make(map[int64]*rpc.Message)
	return sc
}

//发送缓存数据
func (s *S2CHelper) flush() {
	for k, v := range s.cachedata {
		if v != nil {
			c := core.clientList.FindNode(k)
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
func (s *S2CHelper) Call(src rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	request := &share.S2CMsg{}
	reader := NewMessageReader(msg)
	if err := reader.ReadObject(request); err != nil {
		log.LogError(err)
		return nil
	}

	out, err := util.CreateMsg(s.sendbuf, request.Data, share.S2C_RPC)
	if err != nil {
		log.LogError(err)
		return nil
	}

	err = s.call(src, request.To, request.Method, out)
	if err != nil {
		log.LogError(err)
	}

	return nil
}

func (s *S2CHelper) call(src rpc.Mailbox, session int64, method string, out []byte) error {
	c := core.clientList.FindNode(session)
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
	cachedata = rpc.NewMessage(packagesize)
	s.cachedata[session] = cachedata
	cachedata.Body = append(cachedata.Body, out...)
	return nil
}

//处理服务器向客户端的广播
func (s *S2CHelper) Broadcast(src rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	request := &share.S2CBrocast{}
	reader := NewMessageReader(msg)
	if err := reader.ReadObject(request); err != nil {
		log.LogError(err)
		return nil
	}

	out, err := util.CreateMsg(s.sendbuf, request.Data, share.S2C_RPC)
	if err != nil {
		log.LogError(err)
		return nil
	}

	for _, to := range request.To {
		if err = s.call(src, to, request.Method, out); err != nil {
			log.LogError(to, err)
		}
	}
	return nil
}
