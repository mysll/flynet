package rpc

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"libs/log"
	"strings"
	"sync"
	"time"
	"util"
)

var (
	ErrTooLong    = errors.New("message is to long")
	rpccallCache  = make(chan *RpcCall, 256)
	headerCache   = make(chan *Header, 256)
	responseCache = make(chan *Response, 32)
)

type CB func(Mailbox, *Message) *Message

type RpcRegister interface {
	RegisterCallback(Servicer)
}

type Servicer interface {
	RegisterCallback(string, CB)
	ThreadPush(*RpcCall) bool
}

func NewRpcCall() *RpcCall {
	var call *RpcCall
	select {
	case call = <-rpccallCache:
	default:
		call = &RpcCall{}
	}
	return call
}

func NewHeader() *Header {
	var header *Header
	select {
	case header = <-headerCache:
	default:
		header = &Header{}
	}
	return header
}

func NewResponse() *Response {
	var resp *Response
	select {
	case resp = <-responseCache:
	default:
		resp = &Response{}
	}
	return resp
}

type Header struct {
	ServiceMethod string // format: "Service.Method"
	Seq           uint64 // sequence number chosen by client
	Mb            uint64
}

func (h *Header) Free() {
	select {
	case headerCache <- h:
	default:
	}
}

type Response struct {
	Seq   uint64
	Reply *Message
	cb    ReplyCB
}

func (r *Response) Free() {
	if r.Reply != nil {
		r.Reply.Free()
		r.Reply = nil
		r.cb = nil
	}
	select {
	case responseCache <- r:
	default:
	}
}

type RpcCall struct {
	session uint64
	srv     *service
	header  *Header
	message *Message
	reply   *Message
	method  CB
	cb      ReplyCB
}

func (call *RpcCall) Call() error {
	mb := NewMailBoxFromUid(call.header.Mb)
	call.reply = call.method(mb, call.message)
	return nil
}

func (call *RpcCall) GetSrc() Mailbox {
	mb := NewMailBoxFromUid(call.header.Mb)
	return mb
}

func (call *RpcCall) GetMethod() string {
	return call.header.ServiceMethod
}

func (call *RpcCall) IsThreadWork() bool {
	return call.srv.ThreadPush(call)
}

func (call *RpcCall) Free() error {
	call.message.Free()
	call.header.Free()
	call.header = nil
	call.message = nil
	call.srv = nil
	call.cb = nil
	call.session = 0
	if call.reply != nil {
		call.reply.Free()
		call.reply = nil
	}
	select {
	case rpccallCache <- call:
	default:
	}
	return nil
}

func (call *RpcCall) Done() error {
	if call.header.Seq != 0 || call.cb != nil {
		return call.srv.svr.sendResponse(call)
	}
	return nil
}

type service struct {
	svr    *Server
	rcvr   interface{}
	method map[string]CB
}

func (srv *service) RegisterCallback(name string, cb CB) {
	srv.method[name] = cb
}

func (srv *service) ThreadPush(call *RpcCall) bool {
	if t, ok := srv.rcvr.(Threader); ok {
		return t.NewJob(call)
	}
	return false
}

type Session struct {
	rwc       io.ReadWriteCloser
	codec     ServerCodec
	sendQueue chan *Response
	quit      bool
}

type Server struct {
	mutex      sync.RWMutex
	serial     uint64
	serviceMap map[string]*service
	sessions   map[uint64]*Session
	sendQueue  chan *Response
	ch         chan *RpcCall
}

