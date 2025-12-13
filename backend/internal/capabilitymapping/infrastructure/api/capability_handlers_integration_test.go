//go:build integration
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

	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
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

type testContext struct {
	db         *sql.DB
	testID     string
	createdIDs []string
}

func (ctx *testContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func setupTestDB(t *testing.T) (*testContext, func()) {
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

	testID := fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())

	ctx := &testContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	cleanup := func() {
		for _, id := range ctx.createdIDs {
			db.Exec("DELETE FROM capabilities WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *testContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

func (ctx *testContext) createTestCapability(t *testing.T, id, name, level string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO capabilities (id, name, description, level, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		id, name, "", level, testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func setupHandlers(db *sql.DB) *CapabilityHandlers {
	tenantDB := database.NewTenantAwareDB(db)

	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewCapabilityReadModel(tenantDB)

	projector := projectors.NewCapabilityProjector(readModel)
	eventBus.Subscribe("CapabilityCreated", projector)
	eventBus.Subscribe("CapabilityUpdated", projector)
	eventBus.Subscribe("CapabilityMetadataUpdated", projector)
	eventBus.Subscribe("CapabilityExpertAdded", projector)
	eventBus.Subscribe("CapabilityTagAdded", projector)
	eventBus.Subscribe("CapabilityDeleted", projector)

	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	createHandler := handlers.NewCreateCapabilityHandler(capabilityRepo)
	updateHandler := handlers.NewUpdateCapabilityHandler(capabilityRepo)
	updateMetadataHandler := handlers.NewUpdateCapabilityMetadataHandler(capabilityRepo)
	addExpertHandler := handlers.NewAddCapabilityExpertHandler(capabilityRepo)
	addTagHandler := handlers.NewAddCapabilityTagHandler(capabilityRepo)
	deleteHandler := handlers.NewDeleteCapabilityHandler(capabilityRepo, readModel)

	commandBus.Register("CreateCapability", createHandler)
	commandBus.Register("UpdateCapability", updateHandler)
	commandBus.Register("UpdateCapabilityMetadata", updateMetadataHandler)
	commandBus.Register("AddCapabilityExpert", addExpertHandler)
	commandBus.Register("AddCapabilityTag", addTagHandler)
	commandBus.Register("DeleteCapability", deleteHandler)

	return NewCapabilityHandlers(commandBus, readModel, hateoas)
}

func TestCreateCapability_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	reqBody := CreateCapabilityRequest{
		Name:        "Customer Management",
		Description: "Manage customer data and relationships",
		Level:       "L1",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.CreateCapability(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	testCtx.setTenantContext(t)
	var aggregateID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&aggregateID)
	require.NoError(t, err)
	testCtx.trackID(aggregateID)

	var eventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'CapabilityCreated'",
		aggregateID,
	).Scan(&eventData)
	require.NoError(t, err)
	assert.Contains(t, eventData, "Customer Management")

	time.Sleep(100 * time.Millisecond)

	capability, err := handlers.readModel.GetByID(tenantContext(), aggregateID)
	require.NoError(t, err)
	assert.Equal(t, "Customer Management", capability.Name)
	assert.Equal(t, "L1", capability.Level)
}

func TestGetAllCapabilities_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	id1 := fmt.Sprintf("test-cap-1-%d", time.Now().UnixNano())
	id2 := fmt.Sprintf("test-cap-2-%d", time.Now().UnixNano())
	testCtx.createTestCapability(t, id1, "Sales Management", "L1")
	testCtx.createTestCapability(t, id2, "Order Processing", "L2")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.GetAllCapabilities(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.CapabilityDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	foundCapabilities := 0
	for _, cap := range response.Data {
		if cap.ID == id1 || cap.ID == id2 {
			foundCapabilities++
			assert.NotNil(t, cap.Links)
			assert.Contains(t, cap.Links, "self")
		}
	}
	assert.Equal(t, 2, foundCapabilities)
}

func TestGetCapabilityByID_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	capabilityID := fmt.Sprintf("test-capability-%d", time.Now().UnixNano())
	testCtx.createTestCapability(t, capabilityID, "Payment Processing", "L2")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/"+capabilityID, nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", capabilityID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetCapabilityByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.CapabilityDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, capabilityID, response.ID)
	assert.Equal(t, "Payment Processing", response.Name)
	assert.NotNil(t, response.Links)
}

func TestGetCapabilityByID_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/"+nonExistentID, nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetCapabilityByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateCapability_ValidationError_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	reqBody := CreateCapabilityRequest{
		Name:  "",
		Level: "L1",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	handlers.CreateCapability(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var count int
	err := testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM events WHERE created_at > NOW() - INTERVAL '5 seconds'",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestUpdateCapability_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	createReqBody := CreateCapabilityRequest{
		Name:        "Original Name",
		Description: "Original description",
		Level:       "L1",
	}
	createBody, _ := json.Marshal(createReqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withTestTenant(createReq)
	createW := httptest.NewRecorder()

	handlers.CreateCapability(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var capabilityID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	updateReqBody := UpdateCapabilityRequest{
		Name:        "Updated Name",
		Description: "Updated description",
	}
	updateBody, _ := json.Marshal(updateReqBody)

	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/capabilities/"+capabilityID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq = withTestTenant(updateReq)
	updateReq = updateReq.WithContext(context.WithValue(updateReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{capabilityID},
		},
	}))
	updateW := httptest.NewRecorder()

	handlers.UpdateCapability(updateW, updateReq)

	assert.Equal(t, http.StatusOK, updateW.Code)

	testCtx.setTenantContext(t)
	var updateEventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'CapabilityUpdated'",
		capabilityID,
	).Scan(&updateEventData)
	require.NoError(t, err)
	assert.Contains(t, updateEventData, "Updated Name")
	assert.Contains(t, updateEventData, "Updated description")

	time.Sleep(100 * time.Millisecond)

	var name, description string
	err = testCtx.db.QueryRow(
		"SELECT name, description FROM capabilities WHERE id = $1",
		capabilityID,
	).Scan(&name, &description)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", name)
	assert.Equal(t, "Updated description", description)
}

func TestGetCapabilityChildren_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	parentID := fmt.Sprintf("test-parent-%d", time.Now().UnixNano())
	childID1 := fmt.Sprintf("test-child-1-%d", time.Now().UnixNano())
	childID2 := fmt.Sprintf("test-child-2-%d", time.Now().UnixNano())

	testCtx.createTestCapability(t, parentID, "Parent Capability", "L1")

	testCtx.setTenantContext(t)
	_, err := testCtx.db.Exec(
		"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		childID1, "Child 1", "", "L2", parentID, testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)
	testCtx.trackID(childID1)

	_, err = testCtx.db.Exec(
		"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		childID2, "Child 2", "", "L2", parentID, testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)
	testCtx.trackID(childID2)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/"+parentID+"/children", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", parentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetCapabilityChildren(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.CapabilityDTO `json:"data"`
	}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Data))
}

