package middleware

import (
	"net/http"

	"github.com/nuric/go-api-template/templates"
)

/* What's happening here is that we want to check if the response header is 404
 * so we can render a custom 404 page. The interceptor captures the status code
 * and allows us to handle it later. */

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rwi *responseWriterInterceptor) WriteHeader(code int) {
	rwi.statusCode = code
	if code != http.StatusNotFound {
		rwi.ResponseWriter.WriteHeader(code)
		return
	}
	// Fix headers
	rwi.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Clear Content-Length to avoid conflicts
	rwi.Header().Del("Content-Length")
	rwi.ResponseWriter.WriteHeader(code)
	// Write response
	templates.RenderHTML(rwi.ResponseWriter, "404.html", nil)
}

func (rwi *responseWriterInterceptor) Write(b []byte) (int, error) {
	if rwi.statusCode == http.StatusNotFound {
		return 0, nil
	}
	return rwi.ResponseWriter.Write(b)
}

func NotFoundRenderer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response writer interceptor
		interceptor := &responseWriterInterceptor{w, http.StatusOK}
		// Call the next handler
		next.ServeHTTP(interceptor, r)
	})
}
