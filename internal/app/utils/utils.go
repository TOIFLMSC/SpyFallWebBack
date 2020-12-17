package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
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

// TokenGenerator func
func TokenGenerator() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// Contains func
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
