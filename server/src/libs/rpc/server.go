package rpc

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"libs/log"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	pb "github.com/golang/protobuf/proto"
)

type pusher interface {
	Push(call *RpcCall) bool
}

// Precompute the reflect type for error.  Can't use error directly
// because Typeof takes an empty interface value.  This is annoying.
var (
	typeOfError = reflect.TypeOf((*error)(nil)).Elem()
	iface       pusher
	pushIface   = reflect.TypeOf(&iface).Elem()
	typeOfSrc   = reflect.TypeOf((*Mailbox)(nil)).Elem()
)

type Request struct {
	ServiceMethod string
	Count         uint16
}

type RpcCall struct {
	Serial  int32
	Req     Request
	Service *service
	Server  *Server
	Mtype   *methodType
	Src     reflect.Value
	Argvs   []reflect.Value
	next    *RpcCall
}

func (r *RpcCall) Call() error {
	return r.Service.call(r)
}

func (r *RpcCall) Free() {
	r.Server.freeRpcCall(r)
}

type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	SrcType    reflect.Type
	ArgsType   []reflect.Type
	numCalls   uint
}

type service struct {
	sync.Mutex
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
	queue  pusher
}

func (s *service) Push(r *RpcCall) bool {
	if s.queue != nil {
		return s.queue.Push(r)
	}

	return false
}

// Server represents an RPC Server.
type Server struct {
	mu         sync.RWMutex
	serviceMap map[string]*service
	callLock   sync.Mutex
	freeCall   *RpcCall
	serial     int32
	ch         chan *RpcCall
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{serviceMap: make(map[string]*service)}
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
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

// Register publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method
//	- two arguments, both of exported type
//	- the second argument is a pointer
//	- one return value, of type error
// It returns an error if the receiver is not an exported type or has
// no suitable methods. It also logs the error using package log.
// The client accesses each method using a string of the form "Type.Method",
// where Type is the receiver's concrete type.
func (server *Server) Register(rcvr interface{}) error {
	return server.register(rcvr, "", false)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}) error {
	return server.register(rcvr, name, true)
}

func (server *Server) register(rcvr interface{}, name string, useName bool) error {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)

	if s.typ.Implements(pushIface) {
		s.queue = s.rcvr.Interface().(pusher)
	}

	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		s := "rpc.Register: no service name for type " + s.typ.String()
		log.LogDebug(s)
		return errors.New(s)
	}
	if !isExported(sname) && !useName {
		s := "rpc.Register: type " + sname + " is not exported"
		log.LogDebug(s)
		return errors.New(s)
	}
	if _, present := server.serviceMap[sname]; present {
		return errors.New("rpc: service already defined: " + sname)
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		log.LogDebug(str)
		return errors.New(str)
	}
	server.serviceMap[s.name] = s
	return nil
}

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: receiver, src, args
		if mtype.NumIn() < 3 {
			if reportErr {
				log.LogDebug("method ", mname, " has wrong number of ins:", mtype.NumIn())
			}
			continue
		}

		srcType := mtype.In(1)
		if srcType != typeOfSrc {
			if reportErr {
				log.LogDebug("method ", mname, " src", srcType.String(), "not MailBox")
			}
			continue
		}

		argsType := make([]reflect.Type, 0, mtype.NumIn()-2)
		flag := false
		for i := 2; i < mtype.NumIn(); i++ {
			argType := mtype.In(i)
			if !isExportedOrBuiltinType(argType) {
				if reportErr {
					log.LogDebug(mname, " argument type not exported:", argType)
				}
				flag = true
				break
			}
			argsType = append(argsType, argType)
		}

		if flag {
			continue
		}

		// Method needs one out.
		if mtype.NumOut() != 1 {
			if reportErr {
				log.LogDebug("method ", mname, " has wrong number of outs:", mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if reportErr {
				log.LogDebug("method ", mname, " returns", returnType.String(), "not error")
			}
			continue
		}
		methods[mname] = &methodType{method: method, SrcType: srcType, ArgsType: argsType}
	}

	return methods
}

// A value sent as a placeholder for the server's response value when the server
// receives an invalid request. It is never decoded by the client since the Response
// contains an error when it is used.
var invalidRequest = struct{}{}

func (m *methodType) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}

