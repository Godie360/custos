package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	custosmetrics "github.com/iPFSoftwares/custos/internal/api/metrics"
)

// Metrics records Prometheus HTTP metrics: request count and latency histogram.
// It must be mounted AFTER chi's routing middleware so RouteContext is populated.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r)

		// Use the matched route pattern to keep label cardinality bounded.
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unknown"
		}

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(ww.status)

		custosmetrics.HTTPRequestsTotal.WithLabelValues(r.Method, route, status).Inc()
		custosmetrics.HTTPDurationSeconds.WithLabelValues(r.Method, route).Observe(duration)
	})
}

// statusWriter wraps http.ResponseWriter to capture the response status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}
