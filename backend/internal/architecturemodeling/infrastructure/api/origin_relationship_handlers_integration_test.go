//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"easi/backend/internal/architecturemodeling/application/handlers"
	"easi/backend/internal/architecturemodeling/application/projectors"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type originTestContext struct {
	db                       *sql.DB
	originHandlers           *OriginRelationshipHandlers
	acquiredEntityHandlers   interface{ CreateAcquiredEntity(http.ResponseWriter, *http.Request) }
	internalTeamHandlers     interface{ CreateInternalTeam(http.ResponseWriter, *http.Request) }
	componentHandlers        *ComponentHandlers
	createdIDs               []string
}

func setupOriginTestHandlers(db *sql.DB) *originTestContext {
	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	// Setup component handlers
	componentReadModel := readmodels.NewApplicationComponentReadModel(tenantDB)
	componentProjector := projectors.NewApplicationComponentProjector(componentReadModel)
	eventBus.Subscribe("ApplicationComponentCreated", componentProjector)
	eventBus.Subscribe("ApplicationComponentUpdated", componentProjector)
	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	createComponentHandler := handlers.NewCreateApplicationComponentHandler(componentRepo)
	commandBus.Register("CreateApplicationComponent", createComponentHandler)
	componentHandlers := NewComponentHandlers(commandBus, componentReadModel, hateoas)

	// Setup origin relationship handlers
	acquiredViaReadModel := readmodels.NewAcquiredViaRelationshipReadModel(tenantDB)
	purchasedFromReadModel := readmodels.NewPurchasedFromRelationshipReadModel(tenantDB)
	builtByReadModel := readmodels.NewBuiltByRelationshipReadModel(tenantDB)

	originProjector := projectors.NewOriginRelationshipProjector(acquiredViaReadModel, purchasedFromReadModel, builtByReadModel)
	eventBus.Subscribe("AcquiredViaRelationshipSet", originProjector)
	eventBus.Subscribe("AcquiredViaRelationshipReplaced", originProjector)
	eventBus.Subscribe("AcquiredViaNotesUpdated", originProjector)
	eventBus.Subscribe("AcquiredViaRelationshipCleared", originProjector)
	eventBus.Subscribe("PurchasedFromRelationshipSet", originProjector)
	eventBus.Subscribe("PurchasedFromRelationshipReplaced", originProjector)
	eventBus.Subscribe("PurchasedFromNotesUpdated", originProjector)
	eventBus.Subscribe("PurchasedFromRelationshipCleared", originProjector)
	eventBus.Subscribe("BuiltByRelationshipSet", originProjector)
	eventBus.Subscribe("BuiltByRelationshipReplaced", originProjector)
	eventBus.Subscribe("BuiltByNotesUpdated", originProjector)
	eventBus.Subscribe("BuiltByRelationshipCleared", originProjector)
	eventBus.Subscribe("ComponentOriginsDeleted", originProjector)

	componentOriginsRepo := repositories.NewComponentOriginsRepository(eventStore)
	setAcquiredViaHandler := handlers.NewSetAcquiredViaHandler(componentOriginsRepo)
	setBuiltByHandler := handlers.NewSetBuiltByHandler(componentOriginsRepo)
	commandBus.Register("SetAcquiredVia", setAcquiredViaHandler)
	commandBus.Register("SetBuiltBy", setBuiltByHandler)

	originHandlers := NewOriginRelationshipHandlers(
		commandBus,
		acquiredViaReadModel,
		purchasedFromReadModel,
		builtByReadModel,
		hateoas,
	)

	// We also need to setup acquired entity handlers for creating test data
	// For simplicity, we'll create test data directly in the read models

	return &originTestContext{
		db:                db,
		originHandlers:    originHandlers,
		componentHandlers: componentHandlers,
		createdIDs:        make([]string, 0),
	}
}

