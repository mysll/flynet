package log

import (
	"fmt"
	"os"
	"runtime/debug"

	l4g "github.com/mysll/log4go"
)

var (
	log l4g.Logger
)

func init() {
	log = make(l4g.Logger)
	l4g.LogCallerDepth = 3
	lw := l4g.NewFormatLogWriter(os.Stdout, "[%D %T] [%L] (%S) %M")
	log.AddFilter("stdout", l4g.DEBUG, lw)
}

func SetLogLevel(filtname string, lvl int) {
	switch lvl {
	case 0:
		log[filtname].Level = l4g.FINEST
	case 1:
		log[filtname].Level = l4g.FINE
	case 2:
		log[filtname].Level = l4g.DEBUG
	case 3:
		log[filtname].Level = l4g.TRACE
	case 4:
		log[filtname].Level = l4g.INFO
	case 5:
		log[filtname].Level = l4g.WARNING
	case 6:
		log[filtname].Level = l4g.ERROR
	case 7:
		log[filtname].Level = l4g.CRITICAL
	}

}

func WriteToFile(filename string) {
	flw := l4g.NewFileLogWriter(filename, false)
	flw.SetFormat("[%D %T] [%L] (%S) %M")
	flw.SetRotate(false)
	flw.SetRotateSize(0)
	flw.SetRotateLines(0)
	flw.SetRotateDaily(false)
	log.AddFilter("file", l4g.DEBUG, flw)
}

func LogFine(args ...interface{}) {
	log.Fine(fmt.Sprint(args...))
}

func LogError(args ...interface{}) {
	log.Error(fmt.Sprint(args...))
}

func LogMessage(args ...interface{}) {
	log.Trace(fmt.Sprint(args...))
}

func LogInfo(args ...interface{}) {
	log.Info(fmt.Sprint(args...))
}

func LogWarning(args ...interface{}) {
	log.Warn(fmt.Sprint(args...))
}

func LogDebug(args ...interface{}) {
	log.Debug(fmt.Sprint(args...))
}

func TraceInfo(module string, args ...interface{}) {
	log.Trace(fmt.Sprintf("%s> %s", module, fmt.Sprint(args...)))
}

func LogFatalf(args ...interface{}) {
	msg := fmt.Sprint(args...)
	log.Critical(msg)
	log.Critical(string(debug.Stack()))
	log.Close()
	panic(msg)
}

func CloseLogger() {
	log.Close()
}
