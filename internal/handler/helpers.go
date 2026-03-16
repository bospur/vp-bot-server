package handler

import (
	"encoding/json"
	"net/http"
)

// writeJSON отправляет JSON ответ с нужным статусом
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
