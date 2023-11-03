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

package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	networking "istio.io/api/networking/v1alpha3"
	v3 "istio.io/istio/pilot/pkg/xds/v3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/host"
	"istio.io/istio/pkg/config/protocol"
)

func onDemandTrimmedSidecarName(name string) string {
	return "on-demand-trimmed-" + name
}

func (node *Proxy) trimSidecarScopeByOnDemandHosts(ps *PushContext) {
	if node.SidecarScope == nil {
		return
	}

	node.RLock()
	watched, ok := node.WatchedResources[v3.VirtualHostType]
	node.RUnlock()
	if !ok || len(watched.ResourceNames) == 0 {
		// Just left the `OnDemandSidecarScope` as nil if
		// no virtual hosts discovery requests have received.
		return
	}

	baseSidecar := node.SidecarScope.Sidecar
	if baseSidecar == nil {
		baseSidecar = &networking.Sidecar{
			Egress: []*networking.IstioEgressListener{{
				Hosts: []string{"*/*"},
			}},
		}
		if ps.Mesh.OutboundTrafficPolicy != nil {
			baseSidecar.OutboundTrafficPolicy = &networking.OutboundTrafficPolicy{
				Mode: networking.OutboundTrafficPolicy_Mode(ps.Mesh.OutboundTrafficPolicy.Mode),
			}
		}
	}

	hostsByPort, err := getVisableOnDemandHosts(watched.ResourceNames, node.DNSDomain,
		node.SidecarScope.servicesByHostname)
	if err != nil {
		log.Errorf("get visable on-demand host failed, fallback to use the entire scope, err: %s", err.Error())
		node.OnDemandSidecarScope = node.SidecarScope
		return
	}

	trimmedEgressListeners := trimSidecarEgress(baseSidecar.Egress, hostsByPort)
	trimmedSidecar := &config.Config{
		Meta: config.Meta{
			Name:      onDemandTrimmedSidecarName(node.SidecarScope.Name),
			Namespace: node.SidecarScope.Namespace,
		},
		Spec: &networking.Sidecar{
			WorkloadSelector:      baseSidecar.WorkloadSelector,
			Ingress:               baseSidecar.Ingress,
			Egress:                trimmedEgressListeners,
			OutboundTrafficPolicy: baseSidecar.OutboundTrafficPolicy,
		}}

	node.OnDemandSidecarScope = ConvertToSidecarScope(ps, trimmedSidecar,
		node.ConfigNamespace)
}

// ParseVirtualHostResourceName parse on-demand virtual hosts discovery requests.
// For service port with protocol sniffing enabled, routeName is at the format of FQDN:port
// Otherwise, the routeName is identical to port.
func ParseVirtualHostResourceName(resourceName string) (int, string, string, error) {
	// not support wildcard character
	sep := strings.LastIndexByte(resourceName, '/')
	if sep == -1 {
		return -1, "", "", fmt.Errorf("invalid format resource name %s", resourceName)
	}

	routeName := resourceName[:sep]
	vhdsName := resourceName[sep+1:]

	vhdsDomain, _, _ := strings.Cut(vhdsName, ":")
	// routeName may be sniffing.
	_, portStr, found := strings.Cut(routeName, ":")
	if !found {
		portStr = routeName
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid format resource name %s", resourceName)
	}

	// port, request host, host domain name.
	return port, vhdsName, vhdsDomain, nil
}

