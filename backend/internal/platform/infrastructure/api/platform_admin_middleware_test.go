package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlatformAdminMiddleware_ValidAPIKey(t *testing.T) {
	middleware := PlatformAdminMiddleware("secret-api-key")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("POST", "/api/platform/v1/tenants", nil)
	req.Header.Set("X-Platform-Admin-Key", "secret-api-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

func TestPlatformAdminMiddleware_MissingAPIKey(t *testing.T) {
	middleware := PlatformAdminMiddleware("secret-api-key")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/platform/v1/tenants", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPlatformAdminMiddleware_InvalidAPIKey(t *testing.T) {
	middleware := PlatformAdminMiddleware("secret-api-key")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/platform/v1/tenants", nil)
	req.Header.Set("X-Platform-Admin-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPlatformAdminMiddleware_EmptyConfiguredKey(t *testing.T) {
	middleware := PlatformAdminMiddleware("")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/platform/v1/tenants", nil)
	req.Header.Set("X-Platform-Admin-Key", "any-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
