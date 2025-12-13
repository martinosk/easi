package api

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	requests    map[string][]time.Time
	mu          sync.Mutex
	maxRequests int
	windowSecs  int
}

func NewRateLimiter(maxRequests, windowSecs int) *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		windowSecs:  windowSecs,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-time.Duration(rl.windowSecs) * time.Second)

	requests := rl.requests[ip]
	var valid []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.maxRequests {
		rl.requests[ip] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[ip] = valid
	return true
}

func (rl *RateLimiter) WindowSeconds() int {
	return rl.windowSecs
}

func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !limiter.Allow(ip) {
				w.Header().Set("Retry-After", strconv.Itoa(limiter.WindowSeconds()))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
