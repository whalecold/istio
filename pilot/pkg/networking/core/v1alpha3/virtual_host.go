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

package v1alpha3

import (
	"fmt"
	"strconv"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/util/errors"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/core/v1alpha3/envoyfilter"
	istio_route "istio.io/istio/pilot/pkg/networking/core/v1alpha3/route"
	"istio.io/istio/pilot/pkg/networking/util"
	"istio.io/istio/pilot/pkg/util/protoconv"
	"istio.io/istio/pkg/proto"
)

type vhdsRequest struct {
	resourceName string
	vhdsName     string
	vhdsDomain   string
}

// BuildHTTPRoutes produces a list of routes for the proxy
func (configgen *ConfigGeneratorImpl) BuildVirtualHosts(
	node *model.Proxy,
	req *model.PushRequest,
	resourceNames []string,
) ([]*discovery.Resource, model.DeletedResources, model.XdsLogDetails) {
	var vhdsConfigurations model.Resources
	var deletedConfigurations model.DeletedResources
	efw := req.Push.EnvoyFilters(node)
	errs := make([]error, 0, len(resourceNames))
	vhdsRequests := make(map[int][]*vhdsRequest)
	switch node.Type {
	case model.SidecarProxy:
		envoyfilterKeys := efw.KeysApplyingTo(
			networkingv1alpha3.EnvoyFilter_VIRTUAL_HOST,
			networkingv1alpha3.EnvoyFilter_HTTP_ROUTE,
		)

		for _, resourceName := range resourceNames {
			listenerPort, vhdsName, vhdsDomain, err := parseVirtualHostResourceName(resourceName)
			if err != nil {
				deletedConfigurations = append(deletedConfigurations, resourceName)
				errs = append(errs, err)
				continue
			}
			vhdsRequests[listenerPort] = append(vhdsRequests[listenerPort], &vhdsRequest{
				resourceName: resourceName,
				vhdsName:     vhdsName,
				vhdsDomain:   vhdsDomain,
			})
		}

		for port, reqs := range vhdsRequests {
			vhds := buildSidecarOutboundVirtualHosts(node, req, port, reqs, efw, envoyfilterKeys)
			if len(vhds) != 0 {
				vhdsConfigurations = append(vhdsConfigurations, vhds...)
			}
		}

	case model.Router:
		// TODO not-implemented.
	}
	var info string
	if len(errs) != 0 {
		info = errors.NewAggregate(errs).Error()
	}
	return vhdsConfigurations, deletedConfigurations, model.XdsLogDetails{
		AdditionalInfo: info,
	}
}

// buildSidecarOutboundVirtualHosts builds an outbound HTTP Route for sidecar.
// Based on port, will determine all virtual hosts that listen on the port.
func buildSidecarOutboundVirtualHosts(
	node *model.Proxy,
	req *model.PushRequest,
	listenerPort int,
	resources []*vhdsRequest,
	efw *model.EnvoyFilterWrapper,
	efKeys []string,
) []*discovery.Resource {

	routeName := strconv.Itoa(listenerPort)
	var out []*discovery.Resource

	vhosts, _, _ := BuildSidecarOutboundVirtualHosts(node, req.Push, routeName, listenerPort, efKeys, &model.DisabledCache{})

	for _, resource := range resources {
		// TODO
		// 1. use single function or reuse the old one.
		// 2. remove the vhds whose correspond route has been deleted.
		var virtualHost *route.VirtualHost
		for _, vh := range vhosts {
			for _, domain := range vh.Domains {
				if domain == resource.vhdsDomain {
					virtualHost = vh
					break
				}
			}
		}

		domains := make([]string, 1, 2)
		domains[0] = resource.vhdsDomain
		// the vhdsName may be equal with vhdsDomain when port is 80 and it can not be set port explicitly.
		if listenerPort != 80 && resource.vhdsDomain != resource.vhdsName {
			domains = append(domains, resource.vhdsName)
		}

		if virtualHost == nil {
			// build default policy.
			// FIXME: how to distinguish the virtualHost is external or should be deleted.
			virtualHost = buildDefaultVirtualHost(node, resource.vhdsName, domains)
		} else {
			// Only unique values for domains are permitted set in one route.
			// conditions:
			// 1. curl productpage.bookinfo:9080 productpage.bookinfo.svc.cluster.local:9080 separatly,
			//  it will generate two Passthrough vhds if there is no correspond service.
			// 2. after creating correspond service, the domains in the two virtualHost would be duplicate.
			//  Envoy forbids to exist duplicate domain in one route.
			virtualHost.Domains = domains
			virtualHost.Name = resource.vhdsName
		}

		rc := &route.RouteConfiguration{
			Name:             strconv.Itoa(listenerPort),
			ValidateClusters: proto.BoolFalse,
			VirtualHosts:     []*route.VirtualHost{virtualHost},
		}

		// apply envoy filter patches
		// TODO use single function or reuse the old rds patches.
		envoyfilter.ApplyRouteConfigurationPatches(networkingv1alpha3.EnvoyFilter_SIDECAR_OUTBOUND, node, efw, rc)

		out = append(out, &discovery.Resource{
			Name:     resource.resourceName,
			Resource: protoconv.MessageToAny(virtualHost),
		})
	}
	return out
}

