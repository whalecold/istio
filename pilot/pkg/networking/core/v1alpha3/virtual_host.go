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
) ([]*discovery.Resource, model.XdsLogDetails) {
	var vhdsConfigurations model.Resources
	efw := req.Push.EnvoyFilters(node)
	errs := make([]error, 0, len(resourceNames))
	switch node.Type {
	case model.SidecarProxy:
		envoyfilterKeys := efw.KeysApplyingTo(
			networking.EnvoyFilter_VIRTUAL_HOST,
			networking.EnvoyFilter_HTTP_ROUTE,
		)
		for _, resourceName := range resourceNames {
			vhds, err := buildSidecarOutboundVirtualHosts(node, req, resourceName, efw, envoyfilterKeys)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			vhdsConfigurations = append(vhdsConfigurations, vhds)
		}
	case model.Router:
		// TODO to be implemented.
	}
	return vhdsConfigurations, model.XdsLogDetails{
		AdditionalInfo: errors.NewAggregate(errs).Error(),
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
) (*discovery.Resource, error) {
	listenerPort, vhdsName, vhdsDomain, err := parseVirtualHostResourceName(resourceName)
	if err != nil {
		return nil, err
	}

	// TODO use single function or reuse the old rds patches.
	vhosts, _, _ := BuildSidecarOutboundVirtualHosts(node, req.Push, strconv.Itoa(listenerPort), listenerPort, efKeys, &model.DisabledCache{})
	var virtualHost *route.VirtualHost
	for _, vh := range vhosts {
		for _, domain := range vh.Domains {
			if domain == vhdsDomain {
				virtualHost = vh
				break
			}
		}
	}
	if virtualHost == nil {
		// build default policy.
		virtualHost = buildDefaultVirtualHost(node, vhdsName, []string{vhdsDomain, vhdsName})
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
	return resource, nil
}

// parseVirtualHostResourceName the format is routeName/domain:routeName
func parseVirtualHostResourceName(resourceName string) (int, string, string, error) {
	err := fmt.Errorf("invalid format resource name %s", resourceName)
	first := strings.IndexRune(resourceName, '/')
	if first == -1 {
		return 0, "", "", err
	}
	routeName := resourceName[:first]
	last := strings.Index(resourceName, ":"+routeName)
	if last != -1 || first+1 >= last {
		return 0, "", "", err
	}
	port, err := strconv.Atoi(routeName)
	if err != nil {
		return 0, "", "", err
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
