// Code generated by protoc-gen-go.
// source: s2c/respnewday.proto
// DO NOT EDIT!

/*
Package s2c is a generated protocol buffer package.

It is generated from these files:
	s2c/respnewday.proto

It has these top-level messages:
	Respnewday
*/
package s2c

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Respnewday struct {
	XXX_unrecognized []byte `json:"-"`
}

func (m *Respnewday) Reset()         { *m = Respnewday{} }
func (m *Respnewday) String() string { return proto.CompactTextString(m) }
func (*Respnewday) ProtoMessage()    {}
