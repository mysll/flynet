package supervisor

import (
	"os/exec"
)

type Child struct {
	cmd       exec.Cmd
	startargs string
}

func (c *Child) Start(args string) {
	c.startargs = args
}

func (c *Child) Restart() {

}
