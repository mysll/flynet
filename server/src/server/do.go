package server

import (
	"runtime"
	"server/libs/common/event"
	"server/libs/log"
	"server/libs/rpc"
	"time"
)

var (
	busy       bool
	warninglvl = 10 * time.Millisecond
	BeatTime   = time.Millisecond * 333
	Updatetime = time.Millisecond * 200
	Freshtime  = time.Second
)

//进程rpc处理
func RpcProcess(ch chan *rpc.RpcCall) {
	var start_time time.Time
	var delay time.Duration
	for {
		select {
		case call := <-ch:
			if call.IsThreadWork() {
				busy = true
			} else {
				log.LogDebug(call.GetSrc(), " rpc call:", call.GetMethod())
				start_time = time.Now()
				err := call.Call()
				if err != nil {
					log.LogError(err)
				}
				delay = time.Now().Sub(start_time)
				if delay > warninglvl {
					log.LogWarning("rpc call ", call.GetMethod(), " delay:", delay.Nanoseconds()/1000000, "ms")
				}
				err = call.Done()
				if err != nil {
					log.LogError(err)
				}
				call.Free()
				busy = true
			}

		default:
			return
		}
	}
}

//rpc 回调处理
func RpcResponseProcess() {
	applock.RLock()
	for _, app := range RemoteApps {
		if app.RpcClient != nil {
			app.RpcClient.Process()
		}
	}
	applock.RUnlock()
}

//事件执行
func DoEvent(e *event.Event) {

	switch e.Typ {
	case NEWUSERCONN:
		core.apper.OnClientConnected(e.Args["id"].(int64))
		return
	case LOSTUSERCONN:
		core.apper.OnClientLost(e.Args["id"].(int64))
		core.clientList.Remove(e.Args["id"].(int64))
		return
	case NEWAPPREADY:
		core.apper.OnReady(e.Args["id"].(string))
		return
	case APPLOST:
		core.apper.OnLost(e.Args["id"].(string))
		return
	}

	core.apper.OnEvent(e.Typ, e.Args)

}

//事件遍历
func EventProcess(e *event.EventList) {
	var start_time time.Time
	var delay time.Duration
	for {
		evt := e.Pop()
		if evt == nil {
			break
		}
		start_time = time.Now()
		DoEvent(evt)
		delay = time.Now().Sub(start_time)
		if delay > warninglvl {
			log.LogWarning("DoEvent delay:", delay.Nanoseconds()/1000000, "ms")
		}
		busy = true
		e.FreeEvent(evt)
	}

}

//主循环，整个服务器的工作循环，每次循环处理顺序：
//1、事件处理
//2、远程调用处理
//3、固定时间间隔的逻辑处理
func Run(s *Server) {
	s.apper.OnStart()
	now := time.Now()
	s.Time.FrameCount = 0
	s.Time.LastBeatTime = now
	s.Time.LastUpdateTime = now
	s.Time.LastScanTime = now
	s.Time.LastFreshTime = now
	s.Time.RunTime = 0
	begin := now
	for !s.quit {
		now := time.Now()
		s.Time.RunTime = now.Sub(begin)
		s.Time.FrameCount++
		busy = false

		EventProcess(s.Emitter)
		RpcProcess(s.rpcCh)
		RpcResponseProcess()

		if now.Sub(s.Time.LastBeatTime) >= BeatTime {
			//处理心跳
			timer.Pump()
			s.apper.OnBeatRun()
			//场景心跳
			sceneBeat.Pump()

			if dispatcherList.Len() > 0 {
				for e := dispatcherList.Front(); e != nil; e = e.Next() {
					e.Value.(Dispatcher).OnBeatRun()
				}
			}
			s.Time.LastBeatTime = now
		}

		if now.Sub(s.Time.LastUpdateTime) >= Updatetime {
			//准备更新回调
			s.apper.OnBeginUpdate()
			if dispatcherList.Len() > 0 {
				for e := dispatcherList.Front(); e != nil; e = e.Next() {
					e.Value.(Dispatcher).OnBeginUpdate()
				}
			}
			//更新回调
			s.apper.OnUpdate()
			//更新kernel调度器
			s.OnUpdate()
			if dispatcherList.Len() > 0 {
				for e := dispatcherList.Front(); e != nil; e = e.Next() {
					e.Value.(Dispatcher).OnUpdate()
				}
			}

			//更新完成后回调
			s.apper.OnLastUpdate()
			if dispatcherList.Len() > 0 {
				for e := dispatcherList.Front(); e != nil; e = e.Next() {
					e.Value.(Dispatcher).OnLastUpdate()
				}
			}
			s.Time.LastUpdateTime = now
		}

		if now.Sub(s.Time.LastFreshTime) >= Freshtime {
			s.apper.OnFlush()
			if dispatcherList.Len() > 0 {
				for e := dispatcherList.Front(); e != nil; e = e.Next() {
					e.Value.(Dispatcher).OnFlush()
				}
			}
			s.Time.LastFreshTime = now
			s.clientList.Check()
		}

		//删除对象
		s.ObjectFactory.ClearDelete()
		s.apper.OnFrame()
		if dispatcherList.Len() > 0 {
			for e := dispatcherList.Front(); e != nil; e = e.Next() {
				e.Value.(Dispatcher).OnFrame()
			}
		}
		s.s2chelper.flush() //发送缓存数据
		runtime.Gosched()
		if !busy {
			time.Sleep(time.Millisecond * 1)
		}
	}
}
