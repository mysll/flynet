package rpc

import (
	"encoding/gob"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Mailbox struct {
	Address string
	Type    string
	Id      int64
	Uid     int64
}

func (m Mailbox) String() string {
	return fmt.Sprintf("mailbox://%s/%s/%x", m.Address, m.Type, m.Uid)
}

func NewMailBoxFromStr(mb string) (Mailbox, error) {
	mbox := Mailbox{}
	if !strings.HasPrefix(mb, "mailbox://") {
		return mbox, errors.New("mailbox string error")
	}
	vals := strings.Split(mb, "/")
	if len(vals) != 5 {
		return mbox, errors.New("mailbox string error")
	}

	var uid int64
	var err error
	if vals[4] != "" {
		uid, err = strconv.ParseInt(vals[4], 16, 64)
		if err != nil {
			return mbox, err
		}
	}

	mbox.Address = vals[2]
	mbox.Type = vals[3]
	mbox.Id = uid & 0xFFFFFFFFFFFF
	mbox.Uid = uid
	return mbox, nil
}

func NewMailBox(address string, typ string, id int64, appid int32) Mailbox {
	m := Mailbox{}
	m.Address = address
	m.Type = typ
	m.Id = id

	m.Uid = ((int64(appid) << 48) & 0x7FFF000000000000) | id
	return m
}

func init() {
	gob.Register(Mailbox{})
}
