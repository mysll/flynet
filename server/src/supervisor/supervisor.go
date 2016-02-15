package supervisor

import (
	"os"
	"os/signal"
	"syscall"
)

type Supervisor struct {
	quit chan struct{}
}

func (s *Supervisor) Start() {
	exitChan := make(chan int)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		exitChan <- 1
	}()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
}
