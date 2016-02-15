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

func (app *app) SendList() {
	applock.RLock()
	defer applock.RUnlock()

	ms := Context.master
	size := len(ms.app)
	if size == 0 {
		return
	}

	rs := make([]share.AddApp, 0, size)
	for _, v := range ms.app {
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
	heartbeat := time.Tick(5 * time.Second)
	for !app.Shutdown {
		select {
		case <-heartbeat: //send heartbeat pkg
			if time.Now().Sub(app.lastcheck).Seconds() > TIMEOUTSECS {
				app.conn.Close()
				return
			}
			data, err := util.CreateMsg(nil, []byte{}, share.M_HEARTBEAT)
			if err != nil {
				log.LogFatalf(err)
			}
			app.conn.Write(data)

		case <-app.exit:
			return
		}
	}
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
	startapp := GetAppName(create.Type)
	err := Start(startapp, create.AppId, create.Type, create.Args)
	if err != nil {
		data, err := share.CreateAppBakMsg(create.Id, create.AppId, err.Error())
		if err != nil {
			log.LogFatalf(err)
		}
		app.conn.Write(data)
		return
	}

	data, err := share.CreateAppBakMsg(create.Id, create.AppId, "ok")
	if err != nil {
		log.LogFatalf(err)
	}
	app.conn.Write(data)
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
