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
	"strings"

	"k8s.io/client-go/tools/cache"

	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/constants"
	"istio.io/istio/pkg/config/schema/gvk"
)

const (
	linkedIndexerName = "weLinkedToSeIndexer"
	serviceAppLabel   = "registrysynctask.mse.paas.volcengine.com/mseApp"
)

// workloadEntryRefServiceEntryIndexer be used to find the WorkloadEntry that belongs to
// the ServiceEntry has the same name as the service account of WorkloadEntry.
func workloadEntryLinkedToServiceEntryIndexer(obj interface{}) ([]string, error) {
	cfg, ok := obj.(config.Config)
	if !ok {
		return nil, nil
	}
	we, ok := cfg.Spec.(*networking.WorkloadEntry)
	if !ok {
		return nil, nil
	}
	appName, ok := we.Labels[serviceAppLabel]
	if !ok {
		return nil, nil
	}

	// Should be in the same namespace when using refIndexer.
	// Use the service account as it is kept consistent with the ServiceEntry name.
	return []string{
		strings.Join([]string{
			gvk.ServiceEntry.CanonicalGroup(),
			gvk.ServiceEntry.Version,
			gvk.ServiceEntry.Kind,
			cfg.Namespace,
			appName,
		}, "/"),
	}, nil
}

func keyForConfig(cfg config.Config) string {
	return strings.Join([]string{
		cfg.Annotations[constants.MCPServerSource],
		cfg.GroupVersionKind.CanonicalGroup(),
		cfg.GroupVersionKind.Version,
		cfg.GroupVersionKind.Kind,
		cfg.Namespace,
		cfg.Name,
	}, "/")
}

func keyForRefIndexer(cfg *config.Config) string {
	return strings.Join([]string{
		cfg.GroupVersionKind.CanonicalGroup(),
		cfg.GroupVersionKind.Version,
		cfg.GroupVersionKind.Kind,
		cfg.Namespace,
		cfg.Name,
	}, "/")
}

func newStore() *serviceInstancesStore {
	return &serviceInstancesStore{
		indexedStore: cache.NewThreadSafeStore(cache.Indexers{
			linkedIndexerName: workloadEntryLinkedToServiceEntryIndexer,
		}, cache.Indices{}),
	}
}

// stores all the service instances from SE, WLE.
type serviceInstancesStore struct {
	indexedStore cache.ThreadSafeStore
}

func (s *serviceInstancesStore) byWeLinkedToSeIndexer(key string) ([]interface{}, error) {
	return s.indexedStore.ByIndex(linkedIndexerName, key)
}

func (s *serviceInstancesStore) byWeLinkedToSeIndexerNumber(key string) (int, error) {
	keys, err := s.indexedStore.IndexKeys(linkedIndexerName, key)
	return len(keys), err
}

func (s *serviceInstancesStore) Delete(cfg config.Config) {
	s.indexedStore.Delete(keyForConfig(cfg))
}

func (s *serviceInstancesStore) Add(cfg config.Config) {
	s.indexedStore.Add(keyForConfig(cfg), cfg)
}

func (s *serviceInstancesStore) Update(cfg config.Config) {
	s.indexedStore.Update(keyForConfig(cfg), cfg)
}
