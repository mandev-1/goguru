package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error encoding json: %v", err)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]any{
		"success": false,
		"message": message,
	})
}

func WriteSuccess(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": message,
	})
}

func WriteSuccessData(w http.ResponseWriter, data map[string]any) {
	data["success"] = true
	WriteJSON(w, http.StatusOK, data)
}