func (server *Server) getCall(servicemethod string, src Mailbox, cb ReplyCB, args ...interface{}) (*RpcCall, error) {
	var msg *Message
	if len(args) > 0 {
		msg = NewMessage(MAX_BUF_LEN)
		ar := util.NewStoreArchiver(msg.Body)
		for i := 0; i < len(args); i++ {
			ar.Write(args[i])
		}
		msg.Body = msg.Body[:ar.Len()]
	} else {
		msg = NewMessage(1)
	}

	call := NewRpcCall()
	call.header = NewHeader()
	call.message = msg
	call.header.Seq = 0
	call.header.Mb = src.Uid
	call.header.ServiceMethod = servicemethod
	call.cb = cb
	dot := strings.LastIndex(call.header.ServiceMethod, ".")
	if dot < 0 {
		err := fmt.Errorf("rpc: service/method request ill-formed: %s", call.header.ServiceMethod)
		log.LogError(err)
		call.Free()
		return nil, err
	}
	serviceName := call.header.ServiceMethod[:dot]
	methodName := call.header.ServiceMethod[dot+1:]

	call.srv = server.serviceMap[serviceName]
	if call.srv == nil {
		err := fmt.Errorf("rpc: can't find service %s", call.header.ServiceMethod)
		log.LogError(err)
		call.Free()
		return nil, err
	}

	call.method = call.srv.method[methodName]
	if call.method == nil {
		err := fmt.Errorf("rpc: can't find method %s", call.header.ServiceMethod)
		log.LogError(err)
		call.Free()
		return nil, err
	}

	return call, nil
}

func (server *Server) Call(servicemethod string, src Mailbox, args ...interface{}) error {
	call, err := server.getCall(servicemethod, src, nil, args...)
	if call != nil {
		server.ch <- call
	}
	return err
}

func (server *Server) CallBack(servicemethod string, src Mailbox, cb ReplyCB, args ...interface{}) error {
	call, err := server.getCall(servicemethod, src, cb, args...)
	if call != nil {
		server.ch <- call
	}
	return err
}

func (server *Server) ServeConn(conn io.ReadWriteCloser, maxlen uint16) {
	codec := &byteServerCodec{
		rwc:    conn,
		encBuf: bufio.NewWriter(conn),
	}
	server.ServeCodec(codec, maxlen)
}

func (server *Server) ServeCodec(codec ServerCodec, maxlen uint16) {
	var serial uint64
	server.mutex.Lock()
	serial = server.serial
	server.serial++
	session := &Session{rwc: codec.GetConn(), codec: codec, sendQueue: make(chan *Response, 32)}
	server.sessions[serial] = session
	server.mutex.Unlock()
	go session.send()
	log.LogMessage("start new service:", serial)
	for {
		msg, err := codec.ReadRequest(maxlen)
		if err != nil {
			if err != io.EOF &&
				!strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") &&
				!strings.Contains(err.Error(), "use of closed network connection") {
				log.LogError("rpc err:", err)
			} else {
				log.LogMessage("service client closed")
			}
			break
		}

		call := server.createCall(msg)
		call.session = serial
		if call != nil {
			server.ch <- call
		}
	}

	session.quit = true
	codec.Close()
	log.LogMessage("service quit:", serial)
	server.mutex.Lock()
	delete(server.sessions, serial)
	server.mutex.Unlock()
}

