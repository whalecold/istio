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

package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

func TestTrimBySidecarByOnDemandHosts(t *testing.T) {
	//TODO

}

func cmpError(err1, err2 error) string {
	if err1 == nil && err2 != nil {
		return fmt.Sprintf("err1 is nil, err2 is %v", err2)
	}
	if err2 == nil && err1 != nil {
		return fmt.Sprintf("err2 is nil, err1 is %v", err1)
	}
	if err1 != nil && err2 != nil {
		if err1.Error() != err2.Error() && !strings.Contains(err1.Error(), err2.Error()) {
			return fmt.Sprintf("err1 is %v, err2 is %v", err1, err2)
		}
	}
	return ""
}

func TestParseVirtualHostResourceName(t *testing.T) {
	g := gomega.NewWithT(t)

	testCases := []struct {
		err          error
		resourceName string
		port         int
		name         string
		domain       string
	}{
		{
			resourceName: "9080/productpage.bookinfo:9080",
			port:         9080,
			name:         "productpage.bookinfo:9080",
			domain:       "productpage.bookinfo",
		},
		{
			resourceName: "80/www.sample.com",
			port:         80,
			name:         "www.sample.com",
			domain:       "www.sample.com",
		},
		{
			resourceName: "9080/productpage.bookinfo:9081",
			port:         9080,
			name:         "productpage.bookinfo:9081",
			domain:       "productpage.bookinfo",
		},
		{
			resourceName: "908d/productpage.bookinfo:908d",
			err:          fmt.Errorf("invalid format resource name 908d/productpage.bookinfo:908d"),
		},
		{
			resourceName: "9080:9080/productpage.bookinfo:9081",
			err:          fmt.Errorf("invalid format resource name 9080:9080/productpage.bookinfo:9081"),
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			port, name, domain, err := ParseVirtualHostResourceName(tc.resourceName)
			g.Expect(cmpError(err, tc.err)).To(gomega.Equal(""))
			g.Expect(port).To(gomega.Equal(tc.port))
			g.Expect(name).To(gomega.Equal(tc.name))
			g.Expect(domain).To(gomega.Equal(tc.domain))
		})
	}
}
