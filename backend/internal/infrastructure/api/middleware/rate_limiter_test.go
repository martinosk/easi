package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_AllowsRequestsUnderLimit(t *testing.T) {
	limiter := NewRateLimiter(10, 60)
	middleware := RateLimitMiddleware(limiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/platform/v1/tenants", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "request %d should be allowed", i+1)
	}
}

func TestRateLimiter_BlocksRequestsOverLimit(t *testing.T) {
	limiter := NewRateLimiter(3, 60)
	middleware := RateLimitMiddleware(limiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/platform/v1/tenants", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	req := httptest.NewRequest("GET", "/api/platform/v1/tenants", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiter_TracksIPsSeparately(t *testing.T) {
	limiter := NewRateLimiter(2, 60)
	middleware := RateLimitMiddleware(limiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/platform/v1/tenants", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	req := httptest.NewRequest("GET", "/api/platform/v1/tenants", nil)
	req.RemoteAddr = "192.168.1.2:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "different IP should have its own limit")
}

func TestRateLimiter_UsesXForwardedFor(t *testing.T) {
	limiter := NewRateLimiter(2, 60)
	middleware := RateLimitMiddleware(limiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/platform/v1/tenants", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	req := httptest.NewRequest("GET", "/api/platform/v1/tenants", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "should rate limit based on X-Forwarded-For")
}

func TestRateLimiter_IncludesRetryAfterHeader(t *testing.T) {
	limiter := NewRateLimiter(1, 60)
	middleware := RateLimitMiddleware(limiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/platform/v1/tenants", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	req = httptest.NewRequest("GET", "/api/platform/v1/tenants", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NotEmpty(t, w.Header().Get("Retry-After"), "should include Retry-After header")
}
