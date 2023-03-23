package server

import (
	"k8s.io/client-go/tools/cache"
)

// stores all the service instances from SE, WLE.
type serviceInstancesStore struct {
	indexedStore cache.ThreadSafeStore
}

func (s *serviceInstancesStore) instancesBySE(key string) {
	return
}
func (s *serviceInstancesStore) instancesNumberBySE(key string) int {
	return 0
}
