package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

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

func init() {
	prometheus.Register(requestCount)
	prometheus.Register(requestDuration)
}

// Measure http middleware to count requests and measure request duration
func Measure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Now().Sub(start).Seconds()
		proto := strconv.Itoa(r.ProtoMajor)
		proto = fmt.Sprintf("%s.%s", proto, strconv.Itoa(r.ProtoMinor))

		requestCount.WithLabelValues(proto).Inc()
		requestDuration.WithLabelValues(proto).Observe(duration)
	})
}
