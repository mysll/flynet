package util

import (
	"errors"
	"libs/log"
	"reflect"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type FSM struct {
	l          sync.Mutex
	StopReason string
	rcvr       reflect.Value // receiver of methods for the service
	typ        reflect.Type  // type of the receiver
	method     map[string]reflect.Method
	stopmethod *reflect.Method
	quit       chan int
	state      string
	stopped    bool
	delaych    chan int
	isdelay    bool
}

type Event struct {
	event   string
	param   interface{}
	timeout int
}

func (fsm *FSM) IsStopped() bool {
	return fsm.stopped
}

func (fsm *FSM) SendBreakEvent(event string, param interface{}) {
	fsm.Break()
	fsm.SendEvent(event, param)
}

func (fsm *FSM) SendEvent(event string, param interface{}) {
	fsm.l.Lock()
	defer fsm.l.Unlock()
	if fsm.stopped {
		return
	}
	fsm.CallState(Event{event, param, 0})
}

func (fsm *FSM) Break() {
	fsm.l.Lock()
	if fsm.stopped || !fsm.isdelay {
		fsm.l.Unlock()
		return
	}
	fsm.l.Unlock()
	fsm.delaych <- 1
}

func (fsm *FSM) Init(start string) error {
	fsm.stopped = false
	if _, ok := fsm.method[start]; !ok {
		fsm.stopped = true
		log.LogError("fsm init state not fount:", start)
		return errors.New("not found state")
	}

	fsm.state = start
	return nil
}

func (fsm *FSM) CallState(e Event) {
	if function, ok := fsm.method[fsm.state]; ok {
		returnValues := function.Func.Call([]reflect.Value{fsm.rcvr, reflect.ValueOf(e.event), reflect.ValueOf(e.param), reflect.ValueOf(e.timeout)})
		nextstate := returnValues[0].String()
		timeout := returnValues[1].Int()
		errInter := returnValues[2].Interface()
		errmsg := ""

		if errInter != nil {
			errmsg = errInter.(error).Error()
		}

		if nextstate == "stop" {
			fsm.Stop(errmsg)
			return
		}

		if errmsg != "" {
			log.LogError(errmsg)
		}

		fsm.state = nextstate

		if timeout > 0 {
			go fsm.DelayCall(timeout)
		}
	}
}

func (fsm *FSM) DelayCall(timeout int64) {
	defer recover()
	fsm.isdelay = true
	tick := time.NewTicker(time.Duration(timeout) * time.Millisecond)
	select {
	case <-tick.C:
		fsm.l.Lock()
		defer fsm.l.Unlock()
		if fsm.stopped {
			break
		}
		fsm.CallState(Event{"timeout", 0, int(timeout)})
		break
	case <-fsm.delaych:
		break
	}
	tick.Stop()
	fsm.isdelay = false
}

func (fsm *FSM) Stop(message string) {
	fsm.StopReason = message
	if fsm.stopmethod != nil {
		fsm.stopmethod.Func.Call([]reflect.Value{fsm.rcvr, reflect.ValueOf(message)})
	}
	fsm.stopped = true
}

func (fsm *FSM) Close() {
	fsm.Break()
	fsm.l.Lock()
	defer fsm.l.Unlock()
	if fsm.stopped {
		return
	}
	fsm.stopped = true
}

func NewFSM(fsm interface{}) *FSM {
	f := &FSM{typ: reflect.TypeOf(fsm), rcvr: reflect.ValueOf(fsm), quit: make(chan int), delaych: make(chan int)}
	f.method, f.stopmethod = suitableMethods(f.typ, true)
	return f
}

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func suitableMethods(typ reflect.Type, reportErr bool) (methods map[string]reflect.Method, stop *reflect.Method) {
	methods = make(map[string]reflect.Method)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name

		if !isExported(mname) {
			continue
		}

		if mname == "Stopped" && mtype.NumIn() == 2 && mtype.In(1).Kind() == reflect.String {
			stop = &method
			continue
		}

		// Method needs four ins: receiver, string, interface{}, int.
		if mtype.NumIn() != 4 {
			if reportErr {
				log.LogError("method ", mname, " has wrong number of ins:", mtype.NumIn())
			}
			continue
		}

		// First arg must be a string.
		if mtype.In(1).Kind() != reflect.String {
			if reportErr {
				log.LogError("method ", mname, " arg1 type not a string:", mtype.In(1).Kind())
			}
			continue
		}
		// Second arg must be a interface.
		if mtype.In(2).Kind() != reflect.Interface {
			if reportErr {
				log.LogError("method ", mname, " arg2 type not a interface:", mtype.In(2).Kind())
			}
			continue
		}

		// Third arg must be a int.
		if mtype.In(3).Kind() != reflect.Int {
			if reportErr {
				log.LogError("method ", mname, " arg3 type not a int:", mtype.In(3).Kind())
			}
			continue
		}

		// Method needs three out.
		if mtype.NumOut() != 3 {
			if reportErr {
				log.LogError("method ", mname, " has wrong number of outs:", mtype.NumOut())
			}
			continue
		}

		if mtype.Out(0).Kind() != reflect.String {
			if reportErr {
				log.LogError("method ", mname, " out1 type not a string:", mtype.Out(0).Kind())
			}
			continue
		}
		if mtype.Out(1).Kind() != reflect.Int {
			if reportErr {
				log.LogError("method ", mname, " out1 type not a int:", mtype.Out(1).Kind())
			}
			continue
		}
		if mtype.Out(2) != typeOfError {
			if reportErr {
				log.LogError("method ", mname, " out3 type not a error:", mtype.In(2).Kind())
			}
			continue
		}
		methods[mname] = method
	}
	return
}
