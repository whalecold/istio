package server

import (
	"fmt"

	"golang.org/x/time/rate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pilot/pkg/model"
	http2 "istio.io/istio/pkg/adsc/server/http"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/gvk"
	"istio.io/pkg/env"
	istiolog "istio.io/pkg/log"
)

// Server ...
type Server interface {
	Serve() error
}

var (
	mcpServerQPS   = env.RegisterFloatVar("MCP_SERVER_QPS", 100, "")
	mcpServerBurst = env.RegisterIntVar("MCP_SERVER_BURST", 200, "")
)

var (
	mcplog = istiolog.RegisterScope("mcpserver", "mcp http server", 0)
)

const (
	serviceEntryKind  = "serviceentry"
	workloadEntryKind = "workloadentry"
)

type convertFn func(config.GroupVersionKind, *config.Config) metav1.Object

type server struct {
	port  int
	store model.ConfigStoreController
	// stores the relationship with query type and gvk info.
	kinds        map[string]config.GroupVersionKind
	convertFns   map[config.GroupVersionKind]convertFn
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
	}
	s.kinds = map[string]config.GroupVersionKind{
		serviceEntryKind:  gvk.ServiceEntry,
		workloadEntryKind: gvk.WorkloadEntry,
	}
	s.convertFns = map[config.GroupVersionKind]convertFn{
		gvk.ServiceEntry:  convertToK8sServiceEntry,
		gvk.WorkloadEntry: convertToK8sWorkloadEntry,
	}
	s.store.RegisterEventHandler(gvk.ServiceEntry, s.serviceEntryHandler)
	s.store.RegisterEventHandler(gvk.WorkloadEntry, s.workloadEntryHandler)
	return s
}

// Serve ...
func (s *server) Serve() error {
	address := fmt.Sprintf(":%d", s.port)
	mcplog.Infof("starting mcp server at address: %s", address)
	return http2.New(address).
		Use(responser).
		Use(s.Limiter()).
		Use(recovery).
		Use(observer).
		Register("/mcp.istio.io/v1alpha1/resources", s.handleListRequest).
		Register("/mcp.istio.io/v1alpha1/resource", s.handleGetRequest).
		Serve()
}
