package httpapi

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

type readinessResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}
