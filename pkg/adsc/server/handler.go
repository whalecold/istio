package server

import (
	"fmt"
	"net/http"

	"istio.io/istio/pilot/pkg/model"
	http2 "istio.io/istio/pkg/adsc/server/http"
	"istio.io/istio/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (s *server) list(gvk config.GroupVersionKind, convert convertFn, opts *ListOptions) ([]metav1.Object, error) {
	if !opts.isRefList() {
		objs, err := s.store.List(gvk, opts.Namespace())
		if err != nil {
			return nil, err
		}

		ret := make([]metav1.Object, 0, len(objs))
		for i := range objs {
			if opts.skip(objs[i]) {
				continue
			}
			ret = append(ret, convert(gvk, &objs[i]))
		}
		return ret, nil
	}

	confs, err := s.indexedStore.byRefIndexer(opts.getRefKey())
	if err != nil {
		return nil, err
	}

	ret := make([]metav1.Object, 0, len(confs))
	for i := range confs {
		conf := confs[i].(config.Config)
		if opts.skip(conf) {
			continue
		}
		temp := convert(gvk, &conf)
		if temp != nil {
			ret = append(ret, temp)
		}
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

	confs, err := s.list(gvk, convert, opt)
	if err != nil {
		return nil, http2.InternalServerHandler(err)
	}

	sortConfigByCreationTime(confs)
	total, confs := paginateResource(opt, confs)

	annotatedConfigs(s.indexedStore, confs, gvk)

	return &ResourceList{
		Total: total,
		Items: confs,
	}, http2.OKHandler()
}
