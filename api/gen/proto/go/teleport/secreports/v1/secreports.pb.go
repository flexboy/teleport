// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: teleport/secreports/v1/secreports.proto

package secreportsv1

import (
	v1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/header/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// AuditQuery is ...
type AuditQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// header is the header for the resource.
	Header *v1.ResourceHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	// spec is
	Spec *AuditQuerySpec `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *AuditQuery) Reset() {
	*x = AuditQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuditQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuditQuery) ProtoMessage() {}

func (x *AuditQuery) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuditQuery.ProtoReflect.Descriptor instead.
func (*AuditQuery) Descriptor() ([]byte, []int) {
	return file_teleport_secreports_v1_secreports_proto_rawDescGZIP(), []int{0}
}

func (x *AuditQuery) GetHeader() *v1.ResourceHeader {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *AuditQuery) GetSpec() *AuditQuerySpec {
	if x != nil {
		return x.Spec
	}
	return nil
}

// AuditQuery spec is ...
type AuditQuerySpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// query is ...
	Query string `protobuf:"bytes,2,opt,name=query,proto3" json:"query,omitempty"`
	// desc is ...
	Desc string `protobuf:"bytes,3,opt,name=desc,proto3" json:"desc,omitempty"`
}

func (x *AuditQuerySpec) Reset() {
	*x = AuditQuerySpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuditQuerySpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuditQuerySpec) ProtoMessage() {}

func (x *AuditQuerySpec) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuditQuerySpec.ProtoReflect.Descriptor instead.
func (*AuditQuerySpec) Descriptor() ([]byte, []int) {
	return file_teleport_secreports_v1_secreports_proto_rawDescGZIP(), []int{1}
}

func (x *AuditQuerySpec) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

func (x *AuditQuerySpec) GetDesc() string {
	if x != nil {
		return x.Desc
	}
	return ""
}

// AuditQuery is ...
type SecurityReport struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// header is the header for the resource.
	Header *v1.ResourceHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	// spec is
	Spec *SecurityReportSpec `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *SecurityReport) Reset() {
	*x = SecurityReport{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SecurityReport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SecurityReport) ProtoMessage() {}

func (x *SecurityReport) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SecurityReport.ProtoReflect.Descriptor instead.
func (*SecurityReport) Descriptor() ([]byte, []int) {
	return file_teleport_secreports_v1_secreports_proto_rawDescGZIP(), []int{2}
}

func (x *SecurityReport) GetHeader() *v1.ResourceHeader {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *SecurityReport) GetSpec() *SecurityReportSpec {
	if x != nil {
		return x.Spec
	}
	return nil
}

// TODO
type SecurityReportSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// header is the header for the resource.
	// TODO
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// TODO
	Desc string `protobuf:"bytes,3,opt,name=desc,proto3" json:"desc,omitempty"`
	// TODO
	Queries []string `protobuf:"bytes,4,rep,name=queries,proto3" json:"queries,omitempty"`
}

func (x *SecurityReportSpec) Reset() {
	*x = SecurityReportSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SecurityReportSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SecurityReportSpec) ProtoMessage() {}

func (x *SecurityReportSpec) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SecurityReportSpec.ProtoReflect.Descriptor instead.
func (*SecurityReportSpec) Descriptor() ([]byte, []int) {
	return file_teleport_secreports_v1_secreports_proto_rawDescGZIP(), []int{3}
}

func (x *SecurityReportSpec) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SecurityReportSpec) GetDesc() string {
	if x != nil {
		return x.Desc
	}
	return ""
}

func (x *SecurityReportSpec) GetQueries() []string {
	if x != nil {
		return x.Queries
	}
	return nil
}

