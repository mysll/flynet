package master

import (
	"libs/log"
	"net"
	"share"
	"time"
	"util"
)

const (
	TIMEOUTSECS = 180
)

type app struct {
	typ        string
	id         string
	conn       net.Conn
	host       string
	port       int
	clienthost string
	clientport int
	exit       chan int
	lastcheck  time.Time
	ready      bool
	Shutdown   bool
}

func (app *app) Close() {
	if !app.Shutdown {
		data, err := util.CreateMsg(nil, []byte{}, share.M_SHUTDOWN)
		if err != nil {
			log.LogFatalf(err)
		}

		_, err = app.conn.Write(data)
		if err != nil {
			log.LogInfo(app.id, " closed")
		} else {
			log.LogInfo(app.id, " send shutdown")
		}
		app.Shutdown = true
	}

}

func (app *app) Send(data []byte) error {
	_, err := app.conn.Write(data)
	return err
}

func (app *app) SendList() {
	applock.RLock()
	defer applock.RUnlock()

	size := len(context.app)
	if size == 0 {
		return
	}

	rs := make([]share.AddApp, 0, size)
	for _, v := range context.app {
		rs = append(rs, share.AddApp{v.typ, v.id, v.host, v.port, v.clienthost, v.clientport, v.ready})
	}

	outmsg, err := share.CreateServerListMsg(rs)
	if err != nil {
		log.LogFatalf(err)
	}

	app.conn.Write(outmsg)
}

func (app *app) Check() {
	app.lastcheck = time.Now()
	heartbeat := time.NewTicker(5 * time.Second)
	updatelist := time.NewTicker(time.Minute) //每分钟同步一下app列表
	for !app.Shutdown {
		select {
		case <-heartbeat.C: //send heartbeat pkg
			if time.Now().Sub(app.lastcheck).Seconds() > TIMEOUTSECS {
				app.conn.Close()
				return
			}
			data, err := util.CreateMsg(nil, []byte{}, share.M_HEARTBEAT)
			if err != nil {
				log.LogFatalf(err)
			}
			app.conn.Write(data)
		case <-updatelist.C:
			app.SendList()
		case <-app.exit:
			return
		}
	}
	heartbeat.Stop()
	updatelist.Stop()
}

func (app *app) Loop() {
	app.exit = make(chan int)
	go app.Check()
	buffer := make([]byte, 2048)
	for {
		id, body, err := util.ReadPkg(app.conn, buffer)
		if err != nil {
			break
		}

		if err := app.Handle(id, body); err != nil {
			log.LogError(err)
			break
		}
	}
	RemoveApp(app.id)
}

func (app *app) CreateApp(create share.CreateApp) {
	context.CreateApp(create.ReqId, create.AppId, 0, create.Type, create.Args, create.CallApp)
}

func (app *app) Handle(id uint16, body []byte) error {
	switch id {
	case share.M_HEARTBEAT:
		app.lastcheck = time.Now()
		log.LogFine(app.id, " recv app heartbeat")
	case share.M_READY:
		app.ready = true
		Ready(app)
	case share.M_CREATEAPP:
		create := share.CreateApp{}
		if err := share.DecodeMsg(body, &create); err != nil {
			log.LogError(err)
		}
		go app.CreateApp(create)
	}

	return nil
}
