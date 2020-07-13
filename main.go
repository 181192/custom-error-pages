package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	httpListenAddress    = "listen"
	httpListenAddressEnv = "HTTP_LISTEN_ADDRESS"

	debug    = "debug"
	debugEnv = "DEBUG"

	logColor    = "log-color"
	logColorEnv = "LOG_COLOR"

	errFilesPath    = "error-files-path"
	errFilesPathEnv = "ERROR_FILES_PATH"
)

// Options cli options
type Options struct {
	HTTPListenAddress string
	Debug             bool
	ColorLog          bool
	ErrFilesPath      string
}

func main() {

	var opts Options

	flag.BoolVar(&opts.Debug, debug, LookupEnvOrBool(debugEnv, false), "sets log level to debug")
	flag.BoolVar(&opts.ColorLog, logColor, LookupEnvOrBool(logColorEnv, false), "sets log format to human-friendly, colorized output")
	flag.StringVar(&opts.HTTPListenAddress, httpListenAddress, LookupEnvOrString(httpListenAddressEnv, ":8080"), "http server address")
	flag.StringVar(&opts.ErrFilesPath, errFilesPath, LookupEnvOrString(errFilesPathEnv, "./www"), "the location on disk of files served by the handler")

	flag.Parse()

	if opts.ColorLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if opts.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	prometheus.Register(requestCount)
	prometheus.Register(requestDuration)

	http.Handle("/", &opts)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", healthHandler)

	log.Info().Msgf("Config values: %+v", getConfig(flag.CommandLine))
	log.Info().Msgf("Listening on %s", opts.HTTPListenAddress)
	err := http.ListenAndServe(opts.HTTPListenAddress, nil)
	log.Fatal().Msg(err.Error())
}

// LookupEnvOrString lookup env from key or return default value
func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

// LookupEnvOrInt lookup env from key or return default value
func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal().Msgf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

// LookupEnvOrBool lookup env from key or return default value
func LookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatal().Msgf("LookupEnvOrBool[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func getConfig(fs *flag.FlagSet) []string {
	cfg := make([]string, 0, 10)
	fs.VisitAll(func(f *flag.Flag) {
		cfg = append(cfg, fmt.Sprintf("%s:%q", f.Name, f.Value.String()))
	})

	return cfg
}
