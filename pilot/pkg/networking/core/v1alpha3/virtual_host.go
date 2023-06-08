package v1alpha3

import (
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
)

// BuildHTTPRoutes produces a list of routes for the proxy
func (configgen *ConfigGeneratorImpl) BuildVirtualHosts(
	node *model.Proxy,
	req *model.PushRequest,
	virtualHostNames []string,
) ([]*discovery.Resource, model.XdsLogDetails) {
	var vhdsConfigurations model.Resources

	efw := req.Push.EnvoyFilters(node)
	switch node.Type {
	case model.SidecarProxy:
		envoyfilterKeys := efw.KeysApplyingTo(
			networking.EnvoyFilter_VIRTUAL_HOST,
			networking.EnvoyFilter_HTTP_ROUTE,
		)
		for _, vhdsName := range virtualHostNames {
			vhds, _ := configgen.buildSidecarOutboundVirtualHosts(node, req, vhdsName, nil, efw, envoyfilterKeys)
			if vhds == nil {
				continue
			}
			vhdsConfigurations = append(vhdsConfigurations, vhds)
		}
	case model.Router:
		// TODO to be implemented.
	}
	return vhdsConfigurations, model.DefaultXdsLogDetails
}
