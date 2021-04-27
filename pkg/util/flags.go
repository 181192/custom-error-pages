package util

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// HTTPListenAddress address server listens on host:port
	HTTPListenAddress = "http-listen-address"
	// Debug if debug logs should be enabled
	Debug = "debug"
	// LogColor sets log format to human-friendly, colorized output
	LogColor = "log-color"

	// ErrFilesPath the location on disk of files served by the handler
	ErrFilesPath = "error-files-path"

	// HideDetails hide request details in response
	HideDetails = "hide-details"
)

// InitFlags initialize viper and pflags
func InitFlags() {
	pflag.Bool(Debug, false, "enable debug log")
	pflag.Bool(LogColor, false, "sets log format to human-friendly, colorized output")
	pflag.String(HTTPListenAddress, ":8080", "http server address")
	pflag.String(ErrFilesPath, "./themes/knockout", "the location on disk of files served by the handler")
	pflag.Bool(HideDetails, false, "hide request details in response")

	pflag.Parse()
	viper.AutomaticEnv()
	viper.BindPFlags(pflag.CommandLine)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

// ConfigureLogger configure log output and levels
func ConfigureLogger() {
	if viper.GetBool(LogColor) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if viper.GetBool(Debug) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
