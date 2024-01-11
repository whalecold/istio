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

package bootstrap

import (
	"fmt"

	"istio.io/istio/pilot/pkg/serviceregistry/aggregate"
	kubecontroller "istio.io/istio/pilot/pkg/serviceregistry/kube/controller"
	"istio.io/istio/pilot/pkg/serviceregistry/provider"
	"istio.io/istio/pilot/pkg/serviceregistry/serviceentry"
	"istio.io/pkg/log"
)

func (s *Server) ServiceController() *aggregate.Controller {
	return s.environment.ServiceDiscovery.(*aggregate.Controller)
}

// initServiceControllers creates and initializes the service controllers
func (s *Server) initServiceControllers(args *PilotArgs) error {
	serviceControllers := s.ServiceController()
	seControllerOptions := []serviceentry.Option{
		serviceentry.WithClusterID(s.clusterID),
	}

	registered := make(map[provider.ID]bool)
	for _, r := range args.RegistryOptions.Registries {
		serviceRegistry := provider.ID(r)
		if _, exists := registered[serviceRegistry]; exists {
			log.Warnf("%s registry specified multiple times.", r)
			continue
		}
		registered[serviceRegistry] = true
		log.Infof("Adding %s registry adapter", serviceRegistry)
		switch serviceRegistry {
		case provider.Kubernetes:
			controller := s.initKubeRegistry(args)
			seControllerOptions = append(seControllerOptions, serviceentry.WithLocalityGetters(controller))
		default:
			return fmt.Errorf("service registry %s is not supported", r)
		}
	}

	s.serviceEntryController = serviceentry.NewController(
		s.configController, s.XDSServer,
		seControllerOptions...,
	)
	serviceControllers.AddRegistry(s.serviceEntryController)

	// Defer running of the service controllers.
	s.addStartFunc(func(stop <-chan struct{}) error {
		go serviceControllers.Run(stop)
		return nil
	})

	return nil
}

// initKubeRegistry creates all the k8s service controllers under this pilot
func (s *Server) initKubeRegistry(args *PilotArgs) *kubecontroller.Multicluster {
	args.RegistryOptions.KubeOptions.ClusterID = s.clusterID
	args.RegistryOptions.KubeOptions.Metrics = s.environment
	args.RegistryOptions.KubeOptions.XDSUpdater = s.XDSServer
	args.RegistryOptions.KubeOptions.NetworksWatcher = s.environment.NetworksWatcher
	args.RegistryOptions.KubeOptions.MeshWatcher = s.environment.Watcher
	args.RegistryOptions.KubeOptions.SystemNamespace = args.Namespace
	args.RegistryOptions.KubeOptions.MeshServiceController = s.ServiceController()
	// pass namespace to k8s service registry
	args.RegistryOptions.KubeOptions.DiscoveryNamespacesFilter = s.multiclusterController.DiscoveryNamespacesFilter

	ctr := kubecontroller.NewMulticluster(args.PodName,
		s.kubeClient.Kube(),
		args.RegistryOptions.ClusterRegistriesNamespace,
		args.RegistryOptions.KubeOptions,
		s.serviceEntryController,
		s.configController,
		s.istiodCertBundleWatcher,
		args.Revision,
		s.shouldStartNsController(),
		s.environment.ClusterLocal(),
		s.server)
	s.multiclusterController.AddHandler(ctr)
	return ctr
}
