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
	"strings"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/util/protoconv"
	"istio.io/istio/pilot/pkg/xds/proto"
)

type SidsGenerator struct {
	Server *DiscoveryServer
}

var _ model.XdsResourceGenerator = &SidsGenerator{}

func (l SidsGenerator) Generate(proxy *model.Proxy, watch *model.WatchedResource, req *model.PushRequest) (model.Resources, model.XdsLogDetails, error) {
	// if there is no endpoint changed, return directly.
	if !edsNeedsPush(req.ConfigsUpdated) {
		return nil, model.DefaultXdsLogDetails, nil
	}
	resources := model.Resources{}
	for _, resource := range watch.ResourceNames {

		// the format of the resource name is namespace/name.namespace.domain
		hosts := strings.Split(resource, "/")
		if len(hosts) != 2 {
			continue
		}
		namespace, host := hosts[0], hosts[1]
		name, _, _ := strings.Cut(host, "."+namespace)
		shards, ok := l.Server.Env.EndpointIndex.ShardsForService(host, namespace)
		if !ok {
			continue
		}
		var endpoints []*proto.ServiceInstance_Endpoint
		shards.Lock()
		for _, eps := range shards.Shards {
			for _, ep := range eps {
				endpoints = append(endpoints, &proto.ServiceInstance_Endpoint{
					Labels:          ep.Labels,
					Address:         ep.Address,
					ServicePortName: ep.ServicePortName,
					EndpointPort:    int32(ep.EndpointPort),
				})
			}
		}
		shards.Unlock()
		resources = append(resources, &discoveryv3.Resource{
			Resource: protoconv.MessageToAny(&proto.ServiceInstance{
				Service: &proto.ServiceInstance_Service{
					Namespace: namespace,
					Name:      name,
					Host:      host,
				},
				Endpoints: endpoints,
			}),
		})
	}

	return resources, model.DefaultXdsLogDetails, nil
}
