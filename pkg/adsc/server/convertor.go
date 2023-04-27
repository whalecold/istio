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
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	clientv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/gvk"
)

// ConfigConvertor converts config.Config to metav1.Object.
type ConfigConvertor interface {
	Convert(*config.Config) (metav1.Object, error)
}

type configConvertor struct {
	store *serviceInstancesStore
}

func (cc configConvertor) Convert(conf *config.Config) (metav1.Object, error) {
	switch conf.Spec.(type) {
	case *networkingv1alpha3.ServiceEntry:
		cc.annotateInstancesNum(conf)
		return convertToK8sServiceEntry(conf)

	case *networkingv1alpha3.WorkloadEntry:
		return convertToK8sWorkloadEntry(conf)

	default:
		return nil, fmt.Errorf("unsupported config type:%s",
			conf.GroupVersionKind.String())
	}
}

// FIXME(wangjian.pg 20230421), according to the comments and the implementation of the `store`
// type(`cache.ThreadSafeStore`), we should not modify anything returned by `List/Get` method of
// the store since it will break the indexing feature in addition to not being thread safe.
func (cc configConvertor) annotateInstancesNum(conf *config.Config) {
	if conf.GroupVersionKind != gvk.ServiceEntry {
		return
	}
	num, err := cc.store.byWeLinkedToSeIndexerNumber(keyForRefIndexer(conf))
	if err != nil || num == 0 {
		return
	}
	// should deep copy the configs.
	anno := conf.Annotations
	conf.Annotations = make(map[string]string, len(anno)+1)
	conf.Annotations[annotationInstanceNumber] = strconv.Itoa(num)
	for key, val := range anno {
		conf.Annotations[key] = val
	}
}

func convertToK8sWorkloadEntry(res *config.Config) (metav1.Object, error) {
	we, ok := res.Spec.(*networkingv1alpha3.WorkloadEntry)
	if !ok {
		return nil, fmt.Errorf("except type:%s, got:%v",
			gvk.WorkloadEntry.String(), res.Spec)
	}
	return &clientv1alpha3.WorkloadEntry{
		TypeMeta:   convertToK8sType(gvk.WorkloadEntry),
		ObjectMeta: convertToK8sMeta(res.Meta),
		Spec:       *we.DeepCopy(),
	}, nil
}

func convertToK8sServiceEntry(res *config.Config) (metav1.Object, error) {
	se, ok := res.Spec.(*networkingv1alpha3.ServiceEntry)
	if !ok {
		return nil, fmt.Errorf("except type:%s, got:%v",
			gvk.ServiceEntry.String(), res.Spec)
	}
	return &clientv1alpha3.ServiceEntry{
		TypeMeta:   convertToK8sType(gvk.ServiceEntry),
		ObjectMeta: convertToK8sMeta(res.Meta),
		Spec:       *se.DeepCopy(),
	}, nil
}

func convertToK8sType(meta config.GroupVersionKind) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       meta.Kind,
		APIVersion: meta.GroupVersion(),
	}
}

func convertToK8sMeta(meta config.Meta) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace:   meta.Namespace,
		Name:        meta.Name,
		Labels:      meta.Labels,
		Annotations: meta.Annotations,
		CreationTimestamp: metav1.Time{
			Time: meta.CreationTimestamp,
		},
	}
}
