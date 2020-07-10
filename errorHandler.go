package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
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

	// ErrFilesPathVar is the name of the environment variable indicating
	// the location on disk of files served by the handler.
	ErrFilesPathVar = "ERROR_FILES_PATH"

	// JSON content-type for json
	JSON = "application/json"

	// HTML content-type for html
	HTML = "text/html"
)

type errorPageData struct {
	Code    string               `json:"code"`
	Title   string               `json:"title"`
	Message string               `json:"message"`
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

func newErrorPageData(req *http.Request, message string) errorPageData {
	statusCode := req.Header.Get(CodeHeader)
	statusCodeNumber, _ := strconv.Atoi(req.Header.Get(CodeHeader))
	statusText := http.StatusText(statusCodeNumber)

	if message == "" {
		message = statusText
	}

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

func getBaseErrorFilePath() string {
	errFilesPath := "./www"
	if os.Getenv(ErrFilesPathVar) != "" {
		errFilesPath = os.Getenv(ErrFilesPathVar)
	}

	return errFilesPath
}

func getFormat(req *http.Request) string {
	format := "text/html"
	formatHeader := strings.Split(req.Header[FormatHeader][0], ",")

	for i := range formatHeader {
		if formatHeader[i] == JSON {
			format = JSON
			break
		}
	}

	return format
}

func getStatusCode(req *http.Request) int {
	errCode := req.Header.Get(CodeHeader)
	code, err := strconv.Atoi(errCode)
	if err != nil {
		code = 404
		log.Printf("unexpected error reading return code: %v. Using %v", err, code)
	}

	return code
}

// HTMLResponse returns html reponse
func HTMLResponse(w http.ResponseWriter, r *http.Request) {
	path := getBaseErrorFilePath()
	code := getStatusCode(r)

	w.Header().Set(ContentType, HTML)
	w.WriteHeader(code)

	stylesPath := fmt.Sprintf("%v/%v", path, "styles.css")
	styles, err := os.Open(stylesPath)

	file := fmt.Sprintf("%v/%v%v", path, code, ".html")
	f, err := os.Open(file)
	if err != nil {
		log.Printf("unexpected error opening file: %v", err)
		scode := strconv.Itoa(code)
		file := fmt.Sprintf("%v/%cxx%v", path, scode[0], ".html")
		f, err := os.Open(file)
		if err != nil {
			log.Printf("unexpected error opening file: %v", err)
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		log.Printf("serving custom error response for code %v and format %v from file %v", code, HTML, file)
		tmpl := template.Must(template.ParseFiles(f.Name(), styles.Name()))

		data := newErrorPageData(r, "")
		tmpl.Execute(w, data)
		return
	}
	defer f.Close()

	log.Printf("serving custom error response for code %v and format %v from file %v", code, HTML, file)
	tmpl := template.Must(template.ParseFiles(f.Name(), styles.Name()))

	data := newErrorPageData(r, "")
	tmpl.Execute(w, data)
}

// JSONResponse returns json reponse
func JSONResponse(w http.ResponseWriter, r *http.Request) {
	code := getStatusCode(r)
	w.Header().Set(ContentType, JSON)
	w.WriteHeader(code)
	body, _ := json.Marshal(newErrorPageData(r, ""))
	w.Write(body)
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if os.Getenv("DEBUG") != "" {
		w.Header().Set(FormatHeader, r.Header.Get(FormatHeader))
		w.Header().Set(CodeHeader, r.Header.Get(CodeHeader))
		w.Header().Set(ContentType, r.Header.Get(ContentType))
		w.Header().Set(OriginalURI, r.Header.Get(OriginalURI))
		w.Header().Set(Namespace, r.Header.Get(Namespace))
		w.Header().Set(IngressName, r.Header.Get(IngressName))
		w.Header().Set(ServiceName, r.Header.Get(ServiceName))
		w.Header().Set(ServicePort, r.Header.Get(ServicePort))
		w.Header().Set(RequestID, r.Header.Get(RequestID))
	}

	format := getFormat(r)

	switch format {
	case JSON:
		JSONResponse(w, r)
	default:
		HTMLResponse(w, r)
	}

	duration := time.Now().Sub(start).Seconds()

	proto := strconv.Itoa(r.ProtoMajor)
	proto = fmt.Sprintf("%s.%s", proto, strconv.Itoa(r.ProtoMinor))

	requestCount.WithLabelValues(proto).Inc()
	requestDuration.WithLabelValues(proto).Observe(duration)
}
