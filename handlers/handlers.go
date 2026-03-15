package handlers

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
}

func FapHandler(w http.ResponseWriter, r *http.Request) {
	res := Response{
		Message: "Fap fap fap",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	res := Response{
		Message: "Hello from Go server",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
