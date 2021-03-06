// Code generated by protoc-gen-go.
// source: s2c/roleinfo.proto
// DO NOT EDIT!

/*
Package s2c is a generated protocol buffer package.

It is generated from these files:
	s2c/roleinfo.proto

It has these top-level messages:
	Role
	RoleInfo
*/
package s2c

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Role struct {
	Name             *string `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Index            *int32  `protobuf:"varint,2,req,name=index" json:"index,omitempty"`
	Roleinfo         *string `protobuf:"bytes,3,req,name=roleinfo" json:"roleinfo,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Role) Reset()         { *m = Role{} }
func (m *Role) String() string { return proto.CompactTextString(m) }
func (*Role) ProtoMessage()    {}

func (m *Role) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *Role) GetIndex() int32 {
	if m != nil && m.Index != nil {
		return *m.Index
	}
	return 0
}

func (m *Role) GetRoleinfo() string {
	if m != nil && m.Roleinfo != nil {
		return *m.Roleinfo
	}
	return ""
}

type RoleInfo struct {
	UserInfo         []*Role `protobuf:"bytes,1,rep,name=user_info" json:"user_info,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *RoleInfo) Reset()         { *m = RoleInfo{} }
func (m *RoleInfo) String() string { return proto.CompactTextString(m) }
func (*RoleInfo) ProtoMessage()    {}

func (m *RoleInfo) GetUserInfo() []*Role {
	if m != nil {
		return m.UserInfo
	}
	return nil
}
