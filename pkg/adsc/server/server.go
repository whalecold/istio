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

package server

import (
	"fmt"

	"golang.org/x/time/rate"

	"istio.io/istio/pilot/pkg/model"
	http2 "istio.io/istio/pkg/adsc/server/http"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/gvk"
	"istio.io/pkg/env"
	istiolog "istio.io/pkg/log"
)

// Server the http server to query the workloadentry and serviceentry in memory.
type Server interface {
	Serve() error
}

var (
	mcpServerQPS   = env.RegisterFloatVar("MCP_SERVER_QPS", 100, "")
	mcpServerBurst = env.RegisterIntVar("MCP_SERVER_BURST", 200, "")

	mcplog = istiolog.RegisterScope("mcpserver", "mcp http server", 0)
)

const (
	serviceEntryKind  = "serviceentry"
	workloadEntryKind = "workloadentry"
)

type server struct {
	port  int
	store model.ConfigStoreController
	// stores the relationship with retrieved type and gvk info.
	kinds        map[string]config.GroupVersionKind
	indexedStore *serviceInstancesStore
	limiter      *rate.Limiter
}

// New query server.
func New(store model.ConfigStoreController, p int) Server {
	s := &server{
		store:        store,
		port:         p,
		indexedStore: newStore(),
		limiter:      rate.NewLimiter(rate.Limit(mcpServerQPS.Get()), mcpServerBurst.Get()),
		kinds: map[string]config.GroupVersionKind{
			serviceEntryKind:  gvk.ServiceEntry,
			workloadEntryKind: gvk.WorkloadEntry,
		},
	}

	s.store.RegisterEventHandler(gvk.ServiceEntry, s.configStoreHandler)
	s.store.RegisterEventHandler(gvk.WorkloadEntry, s.configStoreHandler)
	return s
}

// Serve the implemention of interface.
func (s *server) Serve() error {
	address := fmt.Sprintf(":%d", s.port)
	mcplog.Infof("starting mcp server at address: %s", address)
	return http2.New(address).
		Use(responser).
		Use(observer).
		Use(s.Limiter()).
		Use(recovery).
		Register("/mcp.istio.io/v1alpha1/resources", s.handleListRequest).
		Register("/mcp.istio.io/v1alpha1/resource", s.handleGetRequest).
		Serve()
}
