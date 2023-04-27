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
	"testing"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pkg/config/constants"
	"istio.io/istio/pkg/kube"
)

const (
	configmapName = "multimcp-conf"
)

func TestRegisterMcpServerAddress(t *testing.T) {
	g := gomega.NewWithT(t)
	cli := kube.NewFakeClient()
	ctx := context.Background()

	r := New(cli.Kube(), "istio-system", configmapName, &Options{})

	// register the first one.
	err := r.Register(ctx, constants.IstioSystemNamespace, configmapName, McpServer{
		ID:      "a1",
		Address: "127.0.0.1:8080",
	})
	g.Expect(err).To(gomega.BeNil())

	cm, err := cli.Kube().CoreV1().ConfigMaps(constants.IstioSystemNamespace).Get(ctx, configmapName, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(cm.Data).To(gomega.Equal(map[string]string{
		"a1": "127.0.0.1:8080",
	}))

	// register the second one.
	err = r.Register(ctx, constants.IstioSystemNamespace, configmapName, McpServer{
		ID:      "a2",
		Address: "127.0.0.1:8181",
	})
	g.Expect(err).To(gomega.BeNil())

	cm, err = cli.Kube().CoreV1().ConfigMaps(constants.IstioSystemNamespace).Get(ctx, configmapName, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(cm.Data).To(gomega.Equal(map[string]string{
		"a1": "127.0.0.1:8080",
		"a2": "127.0.0.1:8181",
	}))

	// log off.
	err = r.DeRegister(ctx, constants.IstioSystemNamespace, configmapName, "a3")
	g.Expect(err).To(gomega.BeNil())

	cm, err = cli.Kube().CoreV1().ConfigMaps(constants.IstioSystemNamespace).Get(ctx, configmapName, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(cm.Data).To(gomega.Equal(map[string]string{
		"a1": "127.0.0.1:8080",
		"a2": "127.0.0.1:8181",
	}))

	err = r.DeRegister(ctx, constants.IstioSystemNamespace, configmapName, "a1")
	g.Expect(err).To(gomega.BeNil())

	cm, err = cli.Kube().CoreV1().ConfigMaps(constants.IstioSystemNamespace).Get(ctx, configmapName, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(cm.Data).To(gomega.Equal(map[string]string{
		"a2": "127.0.0.1:8181",
	}))
}
