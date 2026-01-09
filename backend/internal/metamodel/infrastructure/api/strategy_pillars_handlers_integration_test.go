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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type strategyPillarsTestHandlers struct {
	handlers                *MetaModelHandlers
	strategyPillarsHandlers *StrategyPillarsHandlers
	commandBus              *cqrs.InMemoryCommandBus
	readModel               *readmodels.MetaModelConfigurationReadModel
	sessionManager          *testSessionManager
	tenantID                string
}

func setupStrategyPillarsTestHandlers(db *sql.DB, tenantID string) *strategyPillarsTestHandlers {
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
	eventBus.Subscribe("StrategyPillarAdded", projector)
	eventBus.Subscribe("StrategyPillarUpdated", projector)
	eventBus.Subscribe("StrategyPillarRemoved", projector)

	configRepo := repositories.NewMetaModelConfigurationRepository(eventStore)

	createHandler := handlers.NewCreateMetaModelConfigurationHandler(configRepo)
	updateHandler := handlers.NewUpdateMaturityScaleHandler(configRepo)
	resetHandler := handlers.NewResetMaturityScaleHandler(configRepo)
	addPillarHandler := handlers.NewAddStrategyPillarHandler(configRepo)
	updatePillarHandler := handlers.NewUpdateStrategyPillarHandler(configRepo)
	removePillarHandler := handlers.NewRemoveStrategyPillarHandler(configRepo)
	batchUpdatePillarsHandler := handlers.NewBatchUpdateStrategyPillarsHandler(configRepo)

	commandBus.Register("CreateMetaModelConfiguration", createHandler)
	commandBus.Register("UpdateMaturityScale", updateHandler)
	commandBus.Register("ResetMaturityScale", resetHandler)
	commandBus.Register("AddStrategyPillar", addPillarHandler)
	commandBus.Register("UpdateStrategyPillar", updatePillarHandler)
	commandBus.Register("RemoveStrategyPillar", removePillarHandler)
	commandBus.Register("BatchUpdateStrategyPillars", batchUpdatePillarsHandler)

	sessionMgr := newTestSessionManager()

	return &strategyPillarsTestHandlers{
		handlers:                NewMetaModelHandlers(commandBus, readModel, hateoas, sessionMgr.sessionManager),
		strategyPillarsHandlers: NewStrategyPillarsHandlers(commandBus, readModel, hateoas, sessionMgr.sessionManager),
		commandBus:              commandBus,
		readModel:               readModel,
		sessionManager:          sessionMgr,
		tenantID:                tenantID,
	}
}

func (h *strategyPillarsTestHandlers) createStrategyPillarsRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(h.sessionManager.scsManager.LoadAndSave)

	router.Get("/api/v1/meta-model/strategy-pillars", h.wrapHandler(h.strategyPillarsHandlers.GetStrategyPillars))
	router.Post("/api/v1/meta-model/strategy-pillars", h.wrapHandler(h.strategyPillarsHandlers.CreateStrategyPillar))
	router.Get("/api/v1/meta-model/strategy-pillars/{id}", h.wrapHandler(h.strategyPillarsHandlers.GetStrategyPillarByID))
	router.Put("/api/v1/meta-model/strategy-pillars/{id}", h.wrapHandler(h.strategyPillarsHandlers.UpdateStrategyPillar))
	router.Delete("/api/v1/meta-model/strategy-pillars/{id}", h.wrapHandler(h.strategyPillarsHandlers.DeleteStrategyPillar))
	router.Patch("/api/v1/meta-model/strategy-pillars", h.wrapHandler(h.strategyPillarsHandlers.BatchUpdateStrategyPillars))

	return router
}

func (h *strategyPillarsTestHandlers) wrapHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = withTenant(r, h.tenantID)
		handler(w, r)
	}
}

func (h *strategyPillarsTestHandlers) executeRequest(router chi.Router, method, path string, opts requestOptions) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, opts.body)
	if opts.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, c := range opts.cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func (h *strategyPillarsTestHandlers) tenantContext() context.Context {
	return tenantContextWithID(h.tenantID)
}

func (h *strategyPillarsTestHandlers) createTestConfig(t *testing.T, testCtx *testContext) string {
	cmd := &commands.CreateMetaModelConfiguration{
		TenantID:  h.tenantID,
		CreatedBy: "test@example.com",
	}

	result, err := h.commandBus.Dispatch(h.tenantContext(), cmd)
	require.NoError(t, err)

	testCtx.trackID(result.CreatedID)
	time.Sleep(100 * time.Millisecond)

	return result.CreatedID
}

