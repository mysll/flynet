package util

import (
	"testing"
)

func TestNewRingBuffer(t *testing.T) {
	rb := NewRingBuffer(8)
	t.Log("begin:", rb, rb.msgs.Len())
	rb.Push([]byte{1, 2, 3})
	t.Log("push:", rb, rb.msgs.Len())
	if rb.end != 3 {
		t.Fatal("index failed, must 0, 3", rb.buffer)
	}
	rb.Push([]byte{2, 2, 2})
	t.Log("push:", rb, rb.msgs.Len())
	data, err := rb.Pop()
	if err != nil || len(data) != 3 || data[0] != 1 || data[1] != 2 || data[2] != 3 {
		t.Fatal("data error, must 1,2,3, now:", len(data), data)
	}
	t.Log("pop:", rb, rb.msgs.Len())
	data, err = rb.Pop()
	t.Log("pop:", rb, rb.msgs.Len())
	if rb.first != 6 {
		t.Fatal("index failed, must 3")
	}

	rb.Push([]byte{4, 4, 4})
	t.Log("push", rb, rb.msgs.Len())
	if rb.end != 3 {
		t.Fatal("index failed, must 6", rb.end)
	}
	rb.Push([]byte{4, 5, 6})
	t.Log("push", rb, rb.msgs.Len())
	if rb.end != 6 {
		t.Fatal("index failed, must 3, now:", rb.end, rb.first)
	}
	ret := rb.Push([]byte{7, 8, 9})
	if ret != nil {
		t.Fatal("push failed, must false")
	}

	t.Log("push", rb, rb.msgs.Len())
	data, err = rb.Pop()
	t.Log("pop", rb, rb.msgs.Len())
	if err != nil || len(data) != 3 || data[0] != 4 || data[1] != 4 || data[2] != 4 {
		t.Fatal("data error, now:", data, err)
	}
	if rb.first != 3 {
		t.Fatal("index failed")
	}
	data, err = rb.Pop()
	t.Log("pop", rb, rb.msgs.Len())
	if err != nil || len(data) != 3 || data[0] != 4 || data[1] != 5 || data[2] != 6 {
		t.Fatal("data error, now:", data)
	}
	if rb.first != 6 {
		t.Fatal("index failed")
	}
}
