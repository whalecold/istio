package server

import (
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pilot/pkg/model"
	http2 "istio.io/istio/pkg/adsc/server/http"
	"istio.io/istio/pkg/config"
)

func (s *server) serviceEntryHandler(_, cur config.Config, event model.Event) {
	switch event {
	case model.EventDelete:
		s.indexedStore.Delete(cur)
	case model.EventAdd:
		s.indexedStore.Add(cur)
	case model.EventUpdate:
		s.indexedStore.Update(cur)
	}
}

func (s *server) workloadEntryHandler(_, cur config.Config, event model.Event) {
	switch event {
	case model.EventDelete:
		s.indexedStore.Delete(cur)
	case model.EventAdd:
		s.indexedStore.Add(cur)
	case model.EventUpdate:
		s.indexedStore.Update(cur)
	}
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

func (s *server) list(gvk config.GroupVersionKind, opts *ListOptions) ([]*config.Config, error) {
	if !opts.isRefList() {
		cache := s.cb.Build()
		defer cache.Close()
		cache.AppendFilter(func(conf *config.Config) bool {
			return opts.skip(conf)
		})
		err := s.store.ListWithCache(gvk, model.NamespaceAll, cache)
		if err != nil {
			return nil, err
		}
		return cache.Configs(), nil
	}

	confs, err := s.indexedStore.byRefIndexer(opts.getRefKey())
	if err != nil {
		return nil, err
	}

	ret := make([]*config.Config, 0, len(confs))
	for i := range confs {
		conf := confs[i].(config.Config)
		if opts.skip(&conf) {
			continue
		}
		ret = append(ret, &conf)
	}
	return ret, nil
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

	confs, err := s.list(gvk, opt)
	if err != nil {
		return nil, http2.InternalServerHandler(err)
	}

	sortConfigByCreationTime(confs)
	total, confs := paginateResource(opt, confs)

	annotatedConfigs(s.indexedStore, confs, gvk)

	ret := make([]metav1.Object, len(confs))
	for idx, c := range confs {
		ret[idx] = convert(gvk, c)
	}

	return &ResourceList{
		Total: total,
		Items: confs,
	}, http2.OKHandler()
}
