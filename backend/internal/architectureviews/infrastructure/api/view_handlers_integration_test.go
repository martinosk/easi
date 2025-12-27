//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/architectureviews/application/handlers"
	"easi/backend/internal/architectureviews/application/projectors"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
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

// setTenantContext sets the tenant context for RLS before running raw queries
func (ctx *viewTestContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func setupViewTestDB(t *testing.T) (*viewTestContext, func()) {
	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
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

	ctx := &viewTestContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	cleanup := func() {
		db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
		for _, id := range ctx.createdIDs {
			db.Exec("DELETE FROM view_element_positions WHERE view_id = $1", id)
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
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO architecture_views (id, name, description, tenant_id, created_at) VALUES ($1, $2, $3, $4, NOW())",
		id, name, description, testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

// addTestComponentToView adds a component to a view in the read model for testing
func (ctx *viewTestContext) addTestComponentToView(t *testing.T, viewID, componentID string, x, y float64) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO view_element_positions (view_id, element_id, element_type, x, y, tenant_id, created_at) VALUES ($1, $2, 'component', $3, $4, $5, NOW())",
		viewID, componentID, x, y, testTenantID(),
	)
	require.NoError(t, err)
}

func (ctx *viewTestContext) makeRequest(t *testing.T, method, url string, body []byte, urlParams map[string]string) (*httptest.ResponseRecorder, *http.Request) {
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

func (ctx *viewTestContext) createViewViaAPI(t *testing.T, handlers *ViewHandlers, name, description string) string {
	reqBody := CreateViewRequest{
		Name:        name,
		Description: description,
	}
	body, _ := json.Marshal(reqBody)

	w, req := ctx.makeRequest(t, http.MethodPost, "/api/v1/views", body, nil)
	handlers.CreateView(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	ctx.setTenantContext(t)
	var viewID string
	err := ctx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'ViewCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&viewID)
	require.NoError(t, err)
	ctx.trackID(viewID)

	return viewID
}

func (ctx *viewTestContext) addComponentViaAPI(t *testing.T, handlers *ViewHandlers, viewID, componentID string, x, y float64) {
	reqBody := AddComponentRequest{
		ComponentID: componentID,
		X:           x,
		Y:           y,
	}
	body, _ := json.Marshal(reqBody)

	w, req := ctx.makeRequest(t, http.MethodPost, "/api/v1/views/"+viewID+"/components", body, map[string]string{"id": viewID})
	handlers.AddComponentToView(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func setupViewHandlers(db *sql.DB) (*ViewHandlers, *readmodels.ArchitectureViewReadModel) {
	// Wrap database connection with tenant-aware wrapper for RLS
	tenantDB := database.NewTenantAwareDB(db)

	// Setup event infrastructure
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	// Setup read model and projector
	readModel := readmodels.NewArchitectureViewReadModel(tenantDB)
	projector := projectors.NewArchitectureViewProjector(readModel)

	// Setup event bus and connect it to the event store
	eventBus := events.NewInMemoryEventBus()
	eventBus.SubscribeAll(projector)
	eventStore.SetEventBus(eventBus)

	// Setup repository and handlers
	viewRepo := repositories.NewArchitectureViewRepository(eventStore)
	layoutRepo := repositories.NewViewLayoutRepository(tenantDB)
	createHandler := handlers.NewCreateViewHandler(viewRepo, readModel)
	addComponentHandler := handlers.NewAddComponentToViewHandler(viewRepo, layoutRepo)
	updatePositionHandler := handlers.NewUpdateComponentPositionHandler(layoutRepo)
	renameHandler := handlers.NewRenameViewHandler(viewRepo)
	deleteHandler := handlers.NewDeleteViewHandler(viewRepo)
	removeComponentHandler := handlers.NewRemoveComponentFromViewHandler(viewRepo)
	setDefaultHandler := handlers.NewSetDefaultViewHandler(viewRepo, readModel)
	updateEdgeTypeHandler := handlers.NewUpdateViewEdgeTypeHandler(layoutRepo)
	updateLayoutDirectionHandler := handlers.NewUpdateViewLayoutDirectionHandler(layoutRepo)
	updateColorSchemeHandler := handlers.NewUpdateViewColorSchemeHandler(layoutRepo)
	updateElementColorHandler := handlers.NewUpdateElementColorHandler(layoutRepo)
	clearElementColorHandler := handlers.NewClearElementColorHandler(layoutRepo)

	commandBus.Register("CreateView", createHandler)
	commandBus.Register("AddComponentToView", addComponentHandler)
	commandBus.Register("UpdateComponentPosition", updatePositionHandler)
	commandBus.Register("RenameView", renameHandler)
	commandBus.Register("DeleteView", deleteHandler)
	commandBus.Register("RemoveComponentFromView", removeComponentHandler)
	commandBus.Register("SetDefaultView", setDefaultHandler)
	commandBus.Register("UpdateViewEdgeType", updateEdgeTypeHandler)
	commandBus.Register("UpdateViewLayoutDirection", updateLayoutDirectionHandler)
	commandBus.Register("UpdateViewColorScheme", updateColorSchemeHandler)
	commandBus.Register("UpdateElementColor", updateElementColorHandler)
	commandBus.Register("ClearElementColor", clearElementColorHandler)

	// Setup HTTP handlers
	viewHandlers := NewViewHandlers(commandBus, readModel, layoutRepo, hateoas)

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
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.CreateView(w, req)

	// Assert HTTP response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get the created aggregate ID from the event store
	testCtx.setTenantContext(t)
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
	view, err := readModel.GetByID(tenantContext(), aggregateID)
	require.NoError(t, err)
	assert.NotNil(t, view)
	assert.Equal(t, "System Architecture", view.Name)
	assert.Equal(t, "Overall system architecture view", view.Description)
}

func TestCreateView_ValidationErrors_Integration(t *testing.T) {
	testCases := []struct {
		name        string
		viewName    string
		description string
	}{
		{"EmptyName", "", "Some description"},
		{"NameTooLong", "This is a view name that is one hundred and one characters long and should fail the validation tests!", "Some description"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			handlers, _ := setupViewHandlers(testCtx.db)

			reqBody := CreateViewRequest{
				Name:        tc.viewName,
				Description: tc.description,
			}
			body, _ := json.Marshal(reqBody)

			w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/views", body, nil)
			handlers.CreateView(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var count int
			err := testCtx.db.QueryRow(
				"SELECT COUNT(*) FROM events WHERE event_type = 'ViewCreated' AND created_at > NOW() - INTERVAL '2 seconds'",
			).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 0, count, "No ViewCreated events should be created for invalid request")
		})
	}
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
	req = withTestTenant(req)
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

	viewID := fmt.Sprintf("test-view-%d", time.Now().UnixNano())
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	comp1 := fmt.Sprintf("comp-1-%d", time.Now().UnixNano())
	comp2 := fmt.Sprintf("comp-2-%d", time.Now().UnixNano())
	testCtx.addTestComponentToView(t, viewID, comp1, 100.0, 200.0)
	testCtx.addTestComponentToView(t, viewID, comp2, 300.0, 400.0)

	w, req := testCtx.makeRequest(t, http.MethodGet, "/api/v1/views/"+viewID, nil, map[string]string{"id": viewID})
	handlers.GetViewByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.ArchitectureViewDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, viewID, response.ID)
	assert.Equal(t, "Test View", response.Name)
	assert.Equal(t, "Test Description", response.Description)
	assert.Len(t, response.Components, 2)
	assert.NotNil(t, response.Links)
}

func TestGetViewByID_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	handlers, _ := setupViewHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	w, req := testCtx.makeRequest(t, http.MethodGet, "/api/v1/views/"+nonExistentID, nil, map[string]string{"id": nonExistentID})
	handlers.GetViewByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAddComponentToView_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, _ := setupViewHandlers(testCtx.db)

	viewID := testCtx.createViewViaAPI(t, viewHandlers, "Test View", "Test Description")

	componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
	reqBody := AddComponentRequest{
		ComponentID: componentID,
		X:           150.5,
		Y:           250.5,
	}
	body, _ := json.Marshal(reqBody)

	w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/views/"+viewID+"/components", body, map[string]string{"id": viewID})
	viewHandlers.AddComponentToView(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var count int
	err := testCtx.db.QueryRow(
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

	viewID := testCtx.createViewViaAPI(t, viewHandlers, "Test View", "Test Description")

	componentID := fmt.Sprintf("comp-%d", time.Now().UnixNano())
	testCtx.addComponentViaAPI(t, viewHandlers, viewID, componentID, 100.0, 200.0)

	updateReqBody := UpdatePositionRequest{
		X: 300.0,
		Y: 400.0,
	}
	updateBody, _ := json.Marshal(updateReqBody)

	w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/components/"+componentID+"/position", updateBody, map[string]string{
		"id":          viewID,
		"componentId": componentID,
	})
	viewHandlers.UpdateComponentPosition(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	testCtx.setTenantContext(t)
	var x, y float64
	err := testCtx.db.QueryRow(
		"SELECT x, y FROM view_element_positions WHERE view_id = $1 AND element_id = $2 AND element_type = 'component'",
		viewID, componentID,
	).Scan(&x, &y)
	require.NoError(t, err)
	assert.Equal(t, 300.0, x, "X position should be updated to 300.0")
	assert.Equal(t, 400.0, y, "Y position should be updated to 400.0")
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
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentViewID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	viewHandlers.AddComponentToView(w, req)

	// Should return bad request or not found or internal server error (aggregate not found)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError,
		"Expected 400, 404, or 500 but got %d", w.Code)
}

func TestSetDefaultView_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, readModel := setupViewHandlers(testCtx.db)

	view1ID := testCtx.createViewViaAPI(t, viewHandlers, "View 1", "First view")
	view2ID := testCtx.createViewViaAPI(t, viewHandlers, "View 2", "Second view")

	w1, req1 := testCtx.makeRequest(t, http.MethodPut, "/api/v1/views/"+view1ID+"/default", nil, map[string]string{"id": view1ID})
	viewHandlers.SetDefaultView(w1, req1)
	require.Equal(t, http.StatusNoContent, w1.Code)

	defaultView, err := readModel.GetDefaultView(tenantContext())
	require.NoError(t, err)
	assert.NotNil(t, defaultView)
	assert.Equal(t, view1ID, defaultView.ID)

	w2, req2 := testCtx.makeRequest(t, http.MethodPut, "/api/v1/views/"+view2ID+"/default", nil, map[string]string{"id": view2ID})
	viewHandlers.SetDefaultView(w2, req2)

	assert.Equal(t, http.StatusNoContent, w2.Code)

	defaultView, err = readModel.GetDefaultView(tenantContext())
	require.NoError(t, err)
	assert.NotNil(t, defaultView)
	assert.Equal(t, view2ID, defaultView.ID)

	view1, err := readModel.GetByID(tenantContext(), view1ID)
	require.NoError(t, err)
	assert.False(t, view1.IsDefault)
}

func TestUpdateEdgeType_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, readModel := setupViewHandlers(testCtx.db)

	viewID := "view-edge-test-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	reqBody := UpdateEdgeTypeRequest{EdgeType: "step"}
	body, _ := json.Marshal(reqBody)

	w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/edge-type", body, map[string]string{"id": viewID})
	viewHandlers.UpdateEdgeType(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	view, err := readModel.GetByID(tenantContext(), viewID)
	require.NoError(t, err)
	assert.Equal(t, "step", view.EdgeType)
}

func TestUpdateEdgeType_InvalidValue_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, _ := setupViewHandlers(testCtx.db)

	viewID := "view-edge-invalid-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	reqBody := UpdateEdgeTypeRequest{EdgeType: "invalid"}
	body, _ := json.Marshal(reqBody)

	w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/edge-type", body, map[string]string{"id": viewID})
	viewHandlers.UpdateEdgeType(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateLayoutDirection_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, readModel := setupViewHandlers(testCtx.db)

	viewID := "view-layout-test-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	reqBody := UpdateLayoutDirectionRequest{LayoutDirection: "LR"}
	body, _ := json.Marshal(reqBody)

	w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/layout-direction", body, map[string]string{"id": viewID})
	viewHandlers.UpdateLayoutDirection(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	view, err := readModel.GetByID(tenantContext(), viewID)
	require.NoError(t, err)
	assert.Equal(t, "LR", view.LayoutDirection)
}

func TestUpdateLayoutDirection_InvalidValue_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, _ := setupViewHandlers(testCtx.db)

	viewID := "view-layout-invalid-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	reqBody := UpdateLayoutDirectionRequest{LayoutDirection: "INVALID"}
	body, _ := json.Marshal(reqBody)

	w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/layout-direction", body, map[string]string{"id": viewID})
	viewHandlers.UpdateLayoutDirection(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateMultipleEdgeTypesAndDirections_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, readModel := setupViewHandlers(testCtx.db)

	viewID := "view-multiple-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	edgeTypes := []string{"default", "step", "smoothstep", "straight"}
	for _, edgeType := range edgeTypes {
		reqBody := UpdateEdgeTypeRequest{EdgeType: edgeType}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/views/"+viewID+"/edge-type", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = withTestTenant(req)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", viewID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		viewHandlers.UpdateEdgeType(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		view, err := readModel.GetByID(tenantContext(), viewID)
		require.NoError(t, err)
		assert.Equal(t, edgeType, view.EdgeType)
	}

	directions := []string{"TB", "LR", "BT", "RL"}
	for _, direction := range directions {
		reqBody := UpdateLayoutDirectionRequest{LayoutDirection: direction}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/views/"+viewID+"/layout-direction", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = withTestTenant(req)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", viewID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		viewHandlers.UpdateLayoutDirection(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		view, err := readModel.GetByID(tenantContext(), viewID)
		require.NoError(t, err)
		assert.Equal(t, direction, view.LayoutDirection)
	}
}

func TestUpdateColorScheme_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, readModel := setupViewHandlers(testCtx.db)

	viewID := "view-color-scheme-test-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	reqBody := UpdateColorSchemeRequest{
		ColorScheme: "classic",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/views/"+viewID+"/color-scheme", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", viewID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	viewHandlers.UpdateColorScheme(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		ColorScheme string            `json:"colorScheme"`
		Links       map[string]string `json:"_links"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "classic", response.ColorScheme)
	assert.NotNil(t, response.Links)
	assert.Contains(t, response.Links, "self")
	assert.Equal(t, "/api/v1/views/"+viewID+"/color-scheme", response.Links["self"])
	assert.Contains(t, response.Links, "view")
	assert.Equal(t, "/api/v1/views/"+viewID, response.Links["view"])

	view, err := readModel.GetByID(tenantContext(), viewID)
	require.NoError(t, err)
	assert.Equal(t, "classic", view.ColorScheme)
}

func TestUpdateColorScheme_InvalidValue_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, _ := setupViewHandlers(testCtx.db)

	viewID := "view-color-scheme-invalid-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	reqBody := UpdateColorSchemeRequest{
		ColorScheme: "invalid-scheme",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/views/"+viewID+"/color-scheme", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", viewID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	viewHandlers.UpdateColorScheme(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateColorScheme_AllValidValues_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	viewHandlers, readModel := setupViewHandlers(testCtx.db)

	viewID := "view-color-scheme-all-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	colorSchemes := []string{"maturity", "classic", "custom"}
	for _, colorScheme := range colorSchemes {
		reqBody := UpdateColorSchemeRequest{ColorScheme: colorScheme}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/views/"+viewID+"/color-scheme", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = withTestTenant(req)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", viewID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		viewHandlers.UpdateColorScheme(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		view, err := readModel.GetByID(tenantContext(), viewID)
		require.NoError(t, err)
		assert.Equal(t, colorScheme, view.ColorScheme)
	}
}

func TestGetViewByID_ReturnsColorScheme_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	handlers, _ := setupViewHandlers(testCtx.db)

	viewID := "view-with-color-scheme-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/views/"+viewID, nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", viewID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetViewByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.ArchitectureViewDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, viewID, response.ID)
	assert.NotNil(t, response.ColorScheme)
}

func TestUpdateElementColor_Integration(t *testing.T) {
	testCases := []struct {
		elementType string
		color       string
		urlPath     string
		urlParam    string
		handler     func(http.ResponseWriter, *http.Request)
	}{
		{"component", "#FF5733", "components", "componentId", nil},
		{"capability", "#00FF00", "capabilities", "capabilityId", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.elementType, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			handlers, _ := setupViewHandlers(testCtx.db)

			viewID := "view-" + tc.elementType + "-color-" + testCtx.testID
			testCtx.createTestView(t, viewID, "Test View", "Test Description")

			elementID := tc.elementType[:3] + "-" + testCtx.testID
			if tc.elementType == "component" {
				testCtx.addTestComponentToView(t, viewID, elementID, 100.0, 200.0)
			} else {
				testCtx.setTenantContext(t)
				_, err := testCtx.db.Exec(
					"INSERT INTO view_element_positions (view_id, element_id, element_type, x, y, tenant_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())",
					viewID, elementID, tc.elementType, 150.0, 250.0, testTenantID(),
				)
				require.NoError(t, err)
			}

			reqBody := UpdateElementColorRequest{Color: tc.color}
			body, _ := json.Marshal(reqBody)

			w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/"+tc.urlPath+"/"+elementID+"/color", body, map[string]string{
				"id":        viewID,
				tc.urlParam: elementID,
			})

			if tc.elementType == "component" {
				handlers.UpdateComponentColor(w, req)
			} else {
				handlers.UpdateCapabilityColor(w, req)
			}

			assert.Equal(t, http.StatusNoContent, w.Code)

			testCtx.setTenantContext(t)
			var customColor sql.NullString
			err := testCtx.db.QueryRow(
				"SELECT custom_color FROM view_element_positions WHERE view_id = $1 AND element_id = $2 AND element_type = $3",
				viewID, elementID, tc.elementType,
			).Scan(&customColor)
			require.NoError(t, err)
			assert.True(t, customColor.Valid)
			assert.Equal(t, tc.color, customColor.String)
		})
	}
}

func TestUpdateComponentColor_InvalidValues_Integration(t *testing.T) {
	testCases := []struct {
		name  string
		color string
	}{
		{"InvalidHexColor", "invalid-color"},
		{"MissingHash", "FF5733"},
		{"TooShort", "#FFF"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			handlers, _ := setupViewHandlers(testCtx.db)

			viewID := "view-comp-" + testCtx.testID
			testCtx.createTestView(t, viewID, "Test View", "Test Description")

			componentID := "comp-" + testCtx.testID
			testCtx.addTestComponentToView(t, viewID, componentID, 100.0, 200.0)

			reqBody := UpdateElementColorRequest{Color: tc.color}
			body, _ := json.Marshal(reqBody)

			w, req := testCtx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/components/"+componentID+"/color", body, map[string]string{
				"id":          viewID,
				"componentId": componentID,
			})
			handlers.UpdateComponentColor(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestClearElementColor_Integration(t *testing.T) {
	testCases := []struct {
		elementType string
		color       string
		urlPath     string
		urlParam    string
	}{
		{"component", "#FF5733", "components", "componentId"},
		{"capability", "#00FF00", "capabilities", "capabilityId"},
	}

	for _, tc := range testCases {
		t.Run(tc.elementType, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			handlers, _ := setupViewHandlers(testCtx.db)

			viewID := "view-clear-" + tc.elementType + "-" + testCtx.testID
			testCtx.createTestView(t, viewID, "Test View", "Test Description")

			elementID := tc.elementType[:3] + "-" + testCtx.testID
			testCtx.setTenantContext(t)
			x, y := 100.0, 200.0
			if tc.elementType == "capability" {
				x, y = 150.0, 250.0
			}
			_, err := testCtx.db.Exec(
				"INSERT INTO view_element_positions (view_id, element_id, element_type, x, y, custom_color, tenant_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
				viewID, elementID, tc.elementType, x, y, tc.color, testTenantID(),
			)
			require.NoError(t, err)

			w, req := testCtx.makeRequest(t, http.MethodDelete, "/api/v1/views/"+viewID+"/"+tc.urlPath+"/"+elementID+"/color", nil, map[string]string{
				"id":        viewID,
				tc.urlParam: elementID,
			})

			if tc.elementType == "component" {
				handlers.ClearComponentColor(w, req)
			} else {
				handlers.ClearCapabilityColor(w, req)
			}

			assert.Equal(t, http.StatusNoContent, w.Code)

			testCtx.setTenantContext(t)
			var customColor sql.NullString
			err = testCtx.db.QueryRow(
				"SELECT custom_color FROM view_element_positions WHERE view_id = $1 AND element_id = $2 AND element_type = $3",
				viewID, elementID, tc.elementType,
			).Scan(&customColor)
			require.NoError(t, err)
			assert.False(t, customColor.Valid)
		})
	}
}

func TestGetViewByID_ReturnsCustomColorForElements_Integration(t *testing.T) {
	testCases := []struct {
		elementType   string
		color         string
		checkResponse func(*testing.T, readmodels.ArchitectureViewDTO, string, string)
	}{
		{
			"component",
			"#FF5733",
			func(t *testing.T, response readmodels.ArchitectureViewDTO, elem1, elem2 string) {
				assert.Len(t, response.Components, 2)
				var elem1Found, elem2Found bool
				for _, comp := range response.Components {
					if comp.ComponentID == elem1 {
						elem1Found = true
						assert.NotNil(t, comp.CustomColor)
						assert.Equal(t, "#FF5733", *comp.CustomColor)
					}
					if comp.ComponentID == elem2 {
						elem2Found = true
						assert.Nil(t, comp.CustomColor)
					}
				}
				assert.True(t, elem1Found, "Element 1 should be in response")
				assert.True(t, elem2Found, "Element 2 should be in response")
			},
		},
		{
			"capability",
			"#00FF00",
			func(t *testing.T, response readmodels.ArchitectureViewDTO, elem1, elem2 string) {
				assert.Len(t, response.Capabilities, 2)
				var elem1Found, elem2Found bool
				for _, cap := range response.Capabilities {
					if cap.CapabilityID == elem1 {
						elem1Found = true
						assert.NotNil(t, cap.CustomColor)
						assert.Equal(t, "#00FF00", *cap.CustomColor)
					}
					if cap.CapabilityID == elem2 {
						elem2Found = true
						assert.Nil(t, cap.CustomColor)
					}
				}
				assert.True(t, elem1Found, "Element 1 should be in response")
				assert.True(t, elem2Found, "Element 2 should be in response")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.elementType, func(t *testing.T) {
			testCtx, cleanup := setupViewTestDB(t)
			defer cleanup()

			handlers, _ := setupViewHandlers(testCtx.db)

			viewID := "view-with-" + tc.elementType + "-colors-" + testCtx.testID
			testCtx.createTestView(t, viewID, "Test View", "Test Description")

			elem1 := tc.elementType[:3] + "-1-" + testCtx.testID
			elem2 := tc.elementType[:3] + "-2-" + testCtx.testID
			testCtx.setTenantContext(t)
			_, err := testCtx.db.Exec(
				"INSERT INTO view_element_positions (view_id, element_id, element_type, x, y, custom_color, tenant_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
				viewID, elem1, tc.elementType, 100.0, 200.0, tc.color, testTenantID(),
			)
			require.NoError(t, err)
			_, err = testCtx.db.Exec(
				"INSERT INTO view_element_positions (view_id, element_id, element_type, x, y, tenant_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())",
				viewID, elem2, tc.elementType, 300.0, 400.0, testTenantID(),
			)
			require.NoError(t, err)

			w, req := testCtx.makeRequest(t, http.MethodGet, "/api/v1/views/"+viewID, nil, map[string]string{"id": viewID})
			handlers.GetViewByID(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response readmodels.ArchitectureViewDTO
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, viewID, response.ID)
			tc.checkResponse(t, response, elem1, elem2)
		})
	}
}

func TestGetViewByID_ReturnsHATEOASLinksForColors_Integration(t *testing.T) {
	testCtx, cleanup := setupViewTestDB(t)
	defer cleanup()

	handlers, _ := setupViewHandlers(testCtx.db)

	viewID := "view-with-links-" + testCtx.testID
	testCtx.createTestView(t, viewID, "Test View", "Test Description")

	componentID := "comp-" + testCtx.testID
	testCtx.addTestComponentToView(t, viewID, componentID, 100.0, 200.0)

	capabilityID := "cap-" + testCtx.testID
	testCtx.setTenantContext(t)
	_, err := testCtx.db.Exec(
		"INSERT INTO view_element_positions (view_id, element_id, element_type, x, y, tenant_id, created_at) VALUES ($1, $2, 'capability', $3, $4, $5, NOW())",
		viewID, capabilityID, 150.0, 250.0, testTenantID(),
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/views/"+viewID, nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", viewID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetViewByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.ArchitectureViewDTO
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Components, 1)
	compLinks := response.Components[0].Links
	assert.NotNil(t, compLinks)
	assert.Contains(t, compLinks, "updateColor")
	assert.Contains(t, compLinks, "clearColor")
	assert.Equal(t, "/api/v1/views/"+viewID+"/components/"+componentID+"/color", compLinks["updateColor"])
	assert.Equal(t, "/api/v1/views/"+viewID+"/components/"+componentID+"/color", compLinks["clearColor"])

	assert.Len(t, response.Capabilities, 1)
	capLinks := response.Capabilities[0].Links
	assert.NotNil(t, capLinks)
	assert.Contains(t, capLinks, "updateColor")
	assert.Contains(t, capLinks, "clearColor")
	assert.Equal(t, "/api/v1/views/"+viewID+"/capabilities/"+capabilityID+"/color", capLinks["updateColor"])
	assert.Equal(t, "/api/v1/views/"+viewID+"/capabilities/"+capabilityID+"/color", capLinks["clearColor"])
}
