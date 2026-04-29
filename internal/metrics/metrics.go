package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "event_receiver_events_received_total",
		Help: "Total number of events received from external sources",
	}, []string{"source", "event_type"})

	EventsPublished = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "event_receiver_events_published_total",
		Help: "Total number of events published to Kafka",
	}, []string{"source", "event_type", "status"})

	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "event_receiver_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status_code"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "event_receiver_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
)
