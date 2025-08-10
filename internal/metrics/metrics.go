package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	CacheGetTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_get_total",
			Help: "Cache get operations",
		},
		[]string{"store", "outcome"}, // outcome=hit|miss|error
	)

	TaskGetTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_get_total",
			Help: "Task get operations by source",
		},
		[]string{"source", "outcome"}, // source=cache|db, outcome=hit|miss|error
	)

	TaskCreateTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_create_total",
			Help: "Task create operations",
		},
		[]string{"outcome"}, // outcome=ok|error
	)
)

func init() {
	prometheus.MustRegister(HttpRequestsTotal, HttpRequestDuration, CacheGetTotal, TaskGetTotal, TaskCreateTotal)
}
