package main

import (
	"flag"
	"fmt"
	"io"
	"libs/log"
	"net"
	"pb/c2s"
	"pb/s2c"
	"share"
	"time"
	"util"

	"github.com/golang/protobuf/proto"
)

var (
	count = flag.Int("n", 500, "threads")
)

func ParseProto(id uint16, buf []byte, pb proto.Message) error {
	switch id {
	case share.S2C_RPC:
		rpc := &s2c.Rpc{}
		if err := proto.Unmarshal(buf, rpc); err != nil {
			return err
		}

		return proto.Unmarshal(rpc.GetData(), pb)
	}
	return fmt.Errorf("unknown msg")
}

func CreateMessage(servicemethod string, pb proto.Message) ([]byte, error) {
	data, err := proto.Marshal(pb)
	if err != nil {
		return nil, err
	}
	rpc := &c2s.Rpc{}
	rpc.Node = proto.String(".")
	rpc.Servicemethod = proto.String(servicemethod)
	rpc.Data = data
	data2, err := proto.Marshal(rpc)
	if err != nil {
		return nil, err
	}

	return util.CreateMsg(nil, data2, share.C2S_RPC)
}

type Client struct {
	id   int
	buf  []byte
	user string
}

func (c *Client) Start() {
	log.LogMessage("client start:", c.id)
	conn, err := net.Dial("tcp", "192.168.1.102:5391")
	if err != nil {
		log.LogError(err)
		return
	}

	c.buf = make([]byte, 1024*16)

	mid, data, err := util.ReadPkg(conn, c.buf)
	if err != nil {
		log.LogError(err)
		log.LogError("quit client ", c.id)
		return
	}

	l := &s2c.Login{}
	err = ParseProto(mid, data, l)
	if err != nil {
		log.LogError(err)
		log.LogError("quit client ", c.id)
		return
	}

	conn.Close()
	c.Login(l.GetHost(), l.GetPort())
}

func (c *Client) Login(h string, p int32) {
	log.LogMessage("client login:", c.id)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
	if err != nil {
		log.LogError(err)
		return
	}

	l := &c2s.Loginuser{}
	l.User = proto.String(fmt.Sprintf("test%d", c.id))
	l.Password = proto.String("123")

	c.user = *l.User
	data, err := CreateMessage("Account.Login", l)
	if err != nil {
		log.LogError(err)
		conn.Close()
		return
	}

	conn.Write(data)
	mid, data, err := util.ReadPkg(conn, c.buf)
	if err != nil {
		log.LogError(err)
		log.LogError("quit client ", c.id)
		conn.Close()
		return
	}

	ls := &s2c.Loginsucceed{}
	err = ParseProto(mid, data, ls)
	if err != nil {
		log.LogError(err)
		log.LogError("quit client ", c.id)
		conn.Close()
		return
	}

	conn.Close()
	c.LoginBase(ls.GetHost(), ls.GetPort(), ls.GetKey())
}

func (c *Client) LoginBase(h string, p int32, key int32) {
	log.LogMessage("client loginbase:", c.id)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
	if err != nil {
		log.LogError(err)
		return
	}

	base := &c2s.Enterbase{}
	base.User = proto.String(c.user)
	base.Key = proto.Int32(key)

	data, err := CreateMessage("Account.Login", base)
	if err != nil {
		log.LogError(err)
		conn.Close()
		return
	}

	conn.Write(data)

	for {
		mid, data, err := util.ReadPkg(conn, c.buf)
		if err != nil {
			log.LogError(err)
			log.LogError("quit client ", c.id)
			conn.Close()
			return
		}

		roleinfo := &s2c.RoleInfo{}
		err = ParseProto(mid, data, roleinfo)
		if err != nil {
			log.LogError(err)
			conn.Close()
			return
		}

		if len(roleinfo.UserInfo) > 0 {
			s := &c2s.Selectuser{}
			s.Rolename = roleinfo.UserInfo[0].Name
			s.Roleindex = roleinfo.UserInfo[0].Index
			data, err := CreateMessage("Account.SelectUser", s)
			if err != nil {
				log.LogError(err)
				conn.Close()
				return
			}

			conn.Write(data)
			break
		} else {
			create := &c2s.Create{}
			create.Name = proto.String(c.user)
			create.Index = proto.Int32(1)
			create.Sex = proto.Int32(1)
			create.Roleid = proto.Int32(1)
			data, err := CreateMessage("Account.CreatePlayer", create)
			if err != nil {
				log.LogError(err)
				conn.Close()
				return
			}
			conn.Write(data)
		}
	}

	c.EnterGame(conn)
}

func (c *Client) EnterGame(conn io.ReadWriteCloser) {
	log.LogMessage("client entergame:", c.id)
	co := &s2c.CreateObject{}
	mid, data, err := util.ReadPkg(conn, c.buf)
	if err != nil {
		log.LogError(err)
		log.LogError("quit client ", c.id)
		conn.Close()
		return
	}
	err = ParseProto(mid, data, co)
	if err != nil {
		log.LogError(err)
		conn.Close()
		return
	}

}

func main() {
	flag.Parse()

	start := time.Now()
	wg := &util.WaitGroupWrapper{}
	for i := 0; i < *count; i++ {
		c := &Client{id: i}
		wg.Wrap(c.Start)
	}

	wg.Wait()
	fmt.Println("times: ", time.Now().Sub(start).Seconds())
	time.Sleep(time.Second)
}
