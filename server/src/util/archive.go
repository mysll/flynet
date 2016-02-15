package util

import (
	"bytes"
	. "data/datatype"
	"encoding/binary"
)

type StoreArchive struct {
	buffer *bytes.Buffer
}

func NewStoreArchive() *StoreArchive {
	ar := &StoreArchive{}
	ar.buffer = bytes.NewBuffer(nil)
	return ar
}

func (ar *StoreArchive) Data() []byte {
	return ar.buffer.Bytes()
}

func (ar *StoreArchive) Len() int {
	return ar.buffer.Len()
}

func (ar *StoreArchive) Write(val interface{}) {
	switch val.(type) {
	case int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		binary.Write(ar.buffer, binary.LittleEndian, val)
	case int:
		binary.Write(ar.buffer, binary.LittleEndian, int32(val.(int)))
	case string:
		ar.WriteString(val.(string))
	case ObjectID:
		ar.WriteObject(val.(ObjectID))
	default:
		binary.Write(ar.buffer, binary.LittleEndian, int8(-1))
	}
}

func (ar *StoreArchive) WriteString(val string) {
	data := []byte(val)
	size := len(data)
	binary.Write(ar.buffer, binary.LittleEndian, int16(size))
	ar.buffer.Write(data)
}

func (ar *StoreArchive) WriteObject(val ObjectID) {
	binary.Write(ar.buffer, binary.LittleEndian, val.Index)
	binary.Write(ar.buffer, binary.LittleEndian, val.Serial)
}

type LoadArchive struct {
	reader *bytes.Reader
}

func NewLoadArchiver(data []byte) *LoadArchive {
	ar := &LoadArchive{}
	ar.reader = bytes.NewReader(data)
	return ar
}

func (ar *LoadArchive) Read(val interface{}) error {
	return binary.Read(ar.reader, binary.LittleEndian, val)
}

func (ar *LoadArchive) ReadInt8(val *int8) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadUInt8(val *uint8) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadInt16(val *int16) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadUInt16(val *uint16) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadInt32(val *int32) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadUInt32(val *uint32) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadInt64(val *int64) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadUInt64(val *uint64) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadFloat32(val *float32) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadFloat64(val *float64) error {
	return ar.Read(val)
}

func (ar *LoadArchive) ReadString(val *string) error {
	var size int16
	binary.Read(ar.reader, binary.LittleEndian, &size)
	data := make([]byte, size)
	_, err := ar.reader.Read(data)
	if err != nil {
		return err
	}
	*val = string(data)
	return nil
}

func (ar *LoadArchive) ReadObject(val *ObjectID) error {
	err := ar.ReadInt32(&val.Index)
	if err != nil {
		return err
	}
	err = ar.ReadInt32(&val.Serial)
	return err
}
