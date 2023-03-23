package xdsclient

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	mcp "istio.io/api/mcp/v1alpha1"

	mem "istio.io/istio/pilot/pkg/config/memory"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/adsc/metrics"
	"istio.io/istio/pkg/config"
)

type stream struct {
	// Stream is the GRPC xDSClient stream, allowing direct GRPC send operations.
	// Set after Dial is called.
	stream discovery.AggregatedDiscoveryService_DeltaAggregatedResourcesClient
	// xds client used to create a stream
	client       discovery.AggregatedDiscoveryServiceClient
	respChan     chan *discovery.DeltaDiscoveryResponse
	nodeID       string
	url          string
	upstreamName string
	stop         <-chan struct{}

	resourceLock      sync.RWMutex
	resourceTypeNonce map[string]string
	// Retrieved configurations can be stored using the common istio model interface.
	store model.ConfigStore
}

func newStream(con *grpc.ClientConn, store model.ConfigStore, name, url, nodeID string, stop <-chan struct{}) (*stream, error) {
	s := &stream{
		stop:              stop,
		store:             store,
		url:               url,
		nodeID:            nodeID,
		upstreamName:      name,
		respChan:          make(chan *discovery.DeltaDiscoveryResponse, BufferSize),
		resourceTypeNonce: make(map[string]string),
	}
	var err error
	s.client = discovery.NewAggregatedDiscoveryServiceClient(con)
	s.stream, err = s.client.DeltaAggregatedResources(context.Background())
	if err != nil {
		return nil, errors.Wrapf(err, "delta aggregate resource failed")
	}
	return s, nil
}

func (s *stream) close() {
	close(s.respChan)
}

// send a wrapper of grpc send.
func (s *stream) send(r *discovery.DeltaDiscoveryRequest) error {
	// add default node info.
	if r.Node == nil {
		r.Node = s.node()
	}
	return s.stream.Send(r)
}

