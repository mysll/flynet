package server

import (
	"fmt"
	. "server/data/datatype"
	"testing"
	"time"
)

func cbk(obj ObjectID, t time.Duration, args interface{}) {
	sb := args.(*SceneBeat)
	fmt.Println("heart beat ", t)
	if sb.Find(obj, "test1") {
		sb.Remove(obj, "test1")
	}
}

func TestSceneBeat(t *testing.T) {
	sb := NewSceneBeat()
	t.Log(sb)
	sb.Add(ObjectID{1, 1}, "test1", time.Millisecond*500, 3, cbk, sb)
	sb.Add(ObjectID{1, 1}, "test2", time.Millisecond*300, 3, cbk, sb)
	sb.Add(ObjectID{1, 1}, "test3", time.Millisecond*100, 1, cbk, sb)
	sb.Add(ObjectID{1, 1}, "test1", time.Millisecond*100, 1, cbk, sb)
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 300)
		t.Log(sb)
		sb.Pump()
		t.Log(sb)
	}
}