func (session *Session) send() {
	for {
		select {
		case resp := <-session.sendQueue:
			if resp.Seq != 0 {
				err := session.codec.WriteResponse(resp.Seq, resp.Reply)
				if err != nil {
					resp.Free()
					log.LogError(err)
					return
				}
			} else if resp.cb != nil {
				resp.cb(resp.Reply)
			}
			resp.Free()
		default:
			if session.quit {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func (server *Server) sendResponse(call *RpcCall) error {

	server.mutex.RLock()
	session := server.sessions[call.session]
	server.mutex.RUnlock()
	if session == nil {
		return fmt.Errorf("session not found", call.session)
	}

	resp := NewResponse()
	resp.Seq = call.header.Seq
	resp.cb = call.cb
	var msg *Message
	if call.reply != nil && len(call.reply.Body) > 0 {
		msg = NewMessage(len(call.reply.Body))
		msg.Body = append(msg.Body, call.reply.Body...)
	} else {
		msg = NewMessage(1)
	}
	resp.Reply = msg
	session.sendQueue <- resp
	return nil
}

func (server *Server) createCall(msg *Message) *RpcCall {
	call := NewRpcCall()
	call.header = NewHeader()
	call.message = msg
	ar := util.NewLoadArchiver(msg.Header)
	var err error
	call.header.Seq, err = ar.ReadUInt64()
	if err != nil {
		log.LogError(err)
		call.Free()
		return nil
	}
	call.header.Mb, err = ar.ReadUInt64()
	if err != nil {
		log.LogError(err)
		call.Free()
		return nil
	}
	call.header.ServiceMethod, err = ar.ReadString()
	if err != nil {
		log.LogError(err)
		call.Free()
		return nil
	}
	dot := strings.LastIndex(call.header.ServiceMethod, ".")
	if dot < 0 {
		log.LogError("rpc: service/method request ill-formed: ", call.header.ServiceMethod)
		call.Free()
		return nil
	}
	serviceName := call.header.ServiceMethod[:dot]
	methodName := call.header.ServiceMethod[dot+1:]

	call.srv = server.serviceMap[serviceName]
	if call.srv == nil {
		log.LogError("rpc: can't find service ", call.header.ServiceMethod)
		call.Free()
		return nil
	}

	call.method = call.srv.method[methodName]
	if call.method == nil {
		log.LogError("rpc: can't find method ", call.header.ServiceMethod)
		call.Free()
		return nil
	}
	return call
}

func ReadMessage(rwc io.Reader, maxrx uint16) (*Message, error) {
	var sz uint16
	var headlen uint16
	var err error
	var msg *Message

	if err = binary.Read(rwc, binary.BigEndian, &sz); err != nil {
		return nil, err
	}

	// Limit messages to the maximum receive value, if not
	// unlimited.  This avoids a potential denaial of service.
	if sz < 0 || (maxrx > 0 && sz > maxrx) {
		return nil, ErrTooLong
	}

	if err = binary.Read(rwc, binary.BigEndian, &headlen); err != nil {
		return nil, err
	}
	bodylen := int(sz - headlen)
	msg = NewMessage(bodylen)
	msg.Header = msg.Header[0:headlen]
	if _, err = io.ReadFull(rwc, msg.Header); err != nil {
		msg.Free()
		return nil, err
	}

	if bodylen > 0 {
		msg.Body = msg.Body[0:bodylen]

		if _, err = io.ReadFull(rwc, msg.Body); err != nil {
			msg.Free()
			return nil, err
		}
	}

	return msg, nil
}

func (server *Server) RegisterName(name string, rcvr interface{}) error {
	if reg, ok := rcvr.(RpcRegister); ok {
		srv := &service{}
		srv.svr = server
		srv.rcvr = rcvr
		srv.method = make(map[string]CB, 16)
		reg.RegisterCallback(srv)
		server.serviceMap[name] = srv
		return nil
	}

	return fmt.Errorf("%s is not RpcRegister", name)
}

func (server *Server) GetRpcInfo(name string) []string {
	var ret []string
	if s, ok := server.serviceMap[name]; ok {
		for k, _ := range s.method {
			ret = append(ret, k)
		}
	}
	return ret
}

func NewServer(ch chan *RpcCall) *Server {
	s := &Server{}
	s.serviceMap = make(map[string]*service)
	s.ch = ch
	s.sessions = make(map[uint64]*Session)
	s.sessions[0] = &Session{} //本地回调
	s.sessions[0].sendQueue = make(chan *Response, 32)
	s.serial = 1
	go s.sessions[0].send() //读取本地消息回调
	return s
}

type byteServerCodec struct {
	rwc    io.ReadWriteCloser
	encBuf *bufio.Writer
	closed bool
}

func (c *byteServerCodec) ReadRequest(maxrc uint16) (*Message, error) {
	return ReadMessage(c.rwc, maxrc)
}

func (c *byteServerCodec) WriteResponse(seq uint64, body *Message) (err error) {
	if body == nil {
		body = NewMessage(16)
	}

	w := util.NewStoreArchiver(body.Header)
	w.Write(seq)
	body.Header = body.Header[:w.Len()]
	count := uint16(len(body.Header) + len(body.Body))
	binary.Write(c.encBuf, binary.BigEndian, count)                    //数据大小
	binary.Write(c.encBuf, binary.BigEndian, uint16(len(body.Header))) //头部大小
	c.encBuf.Write(body.Header)
	if len(body.Body) > 0 {
		c.encBuf.Write(body.Body)
	}
	body.Header = body.Header[:0]
	return c.encBuf.Flush()
}

func (c *byteServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

func (c *byteServerCodec) GetConn() io.ReadWriteCloser {
	return c.rwc
}

type ServerCodec interface {
	ReadRequest(maxrc uint16) (*Message, error)
	WriteResponse(seq uint64, body *Message) (err error)
	GetConn() io.ReadWriteCloser
	Close() error
}
