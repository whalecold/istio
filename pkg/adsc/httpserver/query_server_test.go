package httpserver

import (
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/onsi/gomega"

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
			url:  "http://127.0.0.1:18001/mcp.istio.io/v1alpha1/resource?namespace=ns&query=hello&start=10&q=q2&limit=2&kind=serviceentry&name=hello",
			body: `{"a":"apple", "b":"banana"}`,
			want: &ListOptions{
				Kind:      "serviceentry",
				Name:      "hello",
				Namespace: "ns",
				Start:     10,
				Limit:     2,
				Labels: map[string]string{
					"a": "apple",
					"b": "banana",
				},
			},
		},
		{
			url: "http://127.0.0.1:18001/mcp.istio.io/v1alpha1/resource?namespace=ns&query=hello&q=q2&kind=serviceentry&name=hello",
			want: &ListOptions{
				Kind:      "serviceentry",
				Name:      "hello",
				Namespace: "ns",
				Start:     0,
				Limit:     10,
				Labels:    map[string]string{},
			},
		},
	}
	s := &server{}
	for _, tc := range testCases {
		req, err := http.NewRequest("GET", tc.url, strings.NewReader(tc.body))
		if err != nil {
			log.Fatalln(err)
		}
		opts, err := s.parseListOptions(req)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(opts).To(gomega.Equal(tc.want))
	}
}

func TestFilterByOptions(t *testing.T) {
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
						Name: "c1",
						Labels: map[string]string{
							"app":     "c",
							"version": "v1",
						},
					},
				},
			},
			opts: &ListOptions{
				Labels: map[string]string{
					"version": "v1",
				},
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
				Name: "c2",
				Labels: map[string]string{
					"version": "v1",
				},
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
		got := paginateResource(tc.opts, tc.input)
		g.Expect(got).To(gomega.Equal(tc.want))
	}
}
