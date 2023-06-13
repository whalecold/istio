//  Copyright Istio Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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

func (vhds *VhdsGenerator) Generate(proxy *model.Proxy, w *model.WatchedResource,
	req *model.PushRequest,
) (model.Resources, model.XdsLogDetails, error) {
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
