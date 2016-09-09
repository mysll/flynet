package share

import (
	"bytes"
	"encoding/gob"
	"server/util"
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
	Type             string
	Id               int32
	Name             string
	Host             string
	Port             int
	ClientHost       string
	ClientPort       int
	EnableGlobalData bool
}

type RegisterAgent struct {
	AgentId   string
	NoBalance bool
}

type AddApp struct {
	Type             string
	Id               int32
	Name             string
	Host             string
	Port             int
	ClientHost       string
	ClientPort       int
	Ready            bool
	EnableGlobalData bool
}

type RemoveApp struct {
	Id int32
}

type AppInfo struct {
	Apps []AddApp
}

type AppReady struct {
	Id int32
}

type MustAppReady struct {
}

type CreateApp struct {
	Type    string
	ReqId   string
	AppName string
	AppUid  int32
	Args    string
	CallApp int32
}

type CreateAppBak struct {
	Id    string
	AppId int32
	Res   string
}

type SendAppMsg struct {
	Id   int32
	Data []byte
}

func CreateForwardMsg(appid int32, msg []byte) (data []byte, err error) {
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

func CreateAppMsg(typ string, reqid string, appid string, appuid int32, args string, callapp int32) (data []byte, err error) {
	create := &CreateApp{typ, reqid, appid, appuid, args, callapp}
	out, e := EncodeMsg(create)
	if e != nil {
		err = e
		return
	}
	return util.CreateMsg(nil, out, M_CREATEAPP)
}

func CreateAppBakMsg(reqid string, appid int32, res string) (data []byte, err error) {
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

func CreateReadyMsg(id int32) (data []byte, err error) {
	ready := &AppReady{id}
	out, e := EncodeMsg(ready)
	if e != nil {
		err = e
		return
	}
	return util.CreateMsg(nil, out, M_READY)
}

func CreateRegisterAppMsg(typ string, id int32, name string, host string, port int, clienthost string, clientport int, enableglobaldata bool) (data []byte, err error) {
	si := &RegisterApp{typ, id, name, host, port, clienthost, clientport, enableglobaldata}
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

func CreateAddServerMsg(typ string, id int32, name string, host string, port int, clienthost string, clientport int, ready bool, enableglobaldata bool) (data []byte, err error) {
	as := &AddApp{typ, id, name, host, port, clienthost, clientport, ready, enableglobaldata}
	out, e := EncodeMsg(as)
	if e != nil {
		err = e
		return
	}

	return util.CreateMsg(nil, out, M_ADD_SERVER)
}

func CreateRemoveServerMsg(id int32) (data []byte, err error) {
	rs := &RemoveApp{id}
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
