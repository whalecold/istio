package xdsclient

import (
	"context"
	"fmt"
	"testing"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"
	mcp "istio.io/api/mcp/v1alpha1"
	networking "istio.io/api/networking/v1alpha3"

	"istio.io/istio/pilot/pkg/config/memory"
	"istio.io/istio/pkg/config/schema/collections"
	"istio.io/istio/pkg/config/schema/gvk"
)

type fakeStream struct {
}

func (s *fakeStream) Send(*discovery.DeltaDiscoveryRequest) error {
	return nil
}
func (s *fakeStream) Recv() (*discovery.DeltaDiscoveryResponse, error) {
	return nil, nil
}

func (s *fakeStream) Header() (metadata.MD, error) {
	return nil, nil
}
func (s *fakeStream) Trailer() metadata.MD {
	return nil
}
func (s *fakeStream) CloseSend() error {
	return nil
}
func (s *fakeStream) Context() context.Context {
	return nil
}
func (s *fakeStream) SendMsg(m interface{}) error {
	return nil
}
func (s *fakeStream) RecvMsg(m interface{}) error {
	return nil
}

func marshalToResource(pb proto.Message, name, host string) (*discovery.Resource, error) {
	seAny, err := ptypes.MarshalAny(pb)
	if err != nil {
		return nil, err
	}
	mcpRes := &mcp.Resource{
		Metadata: &mcp.Metadata{
			Name: name,
			Labels: map[string]string{
				"host": host,
			},
		},
		Body: seAny,
	}
	resAny, err := ptypes.MarshalAny(mcpRes)
	if err != nil {
		return nil, err
	}
	return &discovery.Resource{
		Resource: resAny,
	}, nil

}

const (
	defaultNamespace = "test-auth-public-default-group"
)

func serviceEntry(name string, host string) (*networking.ServiceEntry, string, string) {
	return &networking.ServiceEntry{
		Hosts: []string{
			host,
		},
		Ports: []*networking.Port{
			{
				Name:     "http",
				Number:   80,
				Protocol: "HTTP",
			},
		},
		Resolution: networking.ServiceEntry_STATIC,
		Endpoints: []*networking.WorkloadEntry{
			{
				Address: "127.0.0.1",
				Ports: map[string]uint32{
					"http": 80,
				},
			},
		},
	}, fmt.Sprintf("%s/%s", defaultNamespace, name), host
}

func TestHandlerDeltaDiscoveryResponse(t *testing.T) {
	g := gomega.NewWithT(t)
	store := memory.MakeSkipValidation(collections.PilotMCP)
	configController := memory.NewController(store)
	s := &stream{
		stream:            &fakeStream{},
		store:             configController,
		resourceTypeNonce: map[string]string{},
	}

	res, err := marshalToResource(serviceEntry("servicea", ""))
	g.Expect(err).To(gomega.BeNil())
	resb, err := marshalToResource(serviceEntry("serviceb", ""))
	g.Expect(err).To(gomega.BeNil())

	// Sotw update
	s.handleDeltaDiscoveryResponse(&discovery.DeltaDiscoveryResponse{
		TypeUrl:   gvk.ServiceEntry.String(),
		Nonce:     "123",
		Resources: []*discovery.Resource{res},
	})

	cfg := s.store.Get(gvk.ServiceEntry, "servicea", defaultNamespace)
	g.Expect(cfg.Name).To(gomega.Equal("servicea"))

	// inc update -- delete
	s.handleDeltaDiscoveryResponse(&discovery.DeltaDiscoveryResponse{
		TypeUrl:          gvk.ServiceEntry.String(),
		Nonce:            "123",
		RemovedResources: []string{defaultNamespace + "/servicea"},
	})

	cfg = s.store.Get(gvk.ServiceEntry, "servicea", defaultNamespace)
	g.Expect(cfg).To(gomega.BeNil())

	// add again.
	s.handleDeltaDiscoveryResponse(&discovery.DeltaDiscoveryResponse{
		TypeUrl:   gvk.ServiceEntry.String(),
		Nonce:     "123",
		Resources: []*discovery.Resource{res},
	})
	cfg = s.store.Get(gvk.ServiceEntry, "servicea", defaultNamespace)
	g.Expect(cfg).NotTo(gomega.BeNil())

	// reconnect full update
	s2 := &stream{
		stream:            &fakeStream{},
		store:             configController,
		resourceTypeNonce: map[string]string{},
	}
	s2.handleDeltaDiscoveryResponse(&discovery.DeltaDiscoveryResponse{
		TypeUrl:   gvk.ServiceEntry.String(),
		Nonce:     "123",
		Resources: []*discovery.Resource{resb},
	})

	cfg = s.store.Get(gvk.ServiceEntry, "servicea", defaultNamespace)
	g.Expect(cfg).To(gomega.BeNil())

	cfg = s.store.Get(gvk.ServiceEntry, "serviceb", defaultNamespace)
	g.Expect(cfg.Name).To(gomega.Equal("serviceb"))

	// inr resource b update
	resbHost, err := marshalToResource(serviceEntry("serviceb", "mesh.volc.cn"))
	g.Expect(err).To(gomega.BeNil())

	s2.handleDeltaDiscoveryResponse(&discovery.DeltaDiscoveryResponse{
		TypeUrl:   gvk.ServiceEntry.String(),
		Nonce:     "123789",
		Resources: []*discovery.Resource{resbHost},
	})
	g.Expect(s2.resourceTypeNonce[gvk.ServiceEntry.String()]).To(gomega.Equal("123789"))
	cfg = s.store.Get(gvk.ServiceEntry, "serviceb", defaultNamespace)
	g.Expect(cfg.Meta.Labels["host"]).To(gomega.Equal("mesh.volc.cn"))
}
