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
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/config"
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

// ListResult list items.
type ListResult struct {
	// Total all the items number includes that are ignored or not in the range indexing.
	Total int `json:"total,omitempty"`
	// Items the items that meet the requests parameter.
	Items []interface{} `json:"items,omitempty"`
}

func configsToListResult(confs []*config.Config, convertor ConfigConvertor, opts *ListOptions) (*ListResult, error) {
	result := &ListResult{}

	total, paginated := paginateResource(opts, confs)
	result.Total = total
	result.Items = make([]interface{}, len(paginated))

	for i, config := range paginated {
		item, err := convertor.Convert(config)
		if err != nil {
			return nil, err
		}
		result.Items[i] = item
	}
	return result, nil
}

// Make sure ListOptions implements model.ConfigFilter at compile time.
var _ model.ConfigFilter = (*ListOptions)(nil)

// ListOptions options of list.
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

	isEmptyFilter bool
}

func (l *ListOptions) isRefList() bool {
	return len(l.refKey) != 0
}

func (l *ListOptions) getRefKey() string {
	return l.refKey
}

func (l *ListOptions) Contains(name string) bool {
	if l.Query == "" {
		return true
	}
	return strings.Contains(name, l.Query)
}

// Matchs return true if the labels matches.
func (l *ListOptions) Matchs(set labels.Labels) bool {
	if l.Selector == nil {
		return true
	}
	return l.Selector.Matches(set)
}

// Namespace if there is only one namespace in the map, return
// it as it can be used to list the resource in specified namespace. If the
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

// Filter checks if the config should in the final list result. Return `true`
// if the input object resides in the list filter.
func (l *ListOptions) Filter(cfg *config.Config) bool {
	if l.isEmptyFilter {
		return true
	}

	return l.InNamespaces(cfg.Namespace) &&
		l.Matchs(labels.Set(cfg.Labels)) &&
		l.Contains(cfg.Name)
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
		t1 := configs[i].CreationTimestamp.Nanosecond()
		t2 := configs[j].CreationTimestamp.Nanosecond()
		if t1 == t2 {
			if configs[i].Namespace == configs[j].Namespace {
				return configs[i].Name < configs[j].Name
			}
			return configs[i].Namespace < configs[j].Namespace
		}
		return t1 > t2
	})
}

func paginateResource(opt *ListOptions, confs []*config.Config) (total int, ret []*config.Config) {
	sortConfigByCreationTime(confs)

	total = len(confs)
	start, limit := opt.Start, opt.Limit
	if start < 0 {
		start = 0
	}

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
	if err != nil {
		return nil, err
	}

	if opts.Selector.Empty() && len(opts.Namespaces) == 0 && opts.Query == "" {
		opts.isEmptyFilter = true
	}

	return opts, nil
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
		// checks start >= 0
		if start < 0 {
			start = 0
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
