package server

import (
	"bytes"
	"io"

	"golang.org/x/net/websocket"
)

type wsconn struct {
	ws      *websocket.Conn
	backend []*bytes.Buffer
	index   int
}

func NewWSConn(ws *websocket.Conn) *wsconn {
	conn := &wsconn{}
	conn.ws = ws
	conn.backend = make([]*bytes.Buffer, 2)
	conn.backend[0] = bytes.NewBuffer(nil)
	conn.backend[1] = bytes.NewBuffer(nil)
	conn.index = 0
	return conn
}

func (conn *wsconn) Read(buf []byte) (n int, err error) {
	r := conn.backend[conn.index]
	min := len(buf)
	for {
		for n < min && err == nil {
			var nn int
			nn, err = r.Read(buf[n:])
			n += nn
		}

		if n >= min {
			err = nil
			conn.index = (conn.index + 1) % len(conn.backend)
			r.WriteTo(conn.backend[conn.index])
			r.Reset()
			return
		} else if err == io.EOF {
			r.Reset()
			var data []byte
			err = websocket.Message.Receive(conn.ws, &data)
			if err != nil {
				return
			}

			r.Write(data)
		}
	}

	return
}

func (conn *wsconn) Write(p []byte) (n int, err error) {
	err = websocket.Message.Send(conn.ws, p)
	n = len(p)
	return
}

func (conn *wsconn) Close() error {
	return conn.ws.Close()
}
