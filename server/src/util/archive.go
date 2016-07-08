package util

import (
	"bytes"
	. "data/datatype"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

type StoreArchive struct {
	buffer *bytes.Buffer
}

func NewStoreArchiver(data []byte) *StoreArchive {
	ar := &StoreArchive{}
	ar.buffer = bytes.NewBuffer(data)
	return ar
}

func (ar *StoreArchive) Data() []byte {
	return ar.buffer.Bytes()
}

func (ar *StoreArchive) Len() int {
	return ar.buffer.Len()
}

func (ar *StoreArchive) WriteAt(offset int, val interface{}) error {
	if offset >= ar.buffer.Len() {
		return fmt.Errorf("offset out of range")
	}

	data := ar.buffer.Bytes()
	tmp := bytes.NewBuffer(data[offset:offset])
	switch val.(type) {
	case int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return binary.Write(tmp, binary.LittleEndian, val)
	case int:
		return binary.Write(tmp, binary.LittleEndian, int32(val.(int)))
	default:
		return fmt.Errorf("unsupport type")
	}
}

func (ar *StoreArchive) Write(val interface{}) error {
	switch val.(type) {
	case int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return binary.Write(ar.buffer, binary.LittleEndian, val)
	case int:
		return binary.Write(ar.buffer, binary.LittleEndian, int32(val.(int)))
	case string:
		return ar.WriteString(val.(string))
	case ObjectID:
		return ar.WriteObjectID(val.(ObjectID))
	case []byte:
		return ar.WriteData(val.([]byte))
	default:
		return ar.WriteObject(val)
	}
}

func (ar *StoreArchive) WriteString(val string) error {
	data := []byte(val)
	size := len(data)
	err := binary.Write(ar.buffer, binary.LittleEndian, int16(size))
	if err != nil {
		return err
	}
	_, err = ar.buffer.Write(data)
	return err
}

func (ar *StoreArchive) WriteObjectID(val ObjectID) error {
	err := binary.Write(ar.buffer, binary.LittleEndian, val.Index)
	if err != nil {
		return err
	}
	err = binary.Write(ar.buffer, binary.LittleEndian, val.Serial)
	return err
}

func (ar *StoreArchive) WriteObject(obj interface{}) error {
	enc := gob.NewEncoder(ar.buffer)
	return enc.Encode(obj)
}

func (ar *StoreArchive) WriteData(data []byte) error {
	err := ar.Write(uint16(len(data)))
	if err != nil {
		return err
	}
	_, err = ar.buffer.Write(data)
	return err
}

type LoadArchive struct {
	reader *bytes.Reader
}

func NewLoadArchiver(data []byte) *LoadArchive {
	ar := &LoadArchive{}
	ar.reader = bytes.NewReader(data)
	return ar
}

func (ar *LoadArchive) Position() int {
	return int(ar.reader.Size()) - ar.reader.Len()
}

func (ar *LoadArchive) AvailableBytes() int {
	return ar.reader.Len()
}

func (ar *LoadArchive) Size() int {
	return int(ar.reader.Size())
}

func (ar *LoadArchive) Seek(offset int, whence int) (int, error) {
	ret, err := ar.reader.Seek(int64(offset), whence)
	return int(ret), err
}

func (ar *LoadArchive) Read(val interface{}) (err error) {
	switch val.(type) {
	case *int8, *int16, *int32, *int64, *uint8, *uint16, *uint32, *uint64, *float32, *float64:
		return binary.Read(ar.reader, binary.LittleEndian, val)
	case *int:
		return binary.Read(ar.reader, binary.LittleEndian, int32(val.(int)))
	case *string:
		inst := val.(*string)
		*inst, err = ar.ReadString()
		return err
	case *ObjectID:
		inst := val.(*ObjectID)
		*inst, err = ar.ReadObjectID()
		return err
	case *[]byte:
		inst := val.(*[]byte)
		*inst, err = ar.ReadData()
		return err
	default:
		return ar.ReadObject(val)
	}
}

func (ar *LoadArchive) ReadInt8() (val int8, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadUInt8() (val uint8, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadInt16() (val int16, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadUInt16() (val uint16, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadInt32() (val int32, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadUInt32() (val uint32, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadInt64() (val int64, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadUInt64() (val uint64, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadFloat32() (val float32, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadFloat64() (val float64, err error) {
	err = ar.Read(&val)
	return
}

func (ar *LoadArchive) ReadString() (val string, err error) {
	var size int16
	binary.Read(ar.reader, binary.LittleEndian, &size)
	if size == 0 {
		val = ""
		return
	}
	data := make([]byte, size)
	_, err = ar.reader.Read(data)
	if err != nil {
		return
	}
	val = string(data)
	return
}

func (ar *LoadArchive) ReadObjectID() (val ObjectID, err error) {
	val.Index, err = ar.ReadInt32()
	if err != nil {
		return
	}
	val.Serial, err = ar.ReadInt32()
	return
}

func (ar *LoadArchive) ReadObject(val interface{}) error {
	dec := gob.NewDecoder(ar.reader)
	return dec.Decode(val)
}

func (ar *LoadArchive) ReadData() (data []byte, err error) {
	var l uint16
	l, err = ar.ReadUInt16()
	data = make([]byte, int(l))
	_, err = ar.reader.Read(data)
	return data, err
}
