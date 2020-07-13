package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// FormatHeader name of the header used to extract the format
	FormatHeader = "X-Format"

	// CodeHeader name of the header used as source of the HTTP status code to return
	CodeHeader = "X-Code"

	// ContentType name of the header that defines the format of the reply
	ContentType = "Content-Type"

	// OriginalURI name of the header with the original URL from NGINX
	OriginalURI = "X-Original-URI"

	// Namespace name of the header that contains information about the Ingress namespace
	Namespace = "X-Namespace"

	// IngressName name of the header that contains the matched Ingress
	IngressName = "X-Ingress-Name"

	// ServiceName name of the header that contains the matched Service in the Ingress
	ServiceName = "X-Service-Name"

	// ServicePort name of the header that contains the matched Service port in the Ingress
	ServicePort = "X-Service-Port"

	// RequestID is a unique ID that identifies the request - same as for backend service
	RequestID = "X-Request-ID"

	// JSON content-type for json
	JSON = "application/json"

	// HTML content-type for html
	HTML = "text/html"
)

type errorPageData struct {
	Code    string               `json:"code"`
	Title   string               `json:"title"`
	Message []string             `json:"message"`
	Details errorPageDataDetails `json:"details,omitempty"`
}

type errorPageDataDetails struct {
	OriginalURI string `json:"originalURI"`
	Namespace   string `json:"namespace"`
	IngressName string `json:"ingressName"`
	ServiceName string `json:"serviceName"`
	ServicePort string `json:"servicePort"`
	RequestID   string `json:"requestId"`
}

func newErrorPageData(req *http.Request, message []string) errorPageData {
	statusCode := req.Header.Get(CodeHeader)
	statusCodeNumber, _ := strconv.Atoi(req.Header.Get(CodeHeader))
	statusText := http.StatusText(statusCodeNumber)

	return errorPageData{
		Code:    statusCode,
		Title:   statusText,
		Message: message,
		Details: errorPageDataDetails{
			OriginalURI: req.Header.Get(OriginalURI),
			Namespace:   req.Header.Get(Namespace),
			IngressName: req.Header.Get(IngressName),
			ServiceName: req.Header.Get(ServiceName),
			ServicePort: req.Header.Get(ServicePort),
			RequestID:   req.Header.Get(RequestID),
		},
	}
}

func getFormat(req *http.Request) string {
	format := HTML
	formatHeader := req.Header[FormatHeader]

	if len(formatHeader) != 0 {

		formatString := strings.Split(formatHeader[0], ",")

		for i := range formatString {
			if formatString[i] == JSON {
				format = JSON
				break
			}
		}
	}

	return format
}

func getStatusCode(req *http.Request) int {
	errCode := req.Header.Get(CodeHeader)
	code, err := strconv.Atoi(errCode)
	if err != nil {
		code = 404
		log.Debug().Msgf("unexpected error reading return code: %v. Using %v", err, code)
	}

	return code
}

func getMessage(code int) []string {
	switch code {
	case http.StatusNotFound:
		return []string{"The page you're looking for could not be found."}
	case http.StatusServiceUnavailable:
		return []string{"Ooops, this shouldn't have happened.", "The server is temporary busy, try again later!"}
	default:
		return []string{http.StatusText(code)}
	}
}

// HTMLResponse returns html reponse
func HTMLResponse(w http.ResponseWriter, r *http.Request, path string) {
	code := getStatusCode(r)
	message := getMessage(code)

	w.Header().Set(ContentType, HTML)
	w.WriteHeader(code)

	stylesPath := fmt.Sprintf("%v/styles.css", path)
	styles, err := os.Open(stylesPath)

	file := fmt.Sprintf("%v/template.html", path)
	f, err := os.Open(file)
	if err != nil {
		log.Warn().Msgf("unexpected error opening file: %v", err)
		JSONResponse(w, r)
		return
	}
	defer f.Close()

	log.Debug().Msgf("serving custom error response for code %v and format %v from file %v", code, HTML, file)
	tmpl := template.Must(template.ParseFiles(f.Name(), styles.Name()))

	data := newErrorPageData(r, message)
	tmpl.Execute(w, data)
}

// JSONResponse returns json reponse
func JSONResponse(w http.ResponseWriter, r *http.Request) {
	code := getStatusCode(r)
	message := getMessage(code)

	w.Header().Set(ContentType, JSON)
	w.WriteHeader(code)
	body, _ := json.Marshal(newErrorPageData(r, message))
	w.Write(body)
}

// ServeHttp error handler
func (opts *Options) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	format := getFormat(r)

	switch format {
	case JSON:
		JSONResponse(w, r)
	default:
		HTMLResponse(w, r, opts.ErrFilesPath)
	}

	duration := time.Now().Sub(start).Seconds()

	proto := strconv.Itoa(r.ProtoMajor)
	proto = fmt.Sprintf("%s.%s", proto, strconv.Itoa(r.ProtoMinor))

	requestCount.WithLabelValues(proto).Inc()
	requestDuration.WithLabelValues(proto).Observe(duration)
}
