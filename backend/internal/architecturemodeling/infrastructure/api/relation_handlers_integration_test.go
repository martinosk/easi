// +build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// relationTestContext holds test-specific state for cleanup
type relationTestContext struct {
	db         *sql.DB
	testID     string
	createdIDs []string
}

// setTenantContext sets the tenant context for RLS before running raw queries
func (ctx *relationTestContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func setupRelationTestDB(t *testing.T) (*relationTestContext, func()) {
	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "change_me_in_production")
	dbName := getEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	// Create unique test ID based on test name and timestamp to avoid collisions
	testID := fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())

	ctx := &relationTestContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	// Clean up only the data created in this specific test
	cleanup := func() {
		// Delete relations and components by tracking the IDs created during the test
		for _, id := range ctx.createdIDs {
			db.Exec("DELETE FROM component_relations WHERE id = $1", id)
			db.Exec("DELETE FROM application_components WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

// trackID adds an aggregate ID to the cleanup list
func (ctx *relationTestContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

type testRelationParams struct {
	ID           string
	SourceID     string
	TargetID     string
	RelationType string
	Name         string
	Description  string
}

func (ctx *relationTestContext) createTestRelation(t *testing.T, params testRelationParams) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO component_relations (id, source_component_id, target_component_id, relation_type, name, description, tenant_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		params.ID, params.SourceID, params.TargetID, params.RelationType, params.Name, params.Description, testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackID(params.ID)
}

func setupRelationHandlers(db *sql.DB) (*RelationHandlers, *readmodels.ComponentRelationReadModel) {
	// Wrap database connection with tenant-aware wrapper for RLS
	tenantDB := database.NewTenantAwareDB(db)

	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	// Setup repository and handlers
	relationRepo := repositories.NewComponentRelationRepository(eventStore)
	createHandler := handlers.NewCreateComponentRelationHandler(relationRepo)
	deleteHandler := handlers.NewDeleteComponentRelationHandler(relationRepo)
	commandBus.Register("CreateComponentRelation", createHandler)
	commandBus.Register("DeleteComponentRelation", deleteHandler)

	// Setup read model
	readModel := readmodels.NewComponentRelationReadModel(tenantDB)

	// Setup HTTP handlers
	relationHandlers := NewRelationHandlers(commandBus, readModel, hateoas)

	return relationHandlers, readModel
}

func TestCreateRelation_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, readModel := setupRelationHandlers(testCtx.db)

	// Create relation via API with unique component IDs (must be valid UUIDs)
	sourceID := uuid.New().String()
	targetID := uuid.New().String()

	reqBody := CreateComponentRelationRequest{
		SourceComponentID: sourceID,
		TargetComponentID: targetID,
		RelationType:      "Triggers",
		Name:              "Triggers API",
		Description:       "Frontend triggers backend API",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/relations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.CreateComponentRelation(w, req)

	// Assert HTTP response
	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get the created aggregate ID from the event store
	testCtx.setTenantContext(t)
	var aggregateID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ComponentRelationCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&aggregateID)
	require.NoError(t, err)
	testCtx.trackID(aggregateID)

	// Verify event data contains expected values
	var eventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'ComponentRelationCreated'",
		aggregateID,
	).Scan(&eventData)
	require.NoError(t, err)
	assert.Contains(t, eventData, sourceID)
	assert.Contains(t, eventData, targetID)
	assert.Contains(t, eventData, "Triggers")

	// Manually insert into read model for testing (simulating event projection)
	_, err = testCtx.db.Exec(
		"INSERT INTO component_relations (id, source_component_id, target_component_id, relation_type, name, description, tenant_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		aggregateID, sourceID, targetID, "Triggers", "Triggers API", "Frontend triggers backend API", testTenantID(),
	)
	require.NoError(t, err)

	// Verify read model contains the relation
	relation, err := readModel.GetByID(tenantContext(), aggregateID)
	require.NoError(t, err)
	assert.NotNil(t, relation)
	assert.Equal(t, sourceID, relation.SourceComponentID)
	assert.Equal(t, targetID, relation.TargetComponentID)
	assert.Equal(t, "Triggers", relation.RelationType)
}

func TestCreateRelation_ValidationError_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)

	// Create relation with invalid relation type (but valid UUIDs)
	reqBody := CreateComponentRelationRequest{
		SourceComponentID: uuid.New().String(),
		TargetComponentID: uuid.New().String(),
		RelationType:      "INVALID_TYPE",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/relations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.CreateComponentRelation(w, req)

	// Assert validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify no event was created in our test's scope
	var count int
	err := testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM events WHERE created_at > NOW() - INTERVAL '5 seconds'",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "No events should be created for invalid request")
}

func TestGetAllRelations_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)

	// Create test data directly in read model with unique IDs (UUIDs required)
	id1 := uuid.New().String()
	id2 := uuid.New().String()
	comp1 := uuid.New().String()
	comp2 := uuid.New().String()
	comp3 := uuid.New().String()

	testCtx.createTestRelation(t, testRelationParams{ID: id1, SourceID: comp1, TargetID: comp2, RelationType: "Triggers", Name: "Triggers", Description: "Description 1"})
	testCtx.createTestRelation(t, testRelationParams{ID: id2, SourceID: comp2, TargetID: comp3, RelationType: "Serves", Name: "Serves", Description: "Description 2"})

	// Test GET all
	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.GetAllRelations(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.ComponentRelationDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Find our test relations in the response
	foundRelations := 0
	for _, rel := range response.Data {
		if rel.ID == id1 || rel.ID == id2 {
			foundRelations++
			assert.NotNil(t, rel.Links)
			assert.Contains(t, rel.Links, "self")
		}
	}
	assert.Equal(t, 2, foundRelations, "Should find both test relations")
}

