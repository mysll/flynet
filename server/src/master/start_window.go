// +build windows

package master

import (
	"fmt"
	"libs/log"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
)

func Start(startapp string, name string, appuid int32, typ string, startargs string) error {
	ferr, err := os.Create(fmt.Sprintf("log/%s_err.log", name))
	if err != nil {
		log.LogError(err)
		return err
	}
	fout, err := os.Create(fmt.Sprintf("log/%s_trace.log", name))
	if err != nil {
		log.LogError(err)
		return err
	}

	cmd := exec.Command(startapp, "-m", fmt.Sprintf("%s:%d", context.Host, context.Port), "-l", context.LocalIP, "-o", context.OuterIP, "-d", strconv.Itoa(int(appuid)), "-t", typ, "-s", startargs)
	cmd.Stdout = fout
	cmd.Stderr = ferr

	err = cmd.Start()
	if err != nil {
		log.LogFatalf(err)
		return err
	}

	log.TraceInfo("master", "app start ", typ, ",", strconv.Itoa(int(appuid)))

	atomic.AddInt32(&Load, 1)
	context.waitGroup.Wrap(func() {
		cmd.Wait()
		ferr.Close()
		fout.Close()
		log.LogMessage(name, " is quit")
		atomic.AddInt32(&Load, -1)
	})
	return nil
}
