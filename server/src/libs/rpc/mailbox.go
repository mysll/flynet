package rpc

import (
	"encoding/gob"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Mailbox struct {
	App  int32
	Flag int8
	Id   int64
	Uid  uint64
}

func (m Mailbox) String() string {
	return fmt.Sprintf("mailbox://%x", m.Uid)
}

func NewMailBoxFromStr(mb string) (Mailbox, error) {
	mbox := Mailbox{}
	if !strings.HasPrefix(mb, "mailbox://") {
		return mbox, errors.New("mailbox string error")
	}
	vals := strings.Split(mb, "/")
	if len(vals) != 3 {
		return mbox, errors.New("mailbox string error")
	}

	var val uint64
	var err error

	val, err = strconv.ParseUint(vals[2], 16, 64)
	if err != nil {
		return mbox, err
	}
	mbox.Uid = val
	mbox.Id = int64(mbox.Uid & 0x7FFFFFFFFFFF)
	mbox.Flag = int8((mbox.Uid >> 47) & 1)
	mbox.App = int32((mbox.Uid >> 48) & 0xFFFF)
	return mbox, nil
}

func NewMailBoxFromUid(val uint64) Mailbox {
	mbox := Mailbox{}
	mbox.Uid = val
	mbox.Id = int64(mbox.Uid & 0x7FFFFFFFFFFF)
	mbox.Flag = int8((mbox.Uid >> 47) & 1)
	mbox.App = int32((mbox.Uid >> 48) & 0xFFFF)
	return mbox
}

func NewMailBox(flag int8, id int64, appid int32) Mailbox {
	if id > 0x7FFFFFFFFFFF || appid > 0xFFFF {
		panic("id is wrong")
	}
	m := Mailbox{}
	m.App = appid
	m.Flag = flag
	m.Id = id
	m.Uid = ((uint64(appid) << 48) & 0xFFFF000000000000) | ((uint64(flag) & 1) << 47) | (uint64(id) & 0x7FFFFFFFFFFF)
	return m
}

func init() {
	gob.Register(Mailbox{})
}
