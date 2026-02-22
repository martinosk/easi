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
	db                     *sql.DB
	originHandlers         *OriginRelationshipHandlers
	acquiredEntityHandlers interface {
		CreateAcquiredEntity(http.ResponseWriter, *http.Request)
	}
	internalTeamHandlers interface {
		CreateInternalTeam(http.ResponseWriter, *http.Request)
	}
	componentHandlers *ComponentHandlers
	createdIDs        []string
}

func setupOriginTestHandlers(db *sql.DB) *originTestContext {
	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	links := NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	componentReadModel := readmodels.NewApplicationComponentReadModel(tenantDB)
	componentProjector := projectors.NewApplicationComponentProjector(componentReadModel)
	eventBus.Subscribe("ApplicationComponentCreated", componentProjector)
	eventBus.Subscribe("ApplicationComponentUpdated", componentProjector)
	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	createComponentHandler := handlers.NewCreateApplicationComponentHandler(componentRepo)
	commandBus.Register("CreateApplicationComponent", createComponentHandler)
	componentHandlers := NewComponentHandlers(commandBus, componentReadModel, links)

	acquiredViaReadModel := readmodels.NewAcquiredViaRelationshipReadModel(tenantDB)
	purchasedFromReadModel := readmodels.NewPurchasedFromRelationshipReadModel(tenantDB)
	builtByReadModel := readmodels.NewBuiltByRelationshipReadModel(tenantDB)

	originProjector := projectors.NewOriginRelationshipProjector(acquiredViaReadModel, purchasedFromReadModel, builtByReadModel)
	eventBus.Subscribe("OriginLinkCreated", originProjector)
	eventBus.Subscribe("OriginLinkSet", originProjector)
	eventBus.Subscribe("OriginLinkReplaced", originProjector)
	eventBus.Subscribe("OriginLinkNotesUpdated", originProjector)
	eventBus.Subscribe("OriginLinkCleared", originProjector)
	eventBus.Subscribe("OriginLinkDeleted", originProjector)

	originLinkRepo := repositories.NewComponentOriginLinkRepository(eventStore)
	commandBus.Register("SetOriginLink", handlers.NewSetOriginLinkHandler(originLinkRepo))
	commandBus.Register("ClearOriginLink", handlers.NewClearOriginLinkHandler(originLinkRepo))

	originHandlers := NewOriginRelationshipHandlersFromConfig(OriginRelationshipHandlersConfig{
		CommandBus: commandBus,
		ReadModels: OriginReadModels{
			AcquiredVia:   acquiredViaReadModel,
			PurchasedFrom: purchasedFromReadModel,
			BuiltBy:       builtByReadModel,
		},
		HATEOAS: links,
	})

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
		ctx.db.Exec("DELETE FROM architecturemodeling.acquired_via_relationships WHERE component_id = $1 OR id = $1", id)
		ctx.db.Exec("DELETE FROM architecturemodeling.built_by_relationships WHERE component_id = $1 OR id = $1", id)
		ctx.db.Exec("DELETE FROM architecturemodeling.purchased_from_relationships WHERE component_id = $1 OR id = $1", id)
		ctx.db.Exec("DELETE FROM architecturemodeling.application_components WHERE id = $1", id)
		ctx.db.Exec("DELETE FROM architecturemodeling.acquired_entities WHERE id = $1", id)
		ctx.db.Exec("DELETE FROM architecturemodeling.internal_teams WHERE id = $1", id)
		ctx.db.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1 OR aggregate_id LIKE $2 OR aggregate_id LIKE $3", id, "component-origins:"+id, "origin-link:%:"+id)
	}
	ctx.db.Close()
}

func (ctx *originTestContext) setTenantContextNoError() {
	ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
}

func (ctx *originTestContext) insertTestRecord(t *testing.T, query string, args ...interface{}) string {
	ctx.setTenantContext(t)
	id := uuid.New().String()
	allArgs := append([]interface{}{id}, args...)
	allArgs = append(allArgs, testTenantID())
	_, err := ctx.db.Exec(query, allArgs...)
	require.NoError(t, err)
	ctx.trackID(id)
	return id
}

