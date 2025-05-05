package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// MetricsResponseWriter wraps an http.ResponseWriter to capture metrics
type MetricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
	length     int
}

// NewMetricsResponseWriter creates a new MetricsResponseWriter
func NewMetricsResponseWriter(w http.ResponseWriter) *MetricsResponseWriter {
	return &MetricsResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default to 200 OK
	}
}

// WriteHeader captures the status code
func (mrw *MetricsResponseWriter) WriteHeader(statusCode int) {
	mrw.statusCode = statusCode
	mrw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response length
func (mrw *MetricsResponseWriter) Write(b []byte) (int, error) {
	n, err := mrw.ResponseWriter.Write(b)
	mrw.length += n
	return n, err
}

// StatusCode returns the captured status code
func (mrw *MetricsResponseWriter) StatusCode() int {
	return mrw.statusCode
}

// Length returns the captured response length
func (mrw *MetricsResponseWriter) Length() int {
	return mrw.length
}

// InstrumentHandler wraps an http handler with metrics instrumentation
func InstrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a wrapped response writer to capture metrics
		mrw := NewMetricsResponseWriter(w)

		// Call the handler with our wrapped response writer
		next.ServeHTTP(mrw, r)

		// Record request duration
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(mrw.StatusCode())

		// Skip tracking metrics for the metrics endpoint itself to avoid recursion
		if r.URL.Path != "/metrics" {
			RequestDuration.WithLabelValues(r.URL.Path, r.Method, statusCode).Observe(duration)
		}
	})
}
