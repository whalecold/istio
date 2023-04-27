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

package mcpdiscovery

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"istio.io/istio/pkg/config/constants"
	"istio.io/pkg/log"
)

var adscLog = log.RegisterScope("adsc", "adsc debugging", 0)

const (
	// maxRetries is the number of times a configmap will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a configmap is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15

	controllerName = "multiADSC"

	resyncInterval = 0
)

// DeRegister log off the mcp server adrress to kubernetes.
func (md *mcpDiscovery) DeRegister(ctx context.Context, namespace, name, id string) error {
	cm, err := md.kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if errors.IsNotFound(err) {
		return nil
	}
	// clone from memory.
	cm = cm.DeepCopy()

	delete(cm.Data, id)
	_, err = md.kubeClient.CoreV1().ConfigMaps(constants.IstioSystemNamespace).Update(ctx, cm, v1.UpdateOptions{})
	return err
}

// Register the mcp server adrress to kubernetes.
func (md *mcpDiscovery) Register(ctx context.Context, namespace, name string, address McpServer) error {
	cm, err := md.kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if errors.IsNotFound(err) {
		cm = &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: constants.IstioSystemNamespace,
			},
			Data: map[string]string{
				address.ID: address.Address,
			},
		}
		_, err = md.kubeClient.CoreV1().ConfigMaps(constants.IstioSystemNamespace).Create(ctx, cm, v1.CreateOptions{})
		return err
	}
	// clone from memory.
	cm = cm.DeepCopy()

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data[address.ID] = address.Address
	_, err = md.kubeClient.CoreV1().ConfigMaps(constants.IstioSystemNamespace).Update(ctx, cm, v1.UpdateOptions{})
	return err
}

// McpServer mcp server address.
type McpServer struct {
	ID      string
	Address string
}

// DisconveryHandler is a callback function that can be invoked when the configuration changed.
type DisconveryHandler interface {
	OnServersUpdate(servers []*McpServer) error
}

// Registry the interface of registry.
type Registry interface {
	Register(context.Context, string, string, McpServer) error
	DeRegister(context.Context, string, string, string) error
}

// Discovery the interface of discovery.
type Discovery interface {
	// RegisterHandler should invoke HasSynced and wait for the `true` return value
	// before invoke RegisterHandler. Only support registry once.
	RegisterHandler(dh DisconveryHandler)
	Run(stopCh <-chan struct{})
}

// DiscoveryRegistry the interface of discovery and registry.
type DiscoveryRegistry interface {
	Discovery
	Registry
}

type mcpDiscovery struct {
	// configMapName the configMap name of the mcp server configuration.
	configMapName string
	// configMapNamespace the configMap namespace of the mcp server configuration.
	configMapNamespace string
	// To allow injection of syncConfigMap for testing.
	syncHandler func(dKey string) error

	// ConfigMap that need to be synced
	queue workqueue.RateLimitingInterface

	// configMapListerSynced returns true if the ConfigMap store has been synced at least once.
	// Added as a member to the struct to allow injection for testing.
	configMapListerSynced cache.InformerSynced
	// configMapLister can list/get configMap from the shared informer's store
	configMapLister corelisters.ConfigMapLister

	mcpSyncHandler  func(servers []*McpServer) error
	kubeClient      kubernetes.Interface
	enableDiscovery bool
	kubeInformer    informers.SharedInformerFactory
}

// Options the options for discovery.
type Options struct {
	EnableDiscovery bool
}

func listMcpConfigMap(opts *v1.ListOptions) {}

// New return a new discovery registry.
func New(cli kubernetes.Interface, namespace, name string, opts *Options) DiscoveryRegistry {
	md := &mcpDiscovery{
		kubeClient:         cli,
		configMapName:      name,
		configMapNamespace: namespace,
		enableDiscovery:    opts.EnableDiscovery,
		queue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mcp-kubernetes-discovery"),
	}
	if !opts.EnableDiscovery {
		return md
	}

	md.kubeInformer = informers.NewFilteredSharedInformerFactory(cli, resyncInterval, constants.IstioSystemNamespace, listMcpConfigMap)
	informer := md.kubeInformer.Core().V1().ConfigMaps()
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    md.addConfigMap,
		UpdateFunc: md.updateConfigMap,
		// This will enter the sync loop and no-op, because the configmap has been deleted from the store.
		DeleteFunc: md.deleteConfigMap,
	})
	md.configMapListerSynced = informer.Informer().HasSynced
	md.configMapLister = informer.Lister()
	md.syncHandler = md.syncConfigMap
	return md
}

