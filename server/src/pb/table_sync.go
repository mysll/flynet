package pb

import (
	"pb/s2c"
	"server/data/datatype"
	"server/libs/log"
	"server/util"

	"github.com/golang/protobuf/proto"
)

type PBTableCodec struct {
}

func (ts *PBTableCodec) GetCodecInfo() string {
	return "use protobuf"
}

func (ts *PBTableCodec) SyncTable(rec datatype.Record) interface{} {
	data, err := rec.Serial()
	if err != nil {
		return nil
	}
	out := &s2c.CreateRecord{}
	out.Record = proto.String(rec.GetName())
	out.Rows = proto.Int32(int32(rec.GetRows()))
	out.Cols = proto.Int32(int32(rec.GetCols()))
	out.Recinfo = data

	return out
}

func (ts *PBTableCodec) RecAppend(rec datatype.Record, row int) interface{} {
	out := &s2c.RecordAddRow{}
	out.Record = proto.String(rec.GetName())
	out.Row = proto.Int32(int32(row))
	data, err := rec.SerialRow(row)
	if err != nil {
		log.LogError(err)
		return nil
	}
	out.Rowinfo = data
	return out
}

func (ts *PBTableCodec) RecDelete(rec datatype.Record, row int) interface{} {
	out := &s2c.RecordDelRow{}
	out.Record = proto.String(rec.GetName())
	out.Row = proto.Int32(int32(row))
	return out
}

func (ts *PBTableCodec) RecClear(rec datatype.Record) interface{} {
	out := &s2c.RecordClear{}
	out.Record = proto.String(rec.GetName())
	return out
}

func (ts *PBTableCodec) RecModify(rec datatype.Record, row, col int) interface{} {
	out := &s2c.RecordGrid{}
	out.Record = proto.String(rec.GetName())
	out.Row = proto.Int32(int32(row))
	out.Col = proto.Int32(int32(col))
	value, _ := rec.Get(row, col)
	ar := util.NewStoreArchiver(nil)
	ar.Write(value)

	out.Gridinfo = ar.Data()
	return out
}

func (ts *PBTableCodec) RecSetRow(rec datatype.Record, row int) interface{} {
	out := &s2c.RecordSetRow{}
	out.Record = proto.String(rec.GetName())
	out.Row = proto.Int32(int32(row))
	data, err := rec.SerialRow(row)
	if err != nil {
		log.LogError(err)
		return nil
	}
	out.Rowinfo = data
	return out
}
