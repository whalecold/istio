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

package aggregate

import (
	"sort"
	"sync"

	"github.com/hashicorp/go-multierror"
	"k8s.io/client-go/tools/cache"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/collection"
)

// ConfigStoreCache the wrapper of model.ConfigStoreCache, add the external AddStore and RemoveStore function to
// implement the dynamic store aggregate.
type ConfigStoreCache interface {
	model.ConfigStoreController
	AddStore(string, model.ConfigStoreController)
	RemoveStore(string)
}

type eventHandler struct {
	kind    config.GroupVersionKind
	handler model.EventHandler
}

// store the implementation of dynamic store aggregate. Only support the readable interface.
type store struct {
	// defaultScheme default scheme. The RegisterEventHandler is invoked before AddStore, as the
	// [storeCache](pkg/config/aggregate/config.go) will judge whether it has the scheme before
	// invoke RegisterEventHandler and ignore store if not exist, so we need add default scheme first,
	// otherwise the handlers will always be empty.
	defaultScheme collection.Schemas
	// schemas is the unified
	schemas collection.Schemas

	// stores is a mapping from config type to stores
	stores map[config.GroupVersionKind]map[string]model.ConfigStore
	// stores is a mapping from instance to a storeCache
	caches map[string]model.ConfigStoreController
	// the cache for RegisterEventHandler, should add it to the new store which is added after the
	// invoking of function RegisterEventHandler.
	handlers []*eventHandler
	// the cache for SetWatchErrorHandler, should add it to the new store which is added after the
	// invoking of function SetWatchErrorHandler.
	errorHandlers []func(r *cache.Reflector, err error)

	lock sync.RWMutex
}

// MakeCache creates an aggregate config store cache. Could add or remove several config store cache dynamically.
func MakeCache(defaultScheme collection.Schemas) ConfigStoreCache {
	s := &store{
		stores: map[config.GroupVersionKind]map[string]model.ConfigStore{},
		caches: map[string]model.ConfigStoreController{},
	}
	s.defaultScheme = defaultScheme
	s.rebuildScheme()
	return s
}

// AddStore add store dynamically. Overwrite the old store which has the same name, should clean the
// store with same name before.
func (cr *store) AddStore(storeName string, csc model.ConfigStoreController) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	if csc == nil {
		return
	}
	for _, s := range csc.Schemas().All() {
		stores, ok := cr.stores[s.Resource().GroupVersionKind()]
		if !ok {
			cr.stores[s.Resource().GroupVersionKind()] = map[string]model.ConfigStore{
				storeName: csc,
			}
			continue
		}
		stores[storeName] = csc
	}
	cr.caches[storeName] = csc

	// need registry the event handlers which registries before the store.
	for _, h := range cr.handlers {
		csc.RegisterEventHandler(h.kind, h.handler)
	}
	for _, f := range cr.errorHandlers {
		csc.SetWatchErrorHandler(f) //nolint
	}
	cr.rebuildScheme()
}

// rebuildScheme rebuild all the scheme
func (cr *store) rebuildScheme() {
	var schemas []collection.Schema
	union := collection.NewSchemasBuilder()
	for _, cache := range cr.caches {
		schemas = append(schemas, cache.Schemas().All()...)
	}
	schemas = append(schemas, cr.defaultScheme.All()...)
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Name() < schemas[j].Name()
	})
	for i := range schemas {
		_ = union.Add(schemas[i])
	}
	cr.schemas = union.Build()
}

// RemoveStore remove store dynamically. If not found the storeName, ignore it.
func (cr *store) RemoveStore(storeName string) {
	cr.lock.Lock()
	defer cr.lock.Unlock()
	cache, ok := cr.caches[storeName]
	if !ok {
		return
	}

	delete(cr.caches, storeName)
	schemas := cache.Schemas().All()
	gvks := make([]config.GroupVersionKind, len(schemas))
	for i, s := range schemas {
		gvk := s.Resource().GroupVersionKind()
		stores, ok := cr.stores[gvk]
		if ok {
			delete(stores, storeName)
		}
		gvks[i] = gvk
	}
	cr.rebuildScheme()
}

