package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

func main() {
	address := ":8080"
	prometheus.Register(requestCount)
	prometheus.Register(requestDuration)

	http.HandleFunc("/", errorHandler)
	http.HandleFunc("/healthz", healthHandler)
	http.Handle("/metrics", promhttp.Handler())

	log.Info().Msgf("Listening on %s", address)
	err := http.ListenAndServe(address, nil)
	log.Fatal().Msg(err.Error())
}
