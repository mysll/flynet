// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"libs/log"
	"net"
	"sync"
	"time"
	"util"
)

const (
	MAX_BUF_LEN = 1024 * 16 //16k
)

var (
	callCache = make(chan *Call, 32)
)

type ReplyCB func(*Message)

// ServerError represents an error that has been returned from
// the remote side of the RPC connection.
type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var ErrShutdown = errors.New("connection is shut down")

// Call represents an active RPC.
type Call struct {
	ServiceMethod string   // The name of the service and method to call.
	Args          *Message // The argument to the function (*struct).
	Reply         *Message // The reply from the function (*struct).
	Error         error    // After completion, the error status.
	CB            ReplyCB  //callback function
	noreply       bool
	mb            Mailbox
}

func NewCall() *Call {
	var call *Call
	select {
	case call = <-callCache:
	default:
		call = &Call{}

	}
	return call
}

func (call *Call) Free() {
	call.Args = nil
	call.Reply = nil
	call.CB = nil
	select {
	case callCache <- call:
	default:
	}
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	codec     ClientCodec
	sending   sync.Mutex
	mutex     sync.Mutex // protects following
	seq       uint64
	pending   map[uint64]*Call
	sendqueue chan *Call
	queue     chan *Call
	closing   bool // user has called Close
	shutdown  bool // server has told us to stop
	freeCall  chan *Call
}

// A ClientCodec implements writing of RPC requests and
// reading of RPC responses for the client side of an RPC session.
// The client calls WriteRequest to write a request to the connection
// and calls ReadResponseHeader and ReadResponseBody in pairs
// to read responses.  The client calls Close when finished with the
// connection. ReadResponseBody may be called with a nil
// argument to force the body of the response to be read and then
// discarded.
type ClientCodec interface {
	// WriteRequest must be safe for concurrent use by multiple goroutines.
	WriteRequest(sync.Mutex, uint64, *Call) error
	ReadMessage() (*Message, error)
	GetAddress() string
	Close() error
}