func (s *service) call(r *RpcCall) error {

	r.Mtype.Lock()
	r.Mtype.numCalls++
	r.Mtype.Unlock()
	function := r.Mtype.method.Func

	args := make([]reflect.Value, 2+len(r.Argvs))
	args[0] = s.rcvr
	args[1] = r.Src
	for k, v := range r.Argvs {
		args[k+2] = v
	}

	s.Lock()
	defer s.Unlock()
	returnValues := function.Call(args)
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()

	if errInter != nil {
		return errInter.(error)
	}

	return nil
}

type gobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	encBuf *bufio.Writer
}

func (c *gobServerCodec) ReadRequestHeader(req *Request) error {
	return c.dec.Decode(req)
}

func (c *gobServerCodec) ReadRequestBody(body interface{}) error {
	if _, ok := body.(pb.Message); ok {
		var buf []byte
		if err := c.dec.Decode(&buf); err != nil {
			return err
		}
		return pb.Unmarshal(buf, body.(pb.Message))
	}
	return c.dec.Decode(body)
}

func (c *gobServerCodec) Close() error {
	return c.rwc.Close()
}

func (c *gobServerCodec) Mailbox() *Mailbox {
	return nil
}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the gob wire format (see package gob) on the
// connection.  To use an alternate codec, use ServeCodec.
func (server *Server) ServeConn(conn io.ReadWriteCloser, ch chan *RpcCall) {
	buf := bufio.NewWriter(conn)
	srv := &gobServerCodec{conn, gob.NewDecoder(conn), buf}
	server.ServeCodec(srv, ch)
}

// ServeCodec is like ServeConn but uses the specified codec to
// decode requests and encode responses.
func (server *Server) ServeCodec(codec ServerCodec, ch chan *RpcCall) {
	server.ch = ch
	for {
		rc, keepReading, err := server.readRequest(codec, codec.Mailbox())
		if err != nil {
			if err != io.EOF &&
				!strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") &&
				!strings.Contains(err.Error(), "use of closed network connection") {
				log.LogError("rpc err:", err)
			}

			if !keepReading {
				log.LogMessage("rpc server closed")
				break
			}
			// send a response if we actually managed to read a header.
			if rc != nil {
				server.freeRpcCall(rc)
			}

			continue
		}
		ch <- rc
	}
	codec.Close()
	log.LogMessage("server loop quit")
}

func (server *Server) getRpcCall() *RpcCall {
	server.callLock.Lock()
	call := server.freeCall
	if call == nil {
		call = new(RpcCall)
	} else {
		server.freeCall = call.next
		*call = RpcCall{}
	}
	server.serial++
	call.Serial = server.serial
	server.callLock.Unlock()
	return call
}

func (server *Server) freeRpcCall(call *RpcCall) {
	server.callLock.Lock()
	call.next = server.freeCall
	server.freeCall = call
	server.callLock.Unlock()
}

func (server *Server) readRequest(codec ServerCodec, src *Mailbox) (rc *RpcCall, keepReading bool, err error) {
	rc, keepReading, err = server.readRequestHeader(codec)
	if err != nil {

		if !keepReading {
			return
		}
		// discard args
		for i := uint16(0); i < rc.Req.Count; i++ {
			codec.ReadRequestBody(nil)
		}
		return
	}

	rc.Server = server
	srcIsValue := false
	if rc.Mtype.SrcType.Kind() == reflect.Ptr {
		rc.Src = reflect.New(rc.Mtype.SrcType.Elem())
	} else {
		rc.Src = reflect.New(rc.Mtype.SrcType)
		srcIsValue = true
	}

	if src == nil {
		// src guaranteed to be a pointer now.
		if err = codec.ReadRequestBody(rc.Src.Interface()); err != nil {
			if err != io.EOF {
				log.LogError("src error", err, ",", rc.Req.ServiceMethod)
			}
			return
		}
		if srcIsValue {
			rc.Src = rc.Src.Elem()
		}
	} else {
		rc.Src = reflect.ValueOf(*src)
	}

	rc.Argvs = make([]reflect.Value, len(rc.Mtype.ArgsType))
	for k, v := range rc.Mtype.ArgsType {
		// Decode the argument value.
		argIsValue := false // if true, need to indirect before calling.
		if v.Kind() == reflect.Ptr {
			rc.Argvs[k] = reflect.New(v.Elem())
		} else {
			rc.Argvs[k] = reflect.New(v)
			argIsValue = true
		}

		if err = codec.ReadRequestBody(rc.Argvs[k].Interface()); err != nil {
			if err != io.EOF &&
				!strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") &&
				!strings.Contains(err.Error(), "use of closed network connection") {
				log.LogError("arg error:", err, ",", rc.Req.ServiceMethod, ", mailbox:", src)
			}
			return
		}
		if argIsValue {
			rc.Argvs[k] = rc.Argvs[k].Elem()
		}
	}
	return
}

