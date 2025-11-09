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

	"easi/backend/internal/architectureviews/application/handlers"
	"easi/backend/internal/architectureviews/application/projectors"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// viewTestContext holds test-specific state for cleanup
type viewTestContext struct {
	db         *sql.DB
	testID     string
	createdIDs []string
}

func setupViewTestDB(t *testing.T) (*viewTestContext, func()) {
	// Connect to test database
	connStr := "host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	// Initialize event store schema (idempotent)
	eventStore := eventstore.NewPostgresEventStore(db)
	err = eventStore.InitializeSchema()
	require.NoError(t, err)

	// Initialize read model schema (idempotent)
	viewReadModel := readmodels.NewArchitectureViewReadModel(db)
	err = viewReadModel.InitializeSchema()
	require.NoError(t, err)

	// Create unique test ID based on test name and timestamp to avoid collisions
	testID := fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())

	ctx := &viewTestContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	// Clean up only the data created in this specific test
	cleanup := func() {
		// Delete views and component positions by tracking the IDs created during the test
		for _, id := range ctx.createdIDs {
			db.Exec("DELETE FROM view_component_positions WHERE view_id = $1", id)
			db.Exec("DELETE FROM architecture_views WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

// trackID adds an aggregate ID to the cleanup list
func (ctx *viewTestContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

// createTestView creates a view directly in the read model for testing
func (ctx *viewTestContext) createTestView(t *testing.T, id, name, description string) {
	_, err := ctx.db.Exec(
		"INSERT INTO architecture_views (id, name, description, created_at) VALUES ($1, $2, $3, NOW())",
		id, name, description,
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

// addTestComponentToView adds a component to a view in the read model for testing
func (ctx *viewTestContext) addTestComponentToView(t *testing.T, viewID, componentID string, x, y float64) {
	_, err := ctx.db.Exec(
		"INSERT INTO view_component_positions (view_id, component_id, x, y, created_at) VALUES ($1, $2, $3, $4, NOW())",
		viewID, componentID, x, y,
	)
	require.NoError(t, err)
}

func setupViewHandlers(db *sql.DB) (*ViewHandlers, *readmodels.ArchitectureViewReadModel) {
	// Setup event infrastructure
	eventStore := eventstore.NewPostgresEventStore(db)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	// Setup read model and projector
	readModel := readmodels.NewArchitectureViewReadModel(db)
	projector := projectors.NewArchitectureViewProjector(readModel)

	// Setup event bus and connect it to the event store
	eventBus := events.NewInMemoryEventBus()
	eventBus.SubscribeAll(projector)
	eventStore.SetEventBus(eventBus)

	// Setup repository and handlers
	viewRepo := repositories.NewArchitectureViewRepository(eventStore)
	createHandler := handlers.NewCreateViewHandler(viewRepo)
	addComponentHandler := handlers.NewAddComponentToViewHandler(viewRepo)
	updatePositionHandler := handlers.NewUpdateComponentPositionHandler(viewRepo)

	commandBus.Register("CreateView", createHandler)
	commandBus.Register("AddComponentToView", addComponentHandler)
	commandBus.Register("UpdateComponentPosition", updatePositionHandler)

	// Setup HTTP handlers
	viewHandlers := NewViewHandlers(commandBus, readModel, hateoas)

	return viewHandlers, readModel
}

func TestCreateView_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	handlers, readModel := setupViewHandlers(testCtx.db)

	// Create view via API
	reqBody := CreateViewRequest{
		Name:        "System Architecture",
		Description: "Overall system architecture view",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/views", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateView(w, req)

	// Assert HTTP response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get the created aggregate ID from the event store
	var aggregateID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ViewCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&aggregateID)
	require.NoError(t, err)
	testCtx.trackID(aggregateID)

	// Verify event data contains expected values
	var eventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'ViewCreated'",
		aggregateID,
	).Scan(&eventData)
	require.NoError(t, err)
	assert.Contains(t, eventData, "System Architecture")
	assert.Contains(t, eventData, "Overall system architecture view")

	// Verify read model contains the view (should be populated by projector)
	view, err := readModel.GetByID(context.Background(), aggregateID)
	require.NoError(t, err)
	assert.NotNil(t, view)
	assert.Equal(t, "System Architecture", view.Name)
	assert.Equal(t, "Overall system architecture view", view.Description)
}

func TestCreateView_ValidationError_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	handlers, _ := setupViewHandlers(testCtx.db)

	// Create view with empty name (should fail validation)
	reqBody := CreateViewRequest{
		Name:        "",
		Description: "Some description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/views", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateView(w, req)

	// Assert validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify no ViewCreated event was created
	var count int
	err := testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM events WHERE event_type = 'ViewCreated' AND created_at > NOW() - INTERVAL '2 seconds'",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "No ViewCreated events should be created for invalid request")
}

func TestGetAllViews_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	handlers, _ := setupViewHandlers(testCtx.db)

	// Create test data directly in read model with unique IDs
	id1 := fmt.Sprintf("view-1-%d", time.Now().UnixNano())
	id2 := fmt.Sprintf("view-2-%d", time.Now().UnixNano())
	testCtx.createTestView(t, id1, "View A", "Description A")
	testCtx.createTestView(t, id2, "View B", "Description B")

	// Test GET all
	req := httptest.NewRequest(http.MethodGet, "/api/v1/views", nil)
	w := httptest.NewRecorder()

	handlers.GetAllViews(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.ArchitectureViewDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Find our test views in the response
	foundViews := 0
	for _, view := range response.Data {
		if view.ID == id1 || view.ID == id2 {
			foundViews++
			assert.NotNil(t, view.Links)
			assert.Contains(t, view.Links, "self")
		}
	}
	assert.Equal(t, 2, foundViews, "Should find both test views")
}

func TestGetViewByID_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	handlers, _ := setupViewHandlers(testCtx.db)

	// Create test data with unique IDs
	viewID := fmt.Sprintf("test-view-%d", time.Now().UnixNano())
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	// Add some components to the view
	comp1 := fmt.Sprintf("comp-1-%d", time.Now().UnixNano())
	comp2 := fmt.Sprintf("comp-2-%d", time.Now().UnixNano())
	testCtx.addTestComponentToView(t, viewID, comp1, 100.0, 200.0)
	testCtx.addTestComponentToView(t, viewID, comp2, 300.0, 400.0)

	// Test GET by ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/views/"+viewID, nil)
	w := httptest.NewRecorder()

	// Add URL param using chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", viewID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetViewByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data readmodels.ArchitectureViewDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, viewID, response.Data.ID)
	assert.Equal(t, "Test View", response.Data.Name)
	assert.Equal(t, "Test Description", response.Data.Description)
	assert.Len(t, response.Data.Components, 2)
	assert.NotNil(t, response.Data.Links)
}

func TestGetViewByID_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	handlers, _ := setupViewHandlers(testCtx.db)

	// Test GET non-existent view with unique ID
	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/views/"+nonExistentID, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetViewByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAddComponentToView_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, _ := setupViewHandlers(testCtx.db)

	// First create a view via event sourcing
	createReqBody := CreateViewRequest{
		Name:        "Test View",
		Description: "Test Description",
	}
	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/views", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	viewHandlers.CreateView(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	// Get the view ID from event store
	var viewID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ViewCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&viewID)
	require.NoError(t, err)
	testCtx.trackID(viewID)

	// Add component to view with unique component ID
	componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
	addReqBody := AddComponentRequest{
		ComponentID: componentID,
		X:           150.5,
		Y:           250.5,
	}
	addBody, _ := json.Marshal(addReqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/views/"+viewID+"/components", bytes.NewReader(addBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add URL param using chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", viewID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	viewHandlers.AddComponentToView(w, req)

	// Assert HTTP response (should be 201 Created)
	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify event was created
	var count int
	err = testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM events WHERE aggregate_id = $1 AND event_type = 'ComponentAddedToView'",
		viewID,
	).Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1, "At least one ComponentAddedToView event should be created")
}

