// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mseconfig

import (
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/cache"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/collection"
	"istio.io/istio/pkg/config/schema/collections"
	"istio.io/istio/pkg/kube"
)

// MseConfiguration used to watch rate limit configuration info
// and signal notify if change
type MseConfiguration struct {
	client     kube.Client
	store      cache.Store
	controller cache.Controller
	handlers   []model.EventHandler
	mutex      sync.Mutex
}

const (
	// LabelConfig used for config label, e.g. rate limit, should be consistent with mse-server
	LabelConfig = "trafficmanager.mse.paas.volcengine.com/config"
)

func NewMseConfiguration(client kube.Client) *MseConfiguration {
	res := &MseConfiguration{
		client: client,
	}

	f := func(event model.Event) {
		res.mutex.Lock()
		handlers := res.handlers
		res.mutex.Unlock()
		for _, h := range handlers {
			h(config.Config{
				Meta: config.Meta{
					GroupVersionKind: collections.MseConfiguration.Resource().GroupVersionKind(),
				},
			}, config.Config{
				Meta: config.Meta{
					GroupVersionKind: collections.MseConfiguration.Resource().GroupVersionKind(),
				},
			}, event)
		}
	}

	store, controller := cache.NewInformer(cache.NewFilteredListWatchFromClient(client.Kube().CoreV1().RESTClient(), "configmaps", metav1.NamespaceAll,
		func(options *metav1.ListOptions) {
			l, err := labels.NewRequirement(LabelConfig, selection.Exists, nil)
			if err != nil {
				panic(err)
			}
			options.LabelSelector = l.String()
		}), &corev1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			f(model.EventAdd)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			f(model.EventUpdate)
		},
		DeleteFunc: func(obj interface{}) {
			f(model.EventDelete)
		},
	})
	res.store = store
	res.controller = controller
	return res
}

func (r *MseConfiguration) RegisterEventHandler(typ config.GroupVersionKind, handler model.EventHandler) {
	if typ != collections.MseConfiguration.Resource().GroupVersionKind() {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.handlers = append(r.handlers, handler)
}

func (r *MseConfiguration) SetWatchErrorHandler(f func(r *cache.Reflector, err error)) error {
	// Todo
	return nil
}

func (r *MseConfiguration) HasStarted() bool {
	return r.client.HasStarted()
}

func (r *MseConfiguration) HasSynced() bool {
	return r.controller.HasSynced()
}

func (r *MseConfiguration) Run(stop <-chan struct{}) {
	r.controller.Run(stop)
}

func (r *MseConfiguration) Schemas() collection.Schemas {
	return collection.SchemasFor(collections.MseConfiguration)
}

func (r *MseConfiguration) Get(typ config.GroupVersionKind, name, namespace string) *config.Config {
	// Todo
	return nil
}

// List all configs, if namespace == "", list all
func (r *MseConfiguration) List(typ config.GroupVersionKind, namespace string) ([]config.Config, error) {
	if typ != collections.MseConfiguration.Resource().GroupVersionKind() {
		return nil, nil
	}

	var res []config.Config
	for _, s := range r.store.List() {
		cm, ok := s.(*corev1.ConfigMap)
		if !ok {
			continue
		}

		res = append(res, config.Config{
			Meta: config.Meta{
				GroupVersionKind: typ,
				Name:             cm.Name,
				Namespace:        cm.Namespace,
				Labels:           cm.Labels,
				Annotations:      cm.Annotations,
			},
			Spec: cm.Data,
		})
	}

	return res, nil
}

func (r *MseConfiguration) Create(config config.Config) (revision string, err error) {
	return "", nil
}

func (r *MseConfiguration) Update(config config.Config) (newRevision string, err error) {
	return "", nil
}

func (r *MseConfiguration) UpdateStatus(config config.Config) (newRevision string, err error) {
	return "", nil
}

func (r *MseConfiguration) Patch(orig config.Config, patchFn config.PatchFunc) (string, error) {
	return "", nil
}

func (r *MseConfiguration) Delete(typ config.GroupVersionKind, name, namespace string, resourceVersion *string) error {
	return nil
}

func (r *MseConfiguration) ListToConfigAppender(typ config.GroupVersionKind, namespace string, appender model.ConfigAppender) error {
	return nil
}