func (s *stream) sendInitRequest(requests []*discovery.DeltaDiscoveryRequest) error {
	// Send the initial requests
	for _, r := range requests {
		err := s.send(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *stream) mcpToPilot(m *mcp.Resource) (*config.Config, error) {
	if m == nil || m.Metadata == nil {
		return &config.Config{}, nil
	}
	c := &config.Config{
		Meta: config.Meta{
			ResourceVersion: m.Metadata.Version,
			Labels:          m.Metadata.Labels,
			Annotations:     m.Metadata.Annotations,
		},
	}

	if c.Meta.Annotations == nil {
		c.Meta.Annotations = make(map[string]string)
	}
	c.Meta.Annotations[MCPServerSource] = s.upstreamName + "~" + s.url
	nsn := strings.Split(m.Metadata.Name, "/")
	if len(nsn) != 2 {
		return nil, fmt.Errorf("invalid name %s", m.Metadata.Name)
	}
	c.Namespace = nsn[0]
	c.Name = nsn[1]
	var err error
	c.CreationTimestamp = m.Metadata.CreateTime.AsTime()

	pb, err := m.Body.UnmarshalNew()
	if err != nil {
		return nil, err
	}
	c.Spec = pb
	return c, nil
}

func (s *stream) updateResource(gvk config.GroupVersionKind, rsc *discovery.Resource, fn func(*config.Config)) error {
	m := &mcp.Resource{}
	err := rsc.Resource.UnmarshalTo(m)
	if err != nil {
		adscLog.Warnf("Error unmarshalling received MCP config %v", err)
		return err
	}
	newCfg, err := s.mcpToPilot(m)
	if err != nil {
		adscLog.Warnf("Error convert resource to pilot config %v", err)
		return err
	}
	if newCfg == nil {
		return nil
	}

	if fn != nil {
		fn(newCfg)
	}

	newCfg.GroupVersionKind = gvk
	oldCfg := s.store.Get(newCfg.GroupVersionKind, newCfg.Name, newCfg.Namespace)
	if oldCfg == nil {
		if _, err = s.store.Create(*newCfg); err != nil {
			adscLog.Warnf("Error adding a new resource to the store %v", err)
			return err
		}
	} else if oldCfg.ResourceVersion != newCfg.ResourceVersion || newCfg.ResourceVersion == "" {
		// update the store only when resource version differs or unset.
		newCfg.Annotations[mem.ResourceVersion] = newCfg.ResourceVersion
		newCfg.ResourceVersion = oldCfg.ResourceVersion
		if _, err = s.store.Update(*newCfg); err != nil {
			adscLog.Warnf("Error updating an existing resource in the store %v", err)
			return err
		}
	}
	return nil
}

func (s *stream) fullUpdateResource(gvk config.GroupVersionKind, resources []*discovery.Resource) {
	existingConfigs, err := s.store.List(gvk, "")
	if err != nil {
		adscLog.Warnf("Error listing existing configs %v", err)
		return
	}
	received := make(map[string]*config.Config)
	for _, rsc := range resources {
		_ = s.updateResource(gvk, rsc, func(c *config.Config) {
			received[c.Namespace+"/"+c.Name] = c
		})
	}
	// remove deleted resources from cache
	for _, cfg := range existingConfigs {
		if _, ok := received[cfg.Namespace+"/"+cfg.Name]; !ok {
			if err := s.store.Delete(cfg.GroupVersionKind, cfg.Name, cfg.Namespace, nil); err != nil {
				adscLog.Warnf("Error deleting an outdated resource from the store %v", err)
				continue
			}
		}
	}
}

func (s *stream) incrementalUpdateResource(gvk config.GroupVersionKind, msg *discovery.DeltaDiscoveryResponse) {
	for _, rsc := range msg.Resources {
		_ = s.updateResource(gvk, rsc, nil)
	}
	// Remove deleted resources from cache
	for _, cfg := range msg.RemovedResources {
		cfgs := strings.SplitN(cfg, "/", 2)
		if len(cfgs) != 2 {
			adscLog.Warnf("Removed resource with the error format %s", cfgs)
			continue
		}
		if err := s.store.Delete(gvk, cfgs[1], cfgs[0], nil); err != nil {
			adscLog.Warnf("Error deleting an outdated resource from the store %v", err)
			continue
		}
	}
}

func (s *stream) handleDeltaDiscoveryResponse(msg *discovery.DeltaDiscoveryResponse) {
	adscLog.Info("Received ", s.url, " type ", msg.TypeUrl,
		" cnt=", len(msg.Resources), " nonce=", msg.Nonce)
	metrics.MCPOverXDSClientReceiveResponseTotal.WithLabelValues(s.upstreamName, s.nodeID, msg.TypeUrl).Inc()
	now := time.Now()
	defer func() {
		t := time.Since(now).Seconds()
		metrics.MCPOverXDSClientReceiveResponseDuration.WithLabelValues(s.upstreamName, s.nodeID, msg.TypeUrl).Observe(t)
	}()
	// Group-value-kind - used for high level api generator.
	gvk := strings.SplitN(msg.TypeUrl, "/", 3)
	if len(gvk) != 3 || s.store == nil {
		return
	}
	groupVersionKind := config.GroupVersionKind{Group: gvk[0], Version: gvk[1], Kind: gvk[2]}
	// Default use incremental update, but full update be used if it's the first response from mcp server.
	// Two condition will happen:
	//   1. The first time connect to the mcp server, full update is same as incremental update.
	//   2. The network do not work temporary and connect again, the changes during the disconnection can't be tracked.
	s.resourceLock.Lock()
	if _, ok := s.resourceTypeNonce[msg.TypeUrl]; ok {
		s.incrementalUpdateResource(groupVersionKind, msg)
	} else {
		s.fullUpdateResource(groupVersionKind, msg.Resources)
	}
	s.resourceTypeNonce[msg.TypeUrl] = msg.Nonce
	s.resourceLock.Unlock()
	s.ack(msg)
}

func (s *stream) node() *v3.Node {
	return &v3.Node{
		Id: s.nodeID,
	}
}

func (s *stream) ack(msg *discovery.DeltaDiscoveryResponse) {
	_ = s.send(&discovery.DeltaDiscoveryRequest{
		ResponseNonce: msg.Nonce,
		TypeUrl:       msg.TypeUrl,
	})
}

func (s *stream) handleResponse() {
	go func() {
		for {
			msg, err := s.stream.Recv()
			if err != nil {
				adscLog.Infof("XDSClient closed for node %v with err: %v", s.nodeID, err)
				return
			}
			s.respChan <- msg
		}
	}()
	go func() {
		select {
		case <-s.stream.Context().Done():
			adscLog.Infof("stream xDSClient terminal %v err %v", s.nodeID, s.stream.Context().Err())
		case <-s.stop:
			adscLog.Infof("receive stop channel signal %v ", s.nodeID)
		}
		s.respChan <- nil
	}()
	for resp := range s.respChan {
		if resp == nil {
			break
		}
		s.handleDeltaDiscoveryResponse(resp)
	}
}
