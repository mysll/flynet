package server

import (
	"server/data/datatype"
	"server/libs/log"
	"server/libs/rpc"
)

var tt TableCodec

type TableCodec interface {
	GetCodecInfo() string
	SyncTable(rec datatype.Recorder) interface{}
	RecAppend(rec datatype.Recorder, row int) interface{}
	RecDelete(rec datatype.Recorder, row int) interface{}
	RecClear(rec datatype.Recorder) interface{}
	RecModify(rec datatype.Recorder, row, col int) interface{}
	RecSetRow(rec datatype.Recorder, row int) interface{}
}

type TableSync struct {
	mailbox rpc.Mailbox
}

func NewTableSync(mb rpc.Mailbox) *TableSync {
	if tt == nil {
		panic("table transporter not set")
	}

	log.LogMessage("table sync proto:", tt.GetCodecInfo())
	ts := &TableSync{}
	ts.mailbox = mb
	return ts
}

func (ts *TableSync) SyncTable(player datatype.Entityer) {

	recs := player.GetRecNames()
	for _, r := range recs {
		rec := player.GetRec(r)
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

func (ts *TableSync) RecAppend(self datatype.Entityer, rec datatype.Recorder, row int) {
	out := tt.RecAppend(rec, row)
	if out == nil {
		return
	}
	err := MailTo(nil, &ts.mailbox, "Entity.RecordRowAdd", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableSync) RecDelete(self datatype.Entityer, rec datatype.Recorder, row int) {
	out := tt.RecDelete(rec, row)
	if out == nil {
		return
	}
	err := MailTo(nil, &ts.mailbox, "Entity.RecordRowDel", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableSync) RecClear(self datatype.Entityer, rec datatype.Recorder) {
	out := tt.RecClear(rec)
	if out == nil {
		return
	}
	err := MailTo(nil, &ts.mailbox, "Entity.RecordClear", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableSync) RecModify(self datatype.Entityer, rec datatype.Recorder, row, col int) {
	out := tt.RecModify(rec, row, col)
	if out == nil {
		return
	}
	err := MailTo(nil, &ts.mailbox, "Entity.RecordGrid", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableSync) RecSetRow(self datatype.Entityer, rec datatype.Recorder, row int) {
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
