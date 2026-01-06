package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/logging"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a new WideEvent for this request
		event := logging.NewWideEvent("http_request")
		ctx := logging.WithContext(r.Context(), event)

		// Enrich with HTTP metadata
		logging.EnrichHTTP(ctx, r.Method, r.RequestURI)
		logging.EnrichHTTPHeader(ctx, "user_agent", r.UserAgent())
		logging.EnrichHTTPHeader(ctx, "content_type", r.Header.Get("Content-Type"))

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Serve the request with enriched context
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Enrich with response metadata
		logging.EnrichHTTPStatus(ctx, rw.statusCode)
		logging.EnrichHTTPDuration(ctx, time.Since(start))

		// Emit the log
		logging.Emit(ctx)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Enrich the WideEvent with panic information
				logging.EnrichPanic(r.Context())
				logging.EnrichMetadata(r.Context(), "panic_value", err)
				logging.Emit(r.Context())

				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

