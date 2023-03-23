package server

import (
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/labels"

	"istio.io/istio/pkg/config"
)

func TestParseListOptions(t *testing.T) {
	g := gomega.NewWithT(t)

	testCases := []struct {
		url  string
		body string
		want *ListOptions
	}{
		{
			url:  "http://127.0.0.1:18001/mcp.istio.io/v1alpha1/resource?namespaces=ns&start=10&q=q2&limit=2&kind=serviceentry&keyword=hello",
			body: `a=apple,b=banana`,
			want: &ListOptions{
				Kind:       "serviceentry",
				Keyword:    "hello",
				Namespaces: map[string]bool{"ns": true},
				Start:      10,
				Limit:      2,
				Selector: labels.SelectorFromSet(labels.Set{
					"a": "apple",
					"b": "banana",
				}),
			},
		},
		{
			url:  "http://127.0.0.1:18001/mcp.istio.io/v1alpha1/resource?namespaces=ns,ns1&query=hello&q=q2&kind=workloadentry&keyword=hello",
			body: `a=apple,b=banana,location=hangzhou`,
			want: &ListOptions{
				Kind:       "workloadentry",
				Keyword:    "hello",
				Namespaces: map[string]bool{"ns": true, "ns1": true},
				Start:      0,
				Limit:      10,
				Selector: labels.SelectorFromSet(labels.Set{
					"a":        "apple",
					"b":        "banana",
					"location": "hangzhou",
				}),
			},
		},
	}
	for _, tc := range testCases {
		req, err := http.NewRequest("GET", tc.url, strings.NewReader(tc.body))
		if err != nil {
			log.Fatalln(err)
		}
		opts, err := parseListOptions(req)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(opts).To(gomega.Equal(tc.want))
	}
}

func TestFilterByOptions(t *testing.T) {
	g := gomega.NewWithT(t)
	selectin, _ := labels.Parse("app in (c1, c2)")
	selectEqual, _ := labels.Parse("registrysynctask.mse.paas.volcengine.com/workspaceId=qbmsqu0j4cd1zr5uoe7s9uqr")
	testCases := []struct {
		input []config.Config
		want  []config.Config
		opts  *ListOptions
	}{
		{
			input: []config.Config{
				{
					Meta: config.Meta{
						Name: "c1",
						Labels: map[string]string{
							"app":     "c",
							"version": "v1",
							"registrysynctask.mse.paas.volcengine.com/workspaceId": "qbmsqu0j4cd1zr5uoe7s9uqr",
						},
					},
				},
				{
					Meta: config.Meta{
						Name: "c2",
						Labels: map[string]string{
							"app":     "c",
							"version": "v2",
						},
					},
				},
			},
			want: []config.Config{
				{
					Meta: config.Meta{
						Name: "c1",
						Labels: map[string]string{
							"app":     "c",
							"version": "v1",
							"registrysynctask.mse.paas.volcengine.com/workspaceId": "qbmsqu0j4cd1zr5uoe7s9uqr",
						},
					},
				},
			},
			opts: &ListOptions{
				Selector:   selectEqual,
				Namespaces: map[string]bool{},
			},
		},
		{
			input: []config.Config{
				{
					Meta: config.Meta{
						Name: "c1",
						Labels: map[string]string{
							"app":     "c",
							"version": "v1",
						},
					},
				},
				{
					Meta: config.Meta{
						Name: "c2",
						Labels: map[string]string{
							"app":     "c",
							"version": "v2",
						},
					},
				},
			},
			want: []config.Config{},
			opts: &ListOptions{
				Keyword: "c2",
				Selector: labels.SelectorFromSet(labels.Set{
					"version": "v1",
				}),
			},
		},
		{
			input: []config.Config{
				{
					Meta: config.Meta{
						Name: "c1",
						Labels: map[string]string{
							"app":     "c1",
							"version": "v1",
						},
					},
				},
				{
					Meta: config.Meta{
						Name: "c2",
						Labels: map[string]string{
							"app":     "c2",
							"version": "v2",
						},
					},
				},
				{
					Meta: config.Meta{
						Name: "c3",
						Labels: map[string]string{
							"app":     "c3",
							"version": "v1",
						},
					},
				},
			},
			want: []config.Config{
				{
					Meta: config.Meta{
						Name: "c1",
						Labels: map[string]string{
							"app":     "c1",
							"version": "v1",
						},
					},
				},
				{
					Meta: config.Meta{
						Name: "c2",
						Labels: map[string]string{
							"app":     "c2",
							"version": "v2",
						},
					},
				},
			},
			opts: &ListOptions{
				Selector: selectin,
			},
		},
	}

	for _, tc := range testCases {
		got := filterByOptions(tc.input, tc.opts)
		g.Expect(got).To(gomega.Equal(tc.want))
	}
}

func TestPaginateResource(t *testing.T) {
	g := gomega.NewWithT(t)
	testCases := []struct {
		input []config.Config
		want  []config.Config
		opts  *ListOptions
	}{
		{
			input: []config.Config{
				{
					Meta: config.Meta{
						Name: "c1",
						Labels: map[string]string{
							"app":     "c",
							"version": "v1",
						},
					},
				},
				{
					Meta: config.Meta{
						Name: "c2",
						Labels: map[string]string{
							"app":     "c",
							"version": "v2",
						},
					},
				},
			},
			want: []config.Config{
				{
					Meta: config.Meta{
						Name: "c2",
						Labels: map[string]string{
							"app":     "c",
							"version": "v2",
						},
					},
				},
			},
			opts: &ListOptions{
				Limit: 1,
				Start: 1,
			},
		},
		{
			input: []config.Config{
				{
					Meta: config.Meta{
						Name: "c1",
						Labels: map[string]string{
							"app":     "c",
							"version": "v1",
						},
					},
				},
				{
					Meta: config.Meta{
						Name: "c2",
						Labels: map[string]string{
							"app":     "c",
							"version": "v2",
						},
					},
				},
			},
			want: []config.Config{},
			opts: &ListOptions{
				Start: 3,
				Limit: 3,
			},
		},
	}

	for _, tc := range testCases {
		_, got := paginateResource(tc.opts, tc.input)
		g.Expect(got).To(gomega.Equal(tc.want))
	}
}