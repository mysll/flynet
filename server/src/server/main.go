package server

import (
	"encoding/gob"
	"server/libs/common/event"
	"server/util"
)

var (
	core *Server
)

func NewServer(app Apper, id int32) *Server {
	s := &Server{}
	core = s
	s.AppId = id
	s.WaitGroup = &util.WaitGroupWrapper{}
	s.exitChannel = make(chan struct{})
	s.shutdown = make(chan struct{})
	s.Eventer = NewEvent()
	s.clientList = NewClientList()
	s.apper = app
	s.Emitter = event.NewEventList()
	s.ObjectFactory = NewFactory()
	s.Kernel = NewKernel(s.ObjectFactory)
	s.channel = make(map[string]*Channel, 32)

	s.s2chelper = NewS2CHelper()
	s.c2shelper = &C2SHelper{}
	s.teleport = &TeleportHelper{}
	s.globalHelper = NewGlobalDataHelper()
	s.modules = make(map[string]Moduler)
	RegisterRemote("S2CHelper", s.s2chelper)
	RegisterRemote("Teleport", s.teleport)
	RegisterRemote("GlobalHelper", s.globalHelper)
	RegisterHandler("C2SHelper", s.c2shelper)
	return s
}

func init() {
	gob.Register([]interface{}{})
}
