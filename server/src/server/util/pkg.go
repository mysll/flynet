package util

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"server/libs/log"
)

func CreateMsg(buffer []byte, data []byte, id int) (out []byte, err error) {
	buf := bytes.NewBuffer(buffer)
	buf.Reset()
	if err = binary.Write(buf, binary.LittleEndian, uint16(len(data)+2)); err != nil {
		return
	}

	if err = binary.Write(buf, binary.LittleEndian, uint16(id)); err != nil {
		return
	}

	if len(data) > 0 {
		if _, err = buf.Write(data); err != nil {
			return
		}
	}

	return buf.Bytes(), nil
}

func ReadPkg(r io.Reader, buff []byte) (id uint16, msgbody []byte, err error) {
	var size uint16
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &id); err != nil {
		return
	}

	if size > uint16(len(buff)) {
		err = errors.New(fmt.Sprintf("msg: %d, size exceed: %d, %d", id, size, len(buff)))
		return
	}

	buf := buff[0 : size-2]
	if _, e := io.ReadFull(r, buf); e != nil {
		log.LogError(fmt.Sprintf("%x, %x, %s", size, id, e.Error()))
		err = e
		return
	}

	msgbody = buf
	return
}