// trimSidecarEgress ...
func trimSidecarEgress(egress []*networking.IstioEgressListener, hostsByPort map[int][]string) []*networking.IstioEgressListener {
	if len(hostsByPort) == 0 {
		return []*networking.IstioEgressListener{{Hosts: []string{denyAll}}}
	}

	out := make([]*networking.IstioEgressListener, 0, len(egress))
	var listenerWithOmittedPort *networking.IstioEgressListener
	for _, e := range egress {
		if e.Port != nil && protocol.Parse(e.Port.Protocol) != protocol.HTTP {
			// Keep the ports with non-HTTP protocol
			// TODO(wangjian.pg 20230926):
			// 1. HTTPS, GRPC, HTTP2 can also work?
			// 2. Protocol with HTTP_PROXY can also be trimmed.
			out = append(out, e)
			continue
		}

		portNumber := int(e.Port.GetNumber())
		if portNumber == 0 {
			// listenerWithOmittedPort SHOULD be the last egress listener and used as the catch call semantic.
			if listenerWithOmittedPort != nil {
				log.Warnf("sidecar: the egress listener with empty port should be the last listener in the list")
			}
			listenerWithOmittedPort = e
			continue
		}

		if hosts, ok := hostsByPort[portNumber]; ok {
			trimmedListener := e.DeepCopy()
			trimmedListener.Hosts = hosts
			out = append(out, trimmedListener)
			delete(hostsByPort, portNumber)
		}
	}

	if len(hostsByPort) > 0 {
		trimmedListener := listenerWithOmittedPort.DeepCopy()
		trimmedListener.Hosts = nil
		for _, hosts := range hostsByPort {
			trimmedListener.Hosts = append(trimmedListener.Hosts, hosts...)
		}
		out = append(out, trimmedListener)
	}

	return out
}

func getVisableOnDemandHosts(onDemandHosts []string, dnsDomain string, visableServices map[host.Name]*Service) (map[int][]string, error) {
	var proxyCurrentNamespace, domainSuffix string
	if idx := strings.Index(dnsDomain, ".svc"); idx == -1 || idx < 1 {
		return nil, errors.Errorf("illegal dnsDomain %s", dnsDomain)
	} else {
		proxyCurrentNamespace = dnsDomain[:idx]
		domainSuffix = dnsDomain[idx:]
	}

	hostsByPort := make(map[int][]string)

	for _, r := range onDemandHosts {
		port, _, hostname, err := ParseVirtualHostResourceName(r)
		if err != nil {
			continue
		}

		shortName := hostname
		if strings.HasSuffix(hostname, ".svc") || strings.HasSuffix(hostname, domainSuffix) {
			shortName, _, _ = strings.Cut(hostname, ".svc")
		}

		var hostNamespace string
		if parts := strings.Split(shortName, "."); len(parts) == 2 {
			shortName, hostNamespace = parts[0], parts[1]
		} else if len(parts) == 1 {
			hostNamespace = proxyCurrentNamespace
		}

		visableAsKubeService := false
		if hostNamespace != "" {
			// The hostname maybe a k8s service name, convert it to FQDN since
			// hostname should in the format of "namespace/FQDN" according to the
			// specification of `networking.Sidecar.IstioEgressListeners.Hosts.`
			fqdn := shortName + "." + hostNamespace + domainSuffix
			if service, visable := visableServices[host.Name(fqdn)]; visable {
				for _, svcPort := range service.Ports {
					// TODO(wangjian.pg 20231030) do we need to check the Protocol of the svcPort?
					if svcPort.Port == port {
						hostsByPort[port] = append(hostsByPort[port], hostNamespace+"/"+fqdn)
						visableAsKubeService = true
						break
					}
				}
			}
		}

		// hostname specified by `ServiceEntry` SHOULD NOT conflict with the one of k8s service
		// and we take the k8s service over `ServiceEntry` here.
		if visableAsKubeService {
			continue
		}

		// hostname maybe a service from an external registry specified by `ServiceEntry`, e.g. foo.bar.remote.cluster
		if service, visable := visableServices[host.Name(hostname)]; visable {
			for _, svcPort := range service.Ports {
				if svcPort.Port == port {
					hostsByPort[port] = append(hostsByPort[port], service.Attributes.Namespace+"/"+hostname)
					break
				}
			}
		}
	}

	return hostsByPort, nil
}
