package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	clientv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pilot/pkg/model"
	adscmetrics "istio.io/istio/pkg/adsc/metrics"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/gvk"
	istiolog "istio.io/pkg/log"
)

var (
	log = istiolog.RegisterScope("mcpserver", "mcp http server", 0)
)

const (
	serviceEntryKind  = "serviceentry"
	workloadEntryKind = "workloadentry"
)

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

func (s *server) serviceEntryHandler(old, cur config.Config, event model.Event) {
	// TODO stores the cache
}

func (s *server) workloadEntryHandler(old, cur config.Config, event model.Event) {
	// TODO stores the cache
}

func handleErrorResponse(w http.ResponseWriter, err error, code int) {
	resp := &Response{
		Error: &Error{
			Code:    code,
			Message: err.Error(),
		},
	}
	out, _ := json.Marshal(resp)
	w.WriteHeader(code)
	w.Write(out)
}

func (s *server) handleResponse(w http.ResponseWriter, gvk config.GroupVersionKind, conf *config.Config) {
	convert, ok := s.convertFns[gvk]
	if !ok {
		handleErrorResponse(w, fmt.Errorf("the kind %v is not support", gvk), http.StatusBadRequest)
		return
	}
	r := &Response{
		Result: convert(gvk, conf),
	}
	out, _ := json.Marshal(r)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

func (s *server) handleResponses(w http.ResponseWriter, gvk config.GroupVersionKind, total int, confs []config.Config) {
	ret := &ResourceList{
		Total: total,
	}
	convert, ok := s.convertFns[gvk]
	if !ok {
		handleErrorResponse(w, fmt.Errorf("the kind %v is not support", gvk), http.StatusBadRequest)
		return
	}

	items := make([]interface{}, 0, len(confs))
	for idx := range confs {
		r := &confs[idx]
		items = append(items, convert(gvk, r))
	}
	ret.Items = items

	response := &Response{
		Result: ret,
	}
	out, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

func (s *server) GetResourceHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		statusCode := http.StatusOK
		t1 := time.Now()
		opt, err := parseGetOption(request)

		defer func() {
			if err != nil {
				handleErrorResponse(w, err, statusCode)
			}
			code := strconv.Itoa(statusCode)
			adscmetrics.MCPServerRequestsDuration.WithLabelValues(opt.Kind, "get", code).Observe(time.Since(t1).Seconds())
			adscmetrics.MCPServerRequestsTotal.WithLabelValues(opt.Kind, "get", code).Inc()
		}()

		if err != nil {
			statusCode = http.StatusBadRequest
			return
		}
		gvk, ok := s.kinds[opt.Kind]
		if !ok {
			err = fmt.Errorf("the kind %s is not support", opt.Kind)
			statusCode = http.StatusBadRequest
			return
		}

		obj := s.store.Get(gvk, opt.Name, opt.Namespace)
		if obj == nil {
			err = fmt.Errorf("%s/%s/%s not found", opt.Kind, opt.Name, opt.Namespace)
			statusCode = http.StatusNotFound
			return
		}
		s.handleResponse(w, gvk, obj)
	})
}

func (s *server) ListResourceHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var (
			err  error
			opt  *ListOptions
			cfgs []config.Config
		)
		statusCode := http.StatusOK
		t1 := time.Now()

		opt, err = parseListOptions(request)

		defer func() {
			if err != nil {
				handleErrorResponse(w, err, statusCode)
			}
			code := strconv.Itoa(statusCode)
			adscmetrics.MCPServerRequestsDuration.WithLabelValues(opt.Kind, "list", code).Observe(time.Since(t1).Seconds())
			adscmetrics.MCPServerRequestsTotal.WithLabelValues(opt.Kind, "list", code).Inc()
		}()

		if err != nil {
			return
		}
		gvk, ok := s.kinds[opt.Kind]
		if !ok {
			err = fmt.Errorf("the kind %s is not support", opt.Kind)
			statusCode = http.StatusBadRequest
			return
		}

		cfgs, err = s.store.List(gvk, opt.Namespace())
		if err != nil {
			statusCode = http.StatusInternalServerError
			return
		}

		cfgs = filterByOptions(cfgs, opt)
		sortConfigByCreationTime(cfgs)
		total, cfgs := paginateResource(opt, cfgs)
		s.handleResponses(w, gvk, total, cfgs)
	})
}

// Run ...
func (s *server) Run() {
	mux := http.NewServeMux()
	mux.Handle("/mcp.istio.io/v1alpha1/resources", s.ListResourceHandler())
	mux.Handle("/mcp.istio.io/v1alpha1/resource", s.GetResourceHandler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), mux)
	if err != nil {
		panic(err)
	}
}

func convertToK8sWorkloadEntry(gvk config.GroupVersionKind, res *config.Config) interface{} {
	we, ok := res.Spec.(*networkingv1alpha3.WorkloadEntry)
	if !ok {
		return nil
	}
	return &clientv1alpha3.WorkloadEntry{
		TypeMeta:   convertToK8sType(gvk),
		ObjectMeta: convertToK8sMate(res.Meta),
		Spec:       *we,
	}
}

func convertToK8sServiceEntry(gvk config.GroupVersionKind, res *config.Config) interface{} {
	se, ok := res.Spec.(*networkingv1alpha3.ServiceEntry)
	if !ok {
		return nil
	}
	return &clientv1alpha3.ServiceEntry{
		TypeMeta:   convertToK8sType(gvk),
		ObjectMeta: convertToK8sMate(res.Meta),
		Spec:       *se,
	}
}

func convertToK8sType(meta config.GroupVersionKind) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       meta.Kind,
		APIVersion: meta.GroupVersion(),
	}
}

func convertToK8sMate(meta config.Meta) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace:   meta.Namespace,
		Name:        meta.Name,
		Labels:      meta.Labels,
		Annotations: meta.Annotations,
		CreationTimestamp: metav1.Time{
			Time: meta.CreationTimestamp,
		},
	}
}
