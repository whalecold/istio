package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	clientv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/pkg/errors"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/gvk"
)

type server struct {
	port  int
	store model.ConfigStore
	// storeKindMap use to store the relation with query kind and gvk info.
	storeKindMap map[string]config.GroupVersionKind
	convertMap   map[config.GroupVersionKind]func(config.GroupVersionKind, *config.Config) interface{}
}

// NewQueryServer new query server.
func NewQueryServer(store model.ConfigStore, port int) *server {
	s := &server{
		store: store,
		port:  port,
	}
	s.storeKindMap = map[string]config.GroupVersionKind{
		"serviceentry":  gvk.ServiceEntry,
		"workloadentry": gvk.WorkloadEntry,
	}
	s.convertMap = map[config.GroupVersionKind]func(config.GroupVersionKind, *config.Config) interface{}{
		gvk.ServiceEntry:  s.convertToK8sServiceEntry,
		gvk.WorkloadEntry: s.convertToK8sWorkloadEntry,
	}
	return s
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

// errorResponseHandler handle the error response.
func (s *server) errorResponseHandler(c http.ResponseWriter, err error, code int) {
	resp := &Response{
		Error: &Error{
			Code:    code,
			Message: err.Error(),
		},
	}
	c.WriteHeader(code)
	out, _ := json.Marshal(resp)
	c.Write(out)
}

// httpGetResourceHandler handle the response.
func (s *server) httpGetResourceHandler(c http.ResponseWriter, gvkInfo config.GroupVersionKind, res *config.Config) {
	handler, ok := s.convertMap[gvkInfo]
	if !ok {
		s.errorResponseHandler(c, fmt.Errorf("the kind %v is not support", gvkInfo), http.StatusBadRequest)
		return
	}
	response := handler(gvkInfo, res)
	out, _ := json.Marshal(response)
	c.WriteHeader(http.StatusOK)
	c.Write(out)
}

// httpResourceHandler handle the response.
func (s *server) httpResourceHandler(c http.ResponseWriter, gvkInfo config.GroupVersionKind, total int, res []config.Config) {
	ret := &ResourceList{
		Total: total,
	}
	handler, ok := s.convertMap[gvkInfo]
	if !ok {
		s.errorResponseHandler(c, fmt.Errorf("the kind %v is not support", gvkInfo), http.StatusBadRequest)
		return
	}

	items := make([]interface{}, 0, len(res))

	for idx := range res {
		r := &res[idx]
		items = append(items, handler(gvkInfo, r))
	}
	ret.Items = items
	response := &Response{
		Result: ret,
	}
	out, _ := json.Marshal(response)
	c.WriteHeader(http.StatusOK)
	c.Write(out)
}

func (s *server) convertToK8sWorkloadEntry(gvkInfo config.GroupVersionKind, res *config.Config) interface{} {
	we, ok := res.Spec.(*networkingv1alpha3.WorkloadEntry)
	if !ok {
		return nil
	}
	return &clientv1alpha3.WorkloadEntry{
		TypeMeta:   convertToK8sType(gvkInfo),
		ObjectMeta: convertToK8sMate(res.Meta),
		Spec:       *we,
	}
}

func (s *server) convertToK8sServiceEntry(gvkInfo config.GroupVersionKind, res *config.Config) interface{} {
	se, ok := res.Spec.(*networkingv1alpha3.ServiceEntry)
	if !ok {
		return nil
	}
	return &clientv1alpha3.ServiceEntry{
		TypeMeta:   convertToK8sType(gvkInfo),
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

func (s *server) parseGetOption(request *http.Request) (*GetOption, error) {
	opts := &GetOption{
		Kind:      request.URL.Query().Get(queryParameterKind),
		Name:      request.URL.Query().Get(queryParameterName),
		Namespace: request.URL.Query().Get(queryParameterNamespace),
	}
	if opts.Kind == "" || opts.Name == "" || opts.Namespace == "" {
		return nil, fmt.Errorf("the parameter should not be empty")
	}
	return opts, nil
}

func (s *server) parseListOptions(request *http.Request) (*ListOptions, error) {
	opts := &ListOptions{
		Kind:       request.URL.Query().Get(queryParameterKind),
		Keyword:    request.URL.Query().Get(queryParameterKeyword),
		Namespaces: map[string]bool{},
	}

	nss := strings.Split(request.URL.Query().Get(queryParameterNamespaces), ",")
	for _, ns := range nss {
		opts.Namespaces[ns] = true
	}

	if opts.Kind == "" {
		// default to serviceentry
		opts.Kind = "serviceentry"
	}
	var err error
	start, limit := request.URL.Query().Get(queryParameterStart), request.URL.Query().Get(queryParameterLimit)
	if start != "" {
		opts.Start, err = strconv.Atoi(start)
		if err != nil {
			return nil, errors.Wrapf(err, "error format of start %s", start)
		}
	}

	if limit != "" {
		opts.Limit, err = strconv.Atoi(limit)
		if err != nil {
			return nil, errors.Wrapf(err, "error format of limit %s", limit)
		}
		if opts.Limit > 100 {
			return nil, errors.Errorf("limit %s should less than 100", limit)
		}
	}
	if opts.Limit == 0 {
		opts.Limit = 10
	}
	defer request.Body.Close()

	b, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "io read all failed")
	}
	if len(b) != 0 {
		opts.Selector, err = labels.Parse(string(b))
		if err != nil {
			return nil, errors.Wrapf(err, "error format of labels %s", string(b))
		}
	}
	return opts, nil
}

func (s *server) GetResourceHandler() http.Handler {
	return http.HandlerFunc(func(c http.ResponseWriter, request *http.Request) {
		opts, err := s.parseGetOption(request)
		if err != nil {
			s.errorResponseHandler(c, err, http.StatusBadRequest)
			return
		}
		gvkInfo, ok := s.storeKindMap[opts.Kind]
		if !ok {
			s.errorResponseHandler(c, fmt.Errorf("the kind %s is not support", opts.Kind), http.StatusBadRequest)
			return
		}

		obj := s.store.Get(gvkInfo, opts.Name, opts.Namespace)
		if obj == nil {
			s.errorResponseHandler(c, fmt.Errorf("%s/%s/%s not found", opts.Kind, opts.Name, opts.Namespace), http.StatusNotFound)
			return
		}
		s.httpGetResourceHandler(c, gvkInfo, obj)
	})
}

func (s *server) ListResourceHandler() http.Handler {
	return http.HandlerFunc(func(c http.ResponseWriter, request *http.Request) {
		options, err := s.parseListOptions(request)
		if err != nil {
			s.errorResponseHandler(c, err, http.StatusBadRequest)
			return
		}
		gvkInfo, ok := s.storeKindMap[options.Kind]
		if !ok {
			s.errorResponseHandler(c, fmt.Errorf("the kind %s is not support", options.Kind), http.StatusBadRequest)
			return
		}

		list, err := s.store.List(gvkInfo, options.Namespace())
		if err != nil {
			s.errorResponseHandler(c, err, http.StatusInternalServerError)
			return
		}

		list = filterByOptions(list, options)
		total := len(list)
		sortConfigByCreationTime(list)
		s.httpResourceHandler(c, gvkInfo, total, paginateResource(options, list))
	})
}

func paginateResource(options *ListOptions, list []config.Config) []config.Config {
	start, limit := options.Start, options.Limit
	if limit == 0 {
		limit = 10
	}
	if start >= len(list) {
		start = len(list)
	}
	end := start + limit
	if end >= len(list) {
		end = len(list)
	}
	return list[start:end]
}

func filterByOptions(configs []config.Config, opts *ListOptions) []config.Config {
	if opts.Selector == nil && opts.Keyword == "" {
		return configs
	}
	var idx int
	for i := range configs {
		if !opts.InNamespaces(configs[i].Namespace) {
			continue
		}
		if opts.Keyword != "" && !strings.Contains(configs[i].Name, opts.Keyword) {
			continue
		}
		if opts.Selector != nil && !opts.Selector.Matches(labels.Set(configs[i].Labels)) {
			continue
		}
		configs[idx] = configs[i]
		idx += 1
	}
	configs = configs[:idx]
	return configs
}

// sortConfigByCreationTime sorts the list of config objects in ascending order by their creation time (if available).
func sortConfigByCreationTime(configs []config.Config) {
	sort.Slice(configs, func(i, j int) bool {
		// If creation time is the same, then behavior is nondeterministic. In this case, we can
		// pick an arbitrary but consistent ordering based on name and namespace, which is unique.
		// CreationTimestamp is stored in seconds, so this is not uncommon.
		if configs[i].CreationTimestamp.Equal(configs[j].CreationTimestamp) {
			in := configs[i].Namespace + "." + configs[i].Name
			jn := configs[j].Namespace + "." + configs[j].Name
			return in < jn
		}
		return configs[i].CreationTimestamp.After(configs[j].CreationTimestamp)
	})
}
