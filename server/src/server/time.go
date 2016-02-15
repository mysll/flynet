package server

import (
	"time"
)

type Time struct {
	FrameCount     int
	RunTime        time.Duration
	LastUpdateTime time.Time
	LastBeatTime   time.Time
	LastScanTime   time.Time
	LastFreshTime  time.Time
}
