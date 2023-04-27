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
	"net/http"

	"istio.io/istio/pilot/pkg/model"
	http2 "istio.io/istio/pkg/adsc/server/http"
	"istio.io/istio/pkg/config"
)

func (s *server) configStoreHandler(_, cur config.Config, event model.Event) {
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

	convertor := configConvertor{store: s.indexedStore}

	conf := s.store.Get(gvk, opt.Name, opt.Namespace)
	if conf == nil {
		return nil, http2.NotFoundHander(fmt.Errorf("%s/%s/%s not found", opt.Kind, opt.Name, opt.Namespace))
	}

	metaObj, err := convertor.Convert(conf)
	if err != nil {
		return nil, http2.InternalServerHandler(err)
	}
	return metaObj, http2.OKHandler()
}

func (s *server) list(gvk config.GroupVersionKind, convertor configConvertor, opts *ListOptions) (*ListResult, error) {
	if !opts.isRefList() {
		filtered := model.NewFilterConfigGetAppender(opts)
		if err := s.store.ListToConfigAppender(gvk, model.NamespaceAll, filtered); err != nil {
			return nil, err
		}
		return configsToListResult(filtered.Configs(), convertor, opts)
	}

	confs, err := s.indexedStore.byWeLinkedToSeIndexer(opts.getRefKey())
	if err != nil {
		return nil, err
	}

	ret := make([]*config.Config, 0, len(confs))
	for i := range confs {
		conf := confs[i].(config.Config)
		if opts.Filter(&conf) {
			ret = append(ret, &conf)
		}
	}

	return configsToListResult(ret, convertor, opts)
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

	convertor := configConvertor{store: s.indexedStore}

	result, err := s.list(gvk, convertor, opt)
	if err != nil {
		return nil, http2.InternalServerHandler(err)
	}
	return result, http2.OKHandler()
}
