package server

import (
	"sync/atomic"
)

type Message struct {
	Body   []byte
	bbuf   []byte
	bsize  int
	refcnt int32
}

type msgCacheInfo struct {
	maxbody int
	cache   chan *Message
}

var messageCache = []msgCacheInfo{
	{maxbody: 64, cache: make(chan *Message, 2048)},   // 128K
	{maxbody: 128, cache: make(chan *Message, 1024)},  // 128K
	{maxbody: 256, cache: make(chan *Message, 1024)},  // 256K
	{maxbody: 512, cache: make(chan *Message, 1024)},  // 512K
	{maxbody: 1024, cache: make(chan *Message, 1024)}, // 1 MB
	{maxbody: 2048, cache: make(chan *Message, 512)},  // 1 MB
	{maxbody: 4096, cache: make(chan *Message, 512)},  // 2 MB
	{maxbody: 8192, cache: make(chan *Message, 256)},  // 2 MB
	{maxbody: 16384, cache: make(chan *Message, 128)}, // 2 MB
	{maxbody: 65536, cache: make(chan *Message, 64)},  // 4 MB
}

func (m *Message) Free() {
	var ch chan *Message
	if v := atomic.AddInt32(&m.refcnt, -1); v > 0 {
		return
	}
	for i := range messageCache {
		if m.bsize == messageCache[i].maxbody {
			ch = messageCache[i].cache
			break
		}
	}
	select {
	case ch <- m:
	default:
	}
}

func (m *Message) Dup() *Message {
	atomic.AddInt32(&m.refcnt, 1)
	return m
}

func NewMessage(sz int) *Message {
	var m *Message
	var ch chan *Message
	for i := range messageCache {
		if sz <= messageCache[i].maxbody {
			ch = messageCache[i].cache
			sz = messageCache[i].maxbody
			break
		}
	}
	select {
	case m = <-ch:
	default:
		m = &Message{}
		m.bbuf = make([]byte, 0, sz)
		m.bsize = sz
	}

	m.refcnt = 1
	m.Body = m.bbuf
	return m
}
