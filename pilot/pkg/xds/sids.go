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
	"fmt"
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

func (l SidsGenerator) getShardsForService(hostname, namespace string) (*model.EndpointShards, string) {
	shards, ok := l.Server.Env.EndpointIndex.ShardsForService(hostname, namespace)
	if ok {
		return shards, hostname
	}

	// If not found, try to use Fully Qualified Domain Name to match
	fqdn := fmt.Sprintf("%s.%s.svc.%s", hostname, namespace, l.Server.Env.DomainSuffix)
	shards, _ = l.Server.Env.EndpointIndex.ShardsForService(fqdn, namespace)
	return shards, fqdn
}

func (l SidsGenerator) Generate(proxy *model.Proxy, watch *model.WatchedResource, req *model.PushRequest) (
	model.Resources, model.XdsLogDetails, error,
) {
	// if there is no endpoint changed, return directly.
	if !edsNeedsPush(req.ConfigsUpdated) {
		return nil, model.DefaultXdsLogDetails, nil
	}
	resources := model.Resources{}
	for _, resource := range watch.ResourceNames {
		// The format of the resource name is namespace/host
		parts := strings.Split(resource, "/")
		if len(parts) != 2 {
			continue
		}
		namespace, hostname := parts[0], parts[1]
		shards, fullhost := l.getShardsForService(hostname, namespace)
		if shards == nil {
			continue
		}
		// TODO the endpoints info are simple and just for lane discovery, can expand more meta info.
		// ref: eds.go
		var endpoints []*proto.ServiceInstance_Endpoint
		shards.RLock()
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
		shards.RUnlock()
		resources = append(resources, &discoveryv3.Resource{
			Resource: protoconv.MessageToAny(&proto.ServiceInstance{
				Service: &proto.ServiceInstance_Service{
					Namespace: namespace,
					Name:      hostname,
					Host:      fullhost,
				},
				Endpoints: endpoints,
			}),
		})
	}

	return resources, model.DefaultXdsLogDetails, nil
}
