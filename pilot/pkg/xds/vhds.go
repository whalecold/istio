package xds

import (
	"fmt"

	"istio.io/istio/pilot/pkg/model"
)

// VhdsGenerator implements the new Generate method for VHDS, using the in-memory, optimized endpoint storage in DiscoveryServer.
type VhdsGenerator struct {
	Server *DiscoveryServer
}

var _ model.XdsResourceGenerator = &VhdsGenerator{}

func (vhds *VhdsGenerator) Generate(proxy *model.Proxy, w *model.WatchedResource, req *model.PushRequest) (model.Resources, model.XdsLogDetails, error) {
	return nil, model.DefaultXdsLogDetails, fmt.Errorf("vhds unsupport sotw ads")
}

func (vhds *VhdsGenerator) GenerateDeltas(proxy *model.Proxy, req *model.PushRequest,
	w *model.WatchedResource,
) (model.Resources, model.DeletedResources, model.XdsLogDetails, bool, error) {
	if !vhdsNeedsPush(req) {
		return nil, nil, model.DefaultXdsLogDetails, true, nil
	}
	resources, deleted, logDetails := vhds.Server.ConfigGenerator.BuildVirtualHosts(proxy, req, w.ResourceNames)
	return resources, deleted, logDetails, true, nil
}

func vhdsNeedsPush(req *model.PushRequest) bool {
	if req == nil {
		return true
	}
	if !req.Full {
		// vhds only handles full push
		return false
	}
	// If none set, we will always push
	if len(req.ConfigsUpdated) == 0 {
		return true
	}

	for config := range req.ConfigsUpdated {
		if _, f := skippedRdsConfigs[config.Kind]; !f {
			return true
		}
	}
	return false
}
