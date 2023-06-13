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
	"sort"
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
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/host"
	"istio.io/istio/pkg/config/protocol"
	"istio.io/istio/pkg/proto"
	"istio.io/istio/pkg/util/sets"
)

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
	switch node.Type {
	case model.SidecarProxy:
		envoyfilterKeys := efw.KeysApplyingTo(
			networkingv1alpha3.EnvoyFilter_VIRTUAL_HOST,
			networkingv1alpha3.EnvoyFilter_HTTP_ROUTE,
		)

		for _, resourceName := range resourceNames {
			vhds, shouldDeleted, err := buildSidecarOutboundVirtualHosts(node, req, resourceName, efw, envoyfilterKeys)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if vhds != nil {
				vhdsConfigurations = append(vhdsConfigurations, vhds)
			}
			if shouldDeleted {
				deletedConfigurations = append(deletedConfigurations, resourceName)
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
	resourceName string,
	efw *model.EnvoyFilterWrapper,
	efKeys []string,
) (*discovery.Resource, bool, error) {
	// not support wildcard character
	if resourceName == "*" {
		return nil, true, nil
	}

	listenerPort, vhdsName, vhdsDomain, err := parseVirtualHostResourceName(resourceName)
	if err != nil {
		return nil, false, err
	}

	routeName := strconv.Itoa(listenerPort)

	// TODO
	// 1. use single function or reuse the old one.
	// 2. remove the vhds whose correspond route has been deleted.
	vhosts := buildSidecarOutboundVirtualHostsBeta(node, req.Push, routeName, listenerPort, efKeys)
	var virtualHost *route.VirtualHost
	for _, vh := range vhosts {
		for _, domain := range vh.Domains {
			if domain == vhdsDomain {
				virtualHost = vh
				break
			}
		}
	}

	domains := make([]string, 1, 2)
	domains[0] = vhdsDomain
	// the vhdsName may be equal with vhdsDomain when port is 80 and it can not be set port explicitly.
	if listenerPort != 80 && vhdsDomain != vhdsName {
		domains = append(domains, vhdsName)
	}

	if virtualHost == nil {
		// build default policy.
		// FIXME: how to distinguish the virtualHost is external or should be deleted.
		virtualHost = buildDefaultVirtualHost(node, vhdsName, domains)
	} else {
		// Only unique values for domains are permitted set in one route.
		// conditions:
		// 1. curl productpage.bookinfo:9080 productpage.bookinfo.svc.cluster.local:9080 separatly,
		//  it will generate two Passthrough vhds if there is no correspond service.
		// 2. after creating correspond service, the domains in the two virtualHost would be duplicate.
		//  Envoy forbids to exist duplicate domain in one route.
		virtualHost.Domains = domains
		virtualHost.Name = vhdsName
	}

	out := &route.RouteConfiguration{
		Name:             strconv.Itoa(listenerPort),
		ValidateClusters: proto.BoolFalse,
		VirtualHosts:     []*route.VirtualHost{virtualHost},
	}

	// apply envoy filter patches
	// TODO use single function or reuse the old rds patches.
	envoyfilter.ApplyRouteConfigurationPatches(networkingv1alpha3.EnvoyFilter_SIDECAR_OUTBOUND, node, efw, out)

	resource := &discovery.Resource{
		Name:     resourceName,
		Resource: protoconv.MessageToAny(virtualHost),
	}
	return resource, false, nil
}

func buildSidecarOutboundVirtualHostsBeta(node *model.Proxy, push *model.PushContext,
	routeName string,
	listenerPort int,
	efKeys []string,
) []*route.VirtualHost {
	var virtualServices []config.Config
	var services []*model.Service

	// Get the services from the egress listener.  When sniffing is enabled, we send
	// route name as foo.bar.com:8080 which is going to match against the wildcard
	// egress listener only. A route with sniffing would not have been generated if there
	// was a sidecar with explicit port (and hence protocol declaration). A route with
	// sniffing is generated only in the case of the catch all egress listener.
	egressListener := node.SidecarScope.GetEgressListenerForRDS(listenerPort, routeName)
	// We should never be getting a nil egress listener because the code that setup this RDS
	// call obviously saw an egress listener
	if egressListener == nil {
		return nil
	}

	services = egressListener.Services()
	// To maintain correctness, we should only use the virtualservices for
	// this listener and not all virtual services accessible to this proxy.
	virtualServices = egressListener.VirtualServices()

	// When generating RDS for ports created via the SidecarScope, we treat ports as HTTP proxy style ports
	// if ports protocol is HTTP_PROXY.
	if egressListener.IstioListener != nil && egressListener.IstioListener.Port != nil &&
		protocol.Parse(egressListener.IstioListener.Port.Protocol) == protocol.HTTP_PROXY {
		listenerPort = 0
	}

	servicesByName := make(map[host.Name]*model.Service)
	for _, svc := range services {
		if listenerPort == 0 {
			// Take all ports when listen port is 0 (http_proxy or uds)
			// Expect virtualServices to resolve to right port
			servicesByName[svc.Hostname] = svc
		} else if svcPort, exists := svc.Ports.GetByPort(listenerPort); exists {
			servicesByName[svc.Hostname] = &model.Service{
				Hostname:       svc.Hostname,
				DefaultAddress: svc.GetAddressForProxy(node),
				MeshExternal:   svc.MeshExternal,
				Resolution:     svc.Resolution,
				Ports:          []*model.Port{svcPort},
				Attributes: model.ServiceAttributes{
					Namespace:       svc.Attributes.Namespace,
					ServiceRegistry: svc.Attributes.ServiceRegistry,
				},
			}
		}
	}

	if listenerPort > 0 {
		services = make([]*model.Service, 0, len(servicesByName))
		// sort services
		for _, svc := range servicesByName {
			services = append(services, svc)
		}
		sort.SliceStable(services, func(i, j int) bool {
			return services[i].Hostname <= services[j].Hostname
		})
	}

	// This is hack to keep consistent with previous behavior.
	if listenerPort != 80 {
		// only select virtualServices that matches a service
		virtualServices = selectVirtualServices(virtualServices, servicesByName)
	}
	// Get list of virtual services bound to the mesh gateway
	virtualHostWrappers := istio_route.BuildSidecarVirtualHostWrapper(nil, node, push, servicesByName, virtualServices, listenerPort)

	vHostPortMap := make(map[int][]*route.VirtualHost)
	vhosts := sets.String{}
	vhdomains := sets.String{}
	knownFQDN := sets.String{}

	buildVirtualHost := func(hostname string, vhwrapper istio_route.VirtualHostWrapper, svc *model.Service) *route.VirtualHost {
		name := util.DomainName(hostname, vhwrapper.Port)
		if vhosts.InsertContains(name) {
			// This means this virtual host has caused duplicate virtual host name.
			var msg string
			if svc == nil {
				msg = fmt.Sprintf("duplicate domain from virtual service: %s", name)
			} else {
				msg = fmt.Sprintf("duplicate domain from service: %s", name)
			}
			push.AddMetric(model.DuplicatedDomains, name, node.ID, msg)
			return nil
		}
		var domains []string
		var altHosts []string
		if svc == nil {
			if SidecarIgnorePort(node) {
				domains = []string{util.IPv6Compliant(hostname)}
			} else {
				domains = []string{util.IPv6Compliant(hostname), name}
			}
		} else {
			domains, altHosts = generateVirtualHostDomains(svc, listenerPort, vhwrapper.Port, node)
		}
		dl := len(domains)
		domains = dedupeDomains(domains, vhdomains, altHosts, knownFQDN)
		if dl != len(domains) {
			var msg string
			if svc == nil {
				msg = fmt.Sprintf("duplicate domain from virtual service: %s", name)
			} else {
				msg = fmt.Sprintf("duplicate domain from service: %s", name)
			}
			// This means this virtual host has caused duplicate virtual host domain.
			push.AddMetric(model.DuplicatedDomains, name, node.ID, msg)
		}
		if len(domains) > 0 {
			return &route.VirtualHost{
				Name:                       name,
				Domains:                    domains,
				Routes:                     vhwrapper.Routes,
				IncludeRequestAttemptCount: true,
			}
		}

		return nil
	}

	for _, virtualHostWrapper := range virtualHostWrappers {
		for _, svc := range virtualHostWrapper.Services {
			name := util.DomainName(string(svc.Hostname), virtualHostWrapper.Port)
			knownFQDN.InsertAll(name, string(svc.Hostname))
		}
	}

	for _, virtualHostWrapper := range virtualHostWrappers {
		// If none of the routes matched by source, skip this virtual host
		if len(virtualHostWrapper.Routes) == 0 {
			continue
		}
		virtualHosts := make([]*route.VirtualHost, 0, len(virtualHostWrapper.VirtualServiceHosts)+len(virtualHostWrapper.Services))

		for _, hostname := range virtualHostWrapper.VirtualServiceHosts {
			if vhost := buildVirtualHost(hostname, virtualHostWrapper, nil); vhost != nil {
				virtualHosts = append(virtualHosts, vhost)
			}
		}

		for _, svc := range virtualHostWrapper.Services {
			if vhost := buildVirtualHost(string(svc.Hostname), virtualHostWrapper, svc); vhost != nil {
				virtualHosts = append(virtualHosts, vhost)
			}
		}
		vHostPortMap[virtualHostWrapper.Port] = append(vHostPortMap[virtualHostWrapper.Port], virtualHosts...)
	}

	var out []*route.VirtualHost
	if listenerPort == 0 {
		out = mergeAllVirtualHosts(vHostPortMap)
	} else {
		out = vHostPortMap[listenerPort]
	}

	return out
}

// parseVirtualHostResourceName the format is routeName/domain:routeName
func parseVirtualHostResourceName(resourceName string) (int, string, string, error) {
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
