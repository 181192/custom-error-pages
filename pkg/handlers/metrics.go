package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics handler
func Metrics() http.Handler {
	return promhttp.Handler()
}
