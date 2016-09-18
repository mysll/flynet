package server

import "server/libs/log"

var (
	workid            int64
	coroutinecomplete = make(chan int64, 16)
	coroutinepending  = make(map[int64]*Coroutines, 16)
)

type Coroutines struct {
	workid int64
	submit func(int, interface{})
	result int
	reply  interface{}
}

func (c *Coroutines) Run(job func(interface{}, interface{}) int, args interface{}, reply interface{}) {
	log.LogMessage("coroutine ", c.workid, " is started")
	ret := job(args, reply)
	c.result = ret
	c.reply = reply
	coroutinecomplete <- c.workid
	log.LogMessage("coroutine ", c.workid, " is complete")
}

//取消一个异步过程,目前只能取消回调,不能中止工作函数
func (kernel *Kernel) CancelCoroutine(id int64) {
	if _, has := coroutinepending[id]; has {
		delete(coroutinepending, id)
	}
}

//运行一个异步过程,job要完成的工作,submit完成后回调函数
func (kernel *Kernel) StartCoroutine(job func(interface{}, interface{}) int, args interface{}, reply interface{}, submit func(int, interface{})) int64 {
	workid++
	c := &Coroutines{}
	c.workid = workid
	c.submit = submit
	coroutinepending[workid] = c
	go c.Run(job, args, reply)
	return workid
}
