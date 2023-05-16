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

package adsc

import (
	"fmt"
	"sync"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	utilserrors "k8s.io/apimachinery/pkg/util/errors"

	"istio.io/istio/pilot/pkg/model"
	mcpaggregate "istio.io/istio/pkg/adsc/aggregate"
	"istio.io/istio/pkg/adsc/mcpdiscovery"
	"istio.io/istio/pkg/adsc/server"
	"istio.io/istio/pkg/adsc/xdsclient"
	"istio.io/istio/pkg/backoff"
	"istio.io/istio/pkg/config/schema/collections"
	"istio.io/istio/pkg/util/sets"
	"istio.io/pkg/env"
)

var (
	InstanceIPVar   = env.RegisterStringVar("INSTANCE_IP", "", "")
	PodNameVar      = env.RegisterStringVar("POD_NAME", "", "pod name")
	PodNamespaceVar = env.RegisterStringVar("POD_NAMESPACE", "", "the pod name space")
	QueryServerPort = env.RegisterIntVar("MCP_SERVER_PORT", 15081, "the port for mcp query server.")
)

// ConfigMultiADSCInitialRequests the multi adsc init requests.
func ConfigMultiADSCInitialRequests() []*discovery.DeltaDiscoveryRequest {
	out := make([]*discovery.DeltaDiscoveryRequest, 0, len(collections.PilotMCP.All())+1)
	for _, sch := range collections.PilotMCP.All() {
		out = append(out, &discovery.DeltaDiscoveryRequest{
			TypeUrl: sch.Resource().GroupVersionKind().String(),
		})
	}
	return out
}

// MultiADSC implements a basic client for MultiADS, for use in stress tests and tools
// or libraries that need to connect to Istio pilot or other ADS servers.
type MultiADSC struct {
	adscLock sync.Mutex
	// adscs maintains the connections to mcp servers.
	adscs map[string]xdsclient.XDSClient

	cfg *Config

	// Retrieved configurations can be stored using the common istio model interface.
	store mcpaggregate.ConfigStoreCache

	buildXDSClient func(string, string, string, []*discovery.DeltaDiscoveryRequest) (xdsclient.XDSClient, error)
	// nodeID the identity of the node.
	nodeID string

	svr server.Server
}

func nodeID() string {
	return fmt.Sprintf("mcp-client~%s/%s~%s", PodNameVar.Get(), PodNamespaceVar.Get(), InstanceIPVar.Get())
}

// NewMultiADSC creates a new multi ADSC manager, maintaining the connections to XDS servers.
// Will:
// - connect to the XDS server specified in ProxyConfig
// - send initial request for watched resources
// - wait for response from XDS server
// - on success, start a background thread to maintain the xDSClient, with exp. backoff.
func NewMultiADSC(cfg *Config) (*MultiADSC, error) {
	if cfg.BackoffPolicy == nil {
		cfg.BackoffPolicy = backoff.NewExponentialBackOff(backoff.DefaultOption())
	}
	store := mcpaggregate.MakeCache(collections.PilotMCP)
	md := &MultiADSC{
		cfg:            cfg,
		adscs:          make(map[string]xdsclient.XDSClient),
		nodeID:         nodeID(),
		store:          store,
		buildXDSClient: xdsclient.New,
		svr:            server.New(store, QueryServerPort.Get()),
	}
	return md, nil
}

// Serve the implemention of interface.
func (ma *MultiADSC) Serve() {
	if err := ma.svr.Serve(); err != nil {
		panic(err)
	}
}

// GetStore return store.
func (ma *MultiADSC) GetStore() model.ConfigStoreController {
	return ma.store
}

// HasSynced returns true if MCP configs have synced
func (ma *MultiADSC) HasSynced() bool {
	if len(ma.adscs) == 0 || ma.cfg == nil || len(ma.cfg.InitialDeltaDiscoveryRequests) == 0 {
		return true
	}
	return ma.store.HasSynced()
}

// OnServersUpdate the handler when services updates.
func (ma *MultiADSC) OnServersUpdate(servers []*mcpdiscovery.McpServer) error {
	// check the duplicated connection url.
	newClients := sets.New[string]()
	for _, server := range servers {
		if newClients.Contains(server.ID) {
			return fmt.Errorf("duplicated val %s for key %s", server.ID, server.Address)
		}
		newClients.Insert(server.ID)
	}

	var added, removed []xdsclient.XDSClient
	var errors []error

	ma.adscLock.Lock()
	for _, server := range servers {
		con, ok := ma.adscs[server.ID]
		if !ok || con.GetURL() != server.Address {
			// delete the xDSClient with different url.
			if con != nil {
				removed = append(removed, con)
			}

			con, err := ma.buildXDSClient(server.ID, server.Address, ma.nodeID, ma.cfg.InitialDeltaDiscoveryRequests)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			ma.adscs[server.ID] = con
			added = append(added, con)
		}
	}

	for k, con := range ma.adscs {
		_, ok := newClients[k]
		if !ok {
			delete(ma.adscs, k)
			removed = append(removed, con)
		}
	}
	ma.adscLock.Unlock()

	for _, con := range removed {
		adscLog.Infof("Remove the old xDSClient (%s/%s)", con.GetID(), con.GetURL())
		ma.store.RemoveStore(con.GetID())
		con.Close()
	}
	for i := range added {
		con := added[i]
		adscLog.Infof("Start establish the connection with (%s/%s)", con.GetID(), con.GetURL())
		ma.store.AddStore(con.GetID(), con.GetStore())
		go func() {
			err := con.Run()
			if err != nil {
				adscLog.Errorf("(%s/%s) run falied %v", con.GetID(), con.GetURL(), err)
			}
		}()
	}

	if len(errors) == 0 {
		ma.store.Initialized()
	}

	return utilserrors.NewAggregate(errors)
}
