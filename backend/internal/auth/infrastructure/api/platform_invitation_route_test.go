package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterPlatformInvitationRoute_RateLimitsRepeatedRequests(t *testing.T) {
	r := chi.NewRouter()
	handlers := NewPlatformInvitationHandlers(cqrs.NewInMemoryCommandBus(), stubTenantCatalog{known: false})
	registerPlatformInvitationRoute(r, "test-key", handlers)

	doRequest := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/invitations", bytes.NewBufferString(`{"tenantId":"acme","email":"admin@acme.com"}`))
		req.RemoteAddr = "203.0.113.7:5555"
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Platform-Admin-Key", "test-key")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for i := 0; i < 100; i++ {
		w := doRequest()
		require.NotEqual(t, http.StatusTooManyRequests, w.Code, "request %d should not be rate limited", i+1)
	}

	w := doRequest()
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