func (ctx *originTestContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func (ctx *originTestContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

func (ctx *originTestContext) cleanup() {
	ctx.setTenantContextNoError()
	for _, id := range ctx.createdIDs {
		ctx.db.Exec("DELETE FROM acquired_via_relationships WHERE component_id = $1 OR id = $1", id)
		ctx.db.Exec("DELETE FROM built_by_relationships WHERE component_id = $1 OR id = $1", id)
		ctx.db.Exec("DELETE FROM purchased_from_relationships WHERE component_id = $1 OR id = $1", id)
		ctx.db.Exec("DELETE FROM application_components WHERE id = $1", id)
		ctx.db.Exec("DELETE FROM acquired_entities WHERE id = $1", id)
		ctx.db.Exec("DELETE FROM internal_teams WHERE id = $1", id)
		ctx.db.Exec("DELETE FROM events WHERE aggregate_id = $1 OR aggregate_id LIKE $2", id, "component-origins:"+id)
	}
	ctx.db.Close()
}

func (ctx *originTestContext) setTenantContextNoError() {
	ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
}

func (ctx *originTestContext) createTestComponent(t *testing.T, name string) string {
	ctx.setTenantContext(t)
	id := uuid.New().String()
	_, err := ctx.db.Exec(
		"INSERT INTO application_components (id, name, description, tenant_id, created_at) VALUES ($1, $2, $3, $4, NOW())",
		id, name, "Test component", testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackID(id)
	return id
}

func (ctx *originTestContext) createTestAcquiredEntity(t *testing.T, name string) string {
	ctx.setTenantContext(t)
	id := uuid.New().String()
	_, err := ctx.db.Exec(
		"INSERT INTO acquired_entities (id, name, tenant_id, created_at) VALUES ($1, $2, $3, NOW())",
		id, name, testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackID(id)
	return id
}

func (ctx *originTestContext) createTestInternalTeam(t *testing.T, name string) string {
	ctx.setTenantContext(t)
	id := uuid.New().String()
	_, err := ctx.db.Exec(
		"INSERT INTO internal_teams (id, name, tenant_id, created_at) VALUES ($1, $2, $3, NOW())",
		id, name, testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackID(id)
	return id
}

func (ctx *originTestContext) makeOriginRequest(t *testing.T, method, url string, body string, componentId string) (*httptest.ResponseRecorder, *http.Request) {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	req = withTestTenant(req)

	if componentId != "" {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("componentId", componentId)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	return httptest.NewRecorder(), req
}

// TestMultipleOriginRelationships reproduces the bug where "acquired via" edge
// disappears when "built by" is added to the same component.
func TestMultipleOriginRelationships(t *testing.T) {
	dbCtx, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := setupOriginTestHandlers(dbCtx.db)
	defer ctx.cleanup()

	// Step 1: Create test data
	t.Log("Step 1: Creating test data...")
	componentID := ctx.createTestComponent(t, "Test Component for Multi-Origin")
	acquiredEntityID := ctx.createTestAcquiredEntity(t, "Test Acquired Entity")
	teamID := ctx.createTestInternalTeam(t, "Test Team")
	t.Logf("  Component ID: %s", componentID)
	t.Logf("  Acquired Entity ID: %s", acquiredEntityID)
	t.Logf("  Internal Team ID: %s", teamID)

	// Step 2: Link component to acquired entity (acquired-via)
	t.Log("Step 2: Linking component to acquired entity (acquired-via)...")
	linkBody := fmt.Sprintf(`{"acquiredEntityId": "%s", "notes": "Test acquired via"}`, acquiredEntityID)
	url := fmt.Sprintf("/api/v1/components/%s/origin/acquired-via", componentID)

	req := httptest.NewRequest(http.MethodPut, url, strings.NewReader(linkBody))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("componentId", componentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	ctx.originHandlers.CreateAcquiredViaRelationship(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Failed to create acquired-via: %s", w.Body.String())
	t.Log("  ✓ Acquired-via relationship created")

	// Wait for projection
	time.Sleep(100 * time.Millisecond)

	// Step 3: Verify acquired-via exists
	t.Log("Step 3: Verifying acquired-via relationship exists...")
	req = httptest.NewRequest(http.MethodGet, "/api/v1/origin-relationships", nil)
	req = withTestTenant(req)
	w = httptest.NewRecorder()
	ctx.originHandlers.GetAllOriginRelationships(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp1 AllOriginRelationshipsDTO
	err := json.NewDecoder(w.Body).Decode(&resp1)
	require.NoError(t, err)

	foundAcquiredVia := false
	for _, rel := range resp1.AcquiredVia {
		if rel.ComponentID == componentID {
			foundAcquiredVia = true
			break
		}
	}
	require.True(t, foundAcquiredVia, "Acquired-via not found before adding built-by")
	t.Logf("  ✓ Acquired-via found (AcquiredVia count: %d)", len(resp1.AcquiredVia))

	// Step 4: Link component to internal team (built-by)
	t.Log("Step 4: Linking component to internal team (built-by)...")
	builtByBody := fmt.Sprintf(`{"internalTeamId": "%s", "notes": "Test built by"}`, teamID)
	url = fmt.Sprintf("/api/v1/components/%s/origin/built-by", componentID)

	req = httptest.NewRequest(http.MethodPut, url, strings.NewReader(builtByBody))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("componentId", componentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	ctx.originHandlers.CreateBuiltByRelationship(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Failed to create built-by: %s", w.Body.String())
	t.Log("  ✓ Built-by relationship created")

	// Wait for projection
	time.Sleep(100 * time.Millisecond)

	// Step 5: Verify BOTH relationships still exist
	t.Log("Step 5: Verifying BOTH relationships still exist...")
	req = httptest.NewRequest(http.MethodGet, "/api/v1/origin-relationships", nil)
	req = withTestTenant(req)
	w = httptest.NewRecorder()
	ctx.originHandlers.GetAllOriginRelationships(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	t.Logf("  Response: %s", w.Body.String())

	var resp2 AllOriginRelationshipsDTO
	err = json.NewDecoder(strings.NewReader(w.Body.String())).Decode(&resp2)
	require.NoError(t, err)

	// Check acquired-via still exists
	foundAcquiredVia = false
	for _, rel := range resp2.AcquiredVia {
		if rel.ComponentID == componentID {
			foundAcquiredVia = true
			t.Logf("  Found acquired-via: componentId=%s, entityId=%s", rel.ComponentID, rel.AcquiredEntityID)
			break
		}
	}

	// Check built-by exists
	foundBuiltBy := false
	for _, rel := range resp2.BuiltBy {
		if rel.ComponentID == componentID {
			foundBuiltBy = true
			t.Logf("  Found built-by: componentId=%s, teamId=%s", rel.ComponentID, rel.InternalTeamID)
			break
		}
	}

	t.Logf("\nResults:")
	t.Logf("  AcquiredVia count: %d, found for component: %v", len(resp2.AcquiredVia), foundAcquiredVia)
	t.Logf("  BuiltBy count: %d, found for component: %v", len(resp2.BuiltBy), foundBuiltBy)

	if !foundAcquiredVia {
		t.Error("BUG REPRODUCED: Acquired-via relationship DISAPPEARED after adding built-by!")
	}
	if !foundBuiltBy {
		t.Error("Built-by relationship not found!")
	}

	if foundAcquiredVia && foundBuiltBy {
		t.Log("\n✅ TEST PASSED: Both relationships exist after adding built-by")
		t.Log("📊 CONCLUSION: Backend correctly maintains both relationships")
	}
}
