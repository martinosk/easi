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
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestTenantID() string {
	return fmt.Sprintf("test-%s", uuid.New().String())
}

type testContext struct {
	db         *sql.DB
	testID     string
	createdIDs []string
}

func (ctx *testContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
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
			db.Exec("DELETE FROM metamodel.meta_model_configurations WHERE id = $1", id)
			db.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

type testHandlers struct {
	handlers       *MetaModelHandlers
	commandBus     *cqrs.InMemoryCommandBus
	readModel      *readmodels.MetaModelConfigurationReadModel
	sessionManager *testSessionManager
	tenantID       string
}

func setupHandlers(db *sql.DB, tenantID string) *testHandlers {
	tenantDB := database.NewTenantAwareDB(db)

	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	links := NewMetaModelLinks(sharedAPI.NewHATEOASLinks("/api/v1"))

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

	sessionMgr := newTestSessionManager()

	return &testHandlers{
		handlers:       NewMetaModelHandlers(commandBus, readModel, links, sessionMgr.sessionManager),
		commandBus:     commandBus,
		readModel:      readModel,
		sessionManager: sessionMgr,
		tenantID:       tenantID,
	}
}

func (h *testHandlers) tenantContext() context.Context {
	return tenantContextWithID(h.tenantID)
}

func (h *testHandlers) createRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(h.sessionManager.scsManager.LoadAndSave)

	router.Put("/api/v1/meta-model/maturity-scale", h.wrapHandler(h.handlers.UpdateMaturityScale))
	router.Post("/api/v1/meta-model/maturity-scale/reset", h.wrapHandler(h.handlers.ResetMaturityScale))

	return router
}

func (h *testHandlers) wrapHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = withTenant(r, h.tenantID)
		handler(w, r)
	}
}

type requestOptions struct {
	body    io.Reader
	cookies []*http.Cookie
}

func (h *testHandlers) executeRequest(router chi.Router, method, path string, opts requestOptions) *httptest.ResponseRecorder {
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

func createTestConfiguration(t *testing.T, testCtx *testContext, h *testHandlers) string {
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

func validSectionsRequest(version int) UpdateMaturityScaleRequest {
	return UpdateMaturityScaleRequest{
		Sections: [4]MaturitySectionRequest{
			{Order: 1, Name: "Section 1", MinValue: 0, MaxValue: 24},
			{Order: 2, Name: "Section 2", MinValue: 25, MaxValue: 49},
			{Order: 3, Name: "Section 3", MinValue: 50, MaxValue: 74},
			{Order: 4, Name: "Section 4", MinValue: 75, MaxValue: 99},
		},
		Version: version,
	}
}

func TestGetMaturityScale_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)
	configID := createTestConfiguration(t, testCtx, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta-model/maturity-scale", nil)
	req = withTenant(req, h.tenantID)
	w := httptest.NewRecorder()

	h.handlers.GetMaturityScale(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, configID, response.ID)
	assert.Equal(t, h.tenantID, response.TenantID)
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

func TestGetMaturityScale_ReturnsDefault_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta-model/maturity-scale", nil)
	req = withTenant(req, h.tenantID)
	w := httptest.NewRecorder()

	h.handlers.GetMaturityScale(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.IsDefault)
	assert.Equal(t, 0, response.Version)
	assert.Len(t, response.Sections, 4)
	assert.Equal(t, "Genesis", response.Sections[0].Name)
}

func TestUpdateMaturityScale_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)
	configID := createTestConfiguration(t, testCtx, h)

	reqBody := UpdateMaturityScaleRequest{
		Sections: [4]MaturitySectionRequest{
			{Order: 1, Name: "Early Stage", MinValue: 0, MaxValue: 19},
			{Order: 2, Name: "Growth", MinValue: 20, MaxValue: 49},
			{Order: 3, Name: "Mature", MinValue: 50, MaxValue: 79},
			{Order: 4, Name: "Sunset", MinValue: 80, MaxValue: 99},
		},
		Version: 1,
	}
	body, _ := json.Marshal(reqBody)

	userEmail := "admin@acme.com"
	cookies := h.sessionManager.getSessionCookies(t, userEmail)
	router := h.createRouter()

	w := h.executeRequest(router, http.MethodPut, "/api/v1/meta-model/maturity-scale", requestOptions{
		body:    bytes.NewReader(body),
		cookies: cookies,
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, configID, response.ID)
	assert.Equal(t, 2, response.Version)
	assert.False(t, response.IsDefault)
	assert.Equal(t, "Early Stage", response.Sections[0].Name)
	assert.Equal(t, "Growth", response.Sections[1].Name)
	assert.Equal(t, "Mature", response.Sections[2].Name)
	assert.Equal(t, "Sunset", response.Sections[3].Name)
	assert.Equal(t, userEmail, response.ModifiedBy)
}

