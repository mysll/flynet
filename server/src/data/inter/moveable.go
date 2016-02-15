package inter

import (
	"bytes"
	. "data/datatype"
	"encoding/gob"
)

type Mover interface {
	GetPos() Vector3
	SetPos(pos Vector3)
	SetOrient(dir float32)
	GetOrient() float32
}

type Moveable struct {
	Position Vector3
	Orient   float32
}

func (m *Moveable) Init() {

}

func (m *Moveable) Clear() {

}

func (m *Moveable) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	var err error

	err = encoder.Encode(m.Position)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(m.Orient)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *Moveable) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	var err error
	err = decoder.Decode(&m.Position)
	if err != nil {
		return err
	}
	err = decoder.Decode(&m.Orient)
	if err != nil {
		return err
	}
	return nil
}

func (m *Moveable) GetPos() Vector3 {
	return m.Position
}

func (m *Moveable) SetPos(p Vector3) {
	m.Position = p
}

func (m *Moveable) SetOrient(dir float32) {
	m.Orient = dir
}

func (m *Moveable) GetOrient() float32 {
	return m.Orient
}
