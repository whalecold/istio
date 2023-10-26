// Copyright Istio Authors. All Rights Reserved.
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
// 	protoc        v3.21.12
// source: pilot/pkg/xds/proto/sids.proto

package proto

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

// ServiceInstance represents an individual instance of a specific version
// of a service. It binds a network endpoint (ip:port), the service
// description (which is oblivious to various versions) and a set of labels
// that describe the service version associated with this instance.
//
// The labels associated with a service instance are unique per a network endpoint.
// There is one well defined set of labels for each service instance network endpoint.
//
// For example, the set of service instances associated with catalog.mystore.com
// are modeled like this
//
//	--> IstioEndpoint(172.16.0.1:8888), Service(catalog.myservice.com), Labels(foo=bar)
//	--> IstioEndpoint(172.16.0.2:8888), Service(catalog.myservice.com), Labels(foo=bar)
//	--> IstioEndpoint(172.16.0.3:8888), Service(catalog.myservice.com), Labels(kitty=cat)
//	--> IstioEndpoint(172.16.0.4:8888), Service(catalog.myservice.com), Labels(kitty=cat)
type ServiceInstance struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Service   *ServiceInstance_Service    `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	Endpoints []*ServiceInstance_Endpoint `protobuf:"bytes,3,rep,name=endpoints,proto3" json:"endpoints,omitempty"`
}

func (x *ServiceInstance) Reset() {
	*x = ServiceInstance{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pilot_pkg_xds_proto_sids_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceInstance) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceInstance) ProtoMessage() {}

func (x *ServiceInstance) ProtoReflect() protoreflect.Message {
	mi := &file_pilot_pkg_xds_proto_sids_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceInstance.ProtoReflect.Descriptor instead.
func (*ServiceInstance) Descriptor() ([]byte, []int) {
	return file_pilot_pkg_xds_proto_sids_proto_rawDescGZIP(), []int{0}
}

func (x *ServiceInstance) GetService() *ServiceInstance_Service {
	if x != nil {
		return x.Service
	}
	return nil
}

func (x *ServiceInstance) GetEndpoints() []*ServiceInstance_Endpoint {
	if x != nil {
		return x.Endpoints
	}
	return nil
}

type ServiceInstance_Service struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name is "destination.service.name" attribute
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Namespace is "destination.service.namespace" attribute
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// Host of the service, e.g. "productpage.bookinfo.svc.cluster.local"
	Host string `protobuf:"bytes,3,opt,name=host,proto3" json:"host,omitempty"`
}

func (x *ServiceInstance_Service) Reset() {
	*x = ServiceInstance_Service{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pilot_pkg_xds_proto_sids_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceInstance_Service) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceInstance_Service) ProtoMessage() {}

func (x *ServiceInstance_Service) ProtoReflect() protoreflect.Message {
	mi := &file_pilot_pkg_xds_proto_sids_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceInstance_Service.ProtoReflect.Descriptor instead.
func (*ServiceInstance_Service) Descriptor() ([]byte, []int) {
	return file_pilot_pkg_xds_proto_sids_proto_rawDescGZIP(), []int{0, 0}
}

func (x *ServiceInstance_Service) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ServiceInstance_Service) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *ServiceInstance_Service) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

type ServiceInstance_Endpoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Labels points to the workload or deployment labels.
	Labels map[string]string `protobuf:"bytes,1,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Address is the address of the instance.
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	// ServicePortName tracks the name of the port.
	ServicePortName string `protobuf:"bytes,3,opt,name=service_port_name,json=servicePortName,proto3" json:"service_port_name,omitempty"`
	// EndpointPort is the port where the workload is listening, can be different
	// from the service port.
	EndpointPort int32 `protobuf:"varint,4,opt,name=endpoint_port,json=endpointPort,proto3" json:"endpoint_port,omitempty"`
}

func (x *ServiceInstance_Endpoint) Reset() {
	*x = ServiceInstance_Endpoint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pilot_pkg_xds_proto_sids_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceInstance_Endpoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceInstance_Endpoint) ProtoMessage() {}

func (x *ServiceInstance_Endpoint) ProtoReflect() protoreflect.Message {
	mi := &file_pilot_pkg_xds_proto_sids_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceInstance_Endpoint.ProtoReflect.Descriptor instead.
func (*ServiceInstance_Endpoint) Descriptor() ([]byte, []int) {
	return file_pilot_pkg_xds_proto_sids_proto_rawDescGZIP(), []int{0, 1}
}

func (x *ServiceInstance_Endpoint) GetLabels() map[string]string {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *ServiceInstance_Endpoint) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ServiceInstance_Endpoint) GetServicePortName() string {
	if x != nil {
		return x.ServicePortName
	}
	return ""
}

