package apix

import (
	"encoding/json"
	"net/http"
)

type responseBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func WriteData(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, responseBody{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func WriteError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, responseBody{
		Code:    status,
		Message: message,
	})
}

func writeJSON(w http.ResponseWriter, status int, data responseBody) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
