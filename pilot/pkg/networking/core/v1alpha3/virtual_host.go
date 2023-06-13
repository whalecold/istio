package v1alpha3

import (
	"fmt"
	"strconv"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	networking "istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/util/errors"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/core/v1alpha3/envoyfilter"
	istio_route "istio.io/istio/pilot/pkg/networking/core/v1alpha3/route"
	"istio.io/istio/pilot/pkg/networking/util"
	"istio.io/istio/pilot/pkg/util/protoconv"
	"istio.io/istio/pkg/proto"
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
			networking.EnvoyFilter_VIRTUAL_HOST,
			networking.EnvoyFilter_HTTP_ROUTE,
		)

		for _, resourceName := range resourceNames {
			// not support wildcard char
			if resourceName == "*" {
				deletedConfigurations = append(deletedConfigurations, resourceName)
				continue
			}

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
		// TODO to be implemented.
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

	listenerPort, vhdsName, vhdsDomain, err := parseVirtualHostResourceName(resourceName)
	if err != nil {
		return nil, false, err
	}

	routeName := strconv.Itoa(listenerPort)

	// TODO
	// 1. use single function or reuse the old one.
	// 2. remove the vhds whose correspond route has been deleted.
	vhosts, _, _ := BuildSidecarOutboundVirtualHosts(node, req.Push, routeName, listenerPort, efKeys, &model.DisabledCache{}, nil)
	var virtualHost *route.VirtualHost
	for _, vh := range vhosts {
		for _, domain := range vh.Domains {
			if domain == vhdsDomain {
				virtualHost = vh
				break
			}
		}
	}

	domains := []string{vhdsDomain, vhdsName}
	if virtualHost == nil {
		// build default policy.
		// FIXME: how to distinguish the virtualHost is external or should be deleted.
		virtualHost = buildDefaultVirtualHost(node, vhdsName, domains)
	} else {
		// Only unique values for domains are permitted set in one route.
		// conditions:
		// 1. curl productpage.bookinfo:9080 productpage.bookinfo.svc.cluster.local:9080 separatly, it will generate two Passthrough vhds
		//	if there is no correspond service.
		// 2. after creating correspond service, the domains in the two virtualHost would be duplicate. Envoy forbids to exist duplicate domain in one route.
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
	envoyfilter.ApplyRouteConfigurationPatches(networking.EnvoyFilter_SIDECAR_OUTBOUND, node, efw, out)

	resource := &discovery.Resource{
		Name:     resourceName,
		Resource: protoconv.MessageToAny(virtualHost),
	}
	return resource, false, nil
}

// parseVirtualHostResourceName the format is routeName/domain:routeName
func parseVirtualHostResourceName(resourceName string) (int, string, string, error) {
	first := strings.IndexRune(resourceName, '/')
	if first == -1 {
		return 0, "", "", fmt.Errorf("invalid format resource name %s", resourceName)
	}
	routeName := resourceName[:first]
	last := strings.Index(resourceName, ":"+routeName)
	if last == -1 || first+1 >= last {
		return 0, "", "", fmt.Errorf("invalid format resource name %s", resourceName)
	}
	port, err := strconv.Atoi(routeName)
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid format resource name %s", resourceName)
	}
	// port vhdsName domain
	return port, resourceName[first+1:], resourceName[first+1 : last], nil
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
