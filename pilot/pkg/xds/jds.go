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

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/jsonpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/util/protoconv"
	"istio.io/istio/pkg/config"
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
// format of resourceName is name.namespace
func buildWatchedResources(resourceNames []string) map[string]sets.String {
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

func isWatchedResource(watches map[string]sets.String, namespace, name string) bool {
	// user does not specify resourceNames, return all
	if len(watches) == 0 {
		return true
	}
	names, ok := watches[namespace]
	if !ok {
		return false
	}

	_, ok = names[name]
	return ok
}

const (
	// MseConfigurationKey ...
	MseConfigurationKey      = "MseConfiguration"
	mseGlobalConfigName      = "global-jds-configuration"
	mseGlobalConfigNamespace = "istio-system"
	mseGlobaleLatestKey      = "mse-global-latest-configuration"
	mseGlobaleLastKey        = "mse-global-last-configuration"
)

func (j *JdsGenerator) needPush(updates model.XdsUpdates) bool {
	log.Info("needPush, ", updates)
	if len(updates) == 0 {
		return true
	}

	for config := range updates {
		if config.Kind == kind.MseConfiguration {
			return true
		}
	}
	return false
}

func getGlobalConfiguration(global *config.Config) (*v3alpha1.GlobalConfiguration, bool, error) {
	if global == nil {
		return nil, false, nil
	}
	data, ok := global.Spec.(map[string]string)
	if !ok {
		return nil, false, fmt.Errorf("error type for jds globale configuration")
	}
	latestConfig := data[mseGlobaleLatestKey]
	changed := latestConfig != data[mseGlobaleLastKey]
	c := &v3alpha1.GlobalConfiguration{}
	err := jsonpb.UnmarshalString(latestConfig, c)
	if err != nil {
		return c, false, fmt.Errorf("unmarshal jds global configuration error:%v", err)
	}
	return c, changed, nil
}

func (j *JdsGenerator) Generate(_ *model.Proxy, w *model.WatchedResource, req *model.PushRequest) (model.Resources, model.XdsLogDetails, error) {
	if !j.needPush(req.ConfigsUpdated) {
		return nil, model.DefaultXdsLogDetails, nil
	}

	resources := make([]*discovery.Resource, 0)
	// get java configuration configmap, if user not set ResourceNames, return all
	watches := buildWatchedResources(w.ResourceNames)

	cmList, err := j.Server.Env.ConfigStore.List(gvk.MseConfiguration, metav1.NamespaceAll)
	if err != nil {
		return nil, model.DefaultXdsLogDetails, err
	}

	globalConfigmap := j.Server.Env.ConfigStore.Get(gvk.MseConfiguration, mseGlobalConfigName, mseGlobalConfigNamespace)
	gc, changed, err := getGlobalConfiguration(globalConfigmap)
	if err != nil {
		log.Warnf("parse global mse configuration error: %v", err)
	}

	for _, cm := range cmList {
		if !isWatchedResource(watches, cm.Namespace, cm.Name) && !changed {
			continue
		}

		data, ok := cm.Spec.(map[string]string)
		if !ok {
			continue
		}

		val, ok := data[MseConfigurationKey]
		if !ok {
			continue
		}
		c := &v3alpha1.Configuration{}
		if err = jsonpb.UnmarshalString(val, c); err != nil {
			return nil, model.DefaultXdsLogDetails, err
		}
		c.Name = cm.Name + "." + cm.Namespace
		c.Global = gc
		resources = append(resources, &discovery.Resource{
			Resource: protoconv.MessageToAny(c),
		})
	}

	return resources, model.DefaultXdsLogDetails, nil
}