func TestUpdateMaturityScale_InvalidSections_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)
	createTestConfiguration(t, testCtx, h)

	reqBody := UpdateMaturityScaleRequest{
		Sections: [4]MaturitySectionRequest{
			{Order: 1, Name: "Section 1", MinValue: 0, MaxValue: 30},
			{Order: 2, Name: "Section 2", MinValue: 31, MaxValue: 50},
			{Order: 3, Name: "Section 3", MinValue: 51, MaxValue: 80},
			{Order: 4, Name: "Section 4", MinValue: 81, MaxValue: 90},
		},
		Version: 1,
	}
	body, _ := json.Marshal(reqBody)

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createRouter()

	w := h.executeRequest(router, http.MethodPut, "/api/v1/meta-model/maturity-scale", requestOptions{
		body:    bytes.NewReader(body),
		cookies: cookies,
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestResetMaturityScale_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)
	configID := createTestConfiguration(t, testCtx, h)

	userEmail := "admin@acme.com"
	cookies := h.sessionManager.getSessionCookies(t, userEmail)
	router := h.createRouter()

	updateReqBody := UpdateMaturityScaleRequest{
		Sections: [4]MaturitySectionRequest{
			{Order: 1, Name: "Custom 1", MinValue: 0, MaxValue: 24},
			{Order: 2, Name: "Custom 2", MinValue: 25, MaxValue: 49},
			{Order: 3, Name: "Custom 3", MinValue: 50, MaxValue: 74},
			{Order: 4, Name: "Custom 4", MinValue: 75, MaxValue: 99},
		},
		Version: 1,
	}
	updateBody, _ := json.Marshal(updateReqBody)

	updateW := h.executeRequest(router, http.MethodPut, "/api/v1/meta-model/maturity-scale", requestOptions{
		body:    bytes.NewReader(updateBody),
		cookies: cookies,
	})
	require.Equal(t, http.StatusOK, updateW.Code)

	time.Sleep(100 * time.Millisecond)

	w := h.executeRequest(router, http.MethodPost, "/api/v1/meta-model/maturity-scale/reset", requestOptions{
		cookies: cookies,
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, configID, response.ID)
	assert.Equal(t, 3, response.Version)
	assert.True(t, response.IsDefault)
	assert.Equal(t, userEmail, response.ModifiedBy)

	assert.Equal(t, "Genesis", response.Sections[0].Name)
	assert.Equal(t, "Custom Built", response.Sections[1].Name)
	assert.Equal(t, "Product", response.Sections[2].Name)
	assert.Equal(t, "Commodity", response.Sections[3].Name)
}

func TestGetMaturityScaleByID_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)
	configID := createTestConfiguration(t, testCtx, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta-model/configurations/"+configID, nil)
	req = withTenant(req, h.tenantID)
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

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta-model/configurations/"+nonExistentID, nil)
	req = withTenant(req, h.tenantID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handlers.GetMaturityScaleByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateMaturityScale_AutoCreatesConfig_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)

	body, _ := json.Marshal(validSectionsRequest(0))
	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createRouter()

	w := h.executeRequest(router, http.MethodPut, "/api/v1/meta-model/maturity-scale", requestOptions{
		body:    bytes.NewReader(body),
		cookies: cookies,
	})

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestResetMaturityScale_AutoCreatesConfig_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)

	cookies := h.sessionManager.getSessionCookies(t, "admin@acme.com")
	router := h.createRouter()

	w := h.executeRequest(router, http.MethodPost, "/api/v1/meta-model/maturity-scale/reset", requestOptions{
		cookies: cookies,
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.IsDefault)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, h.tenantID, response.TenantID)
}

func TestUpdateMaturityScale_Unauthorized_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)
	createTestConfiguration(t, testCtx, h)

	body, _ := json.Marshal(validSectionsRequest(1))
	router := h.createRouter()

	w := h.executeRequest(router, http.MethodPut, "/api/v1/meta-model/maturity-scale", requestOptions{
		body: bytes.NewReader(body),
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestResetMaturityScale_Unauthorized_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)
	createTestConfiguration(t, testCtx, h)

	router := h.createRouter()

	w := h.executeRequest(router, http.MethodPost, "/api/v1/meta-model/maturity-scale/reset", requestOptions{})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMaturityScale_RecordsUserEmail_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := generateTestTenantID()
	h := setupHandlers(testCtx.db, tenantID)
	createTestConfiguration(t, testCtx, h)

	reqBody := UpdateMaturityScaleRequest{
		Sections: [4]MaturitySectionRequest{
			{Order: 1, Name: "Phase 1", MinValue: 0, MaxValue: 24},
			{Order: 2, Name: "Phase 2", MinValue: 25, MaxValue: 49},
			{Order: 3, Name: "Phase 3", MinValue: 50, MaxValue: 74},
			{Order: 4, Name: "Phase 4", MinValue: 75, MaxValue: 99},
		},
		Version: 1,
	}
	body, _ := json.Marshal(reqBody)

	specificEmail := "specific-user@acme.com"
	cookies := h.sessionManager.getSessionCookies(t, specificEmail)
	router := h.createRouter()

	w := h.executeRequest(router, http.MethodPut, "/api/v1/meta-model/maturity-scale", requestOptions{
		body:    bytes.NewReader(body),
		cookies: cookies,
	})

	require.Equal(t, http.StatusOK, w.Code)

	var response readmodels.MetaModelConfigurationDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, specificEmail, response.ModifiedBy, "ModifiedBy should contain the authenticated user's email")
}