// HasSynced returns true if the discovery has been ready.
func (md *mcpDiscovery) HasSynced() bool {
	return md.configMapListerSynced()
}

func (md *mcpDiscovery) shouldEnqueue(cm *corev1.ConfigMap) bool {
	return cm.Namespace == md.configMapNamespace && cm.Name == md.configMapName
}

func (md *mcpDiscovery) addConfigMap(obj interface{}) {
	d := obj.(*corev1.ConfigMap)
	if !md.shouldEnqueue(d) {
		return
	}
	md.enqueueConfigMap(d)
}

func (md *mcpDiscovery) deleteConfigMap(obj interface{}) {
	d, ok := obj.(*corev1.ConfigMap)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*corev1.ConfigMap)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a Deployment %#v", obj))
			return
		}
	}
	if !md.shouldEnqueue(d) {
		return
	}
	md.enqueueConfigMap(d)
}

func (md *mcpDiscovery) updateConfigMap(old, cur interface{}) {
	curCM := cur.(*corev1.ConfigMap)
	if !md.shouldEnqueue(curCM) {
		return
	}
	md.enqueueConfigMap(curCM)
}

func (md *mcpDiscovery) enqueueConfigMap(cm *corev1.ConfigMap) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cm)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", cm, err))
		return
	}
	md.queue.Add(key)
}

func (md *mcpDiscovery) processNextWorkItem() bool {
	key, quit := md.queue.Get()
	if quit {
		return false
	}
	defer md.queue.Done(key)

	err := md.syncHandler(key.(string))
	md.handleErr(err, key)

	return true
}

func (md *mcpDiscovery) handleErr(err error, key interface{}) {
	if err == nil || errors.HasStatusCause(err, corev1.NamespaceTerminatingCause) {
		md.queue.Forget(key)
		return
	}

	ns, name, keyErr := cache.SplitMetaNamespaceKey(key.(string))
	if keyErr != nil {
		adscLog.Errorf("Failed to split meta namespace cache key %s keyErr %v, err %v", key, keyErr, err)
	}

	if md.queue.NumRequeues(key) < maxRetries {
		adscLog.Infof("Error syncing configmap %s/%s error %v", ns, name, err)
		md.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	adscLog.Infof("Dropping configmap out of the queue configmap %s/%s error %v", ns, name, err)
	md.queue.Forget(key)
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (md *mcpDiscovery) worker() {
	for md.processNextWorkItem() {
	}
}

// RegisterHandler registry handler
func (md *mcpDiscovery) RegisterHandler(dh DisconveryHandler) {
	md.mcpSyncHandler = dh.OnServersUpdate
}

func (md *mcpDiscovery) WaitForCacheSync(stopCh <-chan struct{}) bool {
	return cache.WaitForCacheSync(stopCh, md.configMapListerSynced)
}

// Run begins watching and syncing, will maintain the grpc clients by the configuration from configmap.
func (md *mcpDiscovery) Run(stopCh <-chan struct{}) {
	if md.enableDiscovery {
		go md.kubeInformer.Start(stopCh)
		defer utilruntime.HandleCrash()
		defer md.queue.ShutDown()

		adscLog.Infof("Starting controller %v", controllerName)
		defer adscLog.Infof("Shutting down controller %v", controllerName)

		if !cache.WaitForNamedCacheSync(controllerName, stopCh, md.configMapListerSynced) {
			return
		}
		go wait.Until(md.worker, time.Second, stopCh)
	}
	<-stopCh
}

func (md *mcpDiscovery) syncConfigMap(key string) error {
	if md.mcpSyncHandler == nil {
		return fmt.Errorf("synce handler is empty")
	}
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		adscLog.Errorf("Failed to split meta namespace cache key %v", key)
		return err
	}
	startTime := time.Now()
	adscLog.Infof("Started syncing configmap %s/%s startTime: %s", namespace, name, startTime.String())
	defer func() {
		adscLog.Infof("Finished syncing configmap %s/%s duration: %s", namespace, name, time.Since(startTime))
	}()

	cm, err := md.configMapLister.ConfigMaps(namespace).Get(name)
	if errors.IsNotFound(err) {
		return md.mcpSyncHandler(nil)
	}
	if err != nil {
		return err
	}
	servers := make([]*McpServer, 0, len(cm.Data))
	for key, val := range cm.Data {
		servers = append(servers, &McpServer{
			ID:      key,
			Address: val,
		})
	}
	return md.mcpSyncHandler(servers)
}
