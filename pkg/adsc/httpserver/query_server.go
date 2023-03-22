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

const (
	serviceEntryKind  = "serviceentry"
	workloadEntryKind = "workloadentry"
)

type server struct {
	port  int
	store model.ConfigStoreCache
	// stores the relationship with query type and gvk info.
	kinds    map[string]config.GroupVersionKind
	handlers map[config.GroupVersionKind]func(config.GroupVersionKind, *config.Config) interface{}
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
	s.handlers = map[config.GroupVersionKind]func(config.GroupVersionKind, *config.Config) interface{}{
		gvk.ServiceEntry:  convertToK8sServiceEntry,
		gvk.WorkloadEntry: convertToK8sWorkloadEntry,
	}
	s.store.RegisterEventHandler(gvk.ServiceEntry, s.serviceEntryHandler)
	s.store.RegisterEventHandler(gvk.WorkloadEntry, s.workloadEntryHandler)
	return s
}

func (s *server) serviceEntryHandler(old, cur config.Config, event model.Event) {
	fmt.Println("serviceEntryHandler ")
}

func (s *server) workloadEntryHandler(old, cur config.Config, event model.Event) {
	fmt.Println("serviceEntryHandler ")
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
	handler, ok := s.handlers[gvk]
	if !ok {
		handleErrorResponse(w, fmt.Errorf("the kind %v is not support", gvk), http.StatusBadRequest)
		return
	}
	r := handler(gvk, conf)
	out, _ := json.Marshal(r)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

func (s *server) handleResponses(w http.ResponseWriter, gvk config.GroupVersionKind, total int, confs []config.Config) {
	ret := &ResourceList{
		Total: total,
	}
	handler, ok := s.handlers[gvk]
	if !ok {
		handleErrorResponse(w, fmt.Errorf("the kind %v is not support", gvk), http.StatusBadRequest)
		return
	}

	items := make([]interface{}, 0, len(confs))
	for idx := range confs {
		r := &confs[idx]
		items = append(items, handler(gvk, r))
	}
	ret.Items = items

	response := &Response{
		Result: ret,
	}
	out, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
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

func parseGetOption(request *http.Request) (*GetOption, error) {
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

func stringDef(in string, def string) string {
	if in == "" {
		return def
	}
	return in
}

func parseNamespaces(namespaces string) map[string]bool {
	if namespaces == "" {
		return nil
	}
	ret := map[string]bool{}
	nss := strings.Split(namespaces, ",")
	for _, ns := range nss {
		ret[ns] = true
	}
	return ret
}

func parseStartAndLimit(sta, lim string) (start int, limit int, err error) {
	if sta != "" {
		start, err = strconv.Atoi(sta)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "error format of start %s", sta)
		}
	}

	if lim != "" {
		limit, err = strconv.Atoi(lim)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "error format of limit %s", lim)
		}
		if limit > 100 {
			return 0, 0, errors.Errorf("limit %s should less than 100", lim)
		}
	}
	if limit == 0 {
		limit = 10
	}
	return
}

func parseSelector(request *http.Request) (labels.Selector, error) {
	b, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "io read all failed")
	}
	selector, err := labels.Parse(string(b))
	if err != nil {
		return nil, errors.Wrapf(err, "error format of labels %s", string(b))
	}
	return selector, nil
}

func parseListOptions(request *http.Request) (*ListOptions, error) {
	opts := &ListOptions{
		Keyword:    request.URL.Query().Get(queryParameterKeyword),
		Kind:       stringDef(request.URL.Query().Get(queryParameterKind), serviceEntryKind),
		Namespaces: parseNamespaces(request.URL.Query().Get(queryParameterNamespaces)),
	}

	var err error
	opts.Start, opts.Limit, err = parseStartAndLimit(request.URL.Query().Get(queryParameterStart),
		request.URL.Query().Get(queryParameterLimit))
	if err != nil {
		return nil, err
	}

	opts.Selector, err = parseSelector(request)

	return opts, err
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
		fmt.Println("list opts ", opt, "selector ", opt.Selector.String())
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

func paginateResource(opt *ListOptions, cfgs []config.Config) (total int, ret []config.Config) {
	total = len(cfgs)
	start, limit := opt.Start, opt.Limit
	if start >= len(cfgs) {
		start = len(cfgs)
	}
	end := start + limit
	if end >= len(cfgs) {
		end = len(cfgs)
	}
	ret = cfgs[start:end]
	return
}

func filterByOptions(cfgs []config.Config, opts *ListOptions) []config.Config {
	if opts.IsEmpty() {
		return cfgs
	}

	var idx int
	for i := range cfgs {
		cfg := &cfgs[i]
		if !opts.InNamespaces(cfg.Namespace) ||
			!opts.Matchs(labels.Set(cfg.Labels)) ||
			!opts.Contains(cfg.Name) {
			continue
		}
		cfgs[idx] = cfgs[i]
		idx += 1
	}
	return cfgs[:idx]
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
