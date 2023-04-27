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
	"testing"

	"github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	clientv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/gvk"
)

func BenchmarkConvertToK8sServiceEntry(b *testing.B) {
	for i := 0; i < b.N; i++ {
		convertToK8sWorkloadEntry(&config.Config{
			Meta: config.Meta{
				Namespace: "ns",
				Name:      "name",
				Labels: map[string]string{
					"app":     "hello",
					"version": "v1",
				},
			},
			Spec: &networkingv1alpha3.WorkloadEntry{
				Address: "1.1.1.1",
				Ports: map[string]uint32{
					"http": 80,
				},
				Labels: map[string]string{
					"app":     "hello",
					"version": "v1",
				},
			},
		})
	}
}

func TestConvert(t *testing.T) {
	g := gomega.NewWithT(t)

	con := &configConvertor{}

	testCases := []struct {
		conf *config.Config
		want metav1.Object
	}{
		{
			conf: &config.Config{
				Meta: config.Meta{
					Namespace: "ns",
					Name:      "name",
					Labels: map[string]string{
						"app":     "hello",
						"version": "v1",
					},
				},
				Spec: &networkingv1alpha3.WorkloadEntry{
					Address: "1.1.1.1",
					Ports: map[string]uint32{
						"http": 80,
					},
					Labels: map[string]string{
						"app":     "hello",
						"version": "v1",
					},
					ServiceAccount: "sa",
				},
			},
			want: &clientv1alpha3.WorkloadEntry{
				TypeMeta: metav1.TypeMeta{
					Kind:       gvk.WorkloadEntry.Kind,
					APIVersion: gvk.WorkloadEntry.GroupVersion(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "name",
					Labels: map[string]string{
						"app":     "hello",
						"version": "v1",
					},
				},
				Spec: networkingv1alpha3.WorkloadEntry{
					Address: "1.1.1.1",
					Ports: map[string]uint32{
						"http": 80,
					},
					Labels: map[string]string{
						"app":     "hello",
						"version": "v1",
					},
					ServiceAccount: "sa",
				},
			},
		},
		{
			conf: &config.Config{
				Meta: config.Meta{
					Namespace: "ns",
					Name:      "name",
					Labels: map[string]string{
						"app":     "hello",
						"version": "v1",
					},
				},
				Spec: &networkingv1alpha3.ServiceEntry{
					Hosts: []string{"1.1.1.1", "www.baidu.com"},
					Ports: []*networkingv1alpha3.Port{
						{
							Number:   80,
							Protocol: "http",
						},
					},
				},
			},
			want: &clientv1alpha3.ServiceEntry{
				TypeMeta: metav1.TypeMeta{
					Kind:       gvk.ServiceEntry.Kind,
					APIVersion: gvk.ServiceEntry.GroupVersion(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "name",
					Labels: map[string]string{
						"app":     "hello",
						"version": "v1",
					},
				},
				Spec: networkingv1alpha3.ServiceEntry{
					Hosts: []string{"1.1.1.1", "www.baidu.com"},
					Ports: []*networkingv1alpha3.Port{
						{
							Number:   80,
							Protocol: "http",
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		got, err := con.Convert(tc.conf)
		g.Expect(err).To(gomega.BeNil())
		switch v := got.(type) {
		case *clientv1alpha3.WorkloadEntry:
			g.Expect(v.ObjectMeta).To(gomega.Equal(tc.want.(*clientv1alpha3.WorkloadEntry).ObjectMeta))
			g.Expect(v.TypeMeta).To(gomega.Equal(tc.want.(*clientv1alpha3.WorkloadEntry).TypeMeta))
			g.Expect(proto.Equal(&v.Spec, &tc.want.(*clientv1alpha3.WorkloadEntry).Spec)).To(gomega.BeTrue())
		case *clientv1alpha3.ServiceEntry:
			g.Expect(v.ObjectMeta).To(gomega.Equal(tc.want.(*clientv1alpha3.ServiceEntry).ObjectMeta))
			g.Expect(v.TypeMeta).To(gomega.Equal(tc.want.(*clientv1alpha3.ServiceEntry).TypeMeta))
			g.Expect(proto.Equal(&v.Spec, &tc.want.(*clientv1alpha3.ServiceEntry).Spec)).To(gomega.BeTrue())
		}
	}
}
