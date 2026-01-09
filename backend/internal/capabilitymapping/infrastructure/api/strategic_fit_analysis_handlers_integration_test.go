//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	"easi/backend/internal/infrastructure/database"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStrategyPillarsGateway struct {
	pillars map[string]*metamodel.StrategyPillarDTO
}

func newMockPillarsGateway() *mockStrategyPillarsGateway {
	return &mockStrategyPillarsGateway{
		pillars: make(map[string]*metamodel.StrategyPillarDTO),
	}
}

func (m *mockStrategyPillarsGateway) GetStrategyPillars(ctx context.Context) (*metamodel.StrategyPillarsConfigDTO, error) {
	var pillars []metamodel.StrategyPillarDTO
	for _, p := range m.pillars {
		pillars = append(pillars, *p)
	}
	return &metamodel.StrategyPillarsConfigDTO{Pillars: pillars}, nil
}

func (m *mockStrategyPillarsGateway) GetActivePillar(ctx context.Context, pillarID string) (*metamodel.StrategyPillarDTO, error) {
	if p, ok := m.pillars[pillarID]; ok && p.Active {
		return p, nil
	}
	return nil, nil
}

func (m *mockStrategyPillarsGateway) InvalidateCache(tenantID string) {}

func (m *mockStrategyPillarsGateway) addPillar(id, name string, fitScoringEnabled bool) {
	m.pillars[id] = &metamodel.StrategyPillarDTO{
		ID:                id,
		Name:              name,
		Active:            true,
		FitScoringEnabled: fitScoringEnabled,
	}
}

type strategicFitTestContext struct {
	db         *sql.DB
	testID     string
	createdIDs []string
}