func (server *Server) readRequestHeader(codec ServerCodec) (rc *RpcCall, keepReading bool, err error) {
	// Grab the request header.
	rc = server.getRpcCall()
	err = codec.ReadRequestHeader(&rc.Req)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return
		}
		err = errors.New("rpc: server cannot decode request: " + err.Error())
		return
	}

	dot := strings.LastIndex(rc.Req.ServiceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc: service/method request ill-formed: " + rc.Req.ServiceMethod)
		return
	}
	serviceName := rc.Req.ServiceMethod[:dot]
	methodName := rc.Req.ServiceMethod[dot+1:]

	// Look up the request.
	server.mu.RLock()
	rc.Service = server.serviceMap[serviceName]
	server.mu.RUnlock()
	if rc.Service == nil {
		err = errors.New("rpc: can't find service " + rc.Req.ServiceMethod)
		keepReading = true
		return
	}
	rc.Mtype = rc.Service.method[methodName]
	if rc.Mtype == nil {
		err = errors.New("rpc: can't find method " + rc.Req.ServiceMethod)
		keepReading = true
	}
	return
}

func (server *Server) DirectCall(serviceMethod string, src interface{}, argv ...interface{}) error {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		return errors.New("rpc: service/method request ill-formed: " + serviceMethod)
	}
	serviceName := serviceMethod[:dot]
	methodName := serviceMethod[dot+1:]

	// Look up the request.
	server.mu.RLock()
	service := server.serviceMap[serviceName]
	server.mu.RUnlock()
	if service == nil {
		return errors.New("rpc: can't find service " + serviceMethod)
	}
	mtype := service.method[methodName]
	if mtype == nil {
		return errors.New("rpc: can't find method " + serviceMethod)
	}

	mtype.Lock()
	mtype.numCalls++
	mtype.Unlock()

	srcvo := reflect.ValueOf(src)
	if mtype.SrcType.Kind() != srcvo.Kind() {
		return errors.New("rpc: src type not match" + serviceMethod)
	}

	argsvo := make([]reflect.Value, len(argv))
	if len(argv) != len(mtype.ArgsType) {
		return errors.New("rpc args type not match" + serviceMethod)
	}

	for k, v := range argv {
		argsvo[k] = reflect.ValueOf(v)
		if mtype.ArgsType[k].Kind() != argsvo[k].Kind() {
			if len(mtype.ArgsType) == 1 { //try protobuf
				argIsValue := false
				if mtype.ArgsType[k].Kind() == reflect.Ptr {
					argsvo[0] = reflect.New(mtype.ArgsType[k].Elem())
				} else {
					argsvo[0] = reflect.New(mtype.ArgsType[k])
					argIsValue = true
				}

				if buf, ok := argv[0].([]byte); ok {
					if pmsg, ok := argsvo[0].Interface().(pb.Message); ok {
						if err := pb.Unmarshal(buf, pmsg); err == nil {

							if argIsValue {
								argsvo[0] = argsvo[0].Elem()
							}
							break
						}
					}
				}

			}
			return errors.New("rpc argv type not match" + serviceMethod)
		}
	}

	/*
		function := mtype.method.Func
		args := make([]reflect.Value, len(argsvo)+2)
		args[0] = service.rcvr
		args[1] = srcvo
		for k, v := range argsvo {
			args[k+2] = v
		}

		//service.Lock()
		//defer service.Unlock()
		returnValues := function.Call(args)
		errInter := returnValues[0].Interface()
		if errInter != nil {
			return errInter.(error)
		}
	*/

	call := server.getRpcCall()
	call.Req = Request{
		ServiceMethod: serviceMethod,
		Count:         uint16(len(argsvo)),
	}
	call.Service = service
	call.Server = server
	call.Mtype = mtype
	call.Argvs = argsvo
	call.Src = srcvo

	server.ch <- call
	return nil
}

// A ServerCodec implements reading of RPC requests and writing of
// RPC responses for the server side of an RPC session.
// The server calls ReadRequestHeader and ReadRequestBody in pairs
// to read requests from the connection, and it calls WriteResponse to
// write a response back.  The server calls Close when finished with the
// connection. ReadRequestBody may be called with a nil
// argument to force the body of the request to be read and discarded.
type ServerCodec interface {
	ReadRequestHeader(*Request) error
	ReadRequestBody(interface{}) error
	Close() error
	Mailbox() *Mailbox
}
