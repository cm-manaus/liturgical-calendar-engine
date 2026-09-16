package api

import (
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var defaultLatencyBuckets = []float64{0.001, 0.005, 0.010, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0, 2.5, 5.0}

type routeKey struct {
	method string
	path   string
	status int
}

// MetricsCollector tracks real-time operational telemetry for Prometheus scraping.
type MetricsCollector struct {
	inFlight      int64
	requestsTotal uint64
	requests2xx   uint64
	requests4xx   uint64
	requests5xx   uint64
	durationTotal uint64 // in milliseconds

	// SRE Golden Signal: Latency Histogram Buckets (seconds)
	histMu       sync.Mutex
	bucketCounts [11]uint64
	durationSec  float64

	// SRE Golden Signal: Traffic and Error breakdown by Route
	routeMu     sync.RWMutex
	routeCounts map[routeKey]uint64
}

// DefaultMetrics is the singleton metrics recorder for the API instance.
var DefaultMetrics = &MetricsCollector{
	routeCounts: make(map[routeKey]uint64),
}

// IncInFlight increments active concurrent requests.
func (m *MetricsCollector) IncInFlight() {
	atomic.AddInt64(&m.inFlight, 1)
}

// DecInFlight decrements active concurrent requests.
func (m *MetricsCollector) DecInFlight() {
	atomic.AddInt64(&m.inFlight, -1)
}

// Record increments Prometheus counters safely across concurrent goroutines.
func (m *MetricsCollector) Record(method, path string, statusCode int, duration time.Duration) {
	durSec := duration.Seconds()
	durMs := uint64(duration.Milliseconds())

	atomic.AddUint64(&m.requestsTotal, 1)
	atomic.AddUint64(&m.durationTotal, durMs)

	switch {
	case statusCode >= 200 && statusCode < 300:
		atomic.AddUint64(&m.requests2xx, 1)
	case statusCode >= 400 && statusCode < 500:
		atomic.AddUint64(&m.requests4xx, 1)
	case statusCode >= 500:
		atomic.AddUint64(&m.requests5xx, 1)
	}

	// Latency histogram buckets
	m.histMu.Lock()
	m.durationSec += durSec
	for i, b := range defaultLatencyBuckets {
		if durSec <= b {
			m.bucketCounts[i]++
		}
	}
	m.histMu.Unlock()

	// Route traffic & errors
	normPath := path
	if len(normPath) > 64 {
		normPath = normPath[:64]
	}
	key := routeKey{method: method, path: normPath, status: statusCode}
	m.routeMu.Lock()
	if m.routeCounts == nil {
		m.routeCounts = make(map[routeKey]uint64)
	}
	m.routeCounts[key]++
	m.routeMu.Unlock()
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

	fmt.Fprintf(w, "# HELP tesouro_http_requests_in_flight Current number of HTTP requests being served.\n")
	fmt.Fprintf(w, "# TYPE tesouro_http_requests_in_flight gauge\n")
	fmt.Fprintf(w, "tesouro_http_requests_in_flight %d\n\n", atomic.LoadInt64(&DefaultMetrics.inFlight))

	fmt.Fprintf(w, "# HELP tesouro_http_requests_total Total number of HTTP requests processed.\n")
	fmt.Fprintf(w, "# TYPE tesouro_http_requests_total counter\n")
	fmt.Fprintf(w, "tesouro_http_requests_total{status_class=\"2xx\"} %d\n", atomic.LoadUint64(&DefaultMetrics.requests2xx))
	fmt.Fprintf(w, "tesouro_http_requests_total{status_class=\"4xx\"} %d\n", atomic.LoadUint64(&DefaultMetrics.requests4xx))
	fmt.Fprintf(w, "tesouro_http_requests_total{status_class=\"5xx\"} %d\n", atomic.LoadUint64(&DefaultMetrics.requests5xx))
	fmt.Fprintf(w, "tesouro_http_requests_total{status_class=\"all\"} %d\n\n", atomic.LoadUint64(&DefaultMetrics.requestsTotal))

	fmt.Fprintf(w, "# HELP tesouro_http_request_duration_ms_total Cumulative duration of HTTP requests in milliseconds.\n")
	fmt.Fprintf(w, "# TYPE tesouro_http_request_duration_ms_total counter\n")
	fmt.Fprintf(w, "tesouro_http_request_duration_ms_total %d\n\n", atomic.LoadUint64(&DefaultMetrics.durationTotal))

	// Prometheus Standard Latency Histogram (DDIA Cap. 1 & SRE Golden Signal)
	fmt.Fprintf(w, "# HELP tesouro_http_request_duration_seconds HTTP request duration histogram in seconds.\n")
	fmt.Fprintf(w, "# TYPE tesouro_http_request_duration_seconds histogram\n")
	DefaultMetrics.histMu.Lock()
	cumulativeCount := uint64(0)
	for i, b := range defaultLatencyBuckets {
		cumulativeCount = DefaultMetrics.bucketCounts[i]
		fmt.Fprintf(w, "tesouro_http_request_duration_seconds_bucket{le=\"%.3f\"} %d\n", b, cumulativeCount)
	}
	totalCount := atomic.LoadUint64(&DefaultMetrics.requestsTotal)
	totalSec := DefaultMetrics.durationSec
	DefaultMetrics.histMu.Unlock()
	fmt.Fprintf(w, "tesouro_http_request_duration_seconds_bucket{le=\"+Inf\"} %d\n", totalCount)
	fmt.Fprintf(w, "tesouro_http_request_duration_seconds_sum %.6f\n", totalSec)
	fmt.Fprintf(w, "tesouro_http_request_duration_seconds_count %d\n\n", totalCount)

	// Route-level Traffic and Status
	fmt.Fprintf(w, "# HELP tesouro_http_route_requests_total HTTP request counts by method, path and status code.\n")
	fmt.Fprintf(w, "# TYPE tesouro_http_route_requests_total counter\n")
	DefaultMetrics.routeMu.RLock()
	var keys []routeKey
	for k := range DefaultMetrics.routeCounts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].path != keys[j].path {
			return keys[i].path < keys[j].path
		}
		return keys[i].status < keys[j].status
	})
	for _, k := range keys {
		count := DefaultMetrics.routeCounts[k]
		fmt.Fprintf(w, "tesouro_http_route_requests_total{method=\"%s\",path=\"%s\",status=\"%d\"} %d\n", k.method, k.path, k.status, count)
	}
	DefaultMetrics.routeMu.RUnlock()
	fmt.Fprintf(w, "\n")

	// Runtime Saturation Metrics
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