func TestUpdateComponentPosition_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, _ := setupViewHandlers(testCtx.db)

	// First create a view
	createReqBody := CreateViewRequest{
		Name:        "Test View",
		Description: "Test Description",
	}
	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/views", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	viewHandlers.CreateView(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	// Get the view ID
	var viewID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ViewCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&viewID)
	require.NoError(t, err)
	testCtx.trackID(viewID)

	// Add a component first with unique component ID
	componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
	addReqBody := AddComponentRequest{
		ComponentID: componentID,
		X:           100.0,
		Y:           200.0,
	}
	addBody, _ := json.Marshal(addReqBody)
	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/views/"+viewID+"/components", bytes.NewReader(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addW := httptest.NewRecorder()
	rctxAdd := chi.NewRouteContext()
	rctxAdd.URLParams.Add("id", viewID)
	addReq = addReq.WithContext(context.WithValue(addReq.Context(), chi.RouteCtxKey, rctxAdd))
	viewHandlers.AddComponentToView(addW, addReq)
	require.Equal(t, http.StatusCreated, addW.Code)

	// Now update the component position
	updateReqBody := UpdatePositionRequest{
		X: 300.0,
		Y: 400.0,
	}
	updateBody, _ := json.Marshal(updateReqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/views/"+viewID+"/components/"+componentID+"/position", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add URL params using chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", viewID)
	rctx.URLParams.Add("componentId", componentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	viewHandlers.UpdateComponentPosition(w, req)

	// Assert HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify event was created
	var count int
	err = testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM events WHERE aggregate_id = $1 AND event_type = 'ComponentPositionUpdated'",
		viewID,
	).Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1, "At least one ComponentPositionUpdated event should be created")
}

func TestAddComponentToView_ViewNotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, _ := setupViewHandlers(testCtx.db)

	// Try to add component to non-existent view with unique IDs
	nonExistentViewID := fmt.Sprintf("non-existent-view-%d", time.Now().UnixNano())
	componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())

	addReqBody := AddComponentRequest{
		ComponentID: componentID,
		X:           150.5,
		Y:           250.5,
	}
	addBody, _ := json.Marshal(addReqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/views/"+nonExistentViewID+"/components", bytes.NewReader(addBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentViewID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	viewHandlers.AddComponentToView(w, req)

	// Should return bad request or not found or internal server error (aggregate not found)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError,
		"Expected 400, 404, or 500 but got %d", w.Code)
}
