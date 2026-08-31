//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	"easi/backend/internal/shared/types"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type xRelatedFixture struct {
	componentHandlers *ComponentHandlers
	relationHandlers  *RelationHandlers
	componentReadM    *readmodels.ApplicationComponentReadModel
	relationReadM     *readmodels.ComponentRelationReadModel
	db                *sql.DB
}

func newXRelatedFixture(db *sql.DB) *xRelatedFixture {
	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	links := NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))

	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	componentReadM := readmodels.NewApplicationComponentReadModel(tenantDB)
	componentProjector := projectors.NewApplicationComponentProjector(componentReadM)
	eventBus.Subscribe("ApplicationComponentCreated", componentProjector)
	eventBus.Subscribe("ApplicationComponentUpdated", componentProjector)
	eventBus.Subscribe("ApplicationComponentDeleted", componentProjector)

	relationReadM := readmodels.NewComponentRelationReadModel(tenantDB)
	relationProjector := projectors.NewComponentRelationProjector(relationReadM)
	eventBus.Subscribe("ComponentRelationCreated", relationProjector)
	eventBus.Subscribe("ComponentRelationDeleted", relationProjector)

	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	relationRepo := repositories.NewComponentRelationRepository(eventStore)
	createComp := handlers.NewCreateApplicationComponentHandler(componentRepo)
	updateComp := handlers.NewUpdateApplicationComponentHandler(componentRepo)
	deleteComp := handlers.NewDeleteApplicationComponentHandler(componentRepo, relationReadM, commandBus)
	createRel := handlers.NewCreateComponentRelationHandler(relationRepo)
	deleteRel := handlers.NewDeleteComponentRelationHandler(relationRepo)
	commandBus.Register("CreateApplicationComponent", createComp)
	commandBus.Register("UpdateApplicationComponent", updateComp)
	commandBus.Register("DeleteApplicationComponent", deleteComp)
	commandBus.Register("CreateComponentRelation", createRel)
	commandBus.Register("DeleteComponentRelation", deleteRel)

	componentHandlers := NewComponentHandlers(commandBus, componentReadM, links)
	relationHandlers := NewRelationHandlers(commandBus, relationReadM, links)

	return &xRelatedFixture{
		componentHandlers: componentHandlers,
		relationHandlers:  relationHandlers,
		componentReadM:    componentReadM,
		relationReadM:     relationReadM,
		db:                db,
	}
}

func (f *xRelatedFixture) createComponentViaAPI(t *testing.T, testCtx *testContext, name, description string) string {
	t.Helper()
	body, _ := json.Marshal(CreateApplicationComponentRequest{Name: name, Description: description})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/components", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	req = withArchitectActor(req)
	f.componentHandlers.CreateApplicationComponent(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "POST /components failed: %s", w.Body.String())

	testCtx.setTenantContext(t)
	var id string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM infrastructure.events WHERE event_type = 'ApplicationComponentCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&id)
	require.NoError(t, err)
	testCtx.trackID(id)
	require.Eventually(t, func() bool {
		c, err := f.componentReadM.GetByID(tenantContext(), id)
		return err == nil && c != nil && c.Name == name
	}, 2*time.Second, 50*time.Millisecond, "expected component %q to project", name)
	return id
}

func TestXRelated_ComponentToComponent_RoundTrip_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	f := newXRelatedFixture(testCtx.db)

	sourceID := f.createComponentViaAPI(t, testCtx, "Source Service", "Source for x-related test")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/components/"+sourceID, nil)
	req = withTestTenant(req)
	req = withArchitectActor(req)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", sourceID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	f.componentHandlers.GetComponentByID(w, req)
	require.Equal(t, http.StatusOK, w.Code, "GET source component failed: %s", w.Body.String())

	advertised := requireXRelatedEntry(t, w.Body.Bytes(), "component-triggers")
	require.Equal(t, "/api/v1/components", advertised.Href)
	require.Contains(t, advertised.Methods, "POST")
	require.Equal(t, "component", advertised.TargetType)

	targetID := f.createComponentViaAPI(t, testCtx, "Target Service", "Created via x-related advertised href")

	endpoint, ok := types.LookupRelationEndpoint(advertised.RelationType)
	require.True(t, ok, "relationType %q must resolve to a backend endpoint via LookupRelationEndpoint", advertised.RelationType)
	require.Equal(t, "/api/v1/relations", endpoint.Path)
	require.Equal(t, http.MethodPost, endpoint.Method)

	relBody, _ := json.Marshal(CreateComponentRelationRequest{
		SourceComponentID: sourceID,
		TargetComponentID: targetID,
		RelationType:      "Triggers",
		Name:              "x-related round-trip",
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(endpoint.Method, endpoint.Path, bytes.NewReader(relBody))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	req = withArchitectActor(req)
	f.relationHandlers.CreateComponentRelation(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "POST %s failed: %s", endpoint.Path, w.Body.String())

	var relationID string
	testCtx.setTenantContext(t)
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM infrastructure.events WHERE event_type = 'ComponentRelationCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&relationID)
	require.NoError(t, err)
	testCtx.trackID(relationID)

	require.Eventually(t, func() bool {
		rel, err := f.relationReadM.GetByID(tenantContext(), relationID)
		return err == nil && rel != nil && rel.SourceComponentID == sourceID && rel.TargetComponentID == targetID
	}, 2*time.Second, 50*time.Millisecond, "expected relation to be projected")
}

func requireXRelatedEntry(t *testing.T, body []byte, relationType string) types.RelatedLink {
	t.Helper()
	var envelope struct {
		Links struct {
			XRelated []types.RelatedLink `json:"x-related"`
		} `json:"_links"`
	}
	require.NoError(t, json.Unmarshal(body, &envelope))
	for _, e := range envelope.Links.XRelated {
		if e.RelationType == relationType {
			return e
		}
	}
	require.FailNowf(t, "missing x-related entry", "no x-related entry with relationType=%q in body: %s", relationType, string(body))
	return types.RelatedLink{}
}