func (h *strategyPillarsTestHandlers) ensureConfigAndGetPillars(t *testing.T, testCtx *testContext) (*readmodels.MetaModelConfigurationDTO, []readmodels.StrategyPillarDTO) {
	config, err := h.readModel.GetByTenantID(h.tenantContext())
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if config == nil {
		h.createTestConfig(t, testCtx)
		config, err = h.readModel.GetByTenantID(h.tenantContext())
		require.NoError(t, err)
		require.NotNil(t, config)
	}

	return config, config.StrategyPillars
}

func generateUniqueTenantID() string {
	return fmt.Sprintf("test-%s", uuid.New().String())
}

func TestCreateStrategyPillar_Success_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateUniqueTenantID()
	h := setupStrategyPillarsTestHandlers(testCtx.db, tenantID)
	config, initialPillars := h.ensureConfigAndGetPillars(t, testCtx)
	initialCount := len(initialPillars)

	uniqueName := fmt.Sprintf("Innovation-%d", time.Now().UnixNano())
	reqBody := CreateStrategyPillarRequest{
		Name:        uniqueName,
		Description: "Driving innovation across the enterprise",
	}
	body, _ := json.Marshal(reqBody)

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createStrategyPillarsRouter()

	w := h.executeRequest(router, http.MethodPost, "/api/v1/meta-model/strategy-pillars", requestOptions{
		body:    bytes.NewReader(body),
		cookies: cookies,
	})

	require.Equal(t, http.StatusCreated, w.Code, "Response: %s", w.Body.String())

	var response StrategyPillarResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Equal(t, uniqueName, response.Name)
	assert.Equal(t, "Driving innovation across the enterprise", response.Description)
	assert.True(t, response.Active)
	assert.NotEmpty(t, response.Links)

	assert.NotEmpty(t, w.Header().Get("Location"))
	assert.NotEmpty(t, w.Header().Get("ETag"))

	time.Sleep(50 * time.Millisecond)

	updatedConfig, err := h.readModel.GetByTenantID(h.tenantContext())
	require.NoError(t, err)
	assert.Equal(t, config.ID, updatedConfig.ID)
	assert.Equal(t, initialCount+1, len(updatedConfig.StrategyPillars))
}

func TestCreateStrategyPillar_DuplicateName_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateUniqueTenantID()
	h := setupStrategyPillarsTestHandlers(testCtx.db, tenantID)
	_, pillars := h.ensureConfigAndGetPillars(t, testCtx)
	require.NotEmpty(t, pillars, "Need at least one pillar")

	existingName := pillars[0].Name
	reqBody := CreateStrategyPillarRequest{
		Name:        existingName,
		Description: "Duplicate of existing pillar",
	}
	body, _ := json.Marshal(reqBody)

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createStrategyPillarsRouter()

	w := h.executeRequest(router, http.MethodPost, "/api/v1/meta-model/strategy-pillars", requestOptions{
		body:    bytes.NewReader(body),
		cookies: cookies,
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteStrategyPillar_Success_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateUniqueTenantID()
	h := setupStrategyPillarsTestHandlers(testCtx.db, tenantID)
	_, pillars := h.ensureConfigAndGetPillars(t, testCtx)

	activePillars := filterActivePillarsDTO(pillars)
	require.GreaterOrEqual(t, len(activePillars), 2, "Need at least 2 active pillars to delete one")

	pillarToDelete := activePillars[0]

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createStrategyPillarsRouter()

	w := h.executeRequest(router, http.MethodDelete, "/api/v1/meta-model/strategy-pillars/"+pillarToDelete.ID, requestOptions{
		cookies: cookies,
	})

	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(50 * time.Millisecond)

	updatedConfig, err := h.readModel.GetByTenantID(h.tenantContext())
	require.NoError(t, err)

	var deletedPillar *readmodels.StrategyPillarDTO
	for i := range updatedConfig.StrategyPillars {
		if updatedConfig.StrategyPillars[i].ID == pillarToDelete.ID {
			deletedPillar = &updatedConfig.StrategyPillars[i]
			break
		}
	}
	require.NotNil(t, deletedPillar, "Pillar should still exist but be inactive")
	assert.False(t, deletedPillar.Active, "Deleted pillar should be inactive")
}

func TestDeleteStrategyPillar_LastActive_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateUniqueTenantID()
	h := setupStrategyPillarsTestHandlers(testCtx.db, tenantID)
	h.ensureConfigAndGetPillars(t, testCtx)

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createStrategyPillarsRouter()

	config, _ := h.readModel.GetByTenantID(h.tenantContext())
	activePillars := filterActivePillarsDTO(config.StrategyPillars)

	for len(activePillars) < 2 {
		reqBody := CreateStrategyPillarRequest{
			Name:        fmt.Sprintf("Extra-%d", time.Now().UnixNano()),
			Description: "Extra pillar for testing",
		}
		body, _ := json.Marshal(reqBody)
		w := h.executeRequest(router, http.MethodPost, "/api/v1/meta-model/strategy-pillars", requestOptions{
			body:    bytes.NewReader(body),
			cookies: cookies,
		})
		require.Equal(t, http.StatusCreated, w.Code)
		time.Sleep(50 * time.Millisecond)

		config, _ = h.readModel.GetByTenantID(h.tenantContext())
		activePillars = filterActivePillarsDTO(config.StrategyPillars)
	}

	require.GreaterOrEqual(t, len(activePillars), 2, "Need at least 2 active pillars")

	for i := 0; i < len(activePillars)-1; i++ {
		w := h.executeRequest(router, http.MethodDelete, "/api/v1/meta-model/strategy-pillars/"+activePillars[i].ID, requestOptions{
			cookies: cookies,
		})
		require.Equal(t, http.StatusNoContent, w.Code, "Failed to delete pillar %d", i)
		time.Sleep(50 * time.Millisecond)
	}

	lastPillar := activePillars[len(activePillars)-1]
	w := h.executeRequest(router, http.MethodDelete, "/api/v1/meta-model/strategy-pillars/"+lastPillar.ID, requestOptions{
		cookies: cookies,
	})

	assert.Equal(t, http.StatusConflict, w.Code, "Should not be able to delete last active pillar")
}

