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

package adsc

import (
	"testing"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/onsi/gomega"

	"istio.io/istio/pilot/pkg/config/memory"
	"istio.io/istio/pilot/pkg/model"
	mcpaggregate "istio.io/istio/pkg/adsc/aggregate"
	"istio.io/istio/pkg/adsc/mcpdiscovery"
	"istio.io/istio/pkg/adsc/xdsclient"
	"istio.io/istio/pkg/config/schema/collections"
)

type fakeClient struct {
	upstreamName string
	url          string
	nodeName     string
}

func (c *fakeClient) GetURL() string { return c.url }
func (c *fakeClient) GetID() string {
	return c.upstreamName
}

func (c *fakeClient) Run() error {
	return nil
}

func (c *fakeClient) Close() {}

func (c *fakeClient) GetStore() model.ConfigStoreController {
	store := memory.MakeSkipValidation(collections.PilotMCP)
	configController := memory.NewController(store)
	return configController
}

func TestSyncHandler(t *testing.T) {
	g := gomega.NewWithT(t)
	m := &MultiADSC{
		cfg: &Config{},
		buildXDSClient: func(s string, s2 string, s3 string, requests []*discovery.DeltaDiscoveryRequest) (xdsclient.XDSClient, error) {
			return &fakeClient{
				upstreamName: s,
				url:          s2,
				nodeName:     s3,
			}, nil
		},
		adscs: make(map[string]xdsclient.XDSClient),
		store: mcpaggregate.MakeCache(collections.PilotMCP),
	}

	// add
	err := m.OnServersUpdate([]*mcpdiscovery.McpServer{
		{
			ID:      "client",
			Address: "127.0.0.1:8080",
		},
	})
	g.Expect(err).To(gomega.BeNil())
	adsc := m.adscs["client"]
	g.Expect(adsc.GetURL()).To(gomega.Equal("127.0.0.1:8080"))

	// chang
	err = m.OnServersUpdate([]*mcpdiscovery.McpServer{
		{
			ID:      "client",
			Address: "127.0.0.1:8081",
		},
		{
			ID:      "client2",
			Address: "127.0.0.1:8080",
		},
	})
	g.Expect(err).To(gomega.BeNil())
	adsc = m.adscs["client"]
	g.Expect(adsc.GetURL()).To(gomega.Equal("127.0.0.1:8081"))
	adsc = m.adscs["client2"]
	g.Expect(adsc.GetURL()).To(gomega.Equal("127.0.0.1:8080"))

	// remove
	err = m.OnServersUpdate([]*mcpdiscovery.McpServer{
		{
			ID:      "client2",
			Address: "127.0.0.1:8080",
		},
	})
	g.Expect(err).To(gomega.BeNil())
	adsc = m.adscs["client"]
	g.Expect(adsc).To(gomega.BeNil())
}
