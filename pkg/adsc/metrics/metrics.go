package metrics

import "github.com/prometheus/client_golang/prometheus"

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
)

func init() {
	prometheus.DefaultRegisterer.MustRegister(MCPOverXDSClientReceiveResponseTotal)
	prometheus.DefaultRegisterer.MustRegister(MCPOverXDSClientReceiveResponseDuration)
}
