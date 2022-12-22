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
	"istio.io/istio/pkg/config"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	queryParameterName       = "name"
	queryParameterNamespace  = "namespace"
	queryParameterNamespaces = "namespaces"
	queryParameterKeyword    = "keyword"
	queryParameterKind       = "kind"
	queryParameterStart      = "start"
	queryParameterLimit      = "limit"
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
	Kind       string `query:"kind"`
	Keyword    string `query:"keyword"`
	Namespaces map[string]bool
	Selector   labels.Selector
	// If Start and Limit are all zero, return all the resource meet the others conditions.
	Start int `query:"start" default:"0"`
	Limit int `query:"limit" default:"10"`
}

func (l *ListOptions) IsEmpty() bool {
	return l.Selector == nil && l.Keyword == "" && len(l.Namespaces) == 0
}

func (l *ListOptions) Contains(name string) bool {
	if l.Keyword == "" {
		return true
	}
	return strings.Contains(name, l.Keyword)
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

// GetOption get option
type GetOption struct {
	Kind      string `query:"kind"`
	Name      string `query:"name"`
	Namespace string `query:"namespace"`
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