func TestGetRelationByID_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)

	// Create test data with unique IDs (UUIDs required)
	relationID := uuid.New().String()
	comp1 := uuid.New().String()
	comp2 := uuid.New().String()

	testCtx.createTestRelation(t, testRelationParams{ID: relationID, SourceID: comp1, TargetID: comp2, RelationType: "Triggers", Name: "Test Relation", Description: "Test Description"})

	// Test GET by ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations/"+relationID, nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	// Add URL param using chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", relationID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetRelationByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data readmodels.ComponentRelationDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, relationID, response.Data.ID)
	assert.Equal(t, comp1, response.Data.SourceComponentID)
	assert.Equal(t, comp2, response.Data.TargetComponentID)
	assert.Equal(t, "Triggers", response.Data.RelationType)
	assert.NotNil(t, response.Data.Links)
}

func TestGetRelationByID_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)

	// Test GET non-existent relation with unique UUID
	nonExistentID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations/"+nonExistentID, nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetRelationByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func testComponentRelations(t *testing.T, testCtx *relationTestContext, handlers *RelationHandlers, direction string) {
	componentID := uuid.New().String()
	other1 := uuid.New().String()
	other2 := uuid.New().String()
	otherComp := uuid.New().String()

	rel1 := uuid.New().String()
	rel2 := uuid.New().String()
	rel3 := uuid.New().String()

	var sourceIDForMatch, targetIDForMatch string
	if direction == "from" {
		testCtx.createTestRelation(t, testRelationParams{ID: rel1, SourceID: componentID, TargetID: other1, RelationType: "Triggers", Name: "Relation 1", Description: "Description 1"})
		testCtx.createTestRelation(t, testRelationParams{ID: rel2, SourceID: componentID, TargetID: other2, RelationType: "Serves", Name: "Relation 2", Description: "Description 2"})
		testCtx.createTestRelation(t, testRelationParams{ID: rel3, SourceID: otherComp, TargetID: componentID, RelationType: "Triggers", Name: "Relation 3", Description: "Description 3"})
		sourceIDForMatch = componentID
	} else {
		testCtx.createTestRelation(t, testRelationParams{ID: rel1, SourceID: other1, TargetID: componentID, RelationType: "Triggers", Name: "Relation 1", Description: "Description 1"})
		testCtx.createTestRelation(t, testRelationParams{ID: rel2, SourceID: other2, TargetID: componentID, RelationType: "Serves", Name: "Relation 2", Description: "Description 2"})
		testCtx.createTestRelation(t, testRelationParams{ID: rel3, SourceID: componentID, TargetID: otherComp, RelationType: "Triggers", Name: "Relation 3", Description: "Description 3"})
		targetIDForMatch = componentID
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations/"+direction+"/"+componentID, nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("componentId", componentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	if direction == "from" {
		handlers.GetRelationsFromComponent(w, req)
	} else {
		handlers.GetRelationsToComponent(w, req)
	}

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.ComponentRelationDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	foundRelations := 0
	for _, rel := range response.Data {
		if rel.ID == rel1 || rel.ID == rel2 {
			foundRelations++
			if direction == "from" {
				assert.Equal(t, sourceIDForMatch, rel.SourceComponentID)
			} else {
				assert.Equal(t, targetIDForMatch, rel.TargetComponentID)
			}
		}
		assert.NotEqual(t, rel3, rel.ID, "Relation with different "+direction+" component should not be included")
	}
	assert.Equal(t, 2, foundRelations, "Should find exactly 2 relations "+direction+" this component")
}

