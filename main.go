package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	address := ":8080"
	prometheus.Register(requestCount)
	prometheus.Register(requestDuration)

	http.HandleFunc("/", errorHandler)
	http.HandleFunc("/healthz", healthHandler)
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Listening on %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
