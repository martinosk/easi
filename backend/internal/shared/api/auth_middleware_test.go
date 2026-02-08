package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sharedctx "easi/backend/internal/shared/context"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(artifactType, idParam string) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/test/{id}", func(r chi.Router) {
		r.Use(RequireWriteOrEditGrant(artifactType, idParam))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})
	return r
}

func requestWithActor(actor sharedctx.Actor) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/test/artifact-123", nil)
	ctx := sharedctx.WithActor(req.Context(), actor)
	return req.WithContext(ctx)
}

func TestRequireWriteOrEditGrant_NativeWritePermission_PassesThrough(t *testing.T) {
	r := setupTestRouter("capabilities", "id")
	actor := sharedctx.NewActor("user-1", "user@test.com", "architect")

	req := requestWithActor(actor)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireWriteOrEditGrant_NoWriteButHasEditGrant_PassesThrough(t *testing.T) {
	r := setupTestRouter("capabilities", "id")
	actor := sharedctx.NewActor("user-1", "user@test.com", "stakeholder")
	actor = actor.WithEditGrants(map[string]map[string]bool{
		"capability": {"artifact-123": true},
	})

	req := requestWithActor(actor)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireWriteOrEditGrant_NoWriteNoEditGrant_Returns403(t *testing.T) {
	r := setupTestRouter("capabilities", "id")
	actor := sharedctx.NewActor("user-1", "user@test.com", "stakeholder")

	req := requestWithActor(actor)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireWriteOrEditGrant_EditGrantForDifferentArtifact_Returns403(t *testing.T) {
	r := setupTestRouter("capabilities", "id")
	actor := sharedctx.NewActor("user-1", "user@test.com", "stakeholder")
	actor = actor.WithEditGrants(map[string]map[string]bool{
		"capability": {"other-artifact-456": true},
	})

	req := requestWithActor(actor)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireWriteOrEditGrant_EditGrantForDifferentType_Returns403(t *testing.T) {
	r := setupTestRouter("capabilities", "id")
	actor := sharedctx.NewActor("user-1", "user@test.com", "stakeholder")
	actor = actor.WithEditGrants(map[string]map[string]bool{
		"component": {"artifact-123": true},
	})

	req := requestWithActor(actor)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireWriteOrEditGrant_NonexistentURLParam_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/test/{id}", func(r chi.Router) {
		r.Use(RequireWriteOrEditGrant("capabilities", "nonexistent_param"))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})
	actor := sharedctx.NewActor("user-1", "user@test.com", "stakeholder")

	req := requestWithActor(actor)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireWriteOrEditGrant_MissingActor_Returns401(t *testing.T) {
	r := setupTestRouter("capabilities", "id")

	req := httptest.NewRequest(http.MethodGet, "/test/artifact-123", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireWriteOrEditGrant_AdminRole_PassesThrough(t *testing.T) {
	r := setupTestRouter("capabilities", "id")
	actor := sharedctx.NewActor("user-1", "admin@test.com", "admin")

	req := requestWithActor(actor)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
