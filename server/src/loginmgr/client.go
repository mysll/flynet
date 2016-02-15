package loginmgr

import (
	"errors"
	"libs/log"
	"net"
	"pb/s2c"
	"server"
	"share"
	"sort"
	"time"
	"util"

	proto "github.com/golang/protobuf/proto"
)

func GetLogin() (host string, port int, err error) {
	App.l.Lock()
	defer App.l.Unlock()

	App.serial++
	if App.serial < 0 {
		App.serial = 0
	}

	ls := server.GetAppIdsByType("login")
	sort.Sort(sort.StringSlice(ls))
	if len(ls) > 0 {
		idx := App.serial % len(ls)
		a := server.GetApp(ls[idx])
		return a.ClientHost, a.ClientPort, nil
	}

	return "", 0, errors.New("not found login")
}

type handler struct {
	ignore [1]byte
}

func (hd *handler) Handle(conn net.Conn) {

	log.LogInfo("new client: ", conn.RemoteAddr())

	if !App.MustAppReady {
		log.LogError("must app not ready")
		conn.Close()
		return
	}

	r := &s2c.Rpc{}
	r.Sender = proto.String(App.Id)

	h, p, err := GetLogin()
	if err != nil {
		e := &s2c.Error{}
		e.ErrorNo = proto.Int32(share.ERROR_NOLOGIN)
		b, err := proto.Marshal(e)
		if err != nil {
			conn.Close()
			log.LogFatalf(err)
		}

		r.Servicemethod = proto.String("Login.Error")
		r.Data = b
		data, err := proto.Marshal(r)
		if err != nil {
			log.LogFatalf(err)
			return
		}
		out, _ := util.CreateMsg(nil, data, share.S2C_RPC)
		_, err = conn.Write(out)
		if err != nil {
			conn.Close()
			log.LogError(err)
			return
		}
	} else {
		l := &s2c.Login{}
		l.Host = proto.String(h)
		l.Port = proto.Int32(int32(p))
		b, err := proto.Marshal(l)
		if err != nil {
			conn.Close()
			log.LogFatalf(err)
		}
		log.LogInfo("client choose login:", h, ":", p)

		r.Servicemethod = proto.String("Login.LoginInfo")
		r.Data = b
		data, err := proto.Marshal(r)
		if err != nil {
			log.LogFatalf(err)
			return
		}
		out, _ := util.CreateMsg(nil, data, share.S2C_RPC)
		_, err = conn.Write(out)
		if err != nil {
			conn.Close()
			log.LogError(err)
			return
		}
	}
	conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	conn.Read(hd.ignore[:])
	log.LogMessage("client close: ", conn.RemoteAddr())
	conn.Close()
}
