package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"istio.io/istio/pilot/pkg/model"
	adscmetrics "istio.io/istio/pkg/adsc/metrics"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/gvk"
	istiolog "istio.io/pkg/log"
)

var (
	mcplog = istiolog.RegisterScope("mcpserver", "mcp http server", 0)
)

const (
	serviceEntryKind  = "serviceentry"
	workloadEntryKind = "workloadentry"
)

func buildMetricRecorder(w http.ResponseWriter, request *http.Request) *metricsRecorder {
	return &metricsRecorder{
		t:          time.Now(),
		w:          w,
		request:    request,
		statusCode: http.StatusOK,
		kind:       stringDef(request.URL.Query().Get(queryParameterKind), serviceEntryKind),
	}
}

type metricsRecorder struct {
	t          time.Time
	w          http.ResponseWriter
	request    *http.Request
	statusCode int
	kind       string
}

func (mr *metricsRecorder) setStatusCode(code int) {
	mr.statusCode = code
}

func (mr *metricsRecorder) ending() {
	code := strconv.Itoa(mr.statusCode)
	t := time.Since(mr.t).Seconds()
	path := mr.request.URL.Path

	adscmetrics.MCPServerRequestsDuration.WithLabelValues(mr.kind, path, code).Observe(t)
	adscmetrics.MCPServerRequestsTotal.WithLabelValues(mr.kind, path, code).Inc()

	mcplog.Debugf("get resource, kind %v, code: %d req path: %s, query: %s, latency %v", mr.kind, code, path, mr.request.URL.RawQuery, t)
}

type convertFn func(config.GroupVersionKind, *config.Config) interface{}

type server struct {
	port  int
	store model.ConfigStoreCache
	// stores the relationship with query type and gvk info.
	kinds      map[string]config.GroupVersionKind
	convertFns map[config.GroupVersionKind]convertFn
}

// New new query server.
func New(store model.ConfigStoreCache, p int) *server {
	s := &server{
		store: store,
		port:  p,
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

// Server ...
func (s *server) Server() {
	mux := http.NewServeMux()
	mux.Handle("/mcp.istio.io/v1alpha1/resources", s.ListResourceHandler())
	mux.Handle("/mcp.istio.io/v1alpha1/resource", s.GetResourceHandler())
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}
	mcplog.Infof("starting MCP service at %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		panic(err)
	}
}

func handleResponse(w http.ResponseWriter, ret interface{}, err *Error) {
	resp := &Response{
		Error:  err,
		Result: ret,
	}
	out, _ := json.Marshal(resp)
	w.WriteHeader(err.GetCode())
	w.Write(out)
}

type handler func(request *http.Request) (interface{}, *Error)

func handlerBuilder(fn handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		recorder := buildMetricRecorder(w, request)
		defer recorder.ending()

		r, err := fn(request)

		recorder.setStatusCode(err.GetCode())

		handleResponse(w, r, err)
	})
}

func (s *server) ListResourceHandler() http.Handler {
	return handlerBuilder(func(request *http.Request) (interface{}, *Error) {
		opt, err := parseListOptions(request)
		if err != nil {
			return nil, BadRequestHander(err)
		}

		return s.handleListRequest(opt)
	})
}

func (s *server) GetResourceHandler() http.Handler {
	return handlerBuilder(func(request *http.Request) (interface{}, *Error) {
		opt, err := parseGetOption(request)
		if err != nil {
			return nil, BadRequestHander(err)
		}

		return s.handleGetRequest(opt)
	})
}
