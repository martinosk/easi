//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"easi/backend/internal/auth/infrastructure/session"
	sharedcontext "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func withTestTenant(req *http.Request) *http.Request {
	ctx := sharedcontext.WithTenant(req.Context(), sharedvo.DefaultTenantID())
	return req.WithContext(ctx)
}

func testTenantID() string {
	return "default"
}

func tenantContext() context.Context {
	return sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
}

func makeRequest(method, url string, body []byte, urlParams map[string]string) (*httptest.ResponseRecorder, *http.Request) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req = withTestTenant(req)

	if len(urlParams) > 0 {
		rctx := chi.NewRouteContext()
		for key, value := range urlParams {
			rctx.URLParams.Add(key, value)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	return httptest.NewRecorder(), req
}

type testSessionManager struct {
	scsManager     *scs.SessionManager
	sessionManager *session.SessionManager
}

func newTestSessionManager() *testSessionManager {
	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	return &testSessionManager{
		scsManager:     scsManager,
		sessionManager: sessionManager,
	}
}

func createTestAuthSession(tenantID, email string) session.AuthSession {
	userID := uuid.New().String()
	sessionData := fmt.Sprintf(`{
		"tenantId": "%s",
		"userId": "%s",
		"userEmail": "%s",
		"accessToken": "test-token",
		"refreshToken": "test-refresh",
		"tokenExpiry": "%s",
		"authenticated": true
	}`, tenantID, userID, email, time.Now().Add(1*time.Hour).Format(time.RFC3339))

	authSession, _ := session.UnmarshalAuthSession([]byte(sessionData))
	return authSession
}

func (tsm *testSessionManager) setupSession(email string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authSession := createTestAuthSession(testTenantID(), email)
		tsm.sessionManager.StoreAuthenticatedSession(r.Context(), authSession)
		w.WriteHeader(http.StatusOK)
	})
}

func (tsm *testSessionManager) getSessionCookies(t interface {
	Helper()
	Fatalf(string, ...interface{})
}, email string) []*http.Cookie {
	t.Helper()
	router := chi.NewRouter()
	router.Use(tsm.scsManager.LoadAndSave)
	router.Post("/setup", tsm.setupSession(email).ServeHTTP)

	req := httptest.NewRequest(http.MethodPost, "/setup", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Failed to setup session: got status %d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("Expected session cookie but got none")
	}
	return cookies
}
