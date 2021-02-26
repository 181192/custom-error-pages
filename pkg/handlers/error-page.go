package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/181192/custom-error-pages/pkg/metrics"
	"github.com/181192/custom-error-pages/pkg/util"
	"github.com/oxtoacart/bpool"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

var bufpool *bpool.BufferPool

func init() {
	bufpool = bpool.NewBufferPool(64)
}

type errorPageData struct {
	Code     string                `json:"code"`
	Title    string                `json:"title"`
	Messages []string              `json:"messages"`
	Details  *errorPageDataDetails `json:"details,omitempty"`
}

type errorPageDataDetails struct {
	OriginalURI string `json:"originalURI"`
	Namespace   string `json:"namespace"`
	IngressName string `json:"ingressName"`
	ServiceName string `json:"serviceName"`
	ServicePort string `json:"servicePort"`
	RequestID   string `json:"requestId"`
}

func newErrorPageData(req *http.Request, code int, messages []string) errorPageData {
	title := http.StatusText(code)

	data := errorPageData{
		Code:     strconv.Itoa(code),
		Title:    title,
		Messages: messages,
	}

	hideDetails := viper.GetBool(util.HideDetails)
	if !hideDetails {
		data.Details = &errorPageDataDetails{
			OriginalURI: req.Header.Get(OriginalURI),
			Namespace:   req.Header.Get(Namespace),
			IngressName: req.Header.Get(IngressName),
			ServiceName: req.Header.Get(ServiceName),
			ServicePort: req.Header.Get(ServicePort),
			RequestID:   req.Header.Get(RequestID),
		}
	}

	return data
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

func getMessages(code int) []string {
	switch code {
	case http.StatusNotFound:
		return []string{"The page you're looking for could not be found."}
	case http.StatusServiceUnavailable:
		return []string{"Ooops, this shouldn't have happened.", "The server is temporary busy, try again later!"}
	default:
		return []string{http.StatusText(code)}
	}
}

// htmlResponse returns html reponse
func htmlResponse(w http.ResponseWriter, r *http.Request) {
	code := getStatusCode(r)
	messages := getMessages(code)

	buf := bufpool.Get()
	defer bufpool.Put(buf)

	templatesDir := viper.GetString(util.ErrFilesPath) + "/*"
	templates, err := template.ParseGlob(templatesDir)
	if err != nil {
		log.Error().Msgf("Failed to parse template %s", err)
		w.Header().Set(ContentType, JSON)
		w.WriteHeader(http.StatusInternalServerError)
		body, _ := json.Marshal(newErrorPageData(r, http.StatusInternalServerError,
			[]string{"Ups, this should not have happened", "Failed to parse templates"}))
		w.Write(body)
		return
	}

	data := newErrorPageData(r, code, messages)
	err = templates.ExecuteTemplate(buf, "index", data)
	if err != nil {
		log.Error().Msgf("Failed to execute template %s", err)
		w.Header().Set(ContentType, JSON)
		w.WriteHeader(http.StatusInternalServerError)
		body, _ := json.Marshal(newErrorPageData(r, http.StatusInternalServerError,
			[]string{"Ups, this should not have happened", "Failed to execute templates"}))
		w.Write(body)
		return
	}

	w.Header().Set(ContentType, HTML)
	w.WriteHeader(code)
	buf.WriteTo(w)
}

// jsonResponse returns json reponse
func jsonResponse(w http.ResponseWriter, r *http.Request) {
	code := getStatusCode(r)
	message := getMessages(code)

	body, err := json.Marshal(newErrorPageData(r, code, message))
	if err != nil {
		log.Error().Msgf("Failed to marshal json response %s", err)
		w.Header().Set(ContentType, JSON)
		w.WriteHeader(http.StatusInternalServerError)
		body, _ := json.Marshal(newErrorPageData(r, http.StatusInternalServerError,
			[]string{"Ups, this should not have happened", "Failed to marshal json response"}))
		w.Write(body)
		return
	}

	w.Header().Set(ContentType, JSON)
	w.WriteHeader(code)
	w.Write(body)
}

// ErrorPage error handler
func ErrorPage() http.Handler {
	return metrics.Measure(errorPage())
}

func errorPage() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		format := getFormat(r)

		switch format {
		case JSON:
			jsonResponse(w, r)
		default:
			htmlResponse(w, r)
		}
	}
	return http.HandlerFunc(fn)
}
