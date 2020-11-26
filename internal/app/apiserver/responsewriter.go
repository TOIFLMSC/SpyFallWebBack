package apiserver

import "net/http"

// responseWriter struct
type responseWriter struct {
	http.ResponseWriter
	code int
}

// WriteHeader func
func (w *responseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
