package xds

import (
	"strings"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/util/protoconv"
	"istio.io/istio/pilot/pkg/xds/proto"
)

type SidsGenerator struct {
	Server *DiscoveryServer
}

var _ model.XdsResourceGenerator = &SidsGenerator{}

func (l SidsGenerator) Generate(proxy *model.Proxy, watch *model.WatchedResource, req *model.PushRequest) (model.Resources, model.XdsLogDetails, error) {
	// if there is no endpoint or service changed, return directly.
	if !edsNeedsPush(req.ConfigsUpdated) && !rdsNeedsPush(req) {
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
		resources = append(resources, &discovery.Resource{
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
