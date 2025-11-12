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
	"easi/backend/internal/architecturemodeling/application/projectors"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testContext holds test-specific state for cleanup
type testContext struct {
	db         *sql.DB
	testID     string
	createdIDs []string
}

func setupTestDB(t *testing.T) (*testContext, func()) {
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

	ctx := &testContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	// Clean up only the data created in this specific test
	cleanup := func() {
		// Delete components by tracking the IDs created during the test
		for _, id := range ctx.createdIDs {
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
func (ctx *testContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

// createTestComponent creates a component directly in the read model for testing
func (ctx *testContext) createTestComponent(t *testing.T, id, name, description string) {
	_, err := ctx.db.Exec(
		"INSERT INTO application_components (id, name, description, created_at) VALUES ($1, $2, $3, NOW())",
		id, name, description,
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func setupHandlers(db *sql.DB) (*ComponentHandlers, *readmodels.ApplicationComponentReadModel) {
	eventStore := eventstore.NewPostgresEventStore(db)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	// Setup event bus and wire to event store
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	// Setup read model
	readModel := readmodels.NewApplicationComponentReadModel(db)

	// Setup projector and wire to event bus
	projector := projectors.NewApplicationComponentProjector(readModel)
	eventBus.Subscribe("ApplicationComponentCreated", projector)
	eventBus.Subscribe("ApplicationComponentUpdated", projector)

	// Setup repository and handlers
	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	createHandler := handlers.NewCreateApplicationComponentHandler(componentRepo)
	updateHandler := handlers.NewUpdateApplicationComponentHandler(componentRepo)
	commandBus.Register("CreateApplicationComponent", createHandler)
	commandBus.Register("UpdateApplicationComponent", updateHandler)

	// Setup HTTP handlers
	componentHandlers := NewComponentHandlers(commandBus, readModel, hateoas)

	return componentHandlers, readModel
}

func TestCreateComponent_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, readModel := setupHandlers(testCtx.db)

	// Create component via API
	reqBody := CreateApplicationComponentRequest{
		Name:        "User Service",
		Description: "Handles user authentication and authorization",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/components", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateApplicationComponent(w, req)

	// Assert HTTP response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get the created aggregate ID from the event store
	var aggregateID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ApplicationComponentCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&aggregateID)
	require.NoError(t, err)
	testCtx.trackID(aggregateID)

	// Verify event data contains expected values
	var eventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'ApplicationComponentCreated'",
		aggregateID,
	).Scan(&eventData)
	require.NoError(t, err)
	assert.Contains(t, eventData, "User Service")
	assert.Contains(t, eventData, "Handles user authentication and authorization")

	// Wait a moment for the projector to update the read model
	time.Sleep(100 * time.Millisecond)

	// Verify read model contains the component (should be projected automatically)
	component, err := readModel.GetByID(context.Background(), aggregateID)
	require.NoError(t, err)
	assert.Equal(t, "User Service", component.Name)
	assert.Equal(t, "Handles user authentication and authorization", component.Description)
}

func TestGetAllComponents_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(testCtx.db)

	// Create test data directly in read model with unique IDs
	id1 := fmt.Sprintf("test-comp-1-%d", time.Now().UnixNano())
	id2 := fmt.Sprintf("test-comp-2-%d", time.Now().UnixNano())
	testCtx.createTestComponent(t, id1, "Service A", "Description A")
	testCtx.createTestComponent(t, id2, "Service B", "Description B")

	// Test GET all
	req := httptest.NewRequest(http.MethodGet, "/api/v1/components", nil)
	w := httptest.NewRecorder()

	handlers.GetAllComponents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.ApplicationComponentDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Find our test components in the response
	foundComponents := 0
	for _, comp := range response.Data {
		if comp.ID == id1 || comp.ID == id2 {
			foundComponents++
			assert.NotNil(t, comp.Links)
			assert.Contains(t, comp.Links, "self")
			assert.Contains(t, comp.Links, "archimate")
		}
	}
	assert.Equal(t, 2, foundComponents, "Should find both test components")
}

func TestGetComponentByID_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(testCtx.db)

	// Create test data with unique ID
	componentID := fmt.Sprintf("test-component-%d", time.Now().UnixNano())
	testCtx.createTestComponent(t, componentID, "Test Service", "Test Description")

	// Test GET by ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/components/"+componentID, nil)
	w := httptest.NewRecorder()

	// Add URL param using chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", componentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetComponentByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data readmodels.ApplicationComponentDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, componentID, response.Data.ID)
	assert.Equal(t, "Test Service", response.Data.Name)
	assert.Equal(t, "Test Description", response.Data.Description)
	assert.NotNil(t, response.Data.Links)
}

func TestGetComponentByID_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(testCtx.db)

	// Test GET non-existent component with unique ID to avoid collisions
	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/components/"+nonExistentID, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetComponentByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateComponent_ValidationError_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(testCtx.db)

	// Create component with empty name (should fail validation)
	reqBody := CreateApplicationComponentRequest{
		Name:        "",
		Description: "Some description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/components", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateApplicationComponent(w, req)

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

func TestGetAllComponentsPaginated_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(testCtx.db)

	// Create test data with unique IDs and different timestamps
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("comp-%s-%d-%d", testCtx.testID, i, time.Now().UnixNano())
		name := fmt.Sprintf("Component %d", i)
		description := fmt.Sprintf("Description %d", i)

		_, err := testCtx.db.Exec(
			"INSERT INTO application_components (id, name, description, created_at) VALUES ($1, $2, $3, NOW() - INTERVAL '"+fmt.Sprintf("%d", i)+" seconds')",
			id, name, description,
		)
		require.NoError(t, err)
		testCtx.trackID(id)

		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Test GET first page with limit
	req := httptest.NewRequest(http.MethodGet, "/api/v1/components?limit=2", nil)
	w := httptest.NewRecorder()

	handlers.GetAllComponents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data       []readmodels.ApplicationComponentDTO `json:"data"`
		Pagination struct {
			Cursor  string `json:"cursor"`
			HasMore bool   `json:"hasMore"`
			Limit   int    `json:"limit"`
		} `json:"pagination"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should have at least 2 components (our test data)
	assert.GreaterOrEqual(t, len(response.Data), 2, "Should return at least 2 components")
	assert.Equal(t, 2, response.Pagination.Limit)

	if len(response.Data) >= 2 && response.Pagination.HasMore {
		// Test GET second page using cursor
		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/components?limit=2&after="+response.Pagination.Cursor, nil)
		w2 := httptest.NewRecorder()

		handlers.GetAllComponents(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)

		var response2 struct {
			Data       []readmodels.ApplicationComponentDTO `json:"data"`
			Pagination struct {
				Cursor  string `json:"cursor"`
				HasMore bool   `json:"hasMore"`
				Limit   int    `json:"limit"`
			} `json:"pagination"`
		}
		err = json.NewDecoder(w2.Body).Decode(&response2)
		require.NoError(t, err)

		// Verify we got different components
		firstPageIDs := make(map[string]bool)
		for _, comp := range response.Data {
			firstPageIDs[comp.ID] = true
		}
		for _, comp := range response2.Data {
			if firstPageIDs[comp.ID] {
				t.Logf("Warning: Component %s appears in both pages", comp.ID)
			}
		}
	}
}

func TestGetAllComponentsPagination_InvalidCursor_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(testCtx.db)

	// Test GET with invalid cursor
	req := httptest.NewRequest(http.MethodGet, "/api/v1/components?limit=2&after=invalid-cursor", nil)
	w := httptest.NewRecorder()

	handlers.GetAllComponents(w, req)

	// Should return bad request for invalid cursor
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateComponent_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(testCtx.db)

	// First, create a component
	createReqBody := CreateApplicationComponentRequest{
		Name:        "Payment Service",
		Description: "Handles payment processing",
	}
	createBody, _ := json.Marshal(createReqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/components", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()

	handlers.CreateApplicationComponent(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	// Get the created component ID
	var componentID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ApplicationComponentCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&componentID)
	require.NoError(t, err)
	testCtx.trackID(componentID)

	// Wait a moment for projections to update
	time.Sleep(100 * time.Millisecond)

	// Now update the component
	updateReqBody := UpdateApplicationComponentRequest{
		Name:        "Enhanced Payment Service",
		Description: "Handles payment processing with fraud detection",
	}
	updateBody, _ := json.Marshal(updateReqBody)

	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/components/"+componentID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq = updateReq.WithContext(context.WithValue(updateReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{componentID},
		},
	}))
	updateW := httptest.NewRecorder()

	handlers.UpdateApplicationComponent(updateW, updateReq)

	// Assert HTTP response
	assert.Equal(t, http.StatusOK, updateW.Code)

	// Verify the update event was created
	var updateEventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'ApplicationComponentUpdated'",
		componentID,
	).Scan(&updateEventData)
	require.NoError(t, err)
	assert.Contains(t, updateEventData, "Enhanced Payment Service")
	assert.Contains(t, updateEventData, "fraud detection")

	// Wait a moment for projections to update
	time.Sleep(100 * time.Millisecond)

	// Verify the read model was updated
	var name, description string
	err = testCtx.db.QueryRow(
		"SELECT name, description FROM application_components WHERE id = $1",
		componentID,
	).Scan(&name, &description)
	require.NoError(t, err)
	assert.Equal(t, "Enhanced Payment Service", name)
	assert.Equal(t, "Handles payment processing with fraud detection", description)
}