func (ctx *originTestContext) createTestComponent(t *testing.T, name string) string {
	return ctx.insertTestRecord(t,
		"INSERT INTO architecturemodeling.application_components (id, name, description, tenant_id, created_at) VALUES ($1, $2, $3, $4, NOW())",
		name, "Test component")
}

func (ctx *originTestContext) createTestAcquiredEntity(t *testing.T, name string) string {
	return ctx.insertTestRecord(t,
		"INSERT INTO architecturemodeling.acquired_entities (id, name, tenant_id, created_at) VALUES ($1, $2, $3, NOW())",
		name)
}

func (ctx *originTestContext) createTestInternalTeam(t *testing.T, name string) string {
	return ctx.insertTestRecord(t,
		"INSERT INTO architecturemodeling.internal_teams (id, name, tenant_id, created_at) VALUES ($1, $2, $3, NOW())",
		name)
}

type originRequest struct {
	componentID string
	body        string
}

func (ctx *originTestContext) newComponentRequest(opts originRequest) (*httptest.ResponseRecorder, *http.Request) {
	var req *http.Request
	path := fmt.Sprintf("/api/v1/components/%s/origin", opts.componentID)
	if opts.body != "" {
		req = httptest.NewRequest(http.MethodPut, path, strings.NewReader(opts.body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(http.MethodGet, path, nil)
	}
	req = withTestTenant(req)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("componentId", opts.componentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return httptest.NewRecorder(), req
}

func (ctx *originTestContext) getAllOriginRelationships(t *testing.T) AllOriginRelationshipsDTO {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/origin-relationships", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()
	ctx.originHandlers.GetAllOriginRelationships(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp AllOriginRelationshipsDTO
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	return resp
}

func hasAcquiredViaForComponent(rels []readmodels.AcquiredViaRelationshipDTO, componentID string) bool {
	for _, rel := range rels {
		if rel.ComponentID == componentID {
			return true
		}
	}
	return false
}

func hasBuiltByForComponent(rels []readmodels.BuiltByRelationshipDTO, componentID string) bool {
	for _, rel := range rels {
		if rel.ComponentID == componentID {
			return true
		}
	}
	return false
}

func TestMultipleOriginRelationships(t *testing.T) {
	dbCtx, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := setupOriginTestHandlers(dbCtx.db)
	defer ctx.cleanup()

	componentID := ctx.createTestComponent(t, "Test Component for Multi-Origin")
	acquiredEntityID := ctx.createTestAcquiredEntity(t, "Test Acquired Entity")
	teamID := ctx.createTestInternalTeam(t, "Test Team")

	t.Run("set acquired-via relationship", func(t *testing.T) {
		body := fmt.Sprintf(`{"acquiredEntityId": "%s", "notes": "Test acquired via"}`, acquiredEntityID)
		w, req := ctx.newComponentRequest(originRequest{componentID: componentID, body: body})
		ctx.originHandlers.CreateAcquiredViaRelationship(w, req)
		require.Equal(t, http.StatusOK, w.Code, "Failed to create acquired-via: %s", w.Body.String())
	})

	time.Sleep(100 * time.Millisecond)

	t.Run("acquired-via exists before adding built-by", func(t *testing.T) {
		resp := ctx.getAllOriginRelationships(t)
		require.True(t, hasAcquiredViaForComponent(resp.AcquiredVia, componentID),
			"Acquired-via not found before adding built-by")
	})

	t.Run("set built-by relationship", func(t *testing.T) {
		body := fmt.Sprintf(`{"internalTeamId": "%s", "notes": "Test built by"}`, teamID)
		w, req := ctx.newComponentRequest(originRequest{componentID: componentID, body: body})
		ctx.originHandlers.CreateBuiltByRelationship(w, req)
		require.Equal(t, http.StatusOK, w.Code, "Failed to create built-by: %s", w.Body.String())
	})

	time.Sleep(100 * time.Millisecond)

	t.Run("both relationships exist after adding built-by", func(t *testing.T) {
		resp := ctx.getAllOriginRelationships(t)
		require.True(t, hasAcquiredViaForComponent(resp.AcquiredVia, componentID),
			"Acquired-via relationship disappeared after adding built-by")
		require.True(t, hasBuiltByForComponent(resp.BuiltBy, componentID),
			"Built-by relationship not found")
	})
}
