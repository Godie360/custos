package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter captures the status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Logger logs each HTTP request as a structured slog entry.
// 5xx responses are logged at ERROR level; everything else at INFO.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		ms := float64(time.Since(start).Microseconds()) / 1000.0
		attrs := []any{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.status),
			slog.Float64("duration_ms", ms),
			slog.Int("bytes", rw.bytesWritten),
			slog.String("ip", realIP(r)),
			slog.String("request_id", GetRequestID(r.Context())),
		}

		if rw.status >= 500 {
			slog.ErrorContext(r.Context(), "http request", attrs...)
		} else {
			slog.InfoContext(r.Context(), "http request", attrs...)
		}
	})
}

// realIP returns the best-effort client IP, honouring X-Forwarded-For when present.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
