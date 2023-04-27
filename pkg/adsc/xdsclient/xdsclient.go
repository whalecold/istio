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

package xdsclient

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"istio.io/istio/pilot/pkg/config/memory"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/collections"
	"istio.io/istio/pkg/config/schema/gvk"
	"istio.io/pkg/log"
)

var adscLog = log.RegisterScope("adsc", "adsc debugging", 0)

const (
	// BufferSize specifies the buffer size of event channel
	BufferSize = 100

	defaultClientMaxReceiveMessageSize = math.MaxInt32
	defaultInitialConnWindowSize       = 1024 * 1024 // default gRPC InitialWindowSize
	defaultInitialWindowSize           = 1024 * 1024 // default gRPC ConnWindowSize
)

// XDSClient the mcp-xDS client, use for sync configuration from mcp server and push them to discovery server.
type XDSClient interface {
	GetURL() string
	GetID() string
	Run() error
	Close()
	GetStore() model.ConfigStoreController
}

// xDSClient the implementation of XDSClient.
type xDSClient struct {
	initRequests []*discovery.DeltaDiscoveryRequest
	url          string
	upstreamName string
	stopCh       chan struct{}
	nodeID       string

	// Retrieved configurations can be stored using the common istio model interface.
	store model.ConfigStoreController

	// backoffPolicy determines to reconnect policy. Based on MCP client.
	backoffPolicy backoff.BackOff

	syncLock      sync.RWMutex
	hasSyncedFunc func() bool
}

// New client
func New(upstream, url, nodeID string, requests []*discovery.DeltaDiscoveryRequest) (XDSClient, error) {
	c := &xDSClient{
		initRequests:  requests,
		stopCh:        make(chan struct{}),
		url:           url,
		upstreamName:  upstream,
		nodeID:        nodeID,
		backoffPolicy: backoff.NewExponentialBackOff(),
	}
	store := memory.MakeSkipValidation(collections.PilotMCP)
	configController := memory.NewSyncController(store)
	configController.RegisterHasSyncedHandler(c.HasSynced)
	c.store = configController
	return c, nil
}

func (c *xDSClient) GetURL() string {
	return c.url
}

func (c *xDSClient) GetID() string {
	return c.upstreamName
}

func (c *xDSClient) resetSyncedFunc(s *stream) {
	if s == nil {
		return
	}
	c.syncLock.Lock()
	defer c.syncLock.Unlock()

	c.hasSyncedFunc = func() bool {
		s.resourceLock.RLock()
		defer s.resourceLock.RUnlock()

		for _, r := range c.initRequests {
			if _, ok := s.resourceTypeNonce[r.TypeUrl]; !ok {
				return false
			}
		}
		return true
	}
}

func (c *xDSClient) Close() {
	// should clean the store first as the store controller will terminal when closing the c.stop chan.
	c.cleanStore()
	close(c.stopCh)
}

// GetStore get store.
func (c *xDSClient) GetStore() model.ConfigStoreController {
	return c.store
}

func (c *xDSClient) HasSynced() bool {
	c.syncLock.RLock()
	defer c.syncLock.RUnlock()

	if c.hasSyncedFunc == nil {
		return false
	}
	return c.hasSyncedFunc()
}

func (c *xDSClient) Run() error {
	go c.store.Run(c.stopCh)

	for {
		select {
		case <-c.stopCh:
			return nil
		default:
		}

		duration := c.backoffPolicy.NextBackOff()
		if duration == backoff.Stop {
			c.backoffPolicy.Reset()
			continue
		}
		if duration > 0 {
			adscLog.Warnf("Waited for %v as backoff policy, xDSClient(%s/%s)", duration, c.upstreamName, c.url)
		}

		select {
		case <-time.After(duration):
		case <-c.stopCh:
			return nil
		}

		err := c.handleConnection()
		if err != nil {
			adscLog.Warnf("handle connection faield %v", err)
		}
	}
}

// use to trigger the store callback handle of deleting event
// TODO it will cause performance problems if remove stores with a large cache frequently.
func (c *xDSClient) cleanStore() {
	gvks := []*config.GroupVersionKind{&gvk.ServiceEntry, &gvk.WorkloadEntry}
	for _, val := range gvks {
		cfgs, err := c.store.List(*val, model.NamespaceAll)
		if err != nil {
			adscLog.Warnf("List gvks %v failed: %s when clean store of %s", *val, err, c.upstreamName)
			continue
		}
		for _, cfg := range cfgs {
			err = c.store.Delete(cfg.GroupVersionKind, cfg.Name, cfg.Namespace, &cfg.ResourceVersion)
			if err != nil {
				adscLog.Warnf("Delete config %v/%s/%s failed: %s when clean store of %s", cfg.GroupVersionKind, cfg.Namespace, cfg.Name, err, c.upstreamName)
			}
		}
	}
}

func defaultGrpcDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		// TODO(SpecialYang) maybe need to make it configurable.
		grpc.WithInitialWindowSize(int32(defaultInitialWindowSize)),
		grpc.WithInitialConnWindowSize(int32(defaultInitialConnWindowSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(defaultClientMaxReceiveMessageSize)),
	}
}

func (c *xDSClient) handleConnection() error {
	defaultGrpcDialOptions := defaultGrpcDialOptions()
	var grpcDialOptions []grpc.DialOption
	grpcDialOptions = append(grpcDialOptions, defaultGrpcDialOptions...)
	// disable transport security
	grpcDialOptions = append(grpcDialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	con, err := grpc.DialContext(ctx, c.url, grpcDialOptions...)
	if err != nil {
		return errors.Wrapf(err, "grpc dail context url %s failed", c.url)
	}
	defer con.Close()

	adscLog.Infof("start connect to source %s/%s", c.upstreamName, c.url)
	s, err := newStream(con, c.store, c.upstreamName, c.url, c.nodeID, c.stopCh)
	if err != nil {
		return errors.Wrapf(err, "connect to upstream (%s/%s) failed", c.upstreamName, con.Target())
	}
	defer s.close()
	err = s.sendInitRequest(c.initRequests)
	if err != nil {
		return errors.Wrapf(err, "send initial requests to upstream (%s/%s) failed", c.upstreamName, con.Target())
	}
	adscLog.Infof("send init requests to %s/%s successfully", c.upstreamName, c.url)
	c.resetSyncedFunc(s)
	s.handleResponse()
	return nil
}
