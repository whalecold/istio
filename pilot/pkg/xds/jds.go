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

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/jsonpb"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/util/protoconv"
	"istio.io/istio/pkg/config/schema/gvk"
	"istio.io/istio/pkg/config/schema/kind"
	"istio.io/istio/pkg/jdsapi/istio.io/api/v3alpha1"
	"istio.io/istio/pkg/util/sets"
)

type JdsGenerator struct {
	Server *DiscoveryServer
}

var _ model.XdsResourceGenerator = &JdsGenerator{}

const (
	// ServiceFormatLength service format name.namespace
	// and if split by '.', will convert to 2 parts
	ServiceFormatLength = 2
)

// ServiceList store service list
// there may be more than one service in namespace
// format: namespace -> name set
func buildServiceList(resourceNames []string) map[string]sets.String {
	if len(resourceNames) == 0 {
		return nil
	}

	res := map[string]sets.String{}
	for _, r := range resourceNames {
		svc := strings.Split(r, ".")
		if len(svc) != ServiceFormatLength {
			continue
		}
		if res[svc[1]] == nil {
			res[svc[1]] = sets.New(svc[0])
			continue
		}

		res[svc[1]].Insert(svc[0])
	}
	return res
}

func isQueryResource(svcList map[string]sets.String, namespace, name string) bool {
	// user not define resourceNames, return all
	if svcList == nil {
		return true
	}
	nameMap, ok := svcList[namespace]
	if !ok {
		return false
	}

	_, ok = nameMap[name]
	return ok
}

const (
	// LimitingConfigKey ...
	LimitingConfigKey = "MseRateLimit"
)

func (j *JdsGenerator) needPush(updates model.XdsUpdates) bool {
	if len(updates) == 0 {
		return true
	}

	for config := range updates {
		if config.Kind == kind.RateLimit {
			return true
		}
	}
	return false
}

func (j *JdsGenerator) Generate(_ *model.Proxy, w *model.WatchedResource, req *model.PushRequest) (model.Resources, model.XdsLogDetails, error) {
	if !j.needPush(req.ConfigsUpdated) {
		return nil, model.DefaultXdsLogDetails, nil
	}

	var rateLimits model.Resources
	// get java configuration configmap, if user not set ResourceNames, return all
	serviceList := buildServiceList(w.ResourceNames)

	cmList, err := j.Server.Env.ConfigStore.List(gvk.RateLimit, "")
	if err != nil {
		return nil, model.DefaultXdsLogDetails, err
	}

	for _, jc := range cmList {
		if !isQueryResource(serviceList, jc.Namespace, jc.Name) {
			continue
		}

		data, ok := jc.Spec.(map[string]string)
		if !ok {
			continue
		}

		mseRateLimit, ok := data[LimitingConfigKey]
		if !ok {
			continue
		}

		c := &v3alpha1.Configuration{}
		if err := jsonpb.UnmarshalString(mseRateLimit, c); err != nil {
			return nil, model.DefaultXdsLogDetails, err
		}

		c.Name = jc.Name + "." + jc.Namespace
		rateLimits = append(rateLimits, &discovery.Resource{
			Resource: protoconv.MessageToAny(c),
		})
	}

	return rateLimits, model.DefaultXdsLogDetails, nil
}
