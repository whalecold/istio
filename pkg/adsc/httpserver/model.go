package httpserver

import "k8s.io/apimachinery/pkg/labels"

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
	Kind      string `query:"kind"`
	Name      string `query:"name"`
	Namespace string `query:"namespace"`
	// If Start and Limit are all zero, return all the resource meet the others conditions.
	Start    int `query:"start" default:"0"`
	Limit    int `query:"limit" default:"10"`
	Selector labels.Selector
}

// GetOption get option
type GetOption struct {
	Kind      string `query:"kind"`
	Name      string `query:"name"`
	Namespace string `query:"namespace"`
}