func (ctx *strategicFitTestContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func setupStrategicFitTestDB(t *testing.T) (*strategicFitTestContext, func()) {
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

	ctx := &strategicFitTestContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	cleanup := func() {
		ctx.setTenantContext(t)
		for _, id := range ctx.createdIDs {
			db.Exec("DELETE FROM effective_capability_importance WHERE capability_id = $1 OR business_domain_id = $1 OR source_capability_id = $1", id)
			db.Exec("DELETE FROM application_fit_scores WHERE id = $1 OR component_id = $1", id)
			db.Exec("DELETE FROM strategy_importance WHERE id = $1 OR capability_id = $1", id)
			db.Exec("DELETE FROM capability_realizations WHERE id = $1 OR capability_id = $1 OR component_id = $1", id)
			db.Exec("DELETE FROM domain_capability_assignments WHERE capability_id = $1 OR business_domain_id = $1", id)
			db.Exec("DELETE FROM domain_capability_metadata WHERE capability_id = $1 OR business_domain_id = $1", id)
			db.Exec("DELETE FROM capabilities WHERE id = $1", id)
			db.Exec("DELETE FROM business_domains WHERE id = $1", id)
			db.Exec("DELETE FROM application_components WHERE id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *strategicFitTestContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

func (ctx *strategicFitTestContext) createTestCapability(t *testing.T, id, name, level string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO capabilities (id, name, description, level, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		id, name, "", level, testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func (ctx *strategicFitTestContext) createTestBusinessDomain(t *testing.T, id, name string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, NOW())",
		id, testTenantID(), name, "", 0,
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func (ctx *strategicFitTestContext) createTestComponent(t *testing.T, id, name string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, NOW())",
		id, testTenantID(), name, "",
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func (ctx *strategicFitTestContext) createTestRealization(t *testing.T, id, capabilityID, componentID, componentName string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, component_name, realization_level, origin, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		id, testTenantID(), capabilityID, componentID, componentName, "Full", "Direct",
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func (ctx *strategicFitTestContext) createTestDomainAssignment(t *testing.T, assignmentID, domainID, domainName, capabilityID, capabilityName string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		assignmentID, testTenantID(), domainID, domainName, capabilityID, capabilityName, "L1",
	)
	require.NoError(t, err)
	ctx.trackID(assignmentID)
}

func (ctx *strategicFitTestContext) createTestStrategyImportance(t *testing.T, id, domainID, domainName, capabilityID, capabilityName, pillarID, pillarName string, importance int) {
	ctx.setTenantContext(t)
	labels := map[int]string{1: "Very Low", 2: "Low", 3: "Medium", 4: "High", 5: "Very High"}
	_, err := ctx.db.Exec(
		"INSERT INTO strategy_importance (id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, pillar_id, pillar_name, importance, importance_label, set_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())",
		id, testTenantID(), domainID, domainName, capabilityID, capabilityName, pillarID, pillarName, importance, labels[importance],
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func (ctx *strategicFitTestContext) createTestApplicationFitScore(t *testing.T, id, componentID, componentName, pillarID, pillarName string, score int) {
	ctx.setTenantContext(t)
	labels := map[int]string{1: "Critical", 2: "Poor", 3: "Adequate", 4: "Good", 5: "Excellent"}
	_, err := ctx.db.Exec(
		"INSERT INTO application_fit_scores (id, tenant_id, component_id, component_name, pillar_id, pillar_name, score, score_label, scored_at, scored_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), $9)",
		id, testTenantID(), componentID, componentName, pillarID, pillarName, score, labels[score], "test-user",
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func (ctx *strategicFitTestContext) createTestDomainCapabilityMetadata(t *testing.T, capabilityID, capabilityName, level, l1CapabilityID, domainID, domainName string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO domain_capability_metadata (tenant_id, capability_id, capability_name, capability_level, l1_capability_id, business_domain_id, business_domain_name) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		testTenantID(), capabilityID, capabilityName, level, l1CapabilityID, domainID, domainName,
	)
	require.NoError(t, err)
	ctx.trackID(capabilityID)
}

func (ctx *strategicFitTestContext) createTestEffectiveCapabilityImportance(t *testing.T, capabilityID, pillarID, domainID string, importance int, sourceCapabilityID, sourceCapabilityName string, isInherited bool) {
	ctx.setTenantContext(t)
	labels := map[int]string{1: "Very Low", 2: "Low", 3: "Medium", 4: "High", 5: "Very High"}
	_, err := ctx.db.Exec(
		"INSERT INTO effective_capability_importance (tenant_id, capability_id, pillar_id, business_domain_id, effective_importance, importance_label, source_capability_id, source_capability_name, is_inherited, computed_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())",
		testTenantID(), capabilityID, pillarID, domainID, importance, labels[importance], sourceCapabilityID, sourceCapabilityName, isInherited,
	)
	require.NoError(t, err)
}

func setupStrategicFitHandlers(db *sql.DB, pillarsGateway metamodel.StrategyPillarsGateway) (*StrategicFitAnalysisHandlers, *scs.SessionManager) {
	tenantDB := database.NewTenantAwareDB(db)

	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	analysisRM := readmodels.NewStrategicFitAnalysisReadModel(tenantDB)

	return NewStrategicFitAnalysisHandlers(analysisRM, pillarsGateway, sessionManager), scsManager
}

func withAuthenticatedSession(req *http.Request, scsManager *scs.SessionManager) *http.Request {
	ctx, _ := scsManager.Load(req.Context(), "")
	tenantID, _ := sharedvo.NewTenantID(testTenantID())
	preAuth := session.NewPreAuthSession(tenantID, "example.com", "")
	authenticatedSession := preAuth.UpgradeToAuthenticated(
		session.UserInfo{ID: uuid.New(), Email: "test@example.com"},
		session.TokenInfo{AccessToken: "test-token", RefreshToken: "test-refresh", Expiry: time.Now().Add(time.Hour)},
	)
	data, _ := authenticatedSession.Marshal()
	scsManager.Put(ctx, "auth_session", data)
	return req.WithContext(ctx)
}

func TestGetStrategicFitAnalysis_WithData_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	pillarID := uuid.New().String()
	pillarName := "Always On"

	mockGateway := newMockPillarsGateway()
	mockGateway.addPillar(pillarID, pillarName, true)

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	capabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()
	realizationID := uuid.New().String()
	assignmentID := uuid.New().String()
	importanceID := uuid.New().String()
	fitScoreID := uuid.New().String()

	testCtx.createTestCapability(t, capabilityID, "Customer Onboarding", "L1")
	testCtx.createTestBusinessDomain(t, domainID, "Customer Management")
	testCtx.createTestComponent(t, componentID, "CRM System")
	testCtx.createTestRealization(t, realizationID, capabilityID, componentID, "CRM System")
	testCtx.createTestDomainAssignment(t, assignmentID, domainID, "Customer Management", capabilityID, "Customer Onboarding")
	testCtx.createTestDomainCapabilityMetadata(t, capabilityID, "Customer Onboarding", "L1", capabilityID, domainID, "Customer Management")
	testCtx.createTestStrategyImportance(t, importanceID, domainID, "Customer Management", capabilityID, "Customer Onboarding", pillarID, pillarName, 5)
	testCtx.createTestEffectiveCapabilityImportance(t, capabilityID, pillarID, domainID, 5, capabilityID, "Customer Onboarding", false)
	testCtx.createTestApplicationFitScore(t, fitScoreID, componentID, "CRM System", pillarID, pillarName, 2)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+pillarID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", pillarID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var response StrategicFitAnalysisResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, pillarID, response.PillarID)
	assert.Equal(t, pillarName, response.PillarName)
	assert.Equal(t, 1, response.Summary.ScoredRealizations)
	assert.Equal(t, 1, response.Summary.LiabilityCount)
	assert.Len(t, response.Liabilities, 1)

	liability := response.Liabilities[0]
	assert.Equal(t, componentID, liability.ComponentID)
	assert.Equal(t, capabilityID, liability.CapabilityID)
	assert.Equal(t, 5, liability.Importance)
	assert.Equal(t, 2, liability.FitScore)
	assert.Equal(t, 3, liability.Gap)
	assert.Equal(t, "liability", liability.Category)
	assert.Equal(t, capabilityID, liability.ImportanceSourceCapabilityID)
	assert.Equal(t, "Customer Onboarding", liability.ImportanceSourceCapabilityName)
	assert.False(t, liability.IsImportanceInherited)
}

func TestGetStrategicFitAnalysis_FitScoringDisabled_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	pillarID := fmt.Sprintf("test-pillar-%d", time.Now().UnixNano())

	mockGateway := newMockPillarsGateway()
	mockGateway.addPillar(pillarID, "Transform", false)

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+pillarID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", pillarID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetStrategicFitAnalysis_PillarNotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	mockGateway := newMockPillarsGateway()

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	nonExistentID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+nonExistentID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", nonExistentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetStrategicFitAnalysis_InvalidPillarID_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	mockGateway := newMockPillarsGateway()

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/invalid-uuid", nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetStrategicFitAnalysis_QueryExecutes_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	pillarID := uuid.New().String()
	pillarName := "Grow"

	mockGateway := newMockPillarsGateway()
	mockGateway.addPillar(pillarID, pillarName, true)

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+pillarID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", pillarID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var response StrategicFitAnalysisResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, pillarID, response.PillarID)
	assert.Equal(t, pillarName, response.PillarName)
}

func (ctx *strategicFitTestContext) createTestCapabilityWithParent(t *testing.T, id, name, level, parentID string) {
	ctx.setTenantContext(t)
	var err error
	if parentID == "" {
		_, err = ctx.db.Exec(
			"INSERT INTO capabilities (id, name, description, level, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
			id, name, "", level, testTenantID(), "Genesis", "Active",
		)
	} else {
		_, err = ctx.db.Exec(
			"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
			id, name, "", level, parentID, testTenantID(), "Genesis", "Active",
		)
	}
	require.NoError(t, err)
	ctx.trackID(id)
}

func TestGetStrategicFitAnalysis_InheritedImportanceFromParent_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	pillarID := uuid.New().String()
	pillarName := "Always On"

	mockGateway := newMockPillarsGateway()
	mockGateway.addPillar(pillarID, pillarName, true)

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	l1CapabilityID := uuid.New().String()
	l2CapabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()
	realizationID := uuid.New().String()
	assignmentID := uuid.New().String()
	importanceID := uuid.New().String()
	fitScoreID := uuid.New().String()

	testCtx.createTestCapabilityWithParent(t, l1CapabilityID, "Payment Processing", "L1", "")
	testCtx.createTestCapabilityWithParent(t, l2CapabilityID, "Card Payments", "L2", l1CapabilityID)
	testCtx.createTestBusinessDomain(t, domainID, "Finance")
	testCtx.createTestComponent(t, componentID, "Payment Gateway")
	testCtx.createTestRealization(t, realizationID, l2CapabilityID, componentID, "Payment Gateway")
	testCtx.createTestDomainAssignment(t, assignmentID, domainID, "Finance", l1CapabilityID, "Payment Processing")
	testCtx.createTestDomainCapabilityMetadata(t, l2CapabilityID, "Card Payments", "L2", l1CapabilityID, domainID, "Finance")
	testCtx.createTestStrategyImportance(t, importanceID, domainID, "Finance", l1CapabilityID, "Payment Processing", pillarID, pillarName, 4)
	testCtx.createTestEffectiveCapabilityImportance(t, l2CapabilityID, pillarID, domainID, 4, l1CapabilityID, "Payment Processing", true)
	testCtx.createTestApplicationFitScore(t, fitScoreID, componentID, "Payment Gateway", pillarID, pillarName, 2)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+pillarID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", pillarID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var response StrategicFitAnalysisResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 1, response.Summary.ScoredRealizations)
	require.Len(t, response.Liabilities, 1)

	liability := response.Liabilities[0]
	assert.Equal(t, l2CapabilityID, liability.CapabilityID)
	assert.Equal(t, "Card Payments", liability.CapabilityName)
	assert.Equal(t, 4, liability.Importance)
	assert.Equal(t, 2, liability.FitScore)
	assert.Equal(t, 2, liability.Gap)
	assert.Equal(t, l1CapabilityID, liability.ImportanceSourceCapabilityID)
	assert.Equal(t, "Payment Processing", liability.ImportanceSourceCapabilityName)
	assert.True(t, liability.IsImportanceInherited)
}

func TestGetStrategicFitAnalysis_DirectRatingOverridesParent_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	pillarID := uuid.New().String()
	pillarName := "Always On"

	mockGateway := newMockPillarsGateway()
	mockGateway.addPillar(pillarID, pillarName, true)

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	l1CapabilityID := uuid.New().String()
	l2CapabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()
	realizationID := uuid.New().String()
	assignmentID := uuid.New().String()
	importanceID1 := uuid.New().String()
	importanceID2 := uuid.New().String()
	fitScoreID := uuid.New().String()

	testCtx.createTestCapabilityWithParent(t, l1CapabilityID, "Payment Processing", "L1", "")
	testCtx.createTestCapabilityWithParent(t, l2CapabilityID, "Card Payments", "L2", l1CapabilityID)
	testCtx.createTestBusinessDomain(t, domainID, "Finance")
	testCtx.createTestComponent(t, componentID, "Payment Gateway")
	testCtx.createTestRealization(t, realizationID, l2CapabilityID, componentID, "Payment Gateway")
	testCtx.createTestDomainAssignment(t, assignmentID, domainID, "Finance", l1CapabilityID, "Payment Processing")
	testCtx.createTestDomainCapabilityMetadata(t, l2CapabilityID, "Card Payments", "L2", l1CapabilityID, domainID, "Finance")
	testCtx.createTestStrategyImportance(t, importanceID1, domainID, "Finance", l1CapabilityID, "Payment Processing", pillarID, pillarName, 3)
	testCtx.createTestStrategyImportance(t, importanceID2, domainID, "Finance", l2CapabilityID, "Card Payments", pillarID, pillarName, 5)
	testCtx.createTestEffectiveCapabilityImportance(t, l2CapabilityID, pillarID, domainID, 5, l2CapabilityID, "Card Payments", false)
	testCtx.createTestApplicationFitScore(t, fitScoreID, componentID, "Payment Gateway", pillarID, pillarName, 2)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+pillarID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", pillarID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var response StrategicFitAnalysisResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 1, response.Summary.ScoredRealizations)
	require.Len(t, response.Liabilities, 1)

	liability := response.Liabilities[0]
	assert.Equal(t, l2CapabilityID, liability.CapabilityID)
	assert.Equal(t, 5, liability.Importance)
	assert.Equal(t, 2, liability.FitScore)
	assert.Equal(t, 3, liability.Gap)
	assert.Equal(t, l2CapabilityID, liability.ImportanceSourceCapabilityID)
	assert.Equal(t, "Card Payments", liability.ImportanceSourceCapabilityName)
	assert.False(t, liability.IsImportanceInherited)
}

func TestGetStrategicFitAnalysis_MultipleCapabilitiesInSameChain_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	pillarID := uuid.New().String()
	pillarName := "Always On"

	mockGateway := newMockPillarsGateway()
	mockGateway.addPillar(pillarID, pillarName, true)

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	l1CapabilityID := uuid.New().String()
	l2CapabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()
	realizationID1 := uuid.New().String()
	realizationID2 := uuid.New().String()
	assignmentID := uuid.New().String()
	importanceID1 := uuid.New().String()
	importanceID2 := uuid.New().String()
	fitScoreID := uuid.New().String()

	testCtx.createTestCapabilityWithParent(t, l1CapabilityID, "Payment Processing", "L1", "")
	testCtx.createTestCapabilityWithParent(t, l2CapabilityID, "Card Payments", "L2", l1CapabilityID)
	testCtx.createTestBusinessDomain(t, domainID, "Finance")
	testCtx.createTestComponent(t, componentID, "Payment Gateway")
	testCtx.createTestRealization(t, realizationID1, l1CapabilityID, componentID, "Payment Gateway")
	testCtx.createTestRealization(t, realizationID2, l2CapabilityID, componentID, "Payment Gateway")
	testCtx.createTestDomainAssignment(t, assignmentID, domainID, "Finance", l1CapabilityID, "Payment Processing")
	testCtx.createTestDomainCapabilityMetadata(t, l1CapabilityID, "Payment Processing", "L1", l1CapabilityID, domainID, "Finance")
	testCtx.createTestDomainCapabilityMetadata(t, l2CapabilityID, "Card Payments", "L2", l1CapabilityID, domainID, "Finance")
	testCtx.createTestStrategyImportance(t, importanceID1, domainID, "Finance", l1CapabilityID, "Payment Processing", pillarID, pillarName, 4)
	testCtx.createTestStrategyImportance(t, importanceID2, domainID, "Finance", l2CapabilityID, "Card Payments", pillarID, pillarName, 5)
	testCtx.createTestEffectiveCapabilityImportance(t, l1CapabilityID, pillarID, domainID, 4, l1CapabilityID, "Payment Processing", false)
	testCtx.createTestEffectiveCapabilityImportance(t, l2CapabilityID, pillarID, domainID, 5, l2CapabilityID, "Card Payments", false)
	testCtx.createTestApplicationFitScore(t, fitScoreID, componentID, "Payment Gateway", pillarID, pillarName, 2)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+pillarID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", pillarID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var response StrategicFitAnalysisResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Summary.ScoredRealizations)
	assert.Equal(t, 2, response.Summary.LiabilityCount)

	capabilitiesFound := make(map[string]RealizationFitResponse)
	for _, liability := range response.Liabilities {
		capabilitiesFound[liability.CapabilityID] = liability
	}

	l1Entry, ok := capabilitiesFound[l1CapabilityID]
	require.True(t, ok, "Expected L1 capability in results")
	assert.Equal(t, 4, l1Entry.Importance)
	assert.Equal(t, 2, l1Entry.Gap)
	assert.False(t, l1Entry.IsImportanceInherited)

	l2Entry, ok := capabilitiesFound[l2CapabilityID]
	require.True(t, ok, "Expected L2 capability in results")
	assert.Equal(t, 5, l2Entry.Importance)
	assert.Equal(t, 3, l2Entry.Gap)
	assert.False(t, l2Entry.IsImportanceInherited)
}

func TestGetStrategicFitAnalysis_DeepHierarchyInheritance_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	pillarID := uuid.New().String()
	pillarName := "Grow"

	mockGateway := newMockPillarsGateway()
	mockGateway.addPillar(pillarID, pillarName, true)

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	l1CapabilityID := uuid.New().String()
	l2CapabilityID := uuid.New().String()
	l3CapabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()
	realizationID := uuid.New().String()
	assignmentID := uuid.New().String()
	importanceID := uuid.New().String()
	fitScoreID := uuid.New().String()

	testCtx.createTestCapabilityWithParent(t, l1CapabilityID, "Customer Management", "L1", "")
	testCtx.createTestCapabilityWithParent(t, l2CapabilityID, "Customer Onboarding", "L2", l1CapabilityID)
	testCtx.createTestCapabilityWithParent(t, l3CapabilityID, "Identity Verification", "L3", l2CapabilityID)
	testCtx.createTestBusinessDomain(t, domainID, "Customer")
	testCtx.createTestComponent(t, componentID, "KYC System")
	testCtx.createTestRealization(t, realizationID, l3CapabilityID, componentID, "KYC System")
	testCtx.createTestDomainAssignment(t, assignmentID, domainID, "Customer", l1CapabilityID, "Customer Management")
	testCtx.createTestDomainCapabilityMetadata(t, l3CapabilityID, "Identity Verification", "L3", l1CapabilityID, domainID, "Customer")
	testCtx.createTestStrategyImportance(t, importanceID, domainID, "Customer", l1CapabilityID, "Customer Management", pillarID, pillarName, 5)
	testCtx.createTestEffectiveCapabilityImportance(t, l3CapabilityID, pillarID, domainID, 5, l1CapabilityID, "Customer Management", true)
	testCtx.createTestApplicationFitScore(t, fitScoreID, componentID, "KYC System", pillarID, pillarName, 3)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+pillarID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", pillarID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var response StrategicFitAnalysisResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 1, response.Summary.ScoredRealizations)
	require.Len(t, response.Liabilities, 1)

	liability := response.Liabilities[0]
	assert.Equal(t, l3CapabilityID, liability.CapabilityID)
	assert.Equal(t, "Identity Verification", liability.CapabilityName)
	assert.Equal(t, 5, liability.Importance)
	assert.Equal(t, 3, liability.FitScore)
	assert.Equal(t, 2, liability.Gap)
	assert.Equal(t, l1CapabilityID, liability.ImportanceSourceCapabilityID)
	assert.Equal(t, "Customer Management", liability.ImportanceSourceCapabilityName)
	assert.True(t, liability.IsImportanceInherited)
}

func TestGetStrategicFitAnalysis_NoRatingInHierarchy_NoGapEntry_Integration(t *testing.T) {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	defer cleanup()

	pillarID := uuid.New().String()
	pillarName := "Transform"

	mockGateway := newMockPillarsGateway()
	mockGateway.addPillar(pillarID, pillarName, true)

	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)

	l1CapabilityID := uuid.New().String()
	l2CapabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()
	realizationID := uuid.New().String()
	assignmentID := uuid.New().String()
	fitScoreID := uuid.New().String()

	testCtx.createTestCapabilityWithParent(t, l1CapabilityID, "Support Operations", "L1", "")
	testCtx.createTestCapabilityWithParent(t, l2CapabilityID, "Ticket Management", "L2", l1CapabilityID)
	testCtx.createTestBusinessDomain(t, domainID, "Operations")
	testCtx.createTestComponent(t, componentID, "Helpdesk System")
	testCtx.createTestRealization(t, realizationID, l2CapabilityID, componentID, "Helpdesk System")
	testCtx.createTestDomainAssignment(t, assignmentID, domainID, "Operations", l1CapabilityID, "Support Operations")
	testCtx.createTestDomainCapabilityMetadata(t, l2CapabilityID, "Ticket Management", "L2", l1CapabilityID, domainID, "Operations")
	testCtx.createTestApplicationFitScore(t, fitScoreID, componentID, "Helpdesk System", pillarID, pillarName, 4)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+pillarID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", pillarID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var response StrategicFitAnalysisResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 0, response.Summary.ScoredRealizations)
	assert.Empty(t, response.Liabilities)
	assert.Empty(t, response.Concerns)
	assert.Empty(t, response.Aligned)
}
