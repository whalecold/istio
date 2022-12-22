package mcpdiscovery

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	"istio.io/istio/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pkg/config/constants"
)

func TestRegisterMcpServerAddress(t *testing.T) {
	g := gomega.NewWithT(t)
	cli := kube.NewFakeClient()
	ctx := context.Background()

	// register the first one.
	err := RegisterMcpServerAddress(ctx, cli, McpServer{
		ID:      "a1",
		Address: "127.0.0.1:8080",
	})
	g.Expect(err).To(gomega.BeNil())

	cm, err := cli.CoreV1().ConfigMaps(constants.IstioSystemNamespace).Get(ctx, configmapName, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(cm.Data).To(gomega.Equal(map[string]string{
		"a1": "127.0.0.1:8080",
	}))

	// register the second one.
	err = RegisterMcpServerAddress(ctx, cli, McpServer{
		ID:      "a2",
		Address: "127.0.0.1:8181",
	})
	g.Expect(err).To(gomega.BeNil())

	cm, err = cli.CoreV1().ConfigMaps(constants.IstioSystemNamespace).Get(ctx, configmapName, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(cm.Data).To(gomega.Equal(map[string]string{
		"a1": "127.0.0.1:8080",
		"a2": "127.0.0.1:8181",
	}))

	// log off.
	err = LogOffMcpServerAddress(ctx, cli, McpServer{
		ID:      "a3",
		Address: "127.0.0.1:8181",
	})
	g.Expect(err).To(gomega.BeNil())

	cm, err = cli.CoreV1().ConfigMaps(constants.IstioSystemNamespace).Get(ctx, configmapName, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(cm.Data).To(gomega.Equal(map[string]string{
		"a1": "127.0.0.1:8080",
		"a2": "127.0.0.1:8181",
	}))

	err = LogOffMcpServerAddress(ctx, cli, McpServer{
		ID:      "a1",
		Address: "127.0.0.1:8181",
	})
	g.Expect(err).To(gomega.BeNil())

	cm, err = cli.CoreV1().ConfigMaps(constants.IstioSystemNamespace).Get(ctx, configmapName, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(cm.Data).To(gomega.Equal(map[string]string{
		"a2": "127.0.0.1:8181",
	}))
}
