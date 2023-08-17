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

package xds

import (
	"reflect"
	"testing"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"

	meshconfig "istio.io/api/mesh/v1alpha1"
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/config"
)

func TestPopulateFailoverPriorityLabels(t *testing.T) {
	tests := []struct {
		name           string
		dr             *config.Config
		mesh           *meshconfig.MeshConfig
		expectedLabels []byte
	}{
		{
			name:           "no dr",
			expectedLabels: nil,
		},
		{
			name: "simple",
			dr: &config.Config{
				Spec: &networking.DestinationRule{
					TrafficPolicy: &networking.TrafficPolicy{
						OutlierDetection: &networking.OutlierDetection{
							ConsecutiveErrors: 5,
						},
						LoadBalancer: &networking.LoadBalancerSettings{
							LocalityLbSetting: &networking.LocalityLoadBalancerSetting{
								FailoverPriority: []string{
									"a",
									"b",
								},
							},
						},
					},
				},
			},
			expectedLabels: []byte("a:a b:b "),
		},
		{
			name: "no outlier detection",
			dr: &config.Config{
				Spec: &networking.DestinationRule{
					TrafficPolicy: &networking.TrafficPolicy{
						LoadBalancer: &networking.LoadBalancerSettings{
							LocalityLbSetting: &networking.LocalityLoadBalancerSetting{
								FailoverPriority: []string{
									"a",
									"b",
								},
							},
						},
					},
				},
			},
			expectedLabels: nil,
		},
		{
			name: "no failover priority",
			dr: &config.Config{
				Spec: &networking.DestinationRule{
					TrafficPolicy: &networking.TrafficPolicy{
						OutlierDetection: &networking.OutlierDetection{
							ConsecutiveErrors: 5,
						},
					},
				},
			},
			expectedLabels: nil,
		},
		{
			name: "failover priority disabled",
			dr: &config.Config{
				Spec: &networking.DestinationRule{
					TrafficPolicy: &networking.TrafficPolicy{
						OutlierDetection: &networking.OutlierDetection{
							ConsecutiveErrors: 5,
						},
						LoadBalancer: &networking.LoadBalancerSettings{
							LocalityLbSetting: &networking.LocalityLoadBalancerSetting{
								FailoverPriority: []string{
									"a",
									"b",
								},
								Enabled: &wrappers.BoolValue{Value: false},
							},
						},
					},
				},
			},
			expectedLabels: nil,
		},
		{
			name: "mesh LocalityLoadBalancerSetting",
			dr: &config.Config{
				Spec: &networking.DestinationRule{
					TrafficPolicy: &networking.TrafficPolicy{
						OutlierDetection: &networking.OutlierDetection{
							ConsecutiveErrors: 5,
						},
					},
				},
			},
			mesh: &meshconfig.MeshConfig{
				LocalityLbSetting: &networking.LocalityLoadBalancerSetting{
					FailoverPriority: []string{
						"a",
						"b",
					},
				},
			},
			expectedLabels: []byte("a:a b:b "),
		},
		{
			name: "mesh LocalityLoadBalancerSetting(no outlier detection)",
			dr: &config.Config{
				Spec: &networking.DestinationRule{
					TrafficPolicy: &networking.TrafficPolicy{},
				},
			},
			mesh: &meshconfig.MeshConfig{
				LocalityLbSetting: &networking.LocalityLoadBalancerSetting{
					FailoverPriority: []string{
						"a",
						"b",
					},
				},
			},
			expectedLabels: nil,
		},
		{
			name: "mesh LocalityLoadBalancerSetting(no failover priority)",
			dr: &config.Config{
				Spec: &networking.DestinationRule{
					TrafficPolicy: &networking.TrafficPolicy{
						OutlierDetection: &networking.OutlierDetection{
							ConsecutiveErrors: 5,
						},
					},
				},
			},
			mesh: &meshconfig.MeshConfig{
				LocalityLbSetting: &networking.LocalityLoadBalancerSetting{},
			},
			expectedLabels: nil,
		},
		{
			name: "mesh LocalityLoadBalancerSetting(failover priority disabled)",
			dr: &config.Config{
				Spec: &networking.DestinationRule{
					TrafficPolicy: &networking.TrafficPolicy{
						OutlierDetection: &networking.OutlierDetection{
							ConsecutiveErrors: 5,
						},
					},
				},
			},
			mesh: &meshconfig.MeshConfig{
				LocalityLbSetting: &networking.LocalityLoadBalancerSetting{
					FailoverPriority: []string{
						"a",
						"b",
					},
					Enabled: &wrappers.BoolValue{Value: false},
				},
			},
			expectedLabels: nil,
		},
		{
			name: "both dr and mesh LocalityLoadBalancerSetting",
			dr: &config.Config{
				Spec: &networking.DestinationRule{
					TrafficPolicy: &networking.TrafficPolicy{
						OutlierDetection: &networking.OutlierDetection{
							ConsecutiveErrors: 5,
						},
						LoadBalancer: &networking.LoadBalancerSettings{
							LocalityLbSetting: &networking.LocalityLoadBalancerSetting{
								FailoverPriority: []string{
									"a",
									"b",
								},
							},
						},
					},
				},
			},
			mesh: &meshconfig.MeshConfig{
				LocalityLbSetting: &networking.LocalityLoadBalancerSetting{
					FailoverPriority: []string{
						"c",
					},
				},
			},
			expectedLabels: []byte("a:a b:b "),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := EndpointBuilder{
				proxy: &model.Proxy{
					Metadata: &model.NodeMetadata{},
					Labels: map[string]string{
						"app": "foo",
						"a":   "a",
						"b":   "b",
					},
				},
				push: &model.PushContext{
					Mesh: tt.mesh,
				},
			}
			if tt.dr != nil {
				b.destinationRule = model.ConvertConsolidatedDestRule(tt.dr)
			}
			b.populateFailoverPriorityLabels()
			if !reflect.DeepEqual(b.failoverPriorityLabels, tt.expectedLabels) {
				t.Fatalf("expected priorityLabels %v but got %v", tt.expectedLabels, b.failoverPriorityLabels)
			}
		})
	}
}

func TestGetEnvoyLbEndpointPort(t *testing.T) {
	tests := []struct {
		b    *EndpointBuilder
		e    *model.IstioEndpoint
		name string
		port uint32
	}{
		{
			nil,
			&model.IstioEndpoint{
				EndpointPort: 80,
			},
			"nil",
			80,
		},
		{
			&EndpointBuilder{},
			&model.IstioEndpoint{
				EndpointPort: 80,
			},
			"none",
			80,
		},
		{
			&EndpointBuilder{
				service: &model.Service{
					Attributes: model.ServiceAttributes{
						Labels: map[string]string{
							sidecarInboundPort: "port",
						},
					},
				},
			},
			&model.IstioEndpoint{
				EndpointPort: 80,
			},
			"string",
			80,
		},
		{
			&EndpointBuilder{
				service: &model.Service{
					Attributes: model.ServiceAttributes{
						Labels: map[string]string{
							sidecarInboundPort: "16000",
						},
					},
				},
			},
			&model.IstioEndpoint{
				EndpointPort: 80,
			},
			"int",
			16000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := getEnvoyLbEndpointPort(tt.b, tt.e)
			if port != tt.port {
				t.Fatalf("expected port %v but got %v", tt.port, port)
			}
		})
	}
}
