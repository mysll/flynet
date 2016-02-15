package server

import (
	"data/entity"
	"libs/log"
	"libs/rpc"
	"pb/s2c"
	"util"

	"github.com/golang/protobuf/proto"
)

type TableSync struct {
	mailbox rpc.Mailbox
}

func NewTableSync(mb rpc.Mailbox) *TableSync {
	ts := &TableSync{}
	ts.mailbox = mb
	return ts
}

func (ts *TableSync) SyncTable(player entity.Entityer) {

	recs := player.GetRecNames()
	for _, r := range recs {
		rec := player.GetRec(r)
		if !rec.IsVisible() {
			continue
		}

		data, err := rec.Serial()
		if err != nil {
			continue
		}
		out := &s2c.CreateRecord{}
		out.Record = proto.String(r)
		out.Rows = proto.Int32(int32(rec.GetRows()))
		out.Cols = proto.Int32(int32(rec.GetCols()))
		out.Recinfo = data
		err = MailTo(nil, &ts.mailbox, "Entity.RecordInfo", out)
		if err != nil {
			log.LogError(err)
		}
	}
}

func (ts *TableSync) RecAppend(rec entity.Recorder, row int) {
	out := &s2c.RecordAddRow{}
	out.Record = proto.String(rec.GetName())
	out.Row = proto.Int32(int32(row))
	data, err := rec.SerialRow(row)
	if err != nil {
		log.LogError(err)
		return
	}
	out.Rowinfo = data
	err = MailTo(nil, &ts.mailbox, "Entity.RecordRowAdd", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableSync) RecDelete(rec entity.Recorder, row int) {
	out := &s2c.RecordDelRow{}
	out.Record = proto.String(rec.GetName())
	out.Row = proto.Int32(int32(row))
	err := MailTo(nil, &ts.mailbox, "Entity.RecordRowDel", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableSync) RecClear(rec entity.Recorder) {
	out := &s2c.RecordClear{}
	out.Record = proto.String(rec.GetName())
	err := MailTo(nil, &ts.mailbox, "Entity.RecordClear", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableSync) RecModify(rec entity.Recorder, row, col int) {
	out := &s2c.RecordGrid{}
	out.Record = proto.String(rec.GetName())
	out.Row = proto.Int32(int32(row))
	out.Col = proto.Int32(int32(col))
	value, _ := rec.Get(row, col)
	ar := util.NewStoreArchive()
	ar.Write(value)

	out.Gridinfo = ar.Data()
	err := MailTo(nil, &ts.mailbox, "Entity.RecordGrid", out)
	if err != nil {
		log.LogError(err)
	}
}

func (ts *TableSync) RecSetRow(rec entity.Recorder, row int) {
	out := &s2c.RecordSetRow{}
	out.Record = proto.String(rec.GetName())
	out.Row = proto.Int32(int32(row))
	data, err := rec.SerialRow(row)
	if err != nil {
		log.LogError(err)
		return
	}
	out.Rowinfo = data
	err = MailTo(nil, &ts.mailbox, "Entity.RecordRowSet", out)
	if err != nil {
		log.LogError(err)
	}
}
