package main

import (
	"net/http"

	"github.com/181192/custom-error-pages/pkg/handlers"
	"github.com/181192/custom-error-pages/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	gitCommit = "unversioned"
	version   = "unversioned"
	date      = "unversioned"
)

func main() {
	util.InitFlags()
	util.ConfigureLogger()

	http.Handle("/", handlers.ErrorPage())
	http.Handle("/metrics", handlers.Metrics())
	http.Handle("/healthz", handlers.Health())

	httpListenAddress := viper.GetString(util.HTTPListenAddress)
	log.Debug().Msgf("config values: %+v", viper.AllSettings())
	log.Info().Msgf("version=%s, commit=%s, date=%s", version, gitCommit, date)
	log.Info().Msgf("listening on %s", httpListenAddress)
	err := http.ListenAndServe(httpListenAddress, nil)
	log.Fatal().Msg(err.Error())
}
