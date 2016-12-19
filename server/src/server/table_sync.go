package server

import (
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
)

var tt TableCodec

type TableCodec interface {
	GetCodecInfo() string
	SyncTable(rec datatype.Record) interface{}
	RecAppend(rec datatype.Record, row int) interface{}
	RecDelete(rec datatype.Record, row int) interface{}
	RecClear(rec datatype.Record) interface{}
	RecModify(rec datatype.Record, row, col int) interface{}
	RecSetRow(rec datatype.Record, row int) interface{}
}

type TableTrans struct {
	mailbox rpc.Mailbox
}

func NewTableTrans(mb rpc.Mailbox) *TableTrans {
	if tt == nil {
		panic("table transporter not set")
	}

	log.LogMessage("table sync proto:", tt.GetCodecInfo())
	ts := &TableTrans{}
	ts.mailbox = mb
	return ts
}

func (ts *TableTrans) SyncTable(player datatype.Entity) {

	recs := player.RecordNames()
	for _, r := range recs {
		rec := player.FindRec(r)
		if !rec.IsVisible() {
			continue
		}

		out := tt.SyncTable(rec)
		if out == nil {
			continue
		}
		err := MailTo(nil, &ts.mailbox, "Entity.RecordInfo", out)
		if err != nil {
			log.LogError(err)
		}
	}
}

func (ts *TableTrans) RecAppend(self datatype.Entity, rec datatype.Record, row int) {
	out := tt.RecAppend(rec, row)
	if out == nil {
		return
	}
	err := MailTo(nil, &ts.mailbox, "Entity.RecordRowAdd", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableTrans) RecDelete(self datatype.Entity, rec datatype.Record, row int) {
	out := tt.RecDelete(rec, row)
	if out == nil {
		return
	}
	err := MailTo(nil, &ts.mailbox, "Entity.RecordRowDel", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableTrans) RecClear(self datatype.Entity, rec datatype.Record) {
	out := tt.RecClear(rec)
	if out == nil {
		return
	}
	err := MailTo(nil, &ts.mailbox, "Entity.RecordClear", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableTrans) RecModify(self datatype.Entity, rec datatype.Record, row, col int) {
	out := tt.RecModify(rec, row, col)
	if out == nil {
		return
	}
	err := MailTo(nil, &ts.mailbox, "Entity.RecordGrid", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableTrans) RecSetRow(self datatype.Entity, rec datatype.Record, row int) {
	out := tt.RecSetRow(rec, row)
	if out == nil {
		return
	}
	err := MailTo(nil, &ts.mailbox, "Entity.RecordRowSet", out)
	if err != nil {
		log.LogError(err)
	}
}

func RegisterTableCodec(t TableCodec) {
	tt = t
}