func (x *ServiceInstance_Endpoint) GetEndpointPort() int32 {
	if x != nil {
		return x.EndpointPort
	}
	return 0
}

var File_pilot_pkg_xds_proto_sids_proto protoreflect.FileDescriptor

var file_pilot_pkg_xds_proto_sids_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x70, 0x69, 0x6c, 0x6f, 0x74, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x78, 0x64, 0x73, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x69, 0x64, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x13, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x22, 0xfd, 0x03, 0x0a, 0x0f, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x46, 0x0a, 0x07, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x69, 0x73, 0x74,
	0x69, 0x6f, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x4b, 0x0a, 0x09, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x45, 0x6e, 0x64, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x52, 0x09, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x1a, 0x4f,
	0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68,
	0x6f, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x1a,
	0x83, 0x02, 0x0a, 0x08, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x51, 0x0a, 0x06,
	0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x39, 0x2e, 0x69,
	0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e,
	0x63, 0x65, 0x2e, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x2e, 0x4c, 0x61, 0x62, 0x65,
	0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x2a, 0x0a, 0x11, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x6f, 0x72,
	0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x65, 0x6e,
	0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x50, 0x6f, 0x72, 0x74, 0x1a, 0x39, 0x0a, 0x0b, 0x4c, 0x61,
	0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x24, 0x5a, 0x22, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x69,
	0x6f, 0x2f, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2f, 0x70, 0x69, 0x6c, 0x6f, 0x74, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x78, 0x64, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_pilot_pkg_xds_proto_sids_proto_rawDescOnce sync.Once
	file_pilot_pkg_xds_proto_sids_proto_rawDescData = file_pilot_pkg_xds_proto_sids_proto_rawDesc
)

func file_pilot_pkg_xds_proto_sids_proto_rawDescGZIP() []byte {
	file_pilot_pkg_xds_proto_sids_proto_rawDescOnce.Do(func() {
		file_pilot_pkg_xds_proto_sids_proto_rawDescData = protoimpl.X.CompressGZIP(file_pilot_pkg_xds_proto_sids_proto_rawDescData)
	})
	return file_pilot_pkg_xds_proto_sids_proto_rawDescData
}

var file_pilot_pkg_xds_proto_sids_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_pilot_pkg_xds_proto_sids_proto_goTypes = []interface{}{
	(*ServiceInstance)(nil),          // 0: istio.mesh.v1alpha1.ServiceInstance
	(*ServiceInstance_Service)(nil),  // 1: istio.mesh.v1alpha1.ServiceInstance.Service
	(*ServiceInstance_Endpoint)(nil), // 2: istio.mesh.v1alpha1.ServiceInstance.Endpoint
	nil,                              // 3: istio.mesh.v1alpha1.ServiceInstance.Endpoint.LabelsEntry
}
var file_pilot_pkg_xds_proto_sids_proto_depIdxs = []int32{
	1, // 0: istio.mesh.v1alpha1.ServiceInstance.service:type_name -> istio.mesh.v1alpha1.ServiceInstance.Service
	2, // 1: istio.mesh.v1alpha1.ServiceInstance.endpoints:type_name -> istio.mesh.v1alpha1.ServiceInstance.Endpoint
	3, // 2: istio.mesh.v1alpha1.ServiceInstance.Endpoint.labels:type_name -> istio.mesh.v1alpha1.ServiceInstance.Endpoint.LabelsEntry
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_pilot_pkg_xds_proto_sids_proto_init() }
func file_pilot_pkg_xds_proto_sids_proto_init() {
	if File_pilot_pkg_xds_proto_sids_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pilot_pkg_xds_proto_sids_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceInstance); i {
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
		file_pilot_pkg_xds_proto_sids_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceInstance_Service); i {
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
		file_pilot_pkg_xds_proto_sids_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceInstance_Endpoint); i {
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
			RawDescriptor: file_pilot_pkg_xds_proto_sids_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pilot_pkg_xds_proto_sids_proto_goTypes,
		DependencyIndexes: file_pilot_pkg_xds_proto_sids_proto_depIdxs,
		MessageInfos:      file_pilot_pkg_xds_proto_sids_proto_msgTypes,
	}.Build()
	File_pilot_pkg_xds_proto_sids_proto = out.File
	file_pilot_pkg_xds_proto_sids_proto_rawDesc = nil
	file_pilot_pkg_xds_proto_sids_proto_goTypes = nil
	file_pilot_pkg_xds_proto_sids_proto_depIdxs = nil
}