func (client *Client) Go() {
	var err error
	for err == nil {
		select {
		case call := <-client.sendqueue:
			err := client.send(call)
			if err != nil {
				log.LogError("send error:", err)
				log.LogMessage("quit sending loop")
				return
			}
		default:
			if client.shutdown || client.closing {
				log.LogMessage("quit sending loop")
				return
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func (client *Client) send(call *Call) error {
	var seq uint64
	if !call.noreply {
		// Register this call.
		client.mutex.Lock()
		if client.shutdown || client.closing {
			call.Error = ErrShutdown
			return ErrShutdown
		}
		client.seq++
		seq = client.seq
		client.pending[seq] = call
		client.mutex.Unlock()
	}

	// Encode and send the request.
	err := client.codec.WriteRequest(client.sending, seq, call)
	if err != nil {
		call.Error = err
		if !call.noreply {
			client.mutex.Lock()
			delete(client.pending, seq)
			client.mutex.Unlock()
		}

		call.done()
		return err
	}

	if call.noreply {
		call.done()
	}

	return err
}

func (client *Client) input() {
	var err error
	for err == nil {
		message, err := client.codec.ReadMessage()
		if err != nil {
			log.LogError(client.codec.GetAddress(), err)
			break
		}

		ar := util.NewLoadArchiver(message.Header)
		seq, err := ar.ReadUInt64()
		client.mutex.Lock()
		call := client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()

		switch {
		case call == nil:
		default:
			call.Reply = message
			client.queue <- call
		}
	}
	// Terminate pending calls.
	client.mutex.Lock()
	client.shutdown = true
	closing := client.closing
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
	client.mutex.Unlock()

}

func (client *Client) Process() {
	start_time := time.Now()
	for {
		select {
		case call := <-client.queue:
			call.done()
			if time.Now().Sub(start_time) > time.Millisecond*500 { //消息太多,先返回,等下一帧再处理
				return
			}
		default:
			return
		}
	}
}

func (call *Call) done() {

	if call.CB != nil {
		call.CB(call.Reply)
	}

	if call.Args != nil {
		call.Args.Free()
	}

	if call.Reply != nil {
		call.Reply.Free()
	}

	call.Free()
}

// NewClient returns a new Client to handle requests to the
// set of services at the other end of the connection.
// It adds a buffer to the write side of the connection so
// the header and payload are sent as a unit.
func NewClient(conn io.ReadWriteCloser) *Client {
	encBuf := bufio.NewWriter(conn)
	client := &byteClientCodec{conn, encBuf, 1024 * 16}
	return NewClientWithCodec(client)
}

// NewClientWithCodec is like NewClient but uses the specified
// codec to encode requests and decode responses.
func NewClientWithCodec(codec ClientCodec) *Client {
	client := &Client{
		codec:     codec,
		pending:   make(map[uint64]*Call),
		queue:     make(chan *Call, 64),
		freeCall:  make(chan *Call, 32),
		sendqueue: make(chan *Call, 32),
	}
	go client.input()
	go client.Go()
	return client
}

type byteClientCodec struct {
	rwc    io.ReadWriteCloser
	encBuf *bufio.Writer
	maxrx  uint16
}

func (c *byteClientCodec) GetAddress() string {
	return c.rwc.(net.Conn).RemoteAddr().String()
}

func (c *byteClientCodec) WriteRequest(sending sync.Mutex, seq uint64, call *Call) (err error) {
	sending.Lock()
	defer sending.Unlock()
	msg := call.Args
	w := util.NewStoreArchiver(msg.Header)
	w.Write(seq)
	w.Write(call.mb.Uid)
	w.WriteString(call.ServiceMethod)
	msg.Header = msg.Header[:w.Len()]
	count := uint16(len(msg.Header) + len(msg.Body))
	binary.Write(c.encBuf, binary.BigEndian, count)                   //数据大小
	binary.Write(c.encBuf, binary.BigEndian, uint16(len(msg.Header))) //头部大小
	c.encBuf.Write(msg.Header)
	if len(msg.Body) > 0 {
		c.encBuf.Write(msg.Body)
	}
	msg.Header = msg.Header[:0]
	return c.encBuf.Flush()
}

func (c *byteClientCodec) ReadMessage() (*Message, error) {
	return ReadMessage(c.rwc, c.maxrx)
}

func (c *byteClientCodec) Close() error {
	return c.rwc.Close()
}

// Dial connects to an RPC server at the specified network address.
func Dial(network, address string) (*Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

func (client *Client) Close() error {
	client.mutex.Lock()
	if client.closing {
		client.mutex.Unlock()
		return ErrShutdown
	}
	client.closing = true
	client.mutex.Unlock()
	return client.codec.Close()
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) SyncCall(serviceMethod string, src Mailbox, args *Message) error {
	call := NewCall()
	var msg *Message
	if args != nil && len(args.Body) > 0 {
		msg = NewMessage(len(args.Body))
		msg.Body = append(msg.Body, args.Body...)
	} else {
		msg = NewMessage(1)
	}

	call.ServiceMethod = serviceMethod
	call.Args = msg
	call.noreply = true
	call.mb = src
	return client.send(call)
}

func (client *Client) SyncCallBack(serviceMethod string, src Mailbox, args *Message, reply ReplyCB) error {
	call := NewCall()
	var msg *Message
	if args != nil && len(args.Body) > 0 {
		msg = NewMessage(len(args.Body))
		msg.Body = append(msg.Body, args.Body...)
	} else {
		msg = NewMessage(1)
	}
	call.ServiceMethod = serviceMethod
	call.Args = msg
	call.CB = reply
	call.noreply = false
	call.mb = src
	return client.send(call)
}

func (client *Client) CallMessage(serviceMethod string, src Mailbox, args *Message) error {
	call := NewCall()
	var msg *Message
	if args != nil && len(args.Body) > 0 {
		msg = NewMessage(len(args.Body))
		msg.Body = append(msg.Body, args.Body...)
	} else {
		msg = NewMessage(1)
	}
	call.ServiceMethod = serviceMethod
	call.Args = msg
	call.noreply = true
	call.mb = src
	client.sendqueue <- call
	return nil
}

func (client *Client) CallMessageBack(serviceMethod string, src Mailbox, args *Message, reply ReplyCB) error {
	call := NewCall()
	var msg *Message
	if args != nil && len(args.Body) > 0 {
		msg = NewMessage(len(args.Body))
		msg.Body = append(msg.Body, args.Body...)
	} else {
		msg = NewMessage(1)
	}
	call.ServiceMethod = serviceMethod
	call.Args = msg
	call.CB = reply
	call.noreply = false
	call.mb = src
	client.sendqueue <- call
	return nil
}

func (client *Client) Call(serviceMethod string, src Mailbox, args ...interface{}) error {
	var msg *Message
	if len(args) > 0 {
		msg = NewMessage(MAX_BUF_LEN)
		ar := util.NewStoreArchiver(msg.Body)
		for i := 0; i < len(args); i++ {
			err := ar.Write(args[i])
			if err != nil {
				msg.Free()
				return err
			}
		}

		msg.Body = msg.Body[:ar.Len()]
	}

	err := client.CallMessage(serviceMethod, src, msg)
	if msg != nil {
		msg.Free()
	}
	return err
}

func (client *Client) CallBack(serviceMethod string, src Mailbox, reply ReplyCB, args ...interface{}) error {
	var msg *Message
	if len(args) > 0 {
		msg = NewMessage(MAX_BUF_LEN)
		ar := util.NewStoreArchiver(msg.Body)
		for i := 0; i < len(args); i++ {
			err := ar.Write(args[i])
			if err != nil {
				msg.Free()
				return err
			}
		}
		msg.Body = msg.Body[:ar.Len()]
	}

	err := client.CallMessageBack(serviceMethod, src, msg, reply)
	if msg != nil {
		msg.Free()
	}
	return err
}
