package share

import (
	"bytes"
	"encoding/gob"
	"util"
)

const (
	M_SERVER_LIST = 100 + iota
	M_ADD_SERVER
	M_REMOVE_SERVER
	M_REGISTER_APP
	M_HEARTBEAT
	M_SHUTDOWN
	M_READY
	M_MUSTAPPREADY
	M_CREATEAPP
	M_CREATEAPP_BAK
	M_REGISTER_AGENT
	M_FORWARD_MSG
)

type RegisterApp struct {
	Type       string
	AppId      string
	Host       string
	Port       int
	ClientHost string
	ClientPort int
}

type RegisterAgent struct {
	AgentId   string
	NoBalance bool
}

type AddApp struct {
	Type       string
	AppId      string
	Host       string
	Port       int
	ClientHost string
	ClientPort int
	Ready      bool
}

type RemoveApp struct {
	AppId string
}

type AppInfo struct {
	Apps []AddApp
}

type AppReady struct {
	AppId string
}

type MustAppReady struct {
}

type CreateApp struct {
	Type    string
	ReqId   string
	AppId   string
	AppUid  int32
	Args    string
	CallApp string
}

type CreateAppBak struct {
	Id    string
	AppId string
	Res   string
}

type SendAppMsg struct {
	AppId string
	Data  []byte
}

func CreateForwardMsg(appid string, msg []byte) (data []byte, err error) {
	forward := &SendAppMsg{appid, msg}
	out, e := EncodeMsg(forward)
	if e != nil {
		err = e
		return
	}
	return util.CreateMsg(nil, out, M_FORWARD_MSG)
}

func CreateRegisterAgent(id string, nobalance bool) (data []byte, err error) {
	reg := &RegisterAgent{id, nobalance}
	out, e := EncodeMsg(reg)
	if e != nil {
		err = e
		return
	}
	return util.CreateMsg(nil, out, M_REGISTER_AGENT)
}

func CreateAppMsg(typ string, reqid string, appid string, appuid int32, args string, callapp string) (data []byte, err error) {
	create := &CreateApp{typ, reqid, appid, appuid, args, callapp}
	out, e := EncodeMsg(create)
	if e != nil {
		err = e
		return
	}
	return util.CreateMsg(nil, out, M_CREATEAPP)
}

func CreateAppBakMsg(reqid string, appid string, res string) (data []byte, err error) {
	create := &CreateAppBak{reqid, appid, res}
	out, e := EncodeMsg(create)
	if e != nil {
		err = e
		return
	}
	return util.CreateMsg(nil, out, M_CREATEAPP_BAK)
}

func CreateMustAppReadyMsg() (data []byte, err error) {
	ready := &MustAppReady{}
	out, e := EncodeMsg(ready)
	if e != nil {
		err = e
		return
	}
	return util.CreateMsg(nil, out, M_MUSTAPPREADY)
}

func CreateReadyMsg(id string) (data []byte, err error) {
	ready := &AppReady{id}
	out, e := EncodeMsg(ready)
	if e != nil {
		err = e
		return
	}
	return util.CreateMsg(nil, out, M_READY)
}

func CreateRegisterAppMsg(typ string, id string, host string, port int, clienthost string, clientport int) (data []byte, err error) {
	si := &RegisterApp{typ, id, host, port, clienthost, clientport}
	out, e := EncodeMsg(si)
	if e != nil {
		err = e
		return
	}

	return util.CreateMsg(nil, out, M_REGISTER_APP)
}

func CreateServerListMsg(slist []AddApp) (data []byte, err error) {
	si := &AppInfo{slist}
	out, e := EncodeMsg(si)
	if e != nil {
		err = e
		return
	}

	return util.CreateMsg(nil, out, M_SERVER_LIST)
}

func CreateAddServerMsg(typ string, id string, host string, port int, clienthost string, clientport int, ready bool) (data []byte, err error) {
	as := &AddApp{typ, id, host, port, clienthost, clientport, ready}
	out, e := EncodeMsg(as)
	if e != nil {
		err = e
		return
	}

	return util.CreateMsg(nil, out, M_ADD_SERVER)
}

func CreateRemoveServerMsg(id string) (data []byte, err error) {
	rs := &RemoveApp{AppId: id}
	out, e := EncodeMsg(rs)
	if e != nil {
		err = e
		return
	}
	data, err = util.CreateMsg(nil, out, M_REMOVE_SERVER)
	return
}

func EncodeMsg(msg interface{}) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(msg)
	return buff.Bytes(), err
}

func DecodeMsg(data []byte, out interface{}) error {

	buff := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buff)
	return dec.Decode(out)
}
