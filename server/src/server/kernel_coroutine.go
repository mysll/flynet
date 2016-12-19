package server

import "server/libs/log"

type Coroutines struct {
	workid   int64
	submit   func(int, interface{})
	result   int
	reply    interface{}
	complete chan int64
}

func (c *Coroutines) Run(job func(interface{}, interface{}) int, args interface{}, reply interface{}) {
	log.LogMessage("coroutine ", c.workid, " is started")
	ret := job(args, reply)
	c.result = ret
	c.reply = reply
	c.complete <- c.workid
	log.LogMessage("coroutine ", c.workid, " is complete")
}

//取消一个异步过程,目前只能取消回调,不能中止工作函数
func (kernel *Kernel) CancelCoroutine(id int64) {
	if _, has := kernel.coroutinepending[id]; has {
		delete(kernel.coroutinepending, id)
	}
}

func (kernel *Kernel) getCoroutineId() int64 {
	kernel.coroutineworkid++
	return kernel.coroutineworkid
}

//运行一个异步过程,job要完成的工作,submit完成后回调函数
func (kernel *Kernel) StartCoroutine(job func(interface{}, interface{}) int, args interface{}, reply interface{}, submit func(int, interface{})) int64 {

	c := &Coroutines{}
	c.workid = kernel.getCoroutineId()
	c.submit = submit
	c.complete = kernel.coroutinecomplete
	kernel.coroutinepending[c.workid] = c
	go c.Run(job, args, reply)
	return c.workid
}
