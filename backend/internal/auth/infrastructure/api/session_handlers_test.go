package api

import (
	"context"
	"encoding/json"
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
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

type mockUserRepository struct {
	user *UserDTO
	err  error
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*UserDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

type mockTenantRepository struct {
	tenant *TenantDTO
	err    error
}

func (m *mockTenantRepository) GetByID(ctx context.Context, tenantID string) (*TenantDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tenant, nil
}

func setupSessionTestHandler(t *testing.T, userRepo UserRepository, tenantRepo TenantRepository) (*SessionHandlers, *scs.SessionManager, *session.SessionManager) {
	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	handlers := NewSessionHandlers(sessionManager, userRepo, tenantRepo)

	return handlers, scsManager, sessionManager
}

func createAuthenticatedSession(t *testing.T, sessionManager *session.SessionManager, scsManager *scs.SessionManager, tenantID string, email string) []*http.Cookie {
	tenantIDVO, _ := sharedvo.NewTenantID(tenantID)
	preAuth := session.NewPreAuthSession(tenantIDVO, "acme.com", "http://localhost:3000")

	userInfo := session.UserInfo{
		ID:    uuid.New(),
		Email: email,
	}
	tokenInfo := session.TokenInfo{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(8 * time.Hour),
	}
	authSession := preAuth.UpgradeToAuthenticated(userInfo, tokenInfo)

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/setup-session", func(w http.ResponseWriter, r *http.Request) {
		err := sessionManager.StoreAuthenticatedSession(r.Context(), authSession)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/setup-session", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	return rec.Result().Cookies()
}

func TestGetCurrentSession_Authenticated(t *testing.T) {
	userID := uuid.New()
	mockUser := &UserDTO{
		ID:     userID,
		Email:  "john@acme.com",
		Name:   "John Doe",
		Role:   "architect",
		Status: "active",
	}
	mockTenant := &TenantDTO{
		ID:   "acme",
		Name: "Acme Corporation",
	}

	handlers, scsManager, sessionManager := setupSessionTestHandler(t,
		&mockUserRepository{user: mockUser},
		&mockTenantRepository{tenant: mockTenant},
	)

	cookies := createAuthenticatedSession(t, sessionManager, scsManager, "acme", "john@acme.com")

	req := httptest.NewRequest(http.MethodGet, "/auth/sessions/current", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/auth/sessions/current", handlers.GetCurrentSession)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response CurrentSessionResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), response.User.ID)
	assert.Equal(t, "john@acme.com", response.User.Email)
	assert.Equal(t, "John Doe", response.User.Name)
	assert.Equal(t, "architect", response.User.Role)
	assert.Contains(t, response.User.Permissions, "components:read")
	assert.Contains(t, response.User.Permissions, "components:write")

	assert.Equal(t, "acme", response.Tenant.ID)
	assert.Equal(t, "Acme Corporation", response.Tenant.Name)

	assert.Contains(t, response.Links, "self")
	assert.Contains(t, response.Links, "logout")
}

func TestGetCurrentSession_NoSession(t *testing.T) {
	handlers, scsManager, _ := setupSessionTestHandler(t,
		&mockUserRepository{},
		&mockTenantRepository{},
	)

	req := httptest.NewRequest(http.MethodGet, "/auth/sessions/current", nil)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/auth/sessions/current", handlers.GetCurrentSession)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetCurrentSession_AdminPermissions(t *testing.T) {
	userID := uuid.New()
	mockUser := &UserDTO{
		ID:     userID,
		Email:  "admin@acme.com",
		Name:   "Admin User",
		Role:   "admin",
		Status: "active",
	}
	mockTenant := &TenantDTO{
		ID:   "acme",
		Name: "Acme Corporation",
	}

	handlers, scsManager, sessionManager := setupSessionTestHandler(t,
		&mockUserRepository{user: mockUser},
		&mockTenantRepository{tenant: mockTenant},
	)

	cookies := createAuthenticatedSession(t, sessionManager, scsManager, "acme", "admin@acme.com")

	req := httptest.NewRequest(http.MethodGet, "/auth/sessions/current", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/auth/sessions/current", handlers.GetCurrentSession)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response CurrentSessionResponse
	json.NewDecoder(rec.Body).Decode(&response)

	assert.Contains(t, response.User.Permissions, "users:manage")
	assert.Contains(t, response.User.Permissions, "components:delete")
	assert.Contains(t, response.User.Permissions, "invitations:manage")
}

func TestDeleteCurrentSession_Success(t *testing.T) {
	handlers, scsManager, sessionManager := setupSessionTestHandler(t,
		&mockUserRepository{},
		&mockTenantRepository{},
	)

	cookies := createAuthenticatedSession(t, sessionManager, scsManager, "acme", "john@acme.com")

	req := httptest.NewRequest(http.MethodDelete, "/auth/sessions/current", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Delete("/auth/sessions/current", handlers.DeleteCurrentSession)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteCurrentSession_NoSession(t *testing.T) {
	handlers, scsManager, _ := setupSessionTestHandler(t,
		&mockUserRepository{},
		&mockTenantRepository{},
	)

	req := httptest.NewRequest(http.MethodDelete, "/auth/sessions/current", nil)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Delete("/auth/sessions/current", handlers.DeleteCurrentSession)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
