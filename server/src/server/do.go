package server

import (
	"libs/common/event"
	"libs/log"
	"libs/rpc"
	"runtime"
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
		case caller := <-ch:
			if caller.Service.Push(caller) {
				busy = true
			} else {
				log.LogDebug(caller.Src.Interface(), " rpc call:", caller.Req.ServiceMethod)
				start_time = time.Now()
				err := caller.Call()
				delay = time.Now().Sub(start_time)
				if delay > warninglvl {
					log.LogWarning("rpc call ", caller.Req.ServiceMethod, " delay:", delay.Nanoseconds()/1000000, "ms")
				}
				caller.Free()
				busy = true
				if err != nil {
					log.LogError("rpc error:", err)
				}
			}

		default:
			return
		}
	}

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

		if now.Sub(s.Time.LastBeatTime) >= BeatTime {
			//处理心跳
			s.timer.Pump()
			s.apper.OnBeatRun()
			//场景心跳
			s.sceneBeat.Pump()
			s.Time.LastBeatTime = now
		}

		if now.Sub(s.Time.LastUpdateTime) >= Updatetime {
			//准备更新回调
			s.apper.OnBeginUpdate()
			//更新回调
			s.apper.OnUpdate()
			//更新kernel调度器
			s.OnUpdate()
			//更新完成后回调
			s.apper.OnLastUpdate()
			s.Time.LastUpdateTime = now
		}

		if now.Sub(s.Time.LastFreshTime) >= Freshtime {
			s.apper.OnFlush()
			s.Time.LastFreshTime = now
			s.clientList.Check()
		}

		//删除对象
		s.ObjectFactory.ClearDelete()
		s.apper.OnFrame()
		s.s2chelper.flush() //发送缓存数据
		runtime.Gosched()
		if !busy {
			time.Sleep(time.Millisecond * 1)
		}
	}
}
