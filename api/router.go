package api

import "net/http"

// NewRouter constructs and configures the HTTP multiplexer with all routes and middlewares.
func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	// System / Diagnostic & Observability routes
	mux.HandleFunc("GET /", h.HandleRoot)
	mux.HandleFunc("GET /healthz", h.HandleHealthz)
	mux.HandleFunc("GET /api/v1/healthz", h.HandleHealthz)
	mux.HandleFunc("GET /metrics", h.HandleMetrics)

	// Liturgical Day & Month endpoints
	mux.HandleFunc("GET /liturgical-day", h.HandleGetLiturgicalDay)
	mux.HandleFunc("POST /liturgical-day", h.HandlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-day", h.HandleGetLiturgicalDay)
	mux.HandleFunc("POST /api/v1/liturgical-day", h.HandlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-month", h.HandleGetLiturgicalMonth)
	mux.HandleFunc("GET /api/v1/calendar/export", h.HandleExportCalendar)
	mux.HandleFunc("GET /calendar/export", h.HandleExportCalendar)

	// Interactive Documentation (Scalar / OpenAPI)
	mux.HandleFunc("GET /docs", h.HandleDocs)
	mux.HandleFunc("GET /openapi.json", h.HandleOpenAPISpec)
	mux.HandleFunc("GET /swagger", h.HandleSwaggerRedirect)
	mux.HandleFunc("GET /swagger/", h.HandleSwaggerRedirect)

	// Middleware execution chain: CORS -> Panic Recovery -> Request Logging -> Router
	return CorsMiddleware(RecoveryMiddleware(LoggingMiddleware(mux)))
}
