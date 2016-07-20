package util

import (
	"container/list"
	"errors"
	"sync"
)

var (
	ERRRINGBUFFULL = errors.New("ringbuffer full")
)

type Message struct {
	begin int
	size  int
}

type RingBuffer struct {
	buffer []byte
	l      sync.RWMutex
	first  int
	end    int
	size   int
	msgs   *list.List
}

func (r *RingBuffer) require(lens int) int {
	if lens > r.size {
		return -1
	}

	if r.msgs.Len() > 0 && r.first == r.end {
		return -1
	}

	if r.first <= r.end {
		if r.end+lens > r.size {
			r.end = 0
			return r.require(lens)
		}
	} else {
		if r.end+lens > r.first {
			return -1
		}
	}

	return r.end
}

func (r *RingBuffer) Count() int {
	r.l.RLock()
	defer r.l.RUnlock()
	return r.msgs.Len()
}

func (r *RingBuffer) Empty() bool {
	return r.msgs.Len() == 0
}

func (r *RingBuffer) Push(data []byte) error {
	r.l.Lock()
	defer r.l.Unlock()
	datasize := len(data)
	if idx := r.require(datasize); idx != -1 {
		r.end += datasize
		copy(r.buffer[idx:idx+datasize], data)
		r.msgs.PushBack(Message{begin: idx, size: datasize})
		return nil
	}

	return ERRRINGBUFFULL
}

func (r *RingBuffer) Pop() ([]byte, error) {
	r.l.Lock()
	defer r.l.Unlock()
	if r.msgs.Len() == 0 {
		return nil, errors.New("no data")
	}

	e := r.msgs.Front()
	m := e.Value.(Message)
	r.msgs.Remove(e)

	r.first = m.begin + m.size
	return r.buffer[m.begin : m.begin+m.size], nil
}

func (r *RingBuffer) Reset() {
	r.l.Lock()
	defer r.l.Unlock()
	r.first = 0
	r.end = 0
	r.msgs = list.New()
}

func NewRingBuffer(capacity int) *RingBuffer {
	rb := &RingBuffer{}
	rb.buffer = make([]byte, capacity, capacity)
	rb.first = 0
	rb.end = 0
	rb.size = capacity
	rb.msgs = list.New()
	return rb
}
