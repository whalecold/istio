package httpserver

import (
	"strings"

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

// Error ...
type Error struct {
	Code    int               `json:"code,omitempty"`
	Message string            `json:"message,omitempty"`
	Data    map[string]string `json:"data,omitempty"`
}

// Response ...
type Response struct {
	Error  *Error      `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
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
