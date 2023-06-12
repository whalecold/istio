package v1alpha3

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

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
			err:          nil,
		},
		{
			resourceName: "908d/productpage.bookinfo:908d",
			err:          fmt.Errorf("invalid format resource name 908d/productpage.bookinfo:908d"),
		},
		{
			resourceName: "9080/productpage.bookinfo:9081",
			err:          fmt.Errorf("invalid format resource name 9080/productpage.bookinfo:9081"),
		},
		{
			resourceName: "9080:9080/productpage.bookinfo:9081",
			err:          fmt.Errorf("invalid format resource name 9080:9080/productpage.bookinfo:9081"),
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			port, name, domain, err := parseVirtualHostResourceName(tc.resourceName)
			g.Expect(cmpError(err, tc.err)).To(gomega.Equal(""))
			g.Expect(port).To(gomega.Equal(tc.port))
			g.Expect(name).To(gomega.Equal(tc.name))
			g.Expect(domain).To(gomega.Equal(tc.domain))
		})
	}
}