// Schemas return schemas.
func (cr *store) Schemas() collection.Schemas {
	return cr.schemas
}

// Get the first config found in the stores.
func (cr *store) Get(typ config.GroupVersionKind, name, namespace string) *config.Config {
	cr.lock.RLock()
	defer cr.lock.RUnlock()
	for _, s := range cr.stores[typ] {
		config := s.Get(typ, name, namespace)
		if config != nil {
			return config
		}
	}
	return nil
}

// List all configs to the appender.
func (cr *store) ListToConfigAppender(typ config.GroupVersionKind, namespace string, appender model.ConfigAppender) error {
	var stores []model.ConfigStore
	cr.lock.RLock()
	for _, s := range cr.stores[typ] {
		stores = append(stores, s)
	}
	cr.lock.RUnlock()

	if len(stores) == 0 {
		return nil
	}

	var errs *multierror.Error
	for _, s := range stores {
		err := s.ListToConfigAppender(typ, namespace, appender)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
	}
	return errs.ErrorOrNil()
}

// List all configs in the stores.
func (cr *store) List(typ config.GroupVersionKind, namespace string) ([]config.Config, error) {
	var stores []model.ConfigStore
	cr.lock.RLock()
	for _, s := range cr.stores[typ] {
		stores = append(stores, s)
	}
	cr.lock.RUnlock()

	if len(stores) == 0 {
		return nil, nil
	}

	var errs *multierror.Error
	var configs []config.Config
	// Used to remove duplicated config
	configMap := make(map[string]struct{})
	for _, s := range stores {
		storeConfigs, err := s.List(typ, namespace)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		for _, config := range storeConfigs {
			key := config.Namespace + "/" + config.Name
			// ignore the data with the same namespace/name between stores.
			if _, exist := configMap[key]; exist {
				continue
			}
			configs = append(configs, config)
			configMap[key] = struct{}{}
		}
	}
	return configs, errs.ErrorOrNil()
}

// Delete all store invoke the Delete function individually, do nothing.
// The same as Create Update UpdateStatus Patch and HasSynced.
func (cr *store) Delete(typ config.GroupVersionKind, name, namespace string, resourceVersion *string) error {
	return nil
}

func (cr *store) Create(c config.Config) (string, error) {
	return "", nil
}

func (cr *store) Update(c config.Config) (string, error) {
	return "", nil
}

func (cr *store) UpdateStatus(c config.Config) (string, error) {
	return "", nil
}

func (cr *store) Patch(orig config.Config, patchFn config.PatchFunc) (string, error) {
	return "", nil
}

// HasSynced the arregate should always true as the store in it can
// be removed or added dynamically.
func (cr *store) HasSynced() bool {
	cr.lock.Lock()
	defer cr.lock.Unlock()
	for _, s := range cr.caches {
		if !s.HasSynced() {
			return false
		}
	}
	return true
}

// RegisterEventHandler adds a handler to receive config update events for a
// configuration type
func (cr *store) RegisterEventHandler(kind config.GroupVersionKind, handler model.EventHandler) {
	cr.lock.Lock()
	defer cr.lock.Unlock()
	for _, cache := range cr.caches {
		cache.RegisterEventHandler(kind, handler)
	}
	// need cache the event handler as some store may be added latter.
	cr.handlers = append(cr.handlers, &eventHandler{
		kind:    kind,
		handler: handler,
	})
}

func (cr *store) SetWatchErrorHandler(handler func(r *cache.Reflector, err error)) error {
	cr.lock.Lock()
	defer cr.lock.Unlock()
	var errs error
	for _, cache := range cr.caches {
		if err := cache.SetWatchErrorHandler(handler); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	cr.errorHandlers = append(cr.errorHandlers, handler)
	return errs
}

// Run all store invoke the Run function individually, do nothing.
func (cr *store) Run(stop <-chan struct{}) {
	<-stop
}

func (cr *store) HasStarted() bool {
	return cr.HasSynced()
}
