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

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/application/handlers"
	"easi/backend/internal/metamodel/application/projectors"
	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/infrastructure/repositories"
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
			db.Exec("DELETE FROM meta_model_configurations WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *testContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

type testHandlers struct {
	handlers   *MetaModelHandlers
	commandBus *cqrs.InMemoryCommandBus
	readModel  *readmodels.MetaModelConfigurationReadModel
}

func setupHandlers(db *sql.DB) *testHandlers {
	tenantDB := database.NewTenantAwareDB(db)

	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewMetaModelConfigurationReadModel(tenantDB)

	projector := projectors.NewMetaModelConfigurationProjector(readModel)
	eventBus.Subscribe("MetaModelConfigurationCreated", projector)
	eventBus.Subscribe("MaturityScaleConfigUpdated", projector)
	eventBus.Subscribe("MaturityScaleConfigReset", projector)

	configRepo := repositories.NewMetaModelConfigurationRepository(eventStore)
	createHandler := handlers.NewCreateMetaModelConfigurationHandler(configRepo)
	updateHandler := handlers.NewUpdateMaturityScaleHandler(configRepo)
	resetHandler := handlers.NewResetMaturityScaleHandler(configRepo)

	commandBus.Register("CreateMetaModelConfiguration", createHandler)
	commandBus.Register("UpdateMaturityScale", updateHandler)
	commandBus.Register("ResetMaturityScale", resetHandler)

	return &testHandlers{
		handlers:   NewMetaModelHandlers(commandBus, readModel, hateoas),
		commandBus: commandBus,
		readModel:  readModel,
	}
}

func createTestConfiguration(t *testing.T, testCtx *testContext, h *testHandlers) string {
	cmd := &commands.CreateMetaModelConfiguration{
		TenantID:  testTenantID(),
		CreatedBy: "test@example.com",
	}

	err := h.commandBus.Dispatch(tenantContext(), cmd)
	require.NoError(t, err)

	testCtx.trackID(cmd.ID)
	time.Sleep(100 * time.Millisecond)

	return cmd.ID
}

func TestGetMaturityScale_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)
	configID := createTestConfiguration(t, testCtx, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metamodel/maturity-scale", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	h.handlers.GetMaturityScale(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, configID, response.ID)
	assert.Equal(t, testTenantID(), response.TenantID)
	assert.Len(t, response.Sections, 4)
	assert.Equal(t, 1, response.Version)

	assert.Equal(t, "Genesis", response.Sections[0].Name)
	assert.Equal(t, 0, response.Sections[0].MinValue)
	assert.Equal(t, 24, response.Sections[0].MaxValue)

	assert.Equal(t, "Custom Built", response.Sections[1].Name)
	assert.Equal(t, 25, response.Sections[1].MinValue)
	assert.Equal(t, 49, response.Sections[1].MaxValue)

	assert.Equal(t, "Product", response.Sections[2].Name)
	assert.Equal(t, 50, response.Sections[2].MinValue)
	assert.Equal(t, 74, response.Sections[2].MaxValue)

	assert.Equal(t, "Commodity", response.Sections[3].Name)
	assert.Equal(t, 75, response.Sections[3].MinValue)
	assert.Equal(t, 99, response.Sections[3].MaxValue)

	assert.NotNil(t, response.Links)
}

func TestGetMaturityScale_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metamodel/maturity-scale", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	h.handlers.GetMaturityScale(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateMaturityScale_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)
	configID := createTestConfiguration(t, testCtx, h)

	reqBody := UpdateMaturityScaleRequest{
		Sections: [4]MaturitySectionRequest{
			{Order: 1, Name: "Early Stage", MinValue: 0, MaxValue: 19},
			{Order: 2, Name: "Growth", MinValue: 20, MaxValue: 49},
			{Order: 3, Name: "Mature", MinValue: 50, MaxValue: 79},
			{Order: 4, Name: "Sunset", MinValue: 80, MaxValue: 99},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/metamodel/maturity-scale", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	h.handlers.UpdateMaturityScale(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, configID, response.ID)
	assert.Equal(t, 2, response.Version)
	assert.Equal(t, "Early Stage", response.Sections[0].Name)
	assert.Equal(t, "Growth", response.Sections[1].Name)
	assert.Equal(t, "Mature", response.Sections[2].Name)
	assert.Equal(t, "Sunset", response.Sections[3].Name)
}

func TestUpdateMaturityScale_InvalidSections_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)
	createTestConfiguration(t, testCtx, h)

	reqBody := UpdateMaturityScaleRequest{
		Sections: [4]MaturitySectionRequest{
			{Order: 1, Name: "Section 1", MinValue: 0, MaxValue: 30},
			{Order: 2, Name: "Section 2", MinValue: 31, MaxValue: 50},
			{Order: 3, Name: "Section 3", MinValue: 51, MaxValue: 80},
			{Order: 4, Name: "Section 4", MinValue: 81, MaxValue: 90},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/metamodel/maturity-scale", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	h.handlers.UpdateMaturityScale(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestResetMaturityScale_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)
	configID := createTestConfiguration(t, testCtx, h)

	updateReqBody := UpdateMaturityScaleRequest{
		Sections: [4]MaturitySectionRequest{
			{Order: 1, Name: "Custom 1", MinValue: 0, MaxValue: 24},
			{Order: 2, Name: "Custom 2", MinValue: 25, MaxValue: 49},
			{Order: 3, Name: "Custom 3", MinValue: 50, MaxValue: 74},
			{Order: 4, Name: "Custom 4", MinValue: 75, MaxValue: 99},
		},
	}
	updateBody, _ := json.Marshal(updateReqBody)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/metamodel/maturity-scale", bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq = withTestTenant(updateReq)
	updateW := httptest.NewRecorder()
	h.handlers.UpdateMaturityScale(updateW, updateReq)
	require.Equal(t, http.StatusOK, updateW.Code)

	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/metamodel/maturity-scale/reset", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	h.handlers.ResetMaturityScale(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, configID, response.ID)
	assert.Equal(t, 3, response.Version)

	assert.Equal(t, "Genesis", response.Sections[0].Name)
	assert.Equal(t, "Custom Built", response.Sections[1].Name)
	assert.Equal(t, "Product", response.Sections[2].Name)
	assert.Equal(t, "Commodity", response.Sections[3].Name)
}

func TestGetMaturityScaleByID_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)
	configID := createTestConfiguration(t, testCtx, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metamodel/configurations/"+configID, nil)
	req = withTestTenant(req)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", configID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handlers.GetMaturityScaleByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, configID, response.ID)
	assert.NotNil(t, response.Links)
}

func TestGetMaturityScaleByID_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metamodel/configurations/"+nonExistentID, nil)
	req = withTestTenant(req)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handlers.GetMaturityScaleByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateMaturityScale_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	reqBody := UpdateMaturityScaleRequest{
		Sections: [4]MaturitySectionRequest{
			{Order: 1, Name: "Section 1", MinValue: 0, MaxValue: 24},
			{Order: 2, Name: "Section 2", MinValue: 25, MaxValue: 49},
			{Order: 3, Name: "Section 3", MinValue: 50, MaxValue: 74},
			{Order: 4, Name: "Section 4", MinValue: 75, MaxValue: 99},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/metamodel/maturity-scale", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	h.handlers.UpdateMaturityScale(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestResetMaturityScale_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/metamodel/maturity-scale/reset", nil)
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	h.handlers.ResetMaturityScale(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
