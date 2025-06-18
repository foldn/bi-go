package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal Log the total number of HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestDuration Record the latency of HTTP requests
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "myapp_http_request_duration_seconds",
			Help:    "Histogram of HTTP request latencies.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// JobsProcessedTotal Record the total number of jobs that have been processed
	JobsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_jobs_processed_total",
			Help: "Total number of processed jobs.",
		},
		[]string{"status"},
	)

	// ActiveBackgroundWorkers Keep track of the number of currently active back-office workers
	ActiveBackgroundWorkers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "myapp_active_background_workers",
			Help: "Number of currently active background job workers.",
		},
	)
)
