package server

import (
	"fmt"
	"net/http"

	"istio.io/istio/pilot/pkg/model"
	http2 "istio.io/istio/pkg/adsc/server/http"
	"istio.io/istio/pkg/config"
)

func (s *server) serviceEntryHandler(old, cur config.Config, event model.Event) {
	// TODO stores the cache
}

func (s *server) workloadEntryHandler(old, cur config.Config, event model.Event) {
	// TODO stores the cache
}

func (s *server) handleGetRequest(request *http.Request) (interface{}, *http2.Error) {
	opt, err := parseGetOption(request)
	if err != nil {
		return nil, http2.BadRequestHander(err)
	}

	gvk, ok := s.kinds[opt.Kind]
	if !ok {
		return nil, http2.BadRequestHander(fmt.Errorf("the kind %s is not support", opt.Kind))
	}
	convert, ok := s.convertFns[gvk]
	if !ok {
		return nil, http2.BadRequestHander(fmt.Errorf("the kind %v is not support", opt.Kind))
	}

	obj := s.store.Get(gvk, opt.Name, opt.Namespace)
	if obj == nil {
		return nil, http2.NotFoundHander(fmt.Errorf("%s/%s/%s not found", opt.Kind, opt.Name, opt.Namespace))
	}
	return convert(gvk, obj), http2.OKHandler()
}

func (s *server) handleListRequest(request *http.Request) (interface{}, *http2.Error) {
	opt, err := parseListOptions(request)
	if err != nil {
		return nil, http2.BadRequestHander(err)
	}
	gvk, ok := s.kinds[opt.Kind]
	if !ok {
		return nil, http2.BadRequestHander(fmt.Errorf("the kind %s is not support", opt.Kind))
	}

	convert, ok := s.convertFns[gvk]
	if !ok {
		return nil, http2.BadRequestHander(fmt.Errorf("the kind %s is not support", opt.Kind))
	}

	cfgs, err := s.store.List(gvk, opt.Namespace())
	if err != nil {
		return nil, http2.InternalServerHandler(err)
	}

	cfgs = filterByOptions(cfgs, opt)
	sortConfigByCreationTime(cfgs)
	total, cfgs := paginateResource(opt, cfgs)

	items := make([]interface{}, 0, len(cfgs))
	for idx := range cfgs {
		r := &cfgs[idx]
		items = append(items, convert(gvk, r))
	}

	return &ResourceList{
		Total: total,
		Items: items,
	}, http2.OKHandler()
}
