package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

func newTestSessionManager() *SessionManager {
	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	return NewSessionManager(scsManager)
}

func TestSessionManager_StoreAndLoadPreAuthSession(t *testing.T) {
	sm := newTestSessionManager()
	tenantID, _ := sharedvo.NewTenantID("acme")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		preAuth := NewPreAuthSession(tenantID, "acme.com", "")
		err := sm.StorePreAuthSession(ctx, preAuth)
		require.NoError(t, err)

		loaded, err := sm.LoadPreAuthSession(ctx)
		require.NoError(t, err)

		assert.Equal(t, preAuth.TenantID(), loaded.TenantID())
		assert.Equal(t, preAuth.State(), loaded.State())
		assert.Equal(t, preAuth.Nonce(), loaded.Nonce())
		assert.Equal(t, preAuth.CodeVerifier(), loaded.CodeVerifier())
		assert.Equal(t, preAuth.ExpectedEmailDomain(), loaded.ExpectedEmailDomain())
		assert.False(t, loaded.IsAuthenticated())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	sm.LoadAndSave(handler).ServeHTTP(rec, req)
}

func TestSessionManager_UpgradeAndLoadAuthenticatedSession(t *testing.T) {
	sm := newTestSessionManager()
	tenantID, _ := sharedvo.NewTenantID("acme")

	var sessionCookie *http.Cookie

	setupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		preAuth := NewPreAuthSession(tenantID, "acme.com", "")
		err := sm.StorePreAuthSession(ctx, preAuth)
		require.NoError(t, err)

		userID := uuid.New()
		authenticated := preAuth.UpgradeToAuthenticated(
			UserInfo{ID: userID, Email: "user@acme.com"},
			TokenInfo{AccessToken: "access", RefreshToken: "refresh", Expiry: time.Now().Add(time.Hour)},
		)
		err = sm.StoreAuthenticatedSession(ctx, authenticated)
		require.NoError(t, err)

		loaded, err := sm.LoadAuthenticatedSession(ctx)
		require.NoError(t, err)
		assert.True(t, loaded.IsAuthenticated())
		assert.Equal(t, userID, loaded.UserID())
		assert.Equal(t, "user@acme.com", loaded.UserEmail())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	sm.LoadAndSave(setupHandler).ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie, "session cookie should be set")
}

func TestSessionManager_LoadPreAuthSession_NotFound(t *testing.T) {
	sm := newTestSessionManager()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := sm.LoadPreAuthSession(r.Context())
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	sm.LoadAndSave(handler).ServeHTTP(rec, req)
}

func TestSessionManager_ClearSession(t *testing.T) {
	sm := newTestSessionManager()
	tenantID, _ := sharedvo.NewTenantID("acme")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		preAuth := NewPreAuthSession(tenantID, "acme.com", "")
		_ = sm.StorePreAuthSession(ctx, preAuth)

		err := sm.ClearSession(ctx)
		require.NoError(t, err)

		_, err = sm.LoadPreAuthSession(ctx)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	sm.LoadAndSave(handler).ServeHTTP(rec, req)
}
