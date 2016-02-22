package db

import (
	"db/mongodb"
	"db/mysqldb"
	"libs/log"
	"server"
	"time"
)

var (
	App *DBApp
)

type DBer interface {
	InitDB(db string, source string, threads int, entity string, role string, limit int) error
	Close()
	KeepAlive()
}

type DBApp struct {
	*server.Server
	roleEntity string
	dbname     string
	ds         string
	entity     string
	typ        string
	shutdown   int
	DBType     string
	db         DBer
	pools      int
	rolelimit  int
	keepid     server.TimerID
}

func parseArgs() {
	args := App.StartArgs

	if typ, ok := args.CheckGet("type"); ok {
		v, err := typ.String()
		if err != nil {
			log.LogFatalf(err)
		}
		App.typ = v
	} else {
		log.LogFatalf("type not defined")
	}

	if ds, ok := args.CheckGet("datasource"); ok {
		v, err := ds.String()
		if err != nil {
			log.LogFatalf(err)
		}
		App.ds = v
	} else {
		log.LogFatalf("datasource not defined")
	}

	if role, ok := args.CheckGet("role"); ok {
		v, err := role.String()
		if err != nil {
			log.LogFatalf(err)
		}
		App.roleEntity = v
	}

	if dname, ok := args.CheckGet("db"); ok {
		v, err := dname.String()
		if err != nil {
			log.LogFatalf(err)
		}
		App.dbname = v
	} else {
		log.LogFatalf("db not defined")
	}

	if entity, ok := args.CheckGet("entity"); ok {
		v, err := entity.String()
		if err != nil {
			log.LogFatalf(err)
		}
		App.entity = v
	} else {
		log.LogFatalf("entity not defined")
	}

	App.pools = 1
	if p, ok := args.CheckGet("pools"); ok {
		v, err := p.Int()
		if err != nil {
			log.LogFatalf(err)
		}
		if v > 0 {
			App.pools = v
		}

	}

	App.rolelimit = 1
	if l, ok := args.CheckGet("rolelimit"); ok {
		v, err := l.Int()
		if err != nil {
			log.LogFatalf(err)
		}
		if v > 1 {
			App.rolelimit = v
		}
	}
}

func (d *DBApp) OnPrepare() bool {
	parseArgs()
	switch App.typ {
	case "mysql":
		App.db = mysqldb.NewMysqlDB()
	case "mongodb":
		App.db = mongodb.NewMongoDB()
	default:
		panic("unknown database")
	}

	if err := App.db.InitDB(App.dbname, App.ds, App.pools, App.entity, App.roleEntity, App.rolelimit); err != nil {
		panic(err)
	}

	return true
}

func (d *DBApp) OnStart() {
	d.keepid = App.AddTimer(time.Minute, -1, d.DbKeepAlive, nil)
}

func (d *DBApp) DbKeepAlive(intervalid server.TimerID, count int32, args interface{}) {
	App.db.KeepAlive()
}

func (d *DBApp) OnShutdown() bool {
	d.shutdown = 1
	return false
}

func (d *DBApp) OnLost(app string) {
	if server.GetAppCount() == 0 && d.shutdown == 1 {
		d.Exit()
	}
}

func (d *DBApp) Exit() {
	d.shutdown = 2
	App.CancelTimer(d.keepid)
	App.db.Close()
	App.Shutdown()
}

func GetAllHandler() map[string]interface{} {
	return server.GetAllHandler()
}

func init() {
	App = &DBApp{}

}
