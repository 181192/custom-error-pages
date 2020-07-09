package main

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "default_http_backend"
	subsystem = "http"
)

var (
	requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_count_total",
		Help:      "Counter of HTTP requests made.",
	}, []string{"proto"})

	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration_milliseconds",
		Help:      "Histogram of the time (in milliseconds) each request took.",
		Buckets:   append([]float64{.001, .003}, prometheus.DefBuckets...),
	}, []string{"proto"})
)
