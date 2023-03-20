package adsc

import (
	"fmt"
	"sync"

	"github.com/cenkalti/backoff/v4"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"istio.io/pkg/env"
	utilserrors "k8s.io/apimachinery/pkg/util/errors"

	"istio.io/istio/pilot/pkg/model"
	mcpaggregate "istio.io/istio/pkg/adsc/aggregate"
	"istio.io/istio/pkg/adsc/httpserver"
	"istio.io/istio/pkg/adsc/mcpdiscovery"
	"istio.io/istio/pkg/adsc/xdsclient"
	"istio.io/istio/pkg/config/schema/collections"
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
		cfg.BackoffPolicy = backoff.NewExponentialBackOff()
	}
	md := &MultiADSC{
		cfg:            cfg,
		adscs:          make(map[string]xdsclient.XDSClient),
		nodeID:         nodeID(),
		store:          mcpaggregate.MakeCache(),
		buildXDSClient: xdsclient.New,
	}
	return md, nil
}

// RunQueryServer http query server
func (ma *MultiADSC) RunQueryServer() {
	port := QueryServerPort.Get()
	adscLog.Infof("listen mcp query server on port: %d", port)
	s := httpserver.NewQueryServer(ma.store, port)
	s.Run()
}

// GetStore return store.
func (ma *MultiADSC) GetStore() model.ConfigStoreCache {
	return ma.store
}

// HasSynced returns true if MCP configs have synced
func (ma *MultiADSC) HasSynced() bool {
	if len(ma.adscs) == 0 || ma.cfg == nil || len(ma.cfg.InitialDeltaDiscoveryRequests) == 0 {
		return true
	}
	return ma.store.HasSynced()
}

func (ma *MultiADSC) OnServersUpdate(servers []*mcpdiscovery.McpServer) error {
	// check the duplicated connection url.
	clientMap := make(map[string]bool)
	for _, val := range servers {
		if clientMap[val.ID] {
			return fmt.Errorf("duplicated val %s for key %s", val.ID, val.Address)
		}
		clientMap[val.ID] = true
	}

	var add, removed []xdsclient.XDSClient
	var errors []error

	ma.adscLock.Lock()
	for _, val := range servers {
		con, ok := ma.adscs[val.ID]
		if !ok || con.GetURL() != val.Address {
			// delete the xDSClient with different url.
			if con != nil {
				removed = append(removed, con)
			}

			con, err := ma.buildXDSClient(val.ID, val.Address, ma.nodeID, ma.cfg.InitialDeltaDiscoveryRequests)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			ma.adscs[val.ID] = con
			add = append(add, con)
		}
	}

	for k, con := range ma.adscs {
		_, ok := clientMap[k]
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
	for _, con := range add {
		adscLog.Infof("Start establish the connection with (%s/%s)", con.GetID(), con.GetURL())
		ma.store.AddStore(con.GetID(), con.GetStore())
		go con.Run()
	}
	return utilserrors.NewAggregate(errors)
}
