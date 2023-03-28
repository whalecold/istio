package server

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	http2 "istio.io/istio/pkg/adsc/server/http"
	"istio.io/istio/pkg/config/schema/gvk"

	"istio.io/istio/pkg/config"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	queryParameterName       = "name"
	queryParameterNamespace  = "namespace"
	queryParameterNamespaces = "namespaces"
	queryParameterKeyword    = "keyword"
	queryParameterRef        = "ref"
	queryParameterKind       = "kind"
	queryParameterStart      = "start"
	queryParameterLimit      = "limit"

	annotationInstanceNumber = "sidecar.mesh.io/instance-number"
)

// Response ...
type Response struct {
	Error  *http2.Error `json:"error,omitempty"`
	Result interface{}  `json:"result,omitempty"`
}

// ResourceList ...
type ResourceList struct {
	Total int         `json:"total,omitempty"`
	Items interface{} `json:"items,omitempty"`
}

// ListOptions ...
type ListOptions struct {
	Kind string `query:"kind"`
	//
	Query      string
	Namespaces map[string]bool
	Selector   labels.Selector
	// If Start and Limit are all zero, return all the resource meet the others conditions.
	Start int `query:"start" default:"0"`
	Limit int `query:"limit" default:"10"`

	// Query the resource that belongs to the ref.
	refKey string
}

func (l *ListOptions) isRefList() bool {
	return len(l.refKey) != 0
}

func (l *ListOptions) getRefKey() string {
	return l.refKey
}

func (l *ListOptions) refGroupVersionKind() config.GroupVersionKind {
	refs := strings.Split(l.refKey, "/")
	if len(refs) != 5 {
		return config.GroupVersionKind{}
	}
	return config.GroupVersionKind{
		Group:   refs[0],
		Version: refs[1],
		Kind:    refs[2],
	}
}

func (l *ListOptions) Contains(name string) bool {
	if l.Query == "" {
		return true
	}
	return strings.Contains(name, l.Query)
}

// Matchs ...
func (l *ListOptions) Matchs(set labels.Labels) bool {
	if l.Selector == nil {
		return true
	}
	return l.Selector.Matches(set)
}

// Namespace if there is only one namespace in the map, return
// it as it can be used to list the specified resource. If the
// query multiple namespace parameters, should return empty string
// to capture all resources and filter them through the `InNamespaces`
// function.
func (l *ListOptions) Namespace() string {
	if len(l.Namespaces) != 1 {
		return ""
	}
	for ns := range l.Namespaces {
		return ns
	}
	return ""
}

// InNamespaces checks the namespace if in the map.
func (l *ListOptions) InNamespaces(ns string) bool {
	if len(l.Namespaces) == 0 {
		return true
	}
	return l.Namespaces[ns]
}

func (l *ListOptions) skip(cfg *config.Config) bool {
	return !l.InNamespaces(cfg.Namespace) ||
		!l.Matchs(labels.Set(cfg.Labels)) ||
		!l.Contains(cfg.Name)
}

// GetOption get option
type GetOption struct {
	Kind      string `query:"kind"`
	Name      string `query:"name"`
	Namespace string `query:"namespace"`
}

// sortConfigByCreationTime sorts the list of config objects in ascending order by their creation time (if available).
func sortConfigByCreationTime(configs []*config.Config) {
	sort.Slice(configs, func(i, j int) bool {
		// If creation time is the same, then behavior is nondeterministic. In this case, we can
		// pick an arbitrary but consistent ordering based on name and namespace, which is unique.
		// CreationTimestamp is stored in seconds, so this is not uncommon.
		t1 := configs[i].CreationTimestamp
		t2 := configs[j].CreationTimestamp

		if t1.Equal(t2) {
			in := configs[i].Namespace + "." + configs[i].Name
			jn := configs[j].Namespace + "." + configs[j].Name
			return in < jn
		}
		return t1.After(t2)
	})
}

func paginateResource(opt *ListOptions, confs []*config.Config) (total int, ret []*config.Config) {
	total = len(confs)
	start, limit := opt.Start, opt.Limit
	if start >= len(confs) {
		start = len(confs)
	}
	end := start + limit
	if end >= len(confs) {
		end = len(confs)
	}
	ret = confs[start:end]
	return
}

func parseListOptions(request *http.Request) (*ListOptions, error) {
	opts := &ListOptions{
		Query:      request.URL.Query().Get(queryParameterKeyword),
		Kind:       stringDef(request.URL.Query().Get(queryParameterKind), serviceEntryKind),
		Namespaces: parseNamespaces(request.URL.Query().Get(queryParameterNamespaces)),
		refKey:     request.URL.Query().Get(queryParameterRef),
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

func parseGetOption(request *http.Request) (*GetOption, error) {
	opts := &GetOption{
		Kind:      request.URL.Query().Get(queryParameterKind),
		Name:      request.URL.Query().Get(queryParameterName),
		Namespace: request.URL.Query().Get(queryParameterNamespace),
	}
	if opts.Kind == "" || opts.Name == "" || opts.Namespace == "" {
		return nil, fmt.Errorf("the parameter should not be empty %v", opts)
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

func annotatedConfigs(store *serviceInstancesStore, cfgs []*config.Config, cgvk config.GroupVersionKind) {
	if cgvk != gvk.ServiceEntry {
		return
	}
	for i := range cfgs {
		num, err := store.refIndexNumber(keyForRefIndexer(cfgs[i]))
		if err != nil || num == 0 {
			continue
		}
		// should deep copy the configs.
		anno := cfgs[i].Annotations
		cfgs[i].Annotations = map[string]string{}
		cfgs[i].Annotations[annotationInstanceNumber] = strconv.Itoa(num)
		for key, val := range anno {
			cfgs[i].Annotations[key] = val
		}
	}
}
