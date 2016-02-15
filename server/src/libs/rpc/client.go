// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package rpc

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"libs/log"
	"net"
	"sync"
)

// ServerError represents an error that has been returned from
// the remote side of the RPC connection.
type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var ErrShutdown = errors.New("connection is shut down")

// Call represents an active RPC.
type Call struct {
	ServiceMethod string        // The name of the service and method to call.
	Src           Mailbox       // The rpc source
	Args          []interface{} // The argument to the function (*struct).
	Done          chan *Call    // Strobes when call is complete.
	Error         error
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	codec   ClientCodec
	sending sync.Mutex

	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

type ClientCodec interface {
	// WriteRequest must be safe for concurrent use by multiple goroutines.
	WriteRequest(string, Mailbox, []interface{}) error
	Close() error
}

func (client *Client) send(call *Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	err := client.codec.WriteRequest(call.ServiceMethod, call.Src, call.Args)
	call.Error = err
	call.done()
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		log.LogWarning("rpc: discarding Call reply due to insufficient Done chan capacity")

	}
}

// NewClient returns a new Client to handle requests to the
// set of services at the other end of the connection.
// It adds a buffer to the write side of the connection so
// the header and payload are sent as a unit.
func NewClient(conn io.ReadWriteCloser) *Client {
	encBuf := bufio.NewWriter(conn)
	client := &gobClientCodec{conn, gob.NewEncoder(encBuf), encBuf}
	return NewClientWithCodec(client)
}

// NewClientWithCodec is like NewClient but uses the specified
// codec to encode requests and decode responses.
func NewClientWithCodec(codec ClientCodec) *Client {
	client := &Client{
		codec: codec,
	}
	return client
}

type gobClientCodec struct {
	rwc    io.ReadWriteCloser
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func (c *gobClientCodec) WriteRequest(servicemethod string, src Mailbox, body []interface{}) (err error) {
	req := Request{servicemethod, uint16(len(body) + 1)}
	if err = c.enc.Encode(req); err != nil {
		return
	}

	if err = c.enc.Encode(src); err != nil {
		return
	}

	for _, v := range body {
		if err = c.enc.Encode(v); err != nil {
			return
		}
	}

	return c.encBuf.Flush()
}

func (c *gobClientCodec) Close() error {
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
	if client.closing {
		return ErrShutdown
	}
	client.closing = true

	return client.codec.Close()
}

// Go invokes the function asynchronously.  It returns the Call structure representing
// the invocation.  The done channel will signal when the call is complete by returning
// the same Call object.  If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (client *Client) Go(serviceMethod string, src Mailbox, done chan *Call, args ...interface{}) *Call {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Src = src
	call.Args = append(call.Args, args...)
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel.  If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.LogFatalf("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	client.send(call)
	return call
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(serviceMethod string, src Mailbox, args ...interface{}) error {
	call := <-client.Go(serviceMethod, src, make(chan *Call, 1), args...).Done
	return call.Error
}
