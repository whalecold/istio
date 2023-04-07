package server

import (
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	clientv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pkg/config"
)

func convertToK8sWorkloadEntry(gvk config.GroupVersionKind, res *config.Config) metav1.Object {
	we, ok := res.Spec.(*networkingv1alpha3.WorkloadEntry)
	if !ok {
		return nil
	}
	return &clientv1alpha3.WorkloadEntry{
		TypeMeta:   convertToK8sType(gvk),
		ObjectMeta: convertToK8sMate(res.Meta),
		Spec:       *we,
	}
}

func convertToK8sServiceEntry(gvk config.GroupVersionKind, res *config.Config) metav1.Object {
	se, ok := res.Spec.(*networkingv1alpha3.ServiceEntry)
	if !ok {
		return nil
	}
	return &clientv1alpha3.ServiceEntry{
		TypeMeta:   convertToK8sType(gvk),
		ObjectMeta: convertToK8sMate(res.Meta),
		Spec:       *se,
	}
}

func convertToK8sType(meta config.GroupVersionKind) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       meta.Kind,
		APIVersion: meta.GroupVersion(),
	}
}

func convertToK8sMate(meta config.Meta) metav1.ObjectMeta {
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