func TestGetRelationsFromComponent_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)
	testComponentRelations(t, testCtx, handlers, "from")
}

func TestGetRelationsToComponent_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)
	testComponentRelations(t, testCtx, handlers, "to")
}

func TestGetAllRelationsPaginated_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)

	// Create test data with unique IDs and different timestamps (UUIDs required)
	comp1 := uuid.New().String()
	comp2 := uuid.New().String()

	testCtx.setTenantContext(t) // Set tenant context once for all inserts
	for i := 1; i <= 5; i++ {
		id := uuid.New().String()
		name := fmt.Sprintf("Relation %d", i)

		relType := "Triggers"
		if i%2 == 0 {
			relType = "Serves"
		}
		_, err := testCtx.db.Exec(
			"INSERT INTO component_relations (id, source_component_id, target_component_id, relation_type, name, description, tenant_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW() - INTERVAL '"+fmt.Sprintf("%d", i)+" seconds')",
			id, comp1, comp2, relType, name, "Description", testTenantID(),
		)
		require.NoError(t, err)
		testCtx.trackID(id)

		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Test GET first page with limit
	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations?limit=2", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.GetAllRelations(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data       []readmodels.ComponentRelationDTO `json:"data"`
		Pagination struct {
			Cursor  string `json:"cursor"`
			HasMore bool   `json:"hasMore"`
			Limit   int    `json:"limit"`
		} `json:"pagination"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should have at least 2 relations (our test data)
	assert.GreaterOrEqual(t, len(response.Data), 2, "Should return at least 2 relations")
	assert.Equal(t, 2, response.Pagination.Limit)
}

func TestDeleteRelation_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)

	sourceID := uuid.New().String()
	targetID := uuid.New().String()

	reqBody := CreateComponentRelationRequest{
		SourceComponentID: sourceID,
		TargetComponentID: targetID,
		RelationType:      "Triggers",
		Name:              "Test Relation",
		Description:       "This will be deleted",
	}
	body, _ := json.Marshal(reqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/relations", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withTestTenant(createReq)
	createW := httptest.NewRecorder()

	handlers.CreateComponentRelation(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var relationID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ComponentRelationCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&relationID)
	require.NoError(t, err)
	testCtx.trackID(relationID)

	time.Sleep(100 * time.Millisecond)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/relations/"+relationID, nil)
	deleteReq = withTestTenant(deleteReq)
	deleteReq = deleteReq.WithContext(context.WithValue(deleteReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{relationID},
		},
	}))
	deleteW := httptest.NewRecorder()

	handlers.DeleteComponentRelation(deleteW, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteW.Code)

	var deleteEventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'ComponentRelationDeleted'",
		relationID,
	).Scan(&deleteEventData)
	require.NoError(t, err)
	assert.NotEmpty(t, deleteEventData)
}

type cascadeTestDependencies struct {
	componentHandlers *ComponentHandlers
	relationHandlers  *RelationHandlers
	db                *sql.DB
}

func setupCascadeTestDependencies(testCtx *testContext) *cascadeTestDependencies {
	tenantDB := database.NewTenantAwareDB(testCtx.db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	componentReadModel := readmodels.NewApplicationComponentReadModel(tenantDB)
	relationReadModel := readmodels.NewComponentRelationReadModel(tenantDB)

	componentProjector := projectors.NewApplicationComponentProjector(componentReadModel)
	relationProjector := projectors.NewComponentRelationProjector(relationReadModel)
	eventBus.Subscribe("ApplicationComponentCreated", componentProjector)
	eventBus.Subscribe("ApplicationComponentDeleted", componentProjector)
	eventBus.Subscribe("ComponentRelationCreated", relationProjector)
	eventBus.Subscribe("ComponentRelationDeleted", relationProjector)

	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	relationRepo := repositories.NewComponentRelationRepository(eventStore)

	createComponentHandler := handlers.NewCreateApplicationComponentHandler(componentRepo)
	deleteComponentHandler := handlers.NewDeleteApplicationComponentHandler(componentRepo, relationReadModel, commandBus)
	createRelationHandler := handlers.NewCreateComponentRelationHandler(relationRepo)
	deleteRelationHandler := handlers.NewDeleteComponentRelationHandler(relationRepo)

	commandBus.Register("CreateApplicationComponent", createComponentHandler)
	commandBus.Register("DeleteApplicationComponent", deleteComponentHandler)
	commandBus.Register("CreateComponentRelation", createRelationHandler)
	commandBus.Register("DeleteComponentRelation", deleteRelationHandler)

	componentHandlers := NewComponentHandlers(commandBus, componentReadModel, hateoas)
	relationHandlers := NewRelationHandlers(commandBus, relationReadModel, hateoas)

	return &cascadeTestDependencies{
		componentHandlers: componentHandlers,
		relationHandlers:  relationHandlers,
		db:                testCtx.db,
	}
}

func createComponentViaAPI(t *testing.T, handlers *ComponentHandlers, name, description string) *httptest.ResponseRecorder {
	reqBody := CreateApplicationComponentRequest{
		Name:        name,
		Description: description,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/components", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.CreateApplicationComponent(w, req)
	return w
}

func createRelationViaAPI(t *testing.T, handlers *RelationHandlers, sourceID, targetID, relationType, name string) *httptest.ResponseRecorder {
	reqBody := CreateComponentRelationRequest{
		SourceComponentID: sourceID,
		TargetComponentID: targetID,
		RelationType:      relationType,
		Name:              name,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/relations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.CreateComponentRelation(w, req)
	return w
}

func deleteComponentViaAPI(t *testing.T, handlers *ComponentHandlers, componentID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/components/"+componentID, nil)
	req = withTestTenant(req)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{componentID},
		},
	}))
	w := httptest.NewRecorder()

	handlers.DeleteApplicationComponent(w, req)
	return w
}

func TestCascadeDeleteRelations_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	deps := setupCascadeTestDependencies(testCtx)

	createW := createComponentViaAPI(t, deps.componentHandlers, "Component With Relations", "This component has relations that should be deleted")
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var componentID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ApplicationComponentCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&componentID)
	require.NoError(t, err)
	testCtx.trackID(componentID)

	time.Sleep(100 * time.Millisecond)

	targetComponentID := uuid.New().String()
	createRelW := createRelationViaAPI(t, deps.relationHandlers, componentID, targetComponentID, "Triggers", "Test Relation")
	assert.Equal(t, http.StatusCreated, createRelW.Code)

	var relationID string
	err = testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ComponentRelationCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&relationID)
	require.NoError(t, err)
	testCtx.trackID(relationID)

	time.Sleep(100 * time.Millisecond)

	deleteW := deleteComponentViaAPI(t, deps.componentHandlers, componentID)
	assert.Equal(t, http.StatusNoContent, deleteW.Code)

	time.Sleep(200 * time.Millisecond)

	var relationDeleted bool
	err = testCtx.db.QueryRow(
		"SELECT is_deleted FROM component_relations WHERE id = $1",
		relationID,
	).Scan(&relationDeleted)
	require.NoError(t, err)
	assert.True(t, relationDeleted, "Relation should be marked as deleted when component is deleted")
}
