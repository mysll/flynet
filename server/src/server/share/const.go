package share

import "encoding/gob"

//save type
const (
	SAVETYPE_TIMER = 1 + iota
	SAVETYPE_OFFLINE
)

//load type
const (
	LOAD_DB = iota
	LOAD_CONFIG
	LOAD_ARCHIVE
)

func init() {
	gob.Register(LetterInfo{})
	gob.Register([]*LetterInfo{})
}