// parseVirtualHostResourceName the format is routeName/domain:routeName
func parseVirtualHostResourceName(resourceName string) (int, string, string, error) {
	// not support wildcard character
	first := strings.IndexRune(resourceName, '/')
	if first == -1 {
		return 0, "", "", fmt.Errorf("invalid format resource name %s", resourceName)
	}
	var vhdsName, vhdsDomain string
	routeName := resourceName[:first]
	last := strings.Index(resourceName, ":")
	if last != -1 && first+1 >= last {
		return 0, "", "", fmt.Errorf("invalid format resource name %s", resourceName)
	}
	if last == -1 {
		vhdsName = resourceName[first+1:]
		vhdsDomain = resourceName[first+1:]
	} else {
		vhdsName = resourceName[first+1:]
		vhdsDomain = resourceName[first+1 : last]
	}
	port, err := strconv.Atoi(routeName)
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid format resource name %s", resourceName)
	}
	// port vhdsName domain
	return port, vhdsName, vhdsDomain, nil
}

func buildDefaultVirtualHost(node *model.Proxy, name string, domains []string) *route.VirtualHost {
	if util.IsAllowAnyOutbound(node) {
		egressCluster := util.PassthroughCluster
		notimeout := durationpb.New(0)

		// no need to check for nil value as the previous if check has checked
		if node.SidecarScope.OutboundTrafficPolicy.EgressProxy != nil {
			// user has provided an explicit destination for all the unknown traffic.
			// build a cluster out of this destination
			egressCluster = istio_route.GetDestinationCluster(node.SidecarScope.OutboundTrafficPolicy.EgressProxy,
				nil, 0)
		}

		routeAction := &route.RouteAction{
			ClusterSpecifier: &route.RouteAction_Cluster{Cluster: egressCluster},
			// Disable timeout instead of assuming some defaults.
			Timeout: notimeout,
			// Use deprecated value for now as the replacement MaxStreamDuration has some regressions.
			// nolint: staticcheck
			MaxGrpcTimeout: notimeout,
		}

		return &route.VirtualHost{
			Name:    name,
			Domains: domains,
			Routes: []*route.Route{
				{
					Name: util.Passthrough,
					Match: &route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"},
					},
					Action: &route.Route_Route{
						Route: routeAction,
					},
				},
			},
			IncludeRequestAttemptCount: true,
		}
	}

	return &route.VirtualHost{
		Name:    name,
		Domains: domains,
		Routes: []*route.Route{
			{
				Name: util.BlackHole,
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"},
				},
				Action: &route.Route_DirectResponse{
					DirectResponse: &route.DirectResponseAction{
						Status: 502,
					},
				},
			},
		},
		IncludeRequestAttemptCount: true,
	}
}
