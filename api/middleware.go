package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += int64(n)
	return n, err
}

// LoggingMiddleware logs all inbound HTTP requests using slog with execution duration and response metadata.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DefaultMetrics.IncInFlight()
		defer DefaultMetrics.DecInFlight()

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		DefaultMetrics.Record(r.Method, r.URL.Path, rec.statusCode, duration)

		// Suppress spammy successful health checks from logs
		if (r.URL.Path == "/healthz" || r.URL.Path == "/api/v1/healthz" || r.URL.Path == "/metrics") && rec.statusCode == http.StatusOK {
			return
		}

		slog.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"bytes", rec.bytesWritten,
		)
	})
}

// RecoveryMiddleware traps unhandled panics and prevents the server process from crashing.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic_recovered",
					"panic", fmt.Sprintf("%v", rec),
					"path", r.URL.Path,
					"method", r.Method,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "internal_server_error",
					"message": "An unexpected server error occurred",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CorsMiddleware injects standard CORS headers for cross-origin browser requests.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept-Language")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
