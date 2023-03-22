package metrics

import "github.com/prometheus/client_golang/prometheus"

// metrics
var (
	// MCPOverXDSClientReceiveResponseTotal ...
	MCPOverXDSClientReceiveResponseTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_over_xds_client_receive_response_total",
			Help: "the total response number of the client received ",
		},
		[]string{"client", "node", "kind"},
	)

	// MCPOverXDSClientReceiveResponseDuration ...
	MCPOverXDSClientReceiveResponseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "mcp_over_xds_client_receive_response_duration",
			Help: "the duration response of the client deal with the received ",
		},
		[]string{"client", "node", "kind"},
	)

	MCPServerRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_server_requests_total",
			Help: "the total number of requests by the mcp server",
		},
		[]string{"kind", "method", "status"},
	)
	MCPServerRequestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "mcp_server_requests_duration",
			Help: "the duration of requests by the mcp server",
		},
		[]string{"kind", "method", "status"},
	)
)

func init() {
	prometheus.DefaultRegisterer.MustRegister(MCPOverXDSClientReceiveResponseTotal)
	prometheus.DefaultRegisterer.MustRegister(MCPOverXDSClientReceiveResponseDuration)
	prometheus.DefaultRegisterer.MustRegister(MCPServerRequestsTotal)
	prometheus.DefaultRegisterer.MustRegister(MCPServerRequestsDuration)
}