func TestUpdateStrategyPillar_OptimisticLockingConflict_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateUniqueTenantID()
	h := setupStrategyPillarsTestHandlers(testCtx.db, tenantID)
	config, pillars := h.ensureConfigAndGetPillars(t, testCtx)
	require.NotEmpty(t, pillars, "Need at least one pillar")

	pillarToUpdate := pillars[0]
	currentVersion := config.Version

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createStrategyPillarsRouter()

	firstUpdate := UpdateStrategyPillarRequest{
		Name:        fmt.Sprintf("Updated-%d", time.Now().UnixNano()),
		Description: "First update",
	}
	firstBody, _ := json.Marshal(firstUpdate)

	req1 := httptest.NewRequest(http.MethodPut, "/api/v1/meta-model/strategy-pillars/"+pillarToUpdate.ID, bytes.NewReader(firstBody))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("If-Match", fmt.Sprintf(`"%d"`, currentVersion))
	for _, c := range cookies {
		req1.AddCookie(c)
	}
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusOK, w1.Code, "First update should succeed: %s", w1.Body.String())

	time.Sleep(50 * time.Millisecond)

	secondUpdate := UpdateStrategyPillarRequest{
		Name:        "Stale Update",
		Description: "Using old version",
	}
	secondBody, _ := json.Marshal(secondUpdate)

	req2 := httptest.NewRequest(http.MethodPut, "/api/v1/meta-model/strategy-pillars/"+pillarToUpdate.ID, bytes.NewReader(secondBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("If-Match", fmt.Sprintf(`"%d"`, currentVersion))
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusPreconditionFailed, w2.Code, "Second update with stale version should fail")
}

