package util

import (
	. "server/data/datatype"
	"testing"
)

func TestArchive(t *testing.T) {
	ar := NewStoreArchive(nil)
	ar.WriteInt8(-8)
	ar.WriteUint8(8)
	ar.WriteInt16(-1600)
	ar.WriteUint16(1600)
	ar.WriteInt32(-320000)
	ar.WriteUint32(320000)
	ar.WriteFloat32(32.00)
	ar.WriteFloat64(64.00)
	ar.WriteString("我是谁啊")
	ar.WriteObject(ObjectID{1, 1})
	data := ar.Data()
	load := NewLoadArchiver(data)
	var i8 int8
	var ui8 uint8
	var i16 int16
	var ui16 uint16
	var i32 int32
	var ui32 uint32
	var f32 float32
	var f64 float64
	var s string
	var obj ObjectID
	err := load.ReadInt8(&i8)
	if err != nil || i8 != -8 {
		t.Fatalf("test failed: want: -8, get:%d", i8)
	}
	err = load.ReadUint8(&ui8)
	if err != nil || ui8 != 8 {
		t.Fatalf("test failed: want: 8, get:%d", ui8)
	}
	err = load.ReadInt16(&i16)
	if err != nil || i16 != -1600 {
		t.Fatalf("test failed: want: -1600, get:%d", i16)
	}

	err = load.ReadUint16(&ui16)
	if err != nil || ui16 != 1600 {
		t.Fatalf("test failed: want: 1600, get:%d", ui16)
	}

	err = load.ReadInt32(&i32)
	if err != nil || i32 != -320000 {
		t.Fatalf("test failed: want: -320000, get:%d", i32)
	}

	err = load.ReadUint32(&ui32)
	if err != nil || ui32 != 320000 {
		t.Fatalf("test failed: want: 320000, get:%d", ui32)
	}

	err = load.ReadFloat32(&f32)
	if err != nil || f32 != 32.00 {
		t.Fatalf("test failed: want: 320000, get:%f", f32)
	}

	err = load.ReadFloat64(&f64)
	if err != nil || f64 != 64.00 {
		t.Fatalf("test failed: want: 320000, get:%f", f64)
	}

	err = load.ReadString(&s)
	if err != nil || s != "我是谁啊" {
		t.Fatalf("test failed: want: 我是谁啊, get:%s", s)
	}

	err = load.ReadObject(&obj)
	if err != nil || obj.Index != 1 || obj.Serial != 1 {
		t.Fatalf("test failed: want: {1,1}, get:%v", obj)
	}
}
