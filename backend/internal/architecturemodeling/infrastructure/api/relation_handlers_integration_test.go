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
	"os"
	"testing"
	"time"

	"easi/backend/internal/architecturemodeling/application/handlers"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
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

func setupRelationTestDB(t *testing.T) (*relationTestContext, func()) {
	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "easi")
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// trackID adds an aggregate ID to the cleanup list
func (ctx *relationTestContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

// createTestRelation creates a relation directly in the read model for testing
func (ctx *relationTestContext) createTestRelation(t *testing.T, id, sourceID, targetID, relationType, name, description string) {
	_, err := ctx.db.Exec(
		"INSERT INTO component_relations (id, source_component_id, target_component_id, relation_type, name, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())",
		id, sourceID, targetID, relationType, name, description,
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func setupRelationHandlers(db *sql.DB) (*RelationHandlers, *readmodels.ComponentRelationReadModel) {
	eventStore := eventstore.NewPostgresEventStore(db)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	// Setup repository and handlers
	relationRepo := repositories.NewComponentRelationRepository(eventStore)
	createHandler := handlers.NewCreateComponentRelationHandler(relationRepo)
	commandBus.Register("CreateComponentRelation", createHandler)

	// Setup read model
	readModel := readmodels.NewComponentRelationReadModel(db)

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
	w := httptest.NewRecorder()

	handlers.CreateComponentRelation(w, req)

	// Assert HTTP response
	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get the created aggregate ID from the event store
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
		"INSERT INTO component_relations (id, source_component_id, target_component_id, relation_type, name, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())",
		aggregateID, sourceID, targetID, "Triggers", "Triggers API", "Frontend triggers backend API",
	)
	require.NoError(t, err)

	// Verify read model contains the relation
	relation, err := readModel.GetByID(context.Background(), aggregateID)
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

	testCtx.createTestRelation(t, id1, comp1, comp2, "Triggers", "Triggers", "Description 1")
	testCtx.createTestRelation(t, id2, comp2, comp3, "Serves", "Serves", "Description 2")

	// Test GET all
	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations", nil)
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

	testCtx.createTestRelation(t, relationID, comp1, comp2, "Triggers", "Test Relation", "Test Description")

	// Test GET by ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations/"+relationID, nil)
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
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetRelationByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetRelationsFromComponent_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)

	// Create unique component IDs (UUIDs required)
	componentID := uuid.New().String()
	target1 := uuid.New().String()
	target2 := uuid.New().String()
	otherComp := uuid.New().String()

	// Create test data - relations from component
	rel1 := uuid.New().String()
	rel2 := uuid.New().String()
	rel3 := uuid.New().String()

	testCtx.createTestRelation(t, rel1, componentID, target1, "Triggers", "Relation 1", "Description 1")
	testCtx.createTestRelation(t, rel2, componentID, target2, "Serves", "Relation 2", "Description 2")
	// This one should not be included (different source)
	testCtx.createTestRelation(t, rel3, otherComp, componentID, "Triggers", "Relation 3", "Description 3")

	// Test GET relations from component
	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations/from/"+componentID, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("componentId", componentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetRelationsFromComponent(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.ComponentRelationDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Find our test relations in the response (should only find relations FROM this component)
	foundRelations := 0
	for _, rel := range response.Data {
		if rel.ID == rel1 || rel.ID == rel2 {
			foundRelations++
			assert.Equal(t, componentID, rel.SourceComponentID)
		}
		// Make sure rel3 is not in the results
		assert.NotEqual(t, rel3, rel.ID, "Relation with different source should not be included")
	}
	assert.Equal(t, 2, foundRelations, "Should find exactly 2 relations from this component")
}

func TestGetRelationsToComponent_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)

	// Create unique component IDs (UUIDs required)
	componentID := uuid.New().String()
	source1 := uuid.New().String()
	source2 := uuid.New().String()
	otherComp := uuid.New().String()

	// Create test data - relations to component
	rel1 := uuid.New().String()
	rel2 := uuid.New().String()
	rel3 := uuid.New().String()

	testCtx.createTestRelation(t, rel1, source1, componentID, "Triggers", "Relation 1", "Description 1")
	testCtx.createTestRelation(t, rel2, source2, componentID, "Serves", "Relation 2", "Description 2")
	// This one should not be included (different target)
	testCtx.createTestRelation(t, rel3, componentID, otherComp, "Triggers", "Relation 3", "Description 3")

	// Test GET relations to component
	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations/to/"+componentID, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("componentId", componentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetRelationsToComponent(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.ComponentRelationDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Find our test relations in the response (should only find relations TO this component)
	foundRelations := 0
	for _, rel := range response.Data {
		if rel.ID == rel1 || rel.ID == rel2 {
			foundRelations++
			assert.Equal(t, componentID, rel.TargetComponentID)
		}
		// Make sure rel3 is not in the results
		assert.NotEqual(t, rel3, rel.ID, "Relation with different target should not be included")
	}
	assert.Equal(t, 2, foundRelations, "Should find exactly 2 relations to this component")
}

func TestGetAllRelationsPaginated_Integration(t *testing.T) {
	testCtx, cleanup := setupRelationTestDB(t)
	defer cleanup()

	handlers, _ := setupRelationHandlers(testCtx.db)

	// Create test data with unique IDs and different timestamps (UUIDs required)
	comp1 := uuid.New().String()
	comp2 := uuid.New().String()

	for i := 1; i <= 5; i++ {
		id := uuid.New().String()
		name := fmt.Sprintf("Relation %d", i)

		relType := "Triggers"
		if i%2 == 0 {
			relType = "Serves"
		}
		_, err := testCtx.db.Exec(
			"INSERT INTO component_relations (id, source_component_id, target_component_id, relation_type, name, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW() - INTERVAL '"+fmt.Sprintf("%d", i)+" seconds')",
			id, comp1, comp2, relType, name, "Description",
		)
		require.NoError(t, err)
		testCtx.trackID(id)

		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Test GET first page with limit
	req := httptest.NewRequest(http.MethodGet, "/api/v1/relations?limit=2", nil)
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
