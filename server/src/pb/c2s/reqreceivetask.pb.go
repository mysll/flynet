// Code generated by protoc-gen-go.
// source: c2s/reqreceivetask.proto
// DO NOT EDIT!

/*
Package c2s is a generated protocol buffer package.

It is generated from these files:
	c2s/reqreceivetask.proto

It has these top-level messages:
	Reqreceivetask
*/
package c2s

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Reqreceivetask struct {
	Taskid           *string `protobuf:"bytes,1,req,name=taskid" json:"taskid,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Reqreceivetask) Reset()         { *m = Reqreceivetask{} }
func (m *Reqreceivetask) String() string { return proto.CompactTextString(m) }
func (*Reqreceivetask) ProtoMessage()    {}

func (m *Reqreceivetask) GetTaskid() string {
	if m != nil && m.Taskid != nil {
		return *m.Taskid
	}
	return ""
}
