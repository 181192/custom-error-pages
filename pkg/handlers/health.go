package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthResponse health response
type HealthResponse struct {
	Message string `json:"message"`
}

// Health handler
func Health() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(&HealthResponse{
			Message: "OK",
		})

		w.WriteHeader(http.StatusOK)
		w.Header().Set(ContentType, JSON)
		w.Write(data)
	}
	return http.HandlerFunc(fn)
}
