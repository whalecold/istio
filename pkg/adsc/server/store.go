package server

import (
	"strings"

	networking "istio.io/api/networking/v1alpha3"
	"k8s.io/client-go/tools/cache"

	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/constants"
	"istio.io/istio/pkg/config/schema/gvk"
)

const (
	refIndexerName = "refIndexer"
)

func refIndexer(obj interface{}) ([]string, error) {
	cfg, ok := obj.(config.Config)
	if !ok {
		return nil, nil
	}
	we, ok := cfg.Spec.(*networking.WorkloadEntry)
	if !ok {
		return nil, nil
	}
	// Should be in the same namespace when using refIndexer.
	// Use the service account as it is kept consistent with the ServiceEntry name.

	str := []string{gvk.ServiceEntry.CanonicalGroup(), gvk.ServiceEntry.Version,
		gvk.ServiceEntry.Kind, cfg.Namespace, we.ServiceAccount}
	return []string{
		strings.Join(str, "/"),
	}, nil
}

func keyForConfigFunc(cfg config.Config) string {
	source, _ := cfg.Annotations[constants.MCPServerSource]
	str := []string{source, cfg.GroupVersionKind.CanonicalGroup(), cfg.GroupVersionKind.Version,
		cfg.GroupVersionKind.Kind, cfg.Namespace, cfg.Name}
	return strings.Join(str, "/")
}

func keyForRefIndexer(cfg *config.Config) string {
	str := []string{cfg.GroupVersionKind.CanonicalGroup(), cfg.GroupVersionKind.Version,
		cfg.GroupVersionKind.Kind, cfg.Namespace, cfg.Name}
	return strings.Join(str, "/")
}

func newStore() *serviceInstancesStore {
	return &serviceInstancesStore{
		indexedStore: cache.NewThreadSafeStore(cache.Indexers{
			refIndexerName: refIndexer,
		}, cache.Indices{}),
	}
}

// stores all the service instances from SE, WLE.
type serviceInstancesStore struct {
	indexedStore cache.ThreadSafeStore
}

func (s *serviceInstancesStore) byRefIndexer(key string) ([]interface{}, error) {
	return s.indexedStore.ByIndex(refIndexerName, key)
}
func (s *serviceInstancesStore) refIndexNumber(key string) (int, error) {
	keys, err := s.indexedStore.IndexKeys(refIndexerName, key)
	return len(keys), err
}

func (s *serviceInstancesStore) Delete(cfg config.Config) {
	s.indexedStore.Delete(keyForConfigFunc(cfg))
}

func (s *serviceInstancesStore) Add(cfg config.Config) {
	s.indexedStore.Add(keyForConfigFunc(cfg), cfg)
}

func (s *serviceInstancesStore) Update(cfg config.Config) {
	s.indexedStore.Update(keyForConfigFunc(cfg), cfg)
}
