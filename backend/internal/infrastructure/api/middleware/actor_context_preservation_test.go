package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/auth/infrastructure/session"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type mockRoleLookup struct {
	role string
}

func (m *mockRoleLookup) GetRoleByEmail(_ context.Context, _ string) (string, error) {
	return m.role, nil
}

type sessionTestFixture struct {
	scsManager     *scs.SessionManager
	sessionManager *session.SessionManager
}

func newSessionTestFixture() *sessionTestFixture {
	mgr := scs.New()
	mgr.Store = memstore.New()
	mgr.Lifetime = time.Hour
	return &sessionTestFixture{
		scsManager:     mgr,
		sessionManager: session.NewSessionManager(mgr),
	}
}

func (f *sessionTestFixture) storeAuthenticatedSession(t *testing.T, tenantID, email string) []*http.Cookie {
	t.Helper()

	tenantIDVO := sharedvo.MustNewTenantID(tenantID)
	preAuth := session.NewPreAuthSession(tenantIDVO, "acme.com", "http://localhost:3000")
	authSession := preAuth.UpgradeToAuthenticated(
		session.UserInfo{ID: uuid.New(), Email: email},
		session.TokenInfo{
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
			Expiry:       time.Now().Add(8 * time.Hour),
		},
	)

	router := chi.NewRouter()
	router.Use(f.scsManager.LoadAndSave)
	router.Get("/setup", func(w http.ResponseWriter, r *http.Request) {
		err := f.sessionManager.StoreAuthenticatedSession(r.Context(), authSession)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	return rec.Result().Cookies()
}

func TestTenantMiddlewareWithSession_SetsActorOnContext(t *testing.T) {
	fixture := newSessionTestFixture()
	cookies := fixture.storeAuthenticatedSession(t, "acme", "user@acme.com")

	lookup := &mockRoleLookup{role: "architect"}

	var capturedActor sharedctx.Actor
	var actorFound bool

	router := chi.NewRouter()
	router.Use(fixture.scsManager.LoadAndSave)
	router.Use(TenantMiddlewareWithSession(fixture.sessionManager, lookup))
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		capturedActor, actorFound = sharedctx.GetActor(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, actorFound, "Actor should be set by TenantMiddlewareWithSession")
	assert.Equal(t, "user@acme.com", capturedActor.Email)
	assert.Equal(t, sharedctx.RoleArchitect, capturedActor.Role)
}

func TestActorPreservedAfterContextWithTenantOverwrite(t *testing.T) {
	originalActor := sharedctx.NewActor("user-1", "user@acme.com", sharedctx.RoleArchitect)

	setActorMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tenantID := sharedvo.MustNewTenantID("acme")
			ctx = sharedctx.WithTenant(ctx, tenantID)
			ctx = sharedctx.WithActor(ctx, originalActor)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	overwriteTenantAndPreserveActor := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origCtx := r.Context()

			tenantID := sharedvo.MustNewTenantID("acme")
			newCtx := sharedctx.WithTenant(origCtx, tenantID)

			if actor, ok := sharedctx.GetActor(origCtx); ok {
				newCtx = sharedctx.WithActor(newCtx, actor)
			}

			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}

	var capturedActor sharedctx.Actor
	var actorFound bool

	router := chi.NewRouter()
	router.Use(setActorMiddleware)
	router.Use(overwriteTenantAndPreserveActor)
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		capturedActor, actorFound = sharedctx.GetActor(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, actorFound, "Actor must survive context replacement by downstream middleware")
	assert.Equal(t, originalActor.Email, capturedActor.Email)
	assert.Equal(t, originalActor.Role, capturedActor.Role)
	assert.Equal(t, originalActor.ID, capturedActor.ID)
	assert.True(t, capturedActor.HasPermission("components:read"), "Actor permissions must be preserved")
	assert.True(t, capturedActor.HasPermission("components:write"), "Architect write permission must be preserved")
}

func TestActorPreservedWithEditGrants(t *testing.T) {
	originalActor := sharedctx.NewActor("user-1", "stakeholder@acme.com", sharedctx.RoleStakeholder)
	originalActor = originalActor.WithEditGrants(map[string]map[string]bool{
		"component": {"comp-1": true, "comp-2": true},
	})

	setActorMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tenantID := sharedvo.MustNewTenantID("acme")
			ctx = sharedctx.WithTenant(ctx, tenantID)
			ctx = sharedctx.WithActor(ctx, originalActor)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	overwriteTenantAndPreserveActor := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origCtx := r.Context()
			tenantID := sharedvo.MustNewTenantID("acme")
			newCtx := sharedctx.WithTenant(origCtx, tenantID)

			if actor, ok := sharedctx.GetActor(origCtx); ok {
				newCtx = sharedctx.WithActor(newCtx, actor)
			}

			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}

	var capturedActor sharedctx.Actor
	var actorFound bool

	router := chi.NewRouter()
	router.Use(setActorMiddleware)
	router.Use(overwriteTenantAndPreserveActor)
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		capturedActor, actorFound = sharedctx.GetActor(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, actorFound, "Actor must survive context replacement")
	assert.Equal(t, "stakeholder@acme.com", capturedActor.Email)
	assert.True(t, capturedActor.HasEditGrant("components", "comp-1"), "Edit grants must be preserved")
	assert.True(t, capturedActor.HasEditGrant("components", "comp-2"), "Edit grants must be preserved")
}
