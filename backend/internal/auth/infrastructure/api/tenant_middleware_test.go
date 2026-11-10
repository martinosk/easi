package api

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

	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/shared/agenttoken"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type mockRoleLookup struct {
	role     string
	roleByID map[string]string
	err      error
}

func (m *mockRoleLookup) GetActiveRoleByEmail(_ context.Context, _ string) (string, error) {
	return m.role, m.err
}

func (m *mockRoleLookup) GetActiveRoleByUserID(_ context.Context, userID string) (string, error) {
	if m.roleByID != nil {
		return m.roleByID[userID], nil
	}
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

func runActorPreservationPipeline(t *testing.T, originalActor sharedctx.Actor) (sharedctx.Actor, bool) {
	t.Helper()

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
	return capturedActor, actorFound
}

func TestBuildBypassContext_SetsAdminActor(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/views", nil)

	ctx, err := buildBypassContext(req)

	require.NoError(t, err)
	actor, ok := sharedctx.GetActor(ctx)
	require.True(t, ok, "bypass context must carry an actor so command handlers can resolve the acting user")
	assert.Equal(t, sharedctx.RoleAdmin, actor.Role)
	assert.NotEmpty(t, actor.ID)
	assert.NotEmpty(t, actor.Email)
}

func TestActorPreservedAfterContextWithTenantOverwrite(t *testing.T) {
	originalActor := sharedctx.NewActor("user-1", "user@acme.com", sharedctx.RoleArchitect)

	capturedActor, actorFound := runActorPreservationPipeline(t, originalActor)

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

	capturedActor, actorFound := runActorPreservationPipeline(t, originalActor)

	assert.True(t, actorFound, "Actor must survive context replacement")
	assert.Equal(t, "stakeholder@acme.com", capturedActor.Email)
	assert.True(t, capturedActor.HasEditGrant("components", "comp-1"), "Edit grants must be preserved")
	assert.True(t, capturedActor.HasEditGrant("components", "comp-2"), "Edit grants must be preserved")
}

func TestTenantMiddlewareWithAgentToken_SetsRoleFromUserID(t *testing.T) {
	t.Setenv("AGENT_TOKEN_SECRET", "test-secret-key")
	token, err := agenttoken.Mint("9d6f7ca6-5f8f-4cca-96f5-f52e1f2e8e4c", "acme", time.Minute)
	require.NoError(t, err)

	lookup := &mockRoleLookup{
		roleByID: map[string]string{"9d6f7ca6-5f8f-4cca-96f5-f52e1f2e8e4c": "architect"},
	}

	var capturedActor sharedctx.Actor
	var actorFound bool

	router := chi.NewRouter()
	router.Use(TenantMiddlewareWithSession(nil, lookup))
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		capturedActor, actorFound = sharedctx.GetActor(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "AgentToken "+token)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, actorFound)
	assert.Equal(t, sharedctx.RoleArchitect, capturedActor.Role)
	assert.True(t, capturedActor.ViaAgent)
	assert.True(t, capturedActor.HasPermission("assistant:use"))
}

func TestTenantMiddlewareWithAgentToken_RejectsWhenRoleNotFound(t *testing.T) {
	t.Setenv("AGENT_TOKEN_SECRET", "test-secret-key")
	token, err := agenttoken.Mint("9d6f7ca6-5f8f-4cca-96f5-f52e1f2e8e4c", "acme", time.Minute)
	require.NoError(t, err)

	lookup := &mockRoleLookup{roleByID: map[string]string{}}

	router := chi.NewRouter()
	router.Use(TenantMiddlewareWithSession(nil, lookup))
	router.Get("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "AgentToken "+token)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func serveSessionRequest(fixture *sessionTestFixture, cookies []*http.Cookie, lookup UserRoleLookup) (int, bool) {
	reached := false
	router := chi.NewRouter()
	router.Use(fixture.scsManager.LoadAndSave)
	router.Use(TenantMiddlewareWithSession(fixture.sessionManager, lookup))
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code, reached
}

func serveWithSessionLookup(t *testing.T, lookup UserRoleLookup) (int, bool) {
	t.Helper()
	fixture := newSessionTestFixture()
	cookies := fixture.storeAuthenticatedSession(t, "acme", "user@acme.com")
	return serveSessionRequest(fixture, cookies, lookup)
}

func TestTenantMiddlewareWithSession_RejectsSessionOfUserWithoutActiveRole(t *testing.T) {
	status, reached := serveWithSessionLookup(t, &mockRoleLookup{role: ""})

	assert.Equal(t, http.StatusUnauthorized, status)
	assert.False(t, reached, "a disabled or unknown user's session must not reach the handler")
}

func TestTenantMiddlewareWithSession_DestroysSessionOfInactiveUser(t *testing.T) {
	fixture := newSessionTestFixture()
	cookies := fixture.storeAuthenticatedSession(t, "acme", "user@acme.com")

	disabledStatus, _ := serveSessionRequest(fixture, cookies, &mockRoleLookup{role: ""})
	reEnabledStatus, reached := serveSessionRequest(fixture, cookies, &mockRoleLookup{role: "architect"})

	assert.Equal(t, http.StatusUnauthorized, disabledStatus)
	assert.Equal(t, http.StatusUnauthorized, reEnabledStatus, "re-enabling the user must not revive the old session")
	assert.False(t, reached)
}

func TestTenantMiddlewareWithSession_FailsClosedWhenUserLookupFails(t *testing.T) {
	status, reached := serveWithSessionLookup(t, &mockRoleLookup{role: "admin", err: assert.AnError})

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.False(t, reached)
}

func TestTenantMiddlewareWithSession_FailsClosedWithoutUserLookup(t *testing.T) {
	status, reached := serveWithSessionLookup(t, nil)

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.False(t, reached)
}

type stubUserByEmail struct {
	user *readmodels.UserDTO
}

func (s stubUserByEmail) GetByEmail(_ context.Context, _ string) (*readmodels.UserDTO, error) {
	return s.user, nil
}

func serveWithPermissionGate(t *testing.T, user *readmodels.UserDTO) int {
	t.Helper()
	fixture := newSessionTestFixture()
	cookies := fixture.storeAuthenticatedSession(t, "acme", "user@acme.com")
	gate := NewAuthMiddleware(fixture.sessionManager).WithUserReadModel(stubUserByEmail{user: user})

	router := chi.NewRouter()
	router.Use(fixture.scsManager.LoadAndSave)
	router.With(gate.RequirePermission(valueobjects.PermComponentsRead)).Get("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code
}

func TestRequirePermission_AdmitsActiveUserHoldingThePermission(t *testing.T) {
	status := serveWithPermissionGate(t, &readmodels.UserDTO{Role: "architect", Status: "active"})

	assert.Equal(t, http.StatusOK, status)
}

func TestRequirePermission_RejectsDisabledUserHoldingThePermission(t *testing.T) {
	status := serveWithPermissionGate(t, &readmodels.UserDTO{Role: "architect", Status: "disabled"})

	assert.Equal(t, http.StatusUnauthorized, status)
}
