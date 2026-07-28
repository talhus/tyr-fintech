package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// declare global metrics
var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed by the Gin engine",
		},
		[]string{"method", "path", "status"},
	)

	//HTTP request duration in measures request latency duration in seconds
	HTTPReqDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request durations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	ActiveSSEConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tyr_active_sse_connections",
			Help: "Number of active SSE (Server-Sent Events) connections",
		},
	)
)
