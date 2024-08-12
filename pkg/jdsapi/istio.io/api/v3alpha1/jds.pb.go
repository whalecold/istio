// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.2
// source: pkg/jdsapi/jds.proto

// Keep this package for backward compatibility.

package v3alpha1

import (
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

type LaneMode int32

const (
	LaneMode_PERMISSIVE LaneMode = 0
	LaneMode_STRICT     LaneMode = 1
)

// Enum value maps for LaneMode.
var (
	LaneMode_name = map[int32]string{
		0: "PERMISSIVE",
		1: "STRICT",
	}
	LaneMode_value = map[string]int32{
		"PERMISSIVE": 0,
		"STRICT":     1,
	}
)

func (x LaneMode) Enum() *LaneMode {
	p := new(LaneMode)
	*p = x
	return p
}

func (x LaneMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LaneMode) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_jdsapi_jds_proto_enumTypes[0].Descriptor()
}

func (LaneMode) Type() protoreflect.EnumType {
	return &file_pkg_jdsapi_jds_proto_enumTypes[0]
}

func (x LaneMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LaneMode.Descriptor instead.
func (LaneMode) EnumDescriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{0}
}

// Java Configuration
type Configuration struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name         string               `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	MseRateLimit []*MseRateLimit      `protobuf:"bytes,2,rep,name=mse_rate_limit,json=mseRateLimit,proto3" json:"mse_rate_limit,omitempty"`
	MseLane      []*MseLane           `protobuf:"bytes,3,rep,name=mse_lane,json=mseLane,proto3" json:"mse_lane,omitempty"`
	Global       *GlobalConfiguration `protobuf:"bytes,4,opt,name=global,proto3" json:"global,omitempty"`
}

func (x *Configuration) Reset() {
	*x = Configuration{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Configuration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Configuration) ProtoMessage() {}

func (x *Configuration) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Configuration.ProtoReflect.Descriptor instead.
func (*Configuration) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{0}
}

func (x *Configuration) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Configuration) GetMseRateLimit() []*MseRateLimit {
	if x != nil {
		return x.MseRateLimit
	}
	return nil
}

func (x *Configuration) GetMseLane() []*MseLane {
	if x != nil {
		return x.MseLane
	}
	return nil
}

func (x *Configuration) GetGlobal() *GlobalConfiguration {
	if x != nil {
		return x.Global
	}
	return nil
}

type GlobalConfiguration struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MseLaneServices []*MseLaneService `protobuf:"bytes,1,rep,name=mse_lane_services,json=mseLaneServices,proto3" json:"mse_lane_services,omitempty"`
}

func (x *GlobalConfiguration) Reset() {
	*x = GlobalConfiguration{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GlobalConfiguration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GlobalConfiguration) ProtoMessage() {}

func (x *GlobalConfiguration) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GlobalConfiguration.ProtoReflect.Descriptor instead.
func (*GlobalConfiguration) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{1}
}

func (x *GlobalConfiguration) GetMseLaneServices() []*MseLaneService {
	if x != nil {
		return x.MseLaneServices
	}
	return nil
}

type MseRateLimit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Policies []*Policy `protobuf:"bytes,1,rep,name=policies,proto3" json:"policies,omitempty"`
	Limiter  *Limiter  `protobuf:"bytes,2,opt,name=limiter,proto3" json:"limiter,omitempty"`
}

func (x *MseRateLimit) Reset() {
	*x = MseRateLimit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MseRateLimit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MseRateLimit) ProtoMessage() {}

func (x *MseRateLimit) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MseRateLimit.ProtoReflect.Descriptor instead.
func (*MseRateLimit) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{2}
}

func (x *MseRateLimit) GetPolicies() []*Policy {
	if x != nil {
		return x.Policies
	}
	return nil
}

func (x *MseRateLimit) GetLimiter() *Limiter {
	if x != nil {
		return x.Limiter
	}
	return nil
}

type MseLane struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Policies []*Policy `protobuf:"bytes,1,rep,name=policies,proto3" json:"policies,omitempty"`
	Dyeing   *Dyeing   `protobuf:"bytes,2,opt,name=dyeing,proto3" json:"dyeing,omitempty"`
}

func (x *MseLane) Reset() {
	*x = MseLane{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MseLane) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MseLane) ProtoMessage() {}

func (x *MseLane) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MseLane.ProtoReflect.Descriptor instead.
func (*MseLane) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{3}
}

func (x *MseLane) GetPolicies() []*Policy {
	if x != nil {
		return x.Policies
	}
	return nil
}

func (x *MseLane) GetDyeing() *Dyeing {
	if x != nil {
		return x.Dyeing
	}
	return nil
}

type MseLaneService struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Mode     LaneMode   `protobuf:"varint,2,opt,name=mode,proto3,enum=istio.v3alpha1.javaagent.LaneMode" json:"mode,omitempty"`
	Services []*Service `protobuf:"bytes,3,rep,name=services,proto3" json:"services,omitempty"`
}

func (x *MseLaneService) Reset() {
	*x = MseLaneService{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MseLaneService) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MseLaneService) ProtoMessage() {}

func (x *MseLaneService) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MseLaneService.ProtoReflect.Descriptor instead.
func (*MseLaneService) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{4}
}

func (x *MseLaneService) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *MseLaneService) GetMode() LaneMode {
	if x != nil {
		return x.Mode
	}
	return LaneMode_PERMISSIVE
}

func (x *MseLaneService) GetServices() []*Service {
	if x != nil {
		return x.Services
	}
	return nil
}

type Service struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Host      string `protobuf:"bytes,2,opt,name=host,proto3" json:"host,omitempty"`
	Namespace string `protobuf:"bytes,3,opt,name=namespace,proto3" json:"namespace,omitempty"`
}

func (x *Service) Reset() {
	*x = Service{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Service) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Service) ProtoMessage() {}

func (x *Service) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Service.ProtoReflect.Descriptor instead.
func (*Service) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{5}
}

func (x *Service) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Service) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Service) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

type Limiter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MaxTokens     string `protobuf:"bytes,1,opt,name=max_tokens,json=maxTokens,proto3" json:"max_tokens,omitempty"`
	TokensPerFill string `protobuf:"bytes,2,opt,name=tokens_per_fill,json=tokensPerFill,proto3" json:"tokens_per_fill,omitempty"`
	FillInterval  string `protobuf:"bytes,3,opt,name=fill_interval,json=fillInterval,proto3" json:"fill_interval,omitempty"`
}

func (x *Limiter) Reset() {
	*x = Limiter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Limiter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Limiter) ProtoMessage() {}

func (x *Limiter) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Limiter.ProtoReflect.Descriptor instead.
func (*Limiter) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{6}
}

func (x *Limiter) GetMaxTokens() string {
	if x != nil {
		return x.MaxTokens
	}
	return ""
}

func (x *Limiter) GetTokensPerFill() string {
	if x != nil {
		return x.TokensPerFill
	}
	return ""
}

func (x *Limiter) GetFillInterval() string {
	if x != nil {
		return x.FillInterval
	}
	return ""
}

type Dyeing struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MeshEnvKey string `protobuf:"bytes,1,opt,name=mesh_env_key,json=meshEnvKey,proto3" json:"mesh_env_key,omitempty"`
	MeshEnvVal string `protobuf:"bytes,2,opt,name=mesh_env_val,json=meshEnvVal,proto3" json:"mesh_env_val,omitempty"`
}

func (x *Dyeing) Reset() {
	*x = Dyeing{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Dyeing) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Dyeing) ProtoMessage() {}

func (x *Dyeing) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Dyeing.ProtoReflect.Descriptor instead.
func (*Dyeing) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{7}
}

func (x *Dyeing) GetMeshEnvKey() string {
	if x != nil {
		return x.MeshEnvKey
	}
	return ""
}

func (x *Dyeing) GetMeshEnvVal() string {
	if x != nil {
		return x.MeshEnvVal
	}
	return ""
}

type Policy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HttpMatch []*HttpMatch `protobuf:"bytes,1,rep,name=http_match,json=httpMatch,proto3" json:"http_match,omitempty"`
}

func (x *Policy) Reset() {
	*x = Policy{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Policy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Policy) ProtoMessage() {}

func (x *Policy) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Policy.ProtoReflect.Descriptor instead.
func (*Policy) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{8}
}

func (x *Policy) GetHttpMatch() []*HttpMatch {
	if x != nil {
		return x.HttpMatch
	}
	return nil
}

type HttpMatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Headers   map[string]*StringMatch `protobuf:"bytes,1,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Principal *StringMatch            `protobuf:"bytes,2,opt,name=principal,proto3" json:"principal,omitempty"`
	Path      *StringMatch            `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	Method    *StringMatch            `protobuf:"bytes,4,opt,name=method,proto3" json:"method,omitempty"`
}

func (x *HttpMatch) Reset() {
	*x = HttpMatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HttpMatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HttpMatch) ProtoMessage() {}

func (x *HttpMatch) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HttpMatch.ProtoReflect.Descriptor instead.
func (*HttpMatch) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{9}
}

func (x *HttpMatch) GetHeaders() map[string]*StringMatch {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *HttpMatch) GetPrincipal() *StringMatch {
	if x != nil {
		return x.Principal
	}
	return nil
}

func (x *HttpMatch) GetPath() *StringMatch {
	if x != nil {
		return x.Path
	}
	return nil
}

func (x *HttpMatch) GetMethod() *StringMatch {
	if x != nil {
		return x.Method
	}
	return nil
}

type StringMatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Invert bool `protobuf:"varint,1,opt,name=invert,proto3" json:"invert,omitempty"`
	// Types that are assignable to MatchType:
	//
	//	*StringMatch_Exact
	//	*StringMatch_Prefix
	//	*StringMatch_Regex
	//	*StringMatch_Contains
	//	*StringMatch_Greater
	//	*StringMatch_Less
	MatchType isStringMatch_MatchType `protobuf_oneof:"match_type"`
}

func (x *StringMatch) Reset() {
	*x = StringMatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_jdsapi_jds_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringMatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringMatch) ProtoMessage() {}

func (x *StringMatch) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_jdsapi_jds_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StringMatch.ProtoReflect.Descriptor instead.
func (*StringMatch) Descriptor() ([]byte, []int) {
	return file_pkg_jdsapi_jds_proto_rawDescGZIP(), []int{10}
}

func (x *StringMatch) GetInvert() bool {
	if x != nil {
		return x.Invert
	}
	return false
}

func (m *StringMatch) GetMatchType() isStringMatch_MatchType {
	if m != nil {
		return m.MatchType
	}
	return nil
}

func (x *StringMatch) GetExact() string {
	if x, ok := x.GetMatchType().(*StringMatch_Exact); ok {
		return x.Exact
	}
	return ""
}

func (x *StringMatch) GetPrefix() string {
	if x, ok := x.GetMatchType().(*StringMatch_Prefix); ok {
		return x.Prefix
	}
	return ""
}

func (x *StringMatch) GetRegex() string {
	if x, ok := x.GetMatchType().(*StringMatch_Regex); ok {
		return x.Regex
	}
	return ""
}

func (x *StringMatch) GetContains() string {
	if x, ok := x.GetMatchType().(*StringMatch_Contains); ok {
		return x.Contains
	}
	return ""
}

func (x *StringMatch) GetGreater() string {
	if x, ok := x.GetMatchType().(*StringMatch_Greater); ok {
		return x.Greater
	}
	return ""
}

func (x *StringMatch) GetLess() string {
	if x, ok := x.GetMatchType().(*StringMatch_Less); ok {
		return x.Less
	}
	return ""
}

type isStringMatch_MatchType interface {
	isStringMatch_MatchType()
}

type StringMatch_Exact struct {
	// exact string match
	Exact string `protobuf:"bytes,2,opt,name=exact,proto3,oneof"`
}

type StringMatch_Prefix struct {
	// prefix-based match
	Prefix string `protobuf:"bytes,3,opt,name=prefix,proto3,oneof"`
}

type StringMatch_Regex struct {
	// RE2 style regex-based match (https://github.com/google/re2/wiki/Syntax).
	Regex string `protobuf:"bytes,4,opt,name=regex,proto3,oneof"`
}

type StringMatch_Contains struct {
	// contains match
	Contains string `protobuf:"bytes,5,opt,name=contains,proto3,oneof"`
}

type StringMatch_Greater struct {
	// greater than match
	Greater string `protobuf:"bytes,6,opt,name=greater,proto3,oneof"`
}

type StringMatch_Less struct {
	// less than match
	Less string `protobuf:"bytes,7,opt,name=less,proto3,oneof"`
}

func (*StringMatch_Exact) isStringMatch_MatchType() {}

func (*StringMatch_Prefix) isStringMatch_MatchType() {}

func (*StringMatch_Regex) isStringMatch_MatchType() {}

func (*StringMatch_Contains) isStringMatch_MatchType() {}

func (*StringMatch_Greater) isStringMatch_MatchType() {}

func (*StringMatch_Less) isStringMatch_MatchType() {}

var File_pkg_jdsapi_jds_proto protoreflect.FileDescriptor

var file_pkg_jdsapi_jds_proto_rawDesc = []byte{
	0x0a, 0x14, 0x70, 0x6b, 0x67, 0x2f, 0x6a, 0x64, 0x73, 0x61, 0x70, 0x69, 0x2f, 0x6a, 0x64, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e, 0x74,
	0x22, 0xf6, 0x01, 0x0a, 0x0d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x4c, 0x0a, 0x0e, 0x6d, 0x73, 0x65, 0x5f, 0x72, 0x61,
	0x74, 0x65, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26,
	0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x4d, 0x73, 0x65, 0x52, 0x61, 0x74,
	0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x52, 0x0c, 0x6d, 0x73, 0x65, 0x52, 0x61, 0x74, 0x65, 0x4c,
	0x69, 0x6d, 0x69, 0x74, 0x12, 0x3c, 0x0a, 0x08, 0x6d, 0x73, 0x65, 0x5f, 0x6c, 0x61, 0x6e, 0x65,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76,
	0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e,
	0x74, 0x2e, 0x4d, 0x73, 0x65, 0x4c, 0x61, 0x6e, 0x65, 0x52, 0x07, 0x6d, 0x73, 0x65, 0x4c, 0x61,
	0x6e, 0x65, 0x12, 0x45, 0x0a, 0x06, 0x67, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x47, 0x6c,
	0x6f, 0x62, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x06, 0x67, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x22, 0x6b, 0x0a, 0x13, 0x47, 0x6c, 0x6f,
	0x62, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x54, 0x0a, 0x11, 0x6d, 0x73, 0x65, 0x5f, 0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x69, 0x73,
	0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76,
	0x61, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x4d, 0x73, 0x65, 0x4c, 0x61, 0x6e, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x0f, 0x6d, 0x73, 0x65, 0x4c, 0x61, 0x6e, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x22, 0x89, 0x01, 0x0a, 0x0c, 0x4d, 0x73, 0x65, 0x52, 0x61,
	0x74, 0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x3c, 0x0a, 0x08, 0x70, 0x6f, 0x6c, 0x69, 0x63,
	0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x69, 0x73, 0x74, 0x69,
	0x6f, 0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61,
	0x67, 0x65, 0x6e, 0x74, 0x2e, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x08, 0x70, 0x6f, 0x6c,
	0x69, 0x63, 0x69, 0x65, 0x73, 0x12, 0x3b, 0x0a, 0x07, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x65, 0x72,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76,
	0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e,
	0x74, 0x2e, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x65, 0x72, 0x52, 0x07, 0x6c, 0x69, 0x6d, 0x69, 0x74,
	0x65, 0x72, 0x22, 0x81, 0x01, 0x0a, 0x07, 0x4d, 0x73, 0x65, 0x4c, 0x61, 0x6e, 0x65, 0x12, 0x3c,
	0x0a, 0x08, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x20, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x50, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x52, 0x08, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x12, 0x38, 0x0a, 0x06,
	0x64, 0x79, 0x65, 0x69, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x69,
	0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61,
	0x76, 0x61, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x44, 0x79, 0x65, 0x69, 0x6e, 0x67, 0x52, 0x06,
	0x64, 0x79, 0x65, 0x69, 0x6e, 0x67, 0x22, 0x9b, 0x01, 0x0a, 0x0e, 0x4d, 0x73, 0x65, 0x4c, 0x61,
	0x6e, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x36, 0x0a,
	0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x22, 0x2e, 0x69, 0x73,
	0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76,
	0x61, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x4c, 0x61, 0x6e, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x52,
	0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x3d, 0x0a, 0x08, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e,
	0x76, 0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65,
	0x6e, 0x74, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x08, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x22, 0x4f, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x22, 0x75, 0x0a, 0x07, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x65, 0x72,
	0x12, 0x1d, 0x0a, 0x0a, 0x6d, 0x61, 0x78, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6d, 0x61, 0x78, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x12,
	0x26, 0x0a, 0x0f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x5f, 0x70, 0x65, 0x72, 0x5f, 0x66, 0x69,
	0x6c, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x73,
	0x50, 0x65, 0x72, 0x46, 0x69, 0x6c, 0x6c, 0x12, 0x23, 0x0a, 0x0d, 0x66, 0x69, 0x6c, 0x6c, 0x5f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c,
	0x66, 0x69, 0x6c, 0x6c, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x22, 0x4c, 0x0a, 0x06,
	0x44, 0x79, 0x65, 0x69, 0x6e, 0x67, 0x12, 0x20, 0x0a, 0x0c, 0x6d, 0x65, 0x73, 0x68, 0x5f, 0x65,
	0x6e, 0x76, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d, 0x65,
	0x73, 0x68, 0x45, 0x6e, 0x76, 0x4b, 0x65, 0x79, 0x12, 0x20, 0x0a, 0x0c, 0x6d, 0x65, 0x73, 0x68,
	0x5f, 0x65, 0x6e, 0x76, 0x5f, 0x76, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x6d, 0x65, 0x73, 0x68, 0x45, 0x6e, 0x76, 0x56, 0x61, 0x6c, 0x22, 0x4c, 0x0a, 0x06, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x12, 0x42, 0x0a, 0x0a, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x6d, 0x61, 0x74,
	0x63, 0x68, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f,
	0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67,
	0x65, 0x6e, 0x74, 0x2e, 0x48, 0x74, 0x74, 0x70, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x09, 0x68,
	0x74, 0x74, 0x70, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x22, 0xf9, 0x02, 0x0a, 0x09, 0x48, 0x74, 0x74,
	0x70, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x12, 0x4a, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e,
	0x76, 0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65,
	0x6e, 0x74, 0x2e, 0x48, 0x74, 0x74, 0x70, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x48, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x73, 0x12, 0x43, 0x0a, 0x09, 0x70, 0x72, 0x69, 0x6e, 0x63, 0x69, 0x70, 0x61, 0x6c, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e, 0x74,
	0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x09, 0x70, 0x72,
	0x69, 0x6e, 0x63, 0x69, 0x70, 0x61, 0x6c, 0x12, 0x39, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e, 0x74,
	0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x04, 0x70, 0x61,
	0x74, 0x68, 0x12, 0x3d, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x25, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f,
	0x64, 0x1a, 0x61, 0x0a, 0x0c, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x3b, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x25, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x76, 0x33, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2e, 0x6a, 0x61, 0x76, 0x61, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0xcd, 0x01, 0x0a, 0x0b, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x4d,
	0x61, 0x74, 0x63, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x69, 0x6e, 0x76, 0x65, 0x72, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x69, 0x6e, 0x76, 0x65, 0x72, 0x74, 0x12, 0x16, 0x0a, 0x05,
	0x65, 0x78, 0x61, 0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x65,
	0x78, 0x61, 0x63, 0x74, 0x12, 0x18, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x16,
	0x0a, 0x05, 0x72, 0x65, 0x67, 0x65, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52,
	0x05, 0x72, 0x65, 0x67, 0x65, 0x78, 0x12, 0x1c, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69,
	0x6e, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x69, 0x6e, 0x73, 0x12, 0x1a, 0x0a, 0x07, 0x67, 0x72, 0x65, 0x61, 0x74, 0x65, 0x72, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x67, 0x72, 0x65, 0x61, 0x74, 0x65, 0x72,
	0x12, 0x14, 0x0a, 0x04, 0x6c, 0x65, 0x73, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00,
	0x52, 0x04, 0x6c, 0x65, 0x73, 0x73, 0x42, 0x0c, 0x0a, 0x0a, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x5f,
	0x74, 0x79, 0x70, 0x65, 0x2a, 0x26, 0x0a, 0x08, 0x4c, 0x61, 0x6e, 0x65, 0x4d, 0x6f, 0x64, 0x65,
	0x12, 0x0e, 0x0a, 0x0a, 0x50, 0x45, 0x52, 0x4d, 0x49, 0x53, 0x53, 0x49, 0x56, 0x45, 0x10, 0x00,
	0x12, 0x0a, 0x0a, 0x06, 0x53, 0x54, 0x52, 0x49, 0x43, 0x54, 0x10, 0x01, 0x42, 0x1c, 0x50, 0x01,
	0x5a, 0x15, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76,
	0x33, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x88, 0x01, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_pkg_jdsapi_jds_proto_rawDescOnce sync.Once
	file_pkg_jdsapi_jds_proto_rawDescData = file_pkg_jdsapi_jds_proto_rawDesc
)

func file_pkg_jdsapi_jds_proto_rawDescGZIP() []byte {
	file_pkg_jdsapi_jds_proto_rawDescOnce.Do(func() {
		file_pkg_jdsapi_jds_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_jdsapi_jds_proto_rawDescData)
	})
	return file_pkg_jdsapi_jds_proto_rawDescData
}

var file_pkg_jdsapi_jds_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pkg_jdsapi_jds_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_pkg_jdsapi_jds_proto_goTypes = []interface{}{
	(LaneMode)(0),               // 0: istio.v3alpha1.javaagent.LaneMode
	(*Configuration)(nil),       // 1: istio.v3alpha1.javaagent.Configuration
	(*GlobalConfiguration)(nil), // 2: istio.v3alpha1.javaagent.GlobalConfiguration
	(*MseRateLimit)(nil),        // 3: istio.v3alpha1.javaagent.MseRateLimit
	(*MseLane)(nil),             // 4: istio.v3alpha1.javaagent.MseLane
	(*MseLaneService)(nil),      // 5: istio.v3alpha1.javaagent.MseLaneService
	(*Service)(nil),             // 6: istio.v3alpha1.javaagent.Service
	(*Limiter)(nil),             // 7: istio.v3alpha1.javaagent.Limiter
	(*Dyeing)(nil),              // 8: istio.v3alpha1.javaagent.Dyeing
	(*Policy)(nil),              // 9: istio.v3alpha1.javaagent.Policy
	(*HttpMatch)(nil),           // 10: istio.v3alpha1.javaagent.HttpMatch
	(*StringMatch)(nil),         // 11: istio.v3alpha1.javaagent.StringMatch
	nil,                         // 12: istio.v3alpha1.javaagent.HttpMatch.HeadersEntry
}
var file_pkg_jdsapi_jds_proto_depIdxs = []int32{
	3,  // 0: istio.v3alpha1.javaagent.Configuration.mse_rate_limit:type_name -> istio.v3alpha1.javaagent.MseRateLimit
	4,  // 1: istio.v3alpha1.javaagent.Configuration.mse_lane:type_name -> istio.v3alpha1.javaagent.MseLane
	2,  // 2: istio.v3alpha1.javaagent.Configuration.global:type_name -> istio.v3alpha1.javaagent.GlobalConfiguration
	5,  // 3: istio.v3alpha1.javaagent.GlobalConfiguration.mse_lane_services:type_name -> istio.v3alpha1.javaagent.MseLaneService
	9,  // 4: istio.v3alpha1.javaagent.MseRateLimit.policies:type_name -> istio.v3alpha1.javaagent.Policy
	7,  // 5: istio.v3alpha1.javaagent.MseRateLimit.limiter:type_name -> istio.v3alpha1.javaagent.Limiter
	9,  // 6: istio.v3alpha1.javaagent.MseLane.policies:type_name -> istio.v3alpha1.javaagent.Policy
	8,  // 7: istio.v3alpha1.javaagent.MseLane.dyeing:type_name -> istio.v3alpha1.javaagent.Dyeing
	0,  // 8: istio.v3alpha1.javaagent.MseLaneService.mode:type_name -> istio.v3alpha1.javaagent.LaneMode
	6,  // 9: istio.v3alpha1.javaagent.MseLaneService.services:type_name -> istio.v3alpha1.javaagent.Service
	10, // 10: istio.v3alpha1.javaagent.Policy.http_match:type_name -> istio.v3alpha1.javaagent.HttpMatch
	12, // 11: istio.v3alpha1.javaagent.HttpMatch.headers:type_name -> istio.v3alpha1.javaagent.HttpMatch.HeadersEntry
	11, // 12: istio.v3alpha1.javaagent.HttpMatch.principal:type_name -> istio.v3alpha1.javaagent.StringMatch
	11, // 13: istio.v3alpha1.javaagent.HttpMatch.path:type_name -> istio.v3alpha1.javaagent.StringMatch
	11, // 14: istio.v3alpha1.javaagent.HttpMatch.method:type_name -> istio.v3alpha1.javaagent.StringMatch
	11, // 15: istio.v3alpha1.javaagent.HttpMatch.HeadersEntry.value:type_name -> istio.v3alpha1.javaagent.StringMatch
	16, // [16:16] is the sub-list for method output_type
	16, // [16:16] is the sub-list for method input_type
	16, // [16:16] is the sub-list for extension type_name
	16, // [16:16] is the sub-list for extension extendee
	0,  // [0:16] is the sub-list for field type_name
}

func init() { file_pkg_jdsapi_jds_proto_init() }
func file_pkg_jdsapi_jds_proto_init() {
	if File_pkg_jdsapi_jds_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_jdsapi_jds_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Configuration); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GlobalConfiguration); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MseRateLimit); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MseLane); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MseLaneService); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Service); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Limiter); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Dyeing); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Policy); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HttpMatch); i {
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
		file_pkg_jdsapi_jds_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringMatch); i {
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
	file_pkg_jdsapi_jds_proto_msgTypes[10].OneofWrappers = []interface{}{
		(*StringMatch_Exact)(nil),
		(*StringMatch_Prefix)(nil),
		(*StringMatch_Regex)(nil),
		(*StringMatch_Contains)(nil),
		(*StringMatch_Greater)(nil),
		(*StringMatch_Less)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_jdsapi_jds_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_jdsapi_jds_proto_goTypes,
		DependencyIndexes: file_pkg_jdsapi_jds_proto_depIdxs,
		EnumInfos:         file_pkg_jdsapi_jds_proto_enumTypes,
		MessageInfos:      file_pkg_jdsapi_jds_proto_msgTypes,
	}.Build()
	File_pkg_jdsapi_jds_proto = out.File
	file_pkg_jdsapi_jds_proto_rawDesc = nil
	file_pkg_jdsapi_jds_proto_goTypes = nil
	file_pkg_jdsapi_jds_proto_depIdxs = nil
}
