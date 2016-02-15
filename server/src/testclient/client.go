package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"libs/log"
	"net"
	"os"
	"os/signal"
	"pb"
	"pb/c2s"
	"pb/s2c"
	"runtime"
	"share"
	"syscall"
	"time"
	"util"
)

var (
	count = flag.Int("n", 10, "threads")
)

func HandleErr(id uint16, msg []byte) bool {
	if id == share.S2C_ERROR {
		err := &s2c.Error{}
		proto.Unmarshal(msg, err)
		log.LogDebug("err:", err.GetErrorNo())
		return true
	}

	return false
}

func ParseMsg(id uint16, msg []byte) (sender string, servicename string, body []byte, err error) {
	if id == share.S2C_RPC {
		rpc := &s2c.Rpc{}
		if err = proto.Unmarshal(msg, rpc); err != nil {
			return
		}

		body = rpc.Data
		sender = rpc.GetSender()
		servicename = rpc.GetServicemethod()
		return

	}

	err = errors.New(fmt.Sprintf("msg unknown%d", id))
	return
}

func RpcCall(conn net.Conn, node string, servicemethod string, data proto.Message) error {
	rpc := &c2s.Rpc{}
	rpc.Node = proto.String(node)
	rpc.Servicemethod = proto.String(servicemethod)
	var err error
	if rpc.Data, err = pb.Encode(data); err == nil {
		pb.SendMsg(conn, share.C2S_RPC, rpc)
		//log.LogDebug("rpc call: ", node, ":", servicemethod)
		return nil
	}

	log.LogDebug(err)
	return err
}

func ConnectBase(h string, p int32, k int32, uid int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
	if err != nil {
		log.LogDebug(err)
		return
	}
	defer conn.Close()
	//log.LogDebug("connect to base", conn.RemoteAddr().String(), " use key:", k, ", uid:", uid)

	eb := &c2s.Enterbase{}
	user := fmt.Sprintf("test%d", uid)
	eb.User = proto.String(user)
	eb.Key = proto.Int32(k)
	RpcCall(conn, ".", "Account.Login", eb)

	buff := make([]byte, 1024)
	id, msg, _ := util.ReadPkg(conn, buff)

	if sender, method, body, err := ParseMsg(id, msg); err == nil {
		log.LogDebug("base:", sender, method, uid)
		role := &s2c.RoleInfo{}
		proto.Unmarshal(body, role)

		if len(role.UserInfo) == 0 {
			cu := &c2s.Create{}
			cu.Name = proto.String(fmt.Sprintf("test%d", uid))
			cu.Index = proto.Int32(0)
			cu.Sex = proto.Int32(1)
			RpcCall(conn, ".", "Account.CreatePlayer", cu)

			id, msg, _ = util.ReadPkg(conn, buff)
			if sender, method, body, err = ParseMsg(id, msg); err == nil {
				if method == "roleinfo" {
					role := &s2c.RoleInfo{}
					proto.Unmarshal(body, role)
					log.LogDebug(role)
				}

			}

		}

		su := &c2s.Selectuser{}
		su.Rolename = proto.String(fmt.Sprintf("test%d", uid))
		su.Roleindex = proto.Int32(0)
		RpcCall(conn, ".", "Account.SelectUser", su)
		id, msg, _ = util.ReadPkg(conn, buff)
		if sender, method, body, err = ParseMsg(id, msg); err == nil {
			log.LogDebug("select bak:", method, body)
		}
		log.LogDebug(role)
	} else {
		log.LogDebug(err)
	}

	log.LogDebug("down,", uid)
	time.Sleep(time.Second * 60)
}

func ConnectLogin(h string, p int32, uid int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
	if err != nil {
		log.LogDebug(err)
		return
	}
	defer conn.Close()
	//log.LogDebug("connect to server", conn.RemoteAddr().String())

	login := &c2s.Loginuser{}
	user := fmt.Sprintf("test%d", uid)
	pwd := "123"
	login.User = &user
	login.Password = &pwd

	RpcCall(conn, ".", "Account.Login", login)

	buff := make([]byte, 1024)
	id, msg, _ := util.ReadPkg(conn, buff)
	if HandleErr(id, msg) {
		return
	}

	if _, method, body, err := ParseMsg(id, msg); err == nil {
		//log.LogDebug(sender, method, body)
		if method == "loginsucceed" {
			l := &s2c.Loginsucceed{}
			if err := proto.Unmarshal(body, l); err == nil {
				//log.LogDebug("logininfo:", l, uid)
				ConnectBase(l.GetHost(), l.GetPort(), l.GetKey(), uid)
			}
		}

	}
	/*
		succ := &s2c.Loginsucceed{}
		proto.Unmarshal(msg, succ)
		log.LogMessage(succ)
		conn.Close()
		ConnectBase(*succ.Host, *succ.Port, *succ.Key, uid)
	*/
}

func client(uid int) {
	conn, err := net.Dial("tcp", "127.0.0.1:5391")
	if err != nil {
		log.LogDebug(err)
		return
	}
	defer conn.Close()
	buffer := make([]byte, 1024)
	id, body, e := util.ReadPkg(conn, buffer)
	if e != nil {
		log.LogDebug(e)
		return
	}

	switch id {
	case share.S2C_LOGININFO:
		login := &s2c.Login{}
		proto.Unmarshal(body, login)
		ConnectLogin(login.GetHost(), login.GetPort(), uid)
	}
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < *count; i++ {
		go client(i)
		time.Sleep(time.Millisecond * 10)
	}

	exitChan := make(chan int)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		exitChan <- 1
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-exitChan:
		//case <-time.After(time.Second * 4):
	}

}
