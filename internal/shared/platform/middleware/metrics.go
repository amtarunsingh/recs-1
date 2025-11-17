package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/metrics"
)

// MetricsMiddleware is a middleware that records HTTP request metrics
type MetricsMiddleware struct {
	metrics *metrics.Metrics
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(m *metrics.Metrics) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: m,
	}
}

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

// Wrap wraps an HTTP handler with metrics collection
func (m *MetricsMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip metrics endpoint to avoid recursive metrics
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			written:        false,
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)
		m.metrics.RecordHTTPRequest(r.Method, r.URL.Path, status, duration)
	})
}