// todo:
type SecurityReportState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// header is the header for the resource.
	Header *v1.ResourceHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	// spec is
	Spec *SecurityReportStateSpec `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *SecurityReportState) Reset() {
	*x = SecurityReportState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SecurityReportState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SecurityReportState) ProtoMessage() {}

func (x *SecurityReportState) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SecurityReportState.ProtoReflect.Descriptor instead.
func (*SecurityReportState) Descriptor() ([]byte, []int) {
	return file_teleport_secreports_v1_secreports_proto_rawDescGZIP(), []int{4}
}

func (x *SecurityReportState) GetHeader() *v1.ResourceHeader {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *SecurityReportState) GetSpec() *SecurityReportStateSpec {
	if x != nil {
		return x.Spec
	}
	return nil
}

// FOoo
type SecurityReportStateSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// state is
	State string `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
	// updated_at is
	UpdatedAt string `protobuf:"bytes,2,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

func (x *SecurityReportStateSpec) Reset() {
	*x = SecurityReportStateSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SecurityReportStateSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SecurityReportStateSpec) ProtoMessage() {}

func (x *SecurityReportStateSpec) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_secreports_v1_secreports_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SecurityReportStateSpec.ProtoReflect.Descriptor instead.
func (*SecurityReportStateSpec) Descriptor() ([]byte, []int) {
	return file_teleport_secreports_v1_secreports_proto_rawDescGZIP(), []int{5}
}

func (x *SecurityReportStateSpec) GetState() string {
	if x != nil {
		return x.State
	}
	return ""
}

func (x *SecurityReportStateSpec) GetUpdatedAt() string {
	if x != nil {
		return x.UpdatedAt
	}
	return ""
}

var File_teleport_secreports_v1_secreports_proto protoreflect.FileDescriptor

var file_teleport_secreports_v1_secreports_proto_rawDesc = []byte{
	0x0a, 0x27, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x73, 0x65, 0x63, 0x72, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x74, 0x65, 0x6c, 0x65, 0x70,
	0x6f, 0x72, 0x74, 0x2e, 0x73, 0x65, 0x63, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x2e, 0x76,
	0x31, 0x1a, 0x27, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x68, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x84, 0x01, 0x0a, 0x0a, 0x41,
	0x75, 0x64, 0x69, 0x74, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x3a, 0x0a, 0x06, 0x68, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x74, 0x65, 0x6c, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x2e, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x3a, 0x0a, 0x04, 0x73, 0x70, 0x65, 0x63, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x73,
	0x65, 0x63, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75, 0x64,
	0x69, 0x74, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53, 0x70, 0x65, 0x63, 0x52, 0x04, 0x73, 0x70, 0x65,
	0x63, 0x22, 0x3a, 0x0a, 0x0e, 0x41, 0x75, 0x64, 0x69, 0x74, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53,
	0x70, 0x65, 0x63, 0x12, 0x14, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x65, 0x73,
	0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x65, 0x73, 0x63, 0x22, 0x8c, 0x01,
	0x0a, 0x0e, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x12, 0x3a, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x22, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x48, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x3e, 0x0a, 0x04,
	0x73, 0x70, 0x65, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x74, 0x65, 0x6c,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x73, 0x65, 0x63, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x73,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x52, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x53, 0x70, 0x65, 0x63, 0x52, 0x04, 0x73, 0x70, 0x65, 0x63, 0x22, 0x56, 0x0a, 0x12,
	0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x53, 0x70,
	0x65, 0x63, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x65, 0x73, 0x63, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x65, 0x73, 0x63, 0x12, 0x18, 0x0a, 0x07, 0x71, 0x75,
	0x65, 0x72, 0x69, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x71, 0x75, 0x65,
	0x72, 0x69, 0x65, 0x73, 0x22, 0x96, 0x01, 0x0a, 0x13, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74,
	0x79, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x3a, 0x0a, 0x06,
	0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x74,
	0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x43, 0x0a, 0x04, 0x73, 0x70, 0x65, 0x63,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x2e, 0x73, 0x65, 0x63, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e,
	0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x53, 0x70, 0x65, 0x63, 0x52, 0x04, 0x73, 0x70, 0x65, 0x63, 0x22, 0x4e, 0x0a,
	0x17, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x53, 0x70, 0x65, 0x63, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1d,
	0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x42, 0x58, 0x5a,
	0x56, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x72, 0x61, 0x76,
	0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x67, 0x6f, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x73, 0x65, 0x63,
	0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x2f, 0x76, 0x31, 0x3b, 0x73, 0x65, 0x63, 0x72, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x73, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_teleport_secreports_v1_secreports_proto_rawDescOnce sync.Once
	file_teleport_secreports_v1_secreports_proto_rawDescData = file_teleport_secreports_v1_secreports_proto_rawDesc
)

func file_teleport_secreports_v1_secreports_proto_rawDescGZIP() []byte {
	file_teleport_secreports_v1_secreports_proto_rawDescOnce.Do(func() {
		file_teleport_secreports_v1_secreports_proto_rawDescData = protoimpl.X.CompressGZIP(file_teleport_secreports_v1_secreports_proto_rawDescData)
	})
	return file_teleport_secreports_v1_secreports_proto_rawDescData
}

var file_teleport_secreports_v1_secreports_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_teleport_secreports_v1_secreports_proto_goTypes = []interface{}{
	(*AuditQuery)(nil),              // 0: teleport.secreports.v1.AuditQuery
	(*AuditQuerySpec)(nil),          // 1: teleport.secreports.v1.AuditQuerySpec
	(*SecurityReport)(nil),          // 2: teleport.secreports.v1.SecurityReport
	(*SecurityReportSpec)(nil),      // 3: teleport.secreports.v1.SecurityReportSpec
	(*SecurityReportState)(nil),     // 4: teleport.secreports.v1.SecurityReportState
	(*SecurityReportStateSpec)(nil), // 5: teleport.secreports.v1.SecurityReportStateSpec
	(*v1.ResourceHeader)(nil),       // 6: teleport.header.v1.ResourceHeader
}
var file_teleport_secreports_v1_secreports_proto_depIdxs = []int32{
	6, // 0: teleport.secreports.v1.AuditQuery.header:type_name -> teleport.header.v1.ResourceHeader
	1, // 1: teleport.secreports.v1.AuditQuery.spec:type_name -> teleport.secreports.v1.AuditQuerySpec
	6, // 2: teleport.secreports.v1.SecurityReport.header:type_name -> teleport.header.v1.ResourceHeader
	3, // 3: teleport.secreports.v1.SecurityReport.spec:type_name -> teleport.secreports.v1.SecurityReportSpec
	6, // 4: teleport.secreports.v1.SecurityReportState.header:type_name -> teleport.header.v1.ResourceHeader
	5, // 5: teleport.secreports.v1.SecurityReportState.spec:type_name -> teleport.secreports.v1.SecurityReportStateSpec
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_teleport_secreports_v1_secreports_proto_init() }
func file_teleport_secreports_v1_secreports_proto_init() {
	if File_teleport_secreports_v1_secreports_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_teleport_secreports_v1_secreports_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuditQuery); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_teleport_secreports_v1_secreports_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuditQuerySpec); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_teleport_secreports_v1_secreports_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SecurityReport); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_teleport_secreports_v1_secreports_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SecurityReportSpec); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_teleport_secreports_v1_secreports_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SecurityReportState); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_teleport_secreports_v1_secreports_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SecurityReportStateSpec); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_teleport_secreports_v1_secreports_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_teleport_secreports_v1_secreports_proto_goTypes,
		DependencyIndexes: file_teleport_secreports_v1_secreports_proto_depIdxs,
		MessageInfos:      file_teleport_secreports_v1_secreports_proto_msgTypes,
	}.Build()
	File_teleport_secreports_v1_secreports_proto = out.File
	file_teleport_secreports_v1_secreports_proto_rawDesc = nil
	file_teleport_secreports_v1_secreports_proto_goTypes = nil
	file_teleport_secreports_v1_secreports_proto_depIdxs = nil
}
