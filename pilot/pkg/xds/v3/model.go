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

package v3

import (
	"strings"

	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

const (
	envoyTypePrefix = resource.APITypePrefix + "envoy."

	ClusterType                = resource.ClusterType
	EndpointType               = resource.EndpointType
	ListenerType               = resource.ListenerType
	RouteType                  = resource.RouteType
	VirtualHostType            = resource.VirtualHostType
	SecretType                 = resource.SecretType
	ExtensionConfigurationType = resource.ExtensionConfigType

	NameTableType        = resource.APITypePrefix + "istio.networking.nds.v1.NameTable"
	HealthInfoType       = resource.APITypePrefix + "istio.v1.HealthInformation"
	ProxyConfigType      = resource.APITypePrefix + "istio.mesh.v1alpha1.ProxyConfig"
	ServiceInstancesType = resource.APITypePrefix + "istio.mesh.v1alpha1.ServiceInstance"
	// DebugType requests debug info from istio, a secured implementation for istio debug interface.
	DebugType     = "istio.io/debug"
	BootstrapType = resource.APITypePrefix + "envoy.config.bootstrap.v3.Bootstrap"

	// nolint
	HttpProtocolOptionsType = "envoy.extensions.upstreams.http.v3.HttpProtocolOptions"

	// ByteDanceAPITypePrefix ...
	ByteDanceAPITypePrefix = "type.bytedance.com/"
	// JavaConfigurationType ...
	JavaConfigurationType = ByteDanceAPITypePrefix + "istio.v3alpha1.javaagent.Configuration"
)

// GetShortType returns an abbreviated form of a type, useful for logging or human friendly messages
func GetShortType(typeURL string) string {
	switch typeURL {
	case ClusterType:
		return "CDS"
	case ListenerType:
		return "LDS"
	case RouteType:
		return "RDS"
	case VirtualHostType:
		return "VHDS"
	case EndpointType:
		return "EDS"
	case SecretType:
		return "SDS"
	case NameTableType:
		return "NDS"
	case ProxyConfigType:
		return "PCDS"
	case ExtensionConfigurationType:
		return "ECDS"
	case JavaConfigurationType:
		return "JDS"
	default:
		return typeURL
	}
}

// GetMetricType returns the form of a type reported for metrics
func GetMetricType(typeURL string) string {
	switch typeURL {
	case ClusterType:
		return "cds"
	case ListenerType:
		return "lds"
	case RouteType:
		return "rds"
	case VirtualHostType:
		return "vhds"
	case EndpointType:
		return "eds"
	case SecretType:
		return "sds"
	case NameTableType:
		return "nds"
	case ProxyConfigType:
		return "pcds"
	case ExtensionConfigurationType:
		return "ecds"
	case BootstrapType:
		return "bds"
	case JavaConfigurationType:
		return "jds"
	default:
		return typeURL
	}
}

// GetResourceType returns resource form of an abbreviated form
func GetResourceType(shortType string) string {
	s := strings.ToUpper(shortType)
	switch s {
	case "CDS":
		return ClusterType
	case "LDS":
		return ListenerType
	case "RDS":
		return RouteType
	case "VHDS":
		return VirtualHostType
	case "EDS":
		return EndpointType
	case "SDS":
		return SecretType
	case "NDS":
		return NameTableType
	case "PCDS":
		return ProxyConfigType
	case "ECDS":
		return ExtensionConfigurationType
	case "JDS":
		return JavaConfigurationType
	default:
		return shortType
	}
}

// IsEnvoyType checks whether the typeURL is a valid Envoy type.
func IsEnvoyType(typeURL string) bool {
	return strings.HasPrefix(typeURL, envoyTypePrefix)
}
