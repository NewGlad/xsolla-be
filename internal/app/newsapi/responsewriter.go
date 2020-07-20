package newsapi

import "net/http"

// ResponseWriter ...
type ResponseWriter struct {
	http.ResponseWriter
	code int
}

// WriteHeader ...
func (r *ResponseWriter) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}
