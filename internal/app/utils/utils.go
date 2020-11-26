package utils

import (
	"encoding/json"
	"net/http"
)

// Message func
func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

// Respond func
func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Error func
func Error(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	Respond(w, map[string]interface{}{"error": err.Error()})
}
