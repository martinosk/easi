package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	authPL "easi/backend/internal/auth/publishedlanguage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type permissionSetMiddleware struct {
	held     map[string]bool
	required []string
}

func (m *permissionSetMiddleware) RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler {
	m.required = append(m.required, permission.String())
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !m.held[permission.String()] {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

var importRoutes = []struct {
	method string
	path   string
}{
	{http.MethodPost, "/imports"},
	{http.MethodGet, "/imports/session-1"},
	{http.MethodPost, "/imports/session-1/confirm"},
	{http.MethodDelete, "/imports/session-1"},
}

func serveImportRoute(auth *permissionSetMiddleware, method, path string) int {
	router := chi.NewRouter()
	registerRoutes(router, NewImportHandlers(nil, nil), auth)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec.Code
}

func TestImportRoutes_RequireEveryWritePermissionAnImportExercises(t *testing.T) {
	auth := &permissionSetMiddleware{}
	registerRoutes(chi.NewRouter(), NewImportHandlers(nil, nil), auth)

	assert.ElementsMatch(t, []string{"components:write", "capabilities:write", "valuestreams:write"}, auth.required)
}

func TestImportRoutes_ForbiddenForReadOnlyCaller(t *testing.T) {
	for _, route := range importRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			auth := &permissionSetMiddleware{held: map[string]bool{
				"components:read":   true,
				"capabilities:read": true,
				"valuestreams:read": true,
			}}

			assert.Equal(t, http.StatusForbidden, serveImportRoute(auth, route.method, route.path))
		})
	}
}

func TestImportRoutes_ForbiddenWhenOneWritePermissionIsMissing(t *testing.T) {
	for _, missing := range []string{"components:write", "capabilities:write", "valuestreams:write"} {
		t.Run(missing, func(t *testing.T) {
			held := map[string]bool{"components:write": true, "capabilities:write": true, "valuestreams:write": true}
			delete(held, missing)
			auth := &permissionSetMiddleware{held: held}

			assert.Equal(t, http.StatusForbidden, serveImportRoute(auth, http.MethodPost, "/imports/session-1/confirm"))
		})
	}
}
