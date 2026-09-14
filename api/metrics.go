package api

import (
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// MetricsCollector tracks real-time operational telemetry for Prometheus scraping.
type MetricsCollector struct {
	requestsTotal uint64
	requests2xx   uint64
	requests4xx   uint64
	requests5xx   uint64
	durationTotal uint64 // in milliseconds
}

// DefaultMetrics is the singleton metrics recorder for the API instance.
var DefaultMetrics = &MetricsCollector{}

// Record increments Prometheus counters safely across concurrent goroutines.
func (m *MetricsCollector) Record(statusCode int, duration time.Duration) {
	atomic.AddUint64(&m.requestsTotal, 1)
	atomic.AddUint64(&m.durationTotal, uint64(duration.Milliseconds()))
	switch {
	case statusCode >= 200 && statusCode < 300:
		atomic.AddUint64(&m.requests2xx, 1)
	case statusCode >= 400 && statusCode < 500:
		atomic.AddUint64(&m.requests4xx, 1)
	case statusCode >= 500:
		atomic.AddUint64(&m.requests5xx, 1)
	}
}

// HandleMetrics serves Prometheus-compatible telemetry in standard plaintext format.
func (h *Handler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	uptimeSeconds := int64(time.Since(h.startTime).Seconds())

	fmt.Fprintf(w, "# HELP tesouro_uptime_seconds Total time the backend has been running.\n")
	fmt.Fprintf(w, "# TYPE tesouro_uptime_seconds gauge\n")
	fmt.Fprintf(w, "tesouro_uptime_seconds %d\n\n", uptimeSeconds)

	fmt.Fprintf(w, "# HELP tesouro_http_requests_total Total number of HTTP requests processed.\n")
	fmt.Fprintf(w, "# TYPE tesouro_http_requests_total counter\n")
	fmt.Fprintf(w, "tesouro_http_requests_total{status_class=\"2xx\"} %d\n", atomic.LoadUint64(&DefaultMetrics.requests2xx))
	fmt.Fprintf(w, "tesouro_http_requests_total{status_class=\"4xx\"} %d\n", atomic.LoadUint64(&DefaultMetrics.requests4xx))
	fmt.Fprintf(w, "tesouro_http_requests_total{status_class=\"5xx\"} %d\n", atomic.LoadUint64(&DefaultMetrics.requests5xx))
	fmt.Fprintf(w, "tesouro_http_requests_total{status_class=\"all\"} %d\n\n", atomic.LoadUint64(&DefaultMetrics.requestsTotal))

	fmt.Fprintf(w, "# HELP tesouro_http_request_duration_ms_total Cumulative duration of HTTP requests in milliseconds.\n")
	fmt.Fprintf(w, "# TYPE tesouro_http_request_duration_ms_total counter\n")
	fmt.Fprintf(w, "tesouro_http_request_duration_ms_total %d\n\n", atomic.LoadUint64(&DefaultMetrics.durationTotal))

	fmt.Fprintf(w, "# HELP go_goroutines Number of goroutines that currently exist.\n")
	fmt.Fprintf(w, "# TYPE go_goroutines gauge\n")
	fmt.Fprintf(w, "go_goroutines %d\n\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.\n")
	fmt.Fprintf(w, "# TYPE go_memstats_alloc_bytes gauge\n")
	fmt.Fprintf(w, "go_memstats_alloc_bytes %d\n\n", mem.Alloc)

	fmt.Fprintf(w, "# HELP go_memstats_sys_bytes Number of bytes obtained from system.\n")
	fmt.Fprintf(w, "# TYPE go_memstats_sys_bytes gauge\n")
	fmt.Fprintf(w, "go_memstats_sys_bytes %d\n\n", mem.Sys)

	fmt.Fprintf(w, "# HELP go_memstats_num_gc Number of completed GC cycles.\n")
	fmt.Fprintf(w, "# TYPE go_memstats_num_gc counter\n")
	fmt.Fprintf(w, "go_memstats_num_gc %d\n", mem.NumGC)
}
