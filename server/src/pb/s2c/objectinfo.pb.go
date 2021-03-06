// Code generated by protoc-gen-go.
// source: s2c/objectinfo.proto
// DO NOT EDIT!

/*
Package s2c is a generated protocol buffer package.

It is generated from these files:
	s2c/objectinfo.proto

It has these top-level messages:
	Objectinfo
*/
package s2c

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Objectinfo struct {
	Self             *bool  `protobuf:"varint,1,req,name=self" json:"self,omitempty"`
	Index            *int32 `protobuf:"varint,2,opt,name=index" json:"index,omitempty"`
	Serial           *int32 `protobuf:"varint,3,opt,name=serial" json:"serial,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *Objectinfo) Reset()         { *m = Objectinfo{} }
func (m *Objectinfo) String() string { return proto.CompactTextString(m) }
func (*Objectinfo) ProtoMessage()    {}

func (m *Objectinfo) GetSelf() bool {
	if m != nil && m.Self != nil {
		return *m.Self
	}
	return false
}

func (m *Objectinfo) GetIndex() int32 {
	if m != nil && m.Index != nil {
		return *m.Index
	}
	return 0
}

func (m *Objectinfo) GetSerial() int32 {
	if m != nil && m.Serial != nil {
		return *m.Serial
	}
	return 0
}
