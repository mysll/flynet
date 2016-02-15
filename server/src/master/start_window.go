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

func Start(startapp string, id string, typ string, startargs string) error {
	ferr, err := os.Create(fmt.Sprintf("log/%s_err.log", id))
	if err != nil {
		log.LogError(err)
		return err
	}
	fout, err := os.Create(fmt.Sprintf("log/%s_trace.log", id))
	if err != nil {
		log.LogError(err)
		return err
	}
	atomic.AddInt32(&AppId, 1)
	cmd := exec.Command(startapp, "-m", fmt.Sprintf("%s:%d", Context.master.Host, Context.master.Port), "-d", strconv.Itoa(int(AppId)), "-t", typ, "-s", startargs)
	cmd.Stdout = fout
	cmd.Stderr = ferr

	err = cmd.Start()
	if err != nil {
		log.LogFatalf(err)
		return err
	}

	log.TraceInfo("master", "app start ", typ, ",", strconv.Itoa(int(AppId)))
	Context.master.waitGroup.Wrap(func() {
		cmd.Wait()
		ferr.Close()
		fout.Close()
		log.LogMessage(id, " is quit")
	})
	return nil
}
