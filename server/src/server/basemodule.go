package server

import (
	"errors"
	"server/libs/log"
)

var (
	modules = make(map[string]Moduler)
)

type BaseModule struct {
}

func (m *BaseModule) Init() {

}

func (m *BaseModule) Load() error {
	return nil
}

func (m *BaseModule) UnLoad() {
}

func (m *BaseModule) OnCommand(module string, args interface{}) error {
	return nil
}

func (m *BaseModule) GetCore() *Server {
	return core
}

type Moduler interface {
	Init()
	Load() error
	UnLoad()
	OnCommand(module string, args interface{}) error
}

//注册模块
func RegisterModule(name string, module Moduler) {
	if module == nil {
		log.LogFatalf(name, " module is nil")
	}

	if _, dup := modules[name]; dup {
		log.LogFatalf(name, " is register")
	}

	modules[name] = module
	log.LogMessage("register module:", name)
}

//通过模块名获取模块对象
func FindModule(name string) Moduler {
	if module, exist := modules[name]; exist {
		return module
	}
	return nil
}

//向模块发送消息
func SendToModule(src string, dest string, args interface{}) error {
	if module, exist := modules[dest]; exist {
		return module.OnCommand(src, args)
	}

	return errors.New("module not found")
}

func initAllModules() {
	for name, module := range modules {
		module.Init()
		log.LogMessage(name, " has inited")
	}
}

func loadAllModules() {
	for name, module := range modules {
		err := module.Load()
		if err != nil {
			log.LogFatalf("load module failed,", err)
		}

		log.LogMessage(name, " has loaded")
	}
}

func unloadAllModules() {
	for name, module := range modules {
		module.UnLoad()
		log.LogMessage(name, " has unloaded")
	}
}
