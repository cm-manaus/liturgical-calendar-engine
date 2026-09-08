package api

import "net/http"

// NewRouter constructs and configures the HTTP multiplexer with all routes and middlewares.
func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	// System / Diagnostic routes
	mux.HandleFunc("GET /", h.HandleRoot)
	mux.HandleFunc("GET /healthz", h.HandleHealthz)
	mux.HandleFunc("GET /api/v1/healthz", h.HandleHealthz)

	// Liturgical Day & Month endpoints
	mux.HandleFunc("GET /liturgical-day", h.HandleGetLiturgicalDay)
	mux.HandleFunc("POST /liturgical-day", h.HandlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-day", h.HandleGetLiturgicalDay)
	mux.HandleFunc("POST /api/v1/liturgical-day", h.HandlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-month", h.HandleGetLiturgicalMonth)

	// Middleware execution chain: CORS -> Panic Recovery -> Request Logging -> Router
	return CorsMiddleware(RecoveryMiddleware(LoggingMiddleware(mux)))
}