func TestDeleteCapability_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	createReqBody := CreateCapabilityRequest{
		Name:        "Capability To Delete",
		Description: "This will be deleted",
		Level:       "L1",
	}
	createBody, _ := json.Marshal(createReqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withTestTenant(createReq)
	createW := httptest.NewRecorder()

	handlers.CreateCapability(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var capabilityID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	capability, err := handlers.readModel.GetByID(tenantContext(), capabilityID)
	require.NoError(t, err)
	assert.NotNil(t, capability)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/capabilities/"+capabilityID, nil)
	deleteReq = withTestTenant(deleteReq)
	deleteReq = deleteReq.WithContext(context.WithValue(deleteReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{capabilityID},
		},
	}))
	deleteW := httptest.NewRecorder()

	handlers.DeleteCapability(deleteW, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteW.Code)

	time.Sleep(100 * time.Millisecond)

	deletedCapability, err := handlers.readModel.GetByID(tenantContext(), capabilityID)
	require.NoError(t, err)
	assert.Nil(t, deletedCapability)

	testCtx.setTenantContext(t)
	var eventCount int
	err = testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM events WHERE aggregate_id = $1 AND event_type = 'CapabilityDeleted'",
		capabilityID,
	).Scan(&eventCount)
	require.NoError(t, err)
	assert.Equal(t, 1, eventCount)
}

func TestDeleteCapability_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/capabilities/"+nonExistentID, nil)
	deleteReq = withTestTenant(deleteReq)
	deleteReq = deleteReq.WithContext(context.WithValue(deleteReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{nonExistentID},
		},
	}))
	deleteW := httptest.NewRecorder()

	handlers.DeleteCapability(deleteW, deleteReq)

	assert.Equal(t, http.StatusNotFound, deleteW.Code)
}

func TestDeleteCapability_HasChildren_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	parentID := fmt.Sprintf("test-parent-%d", time.Now().UnixNano())
	childID := fmt.Sprintf("test-child-%d", time.Now().UnixNano())

	testCtx.createTestCapability(t, parentID, "Parent Capability", "L1")

	testCtx.setTenantContext(t)
	_, err := testCtx.db.Exec(
		"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		childID, "Child Capability", "", "L2", parentID, testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)
	testCtx.trackID(childID)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/capabilities/"+parentID, nil)
	deleteReq = withTestTenant(deleteReq)
	deleteReq = deleteReq.WithContext(context.WithValue(deleteReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{parentID},
		},
	}))
	deleteW := httptest.NewRecorder()

	handlers.DeleteCapability(deleteW, deleteReq)

	assert.Equal(t, http.StatusConflict, deleteW.Code)

	var response struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	err = json.NewDecoder(deleteW.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response.Message, "Cannot delete capability with children")

	parentCapability, err := handlers.readModel.GetByID(tenantContext(), parentID)
	require.NoError(t, err)
	assert.NotNil(t, parentCapability)
}