func TestBatchUpdateStrategyPillars_AtomicRollback_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateUniqueTenantID()
	h := setupStrategyPillarsTestHandlers(testCtx.db, tenantID)
	config, pillars := h.ensureConfigAndGetPillars(t, testCtx)
	require.NotEmpty(t, pillars, "Need at least one pillar")
	originalPillarCount := len(pillars)
	originalVersion := config.Version

	existingName := pillars[0].Name

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createStrategyPillarsRouter()

	reqBody := BatchUpdateStrategyPillarsRequest{
		Changes: []PillarChangeRequest{
			{Operation: "add", Name: fmt.Sprintf("Valid-%d", time.Now().UnixNano()), Description: "Valid"},
			{Operation: "add", Name: existingName, Description: "Duplicate - should cause rollback"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/meta-model/strategy-pillars", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("If-Match", fmt.Sprintf(`"%d"`, originalVersion))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Batch with duplicate should fail")

	time.Sleep(50 * time.Millisecond)

	updatedConfig, err := h.readModel.GetByTenantID(h.tenantContext())
	require.NoError(t, err)
	assert.Equal(t, originalPillarCount, len(updatedConfig.StrategyPillars), "No changes should be persisted on error")
	assert.Equal(t, originalVersion, updatedConfig.Version, "Version should not have changed")
}

func TestGetStrategyPillars_FilterInactive_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateUniqueTenantID()
	h := setupStrategyPillarsTestHandlers(testCtx.db, tenantID)
	h.ensureConfigAndGetPillars(t, testCtx)

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createStrategyPillarsRouter()

	config, _ := h.readModel.GetByTenantID(h.tenantContext())
	activePillars := filterActivePillarsDTO(config.StrategyPillars)

	for len(activePillars) < 2 {
		reqBody := CreateStrategyPillarRequest{
			Name:        fmt.Sprintf("Extra-%d", time.Now().UnixNano()),
			Description: "Extra pillar for testing",
		}
		body, _ := json.Marshal(reqBody)
		w := h.executeRequest(router, http.MethodPost, "/api/v1/meta-model/strategy-pillars", requestOptions{
			body:    bytes.NewReader(body),
			cookies: cookies,
		})
		require.Equal(t, http.StatusCreated, w.Code)
		time.Sleep(50 * time.Millisecond)

		config, _ = h.readModel.GetByTenantID(h.tenantContext())
		activePillars = filterActivePillarsDTO(config.StrategyPillars)
	}

	pillarToDelete := activePillars[0]
	deleteW := h.executeRequest(router, http.MethodDelete, "/api/v1/meta-model/strategy-pillars/"+pillarToDelete.ID, requestOptions{
		cookies: cookies,
	})
	require.Equal(t, http.StatusNoContent, deleteW.Code)
	time.Sleep(50 * time.Millisecond)

	getActiveW := h.executeRequest(router, http.MethodGet, "/api/v1/meta-model/strategy-pillars", requestOptions{})
	require.Equal(t, http.StatusOK, getActiveW.Code)

	var activeResponse struct {
		Data []StrategyPillarResponse `json:"data"`
	}
	err := json.NewDecoder(getActiveW.Body).Decode(&activeResponse)
	require.NoError(t, err)

	for _, p := range activeResponse.Data {
		assert.True(t, p.Active, "All returned pillars should be active")
		assert.NotEqual(t, pillarToDelete.ID, p.ID, "Deleted pillar should not be in active list")
	}

	getAllW := h.executeRequest(router, http.MethodGet, "/api/v1/meta-model/strategy-pillars?includeInactive=true", requestOptions{})
	require.Equal(t, http.StatusOK, getAllW.Code)

	var allResponse struct {
		Data []StrategyPillarResponse `json:"data"`
	}
	err = json.NewDecoder(getAllW.Body).Decode(&allResponse)
	require.NoError(t, err)

	assert.Greater(t, len(allResponse.Data), len(activeResponse.Data), "Including inactive should return more pillars")

	foundDeleted := false
	for _, p := range allResponse.Data {
		if p.ID == pillarToDelete.ID {
			foundDeleted = true
			assert.False(t, p.Active, "Deleted pillar should be inactive")
		}
	}
	assert.True(t, foundDeleted, "Deleted pillar should be in list when including inactive")
}

func TestUpdateStrategyPillar_MissingIfMatchHeader_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateUniqueTenantID()
	h := setupStrategyPillarsTestHandlers(testCtx.db, tenantID)
	_, pillars := h.ensureConfigAndGetPillars(t, testCtx)
	require.NotEmpty(t, pillars, "Need at least one pillar")

	pillarToUpdate := pillars[0]

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createStrategyPillarsRouter()

	reqBody := UpdateStrategyPillarRequest{
		Name:        "Updated Name",
		Description: "Updated description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/meta-model/strategy-pillars/"+pillarToUpdate.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateStrategyPillar_Unauthorized_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateUniqueTenantID()
	h := setupStrategyPillarsTestHandlers(testCtx.db, tenantID)
	h.ensureConfigAndGetPillars(t, testCtx)

	reqBody := CreateStrategyPillarRequest{
		Name:        fmt.Sprintf("New-%d", time.Now().UnixNano()),
		Description: "Description",
	}
	body, _ := json.Marshal(reqBody)

	router := h.createStrategyPillarsRouter()

	w := h.executeRequest(router, http.MethodPost, "/api/v1/meta-model/strategy-pillars", requestOptions{
		body: bytes.NewReader(body),
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func filterActivePillarsDTO(pillars []readmodels.StrategyPillarDTO) []readmodels.StrategyPillarDTO {
	result := make([]readmodels.StrategyPillarDTO, 0)
	for _, p := range pillars {
		if p.Active {
			result = append(result, p)
		}
	}
	return result
}
