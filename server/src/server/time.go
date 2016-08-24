package server

import (
	"time"
)

type TimeInfo struct {
	FrameCount     int
	RunTime        time.Duration //总运行时间
	StartTime      time.Time     //开始运行时间
	DeltaTime      time.Duration //update间隔照章
	LastUpdateTime time.Time     //最后一次更新时间
	LastBeatTime   time.Time     //最后一次心跳时间
	LastScanTime   time.Time     //最后一次扫描时间
	LastFreshTime  time.Time     //最后一次刷新时间
}
