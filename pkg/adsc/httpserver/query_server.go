package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	clientv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/gvk"
)

const (
	serviceEntryKind  = "serviceentry"
	workloadEntryKind = "workloadentry"
)

type server struct {
	port  int
	store model.ConfigStoreCache
	// stores the relationship with query type and gvk info.
	kinds      map[string]config.GroupVersionKind
	convertFns map[config.GroupVersionKind]func(config.GroupVersionKind, *config.Config) interface{}
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
	s.convertFns = map[config.GroupVersionKind]func(config.GroupVersionKind, *config.Config) interface{}{
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
	w.WriteHeader(code)
	out, _ := json.Marshal(resp)
	w.Write(out)
}

func (s *server) handleResponse(w http.ResponseWriter, gvk config.GroupVersionKind, conf *config.Config) {
	convert, ok := s.convertFns[gvk]
	if !ok {
		handleErrorResponse(w, fmt.Errorf("the kind %v is not support", gvk), http.StatusBadRequest)
		return
	}
	r := convert(gvk, conf)
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
		opt, err := parseGetOption(request)
		if err != nil {
			handleErrorResponse(w, err, http.StatusBadRequest)
			return
		}
		gvk, ok := s.kinds[opt.Kind]
		if !ok {
			handleErrorResponse(w, fmt.Errorf("the kind %s is not support", opt.Kind), http.StatusBadRequest)
			return
		}

		obj := s.store.Get(gvk, opt.Name, opt.Namespace)
		if obj == nil {
			handleErrorResponse(w, fmt.Errorf("%s/%s/%s not found", opt.Kind, opt.Name, opt.Namespace), http.StatusNotFound)
			return
		}
		s.handleResponse(w, gvk, obj)
	})
}

func (s *server) ListResourceHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		opt, err := parseListOptions(request)
		if err != nil {
			handleErrorResponse(w, err, http.StatusBadRequest)
			return
		}
		gvk, ok := s.kinds[opt.Kind]
		if !ok {
			handleErrorResponse(w, fmt.Errorf("the kind %s is not support", opt.Kind), http.StatusBadRequest)
			return
		}

		cfgs, err := s.store.List(gvk, opt.Namespace())
		if err != nil {
			handleErrorResponse(w, err, http.StatusInternalServerError)
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
