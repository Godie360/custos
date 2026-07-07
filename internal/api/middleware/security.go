package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// SecurityHeaders sets defensive HTTP response headers on every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-XSS-Protection", "0") // modern browsers: rely on CSP instead
		next.ServeHTTP(w, r)
	})
}

// ipLimiter wraps a per-IP token-bucket rate limiter.
type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func newIPLimiter(r rate.Limit, b int) *ipLimiter {
	return &ipLimiter{limiters: make(map[string]*rate.Limiter), r: r, b: b}
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim, ok := l.limiters[ip]
	if !ok {
		lim = rate.NewLimiter(l.r, l.b)
		l.limiters[ip] = lim
	}
	return lim
}

// globalLimiter is a process-wide shared IP → limiter map: 60 req/s per IP, burst 120.
var globalLimiter = newIPLimiter(60, 120)

// RateLimit rejects requests that exceed 60 req/s per client IP with 429.
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !globalLimiter.get(ip).Allow() {
			http.Error(w, `{"error":"rate_limited","message":"too many requests"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
