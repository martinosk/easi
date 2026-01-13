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

type capabilityData struct {
	id       string
	name     string
	level    string
	parentID string
}

type domainData struct {
	id   string
	name string
}

type componentData struct {
	id   string
	name string
}

type realizationData struct {
	id           string
	capabilityID string
	componentID  string
}

type domainAssignmentData struct {
	id             string
	domainID       string
	domainName     string
	capabilityID   string
	capabilityName string
}

type strategyImportanceData struct {
	id             string
	domainID       string
	domainName     string
	capabilityID   string
	capabilityName string
	pillarID       string
	pillarName     string
	importance     int
}

type applicationFitScoreData struct {
	id            string
	componentID   string
	componentName string
	pillarID      string
	pillarName    string
	score         int
}

type domainCapabilityMetadataData struct {
	capabilityID   string
	capabilityName string
	level          string
	l1CapabilityID string
	domainID       string
	domainName     string
}

type effectiveImportanceData struct {
	capabilityID         string
	pillarID             string
	domainID             string
	importance           int
	sourceCapabilityID   string
	sourceCapabilityName string
	isInherited          bool
}

type testDataBuilder struct {
	ctx                       *strategicFitTestContext
	t                         *testing.T
	capabilities              []capabilityData
	domains                   []domainData
	components                []componentData
	realizations              []realizationData
	domainAssignments         []domainAssignmentData
	strategyImportances       []strategyImportanceData
	applicationFitScores      []applicationFitScoreData
	domainCapabilityMetadatas []domainCapabilityMetadataData
	effectiveImportances      []effectiveImportanceData
}

func newTestDataBuilder(ctx *strategicFitTestContext, t *testing.T) *testDataBuilder {
	return &testDataBuilder{ctx: ctx, t: t}
}

func (b *testDataBuilder) withCapability(id, name, level string) *testDataBuilder {
	b.capabilities = append(b.capabilities, capabilityData{id: id, name: name, level: level})
	return b
}

func (b *testDataBuilder) withCapabilityParent(id, name, level, parentID string) *testDataBuilder {
	b.capabilities = append(b.capabilities, capabilityData{id: id, name: name, level: level, parentID: parentID})
	return b
}

func (b *testDataBuilder) withDomain(id, name string) *testDataBuilder {
	b.domains = append(b.domains, domainData{id: id, name: name})
	return b
}

func (b *testDataBuilder) withComponent(id, name string) *testDataBuilder {
	b.components = append(b.components, componentData{id: id, name: name})
	return b
}

func (b *testDataBuilder) withRealization(id, capabilityID, componentID string) *testDataBuilder {
	b.realizations = append(b.realizations, realizationData{id: id, capabilityID: capabilityID, componentID: componentID})
	return b
}

func (b *testDataBuilder) withDomainAssignment(id, domainID, domainName, capabilityID, capabilityName string) *testDataBuilder {
	b.domainAssignments = append(b.domainAssignments, domainAssignmentData{
		id: id, domainID: domainID, domainName: domainName, capabilityID: capabilityID, capabilityName: capabilityName,
	})
	return b
}

func (b *testDataBuilder) withStrategyImportance(id, domainID, domainName, capabilityID, capabilityName, pillarID, pillarName string, importance int) *testDataBuilder {
	b.strategyImportances = append(b.strategyImportances, strategyImportanceData{
		id: id, domainID: domainID, domainName: domainName, capabilityID: capabilityID, capabilityName: capabilityName,
		pillarID: pillarID, pillarName: pillarName, importance: importance,
	})
	return b
}

func (b *testDataBuilder) withApplicationFitScore(id, componentID, componentName, pillarID, pillarName string, score int) *testDataBuilder {
	b.applicationFitScores = append(b.applicationFitScores, applicationFitScoreData{
		id: id, componentID: componentID, componentName: componentName, pillarID: pillarID, pillarName: pillarName, score: score,
	})
	return b
}

func (b *testDataBuilder) withDomainCapabilityMetadata(capabilityID, capabilityName, level, l1CapabilityID, domainID, domainName string) *testDataBuilder {
	b.domainCapabilityMetadatas = append(b.domainCapabilityMetadatas, domainCapabilityMetadataData{
		capabilityID: capabilityID, capabilityName: capabilityName, level: level, l1CapabilityID: l1CapabilityID, domainID: domainID, domainName: domainName,
	})
	return b
}

func (b *testDataBuilder) withEffectiveImportance(capabilityID, pillarID, domainID string, importance int, sourceCapabilityID, sourceCapabilityName string, isInherited bool) *testDataBuilder {
	b.effectiveImportances = append(b.effectiveImportances, effectiveImportanceData{
		capabilityID: capabilityID, pillarID: pillarID, domainID: domainID, importance: importance,
		sourceCapabilityID: sourceCapabilityID, sourceCapabilityName: sourceCapabilityName, isInherited: isInherited,
	})
	return b
}

func (b *testDataBuilder) build() {
	for _, c := range b.capabilities {
		b.createCapability(c)
	}
	for _, d := range b.domains {
		b.createDomain(d)
	}
	for _, c := range b.components {
		b.createComponent(c)
	}
	for _, r := range b.realizations {
		b.createRealization(r)
	}
	for _, da := range b.domainAssignments {
		b.createDomainAssignment(da)
	}
	for _, si := range b.strategyImportances {
		b.createStrategyImportance(si)
	}
	for _, afs := range b.applicationFitScores {
		b.createApplicationFitScore(afs)
	}
	for _, dcm := range b.domainCapabilityMetadatas {
		b.createDomainCapabilityMetadata(dcm)
	}
	for _, ei := range b.effectiveImportances {
		b.createEffectiveImportance(ei)
	}
}

func (b *testDataBuilder) createCapability(c capabilityData) {
	b.ctx.setTenantContext(b.t)
	var err error
	if c.parentID == "" {
		_, err = b.ctx.db.Exec(
			"INSERT INTO capabilities (id, name, description, level, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
			c.id, c.name, "", c.level, testTenantID(), "Genesis", "Active",
		)
	} else {
		_, err = b.ctx.db.Exec(
			"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
			c.id, c.name, "", c.level, c.parentID, testTenantID(), "Genesis", "Active",
		)
	}
	require.NoError(b.t, err)
	b.ctx.trackID(c.id)
}

func (b *testDataBuilder) createDomain(d domainData) {
	b.ctx.setTenantContext(b.t)
	_, err := b.ctx.db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, NOW())",
		d.id, testTenantID(), d.name, "", 0,
	)
	require.NoError(b.t, err)
	b.ctx.trackID(d.id)
}

func (b *testDataBuilder) createComponent(c componentData) {
	b.ctx.setTenantContext(b.t)
	_, err := b.ctx.db.Exec(
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, NOW())",
		c.id, testTenantID(), c.name, "",
	)
	require.NoError(b.t, err)
	b.ctx.trackID(c.id)
}

func (b *testDataBuilder) createRealization(r realizationData) {
	b.ctx.setTenantContext(b.t)
	_, err := b.ctx.db.Exec(
		"INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, component_name, realization_level, origin, linked_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		r.id, testTenantID(), r.capabilityID, r.componentID, "", "Full", "Direct",
	)
	require.NoError(b.t, err)
	b.ctx.trackID(r.id)
}

func (b *testDataBuilder) createDomainAssignment(da domainAssignmentData) {
	b.ctx.setTenantContext(b.t)
	_, err := b.ctx.db.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		da.id, testTenantID(), da.domainID, da.domainName, da.capabilityID, da.capabilityName, "L1",
	)
	require.NoError(b.t, err)
	b.ctx.trackID(da.id)
}

func (b *testDataBuilder) createStrategyImportance(si strategyImportanceData) {
	b.ctx.setTenantContext(b.t)
	labels := map[int]string{1: "Very Low", 2: "Low", 3: "Medium", 4: "High", 5: "Very High"}
	_, err := b.ctx.db.Exec(
		"INSERT INTO strategy_importance (id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, pillar_id, pillar_name, importance, importance_label, set_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())",
		si.id, testTenantID(), si.domainID, si.domainName, si.capabilityID, si.capabilityName, si.pillarID, si.pillarName, si.importance, labels[si.importance],
	)
	require.NoError(b.t, err)
	b.ctx.trackID(si.id)
}

func (b *testDataBuilder) createApplicationFitScore(afs applicationFitScoreData) {
	b.ctx.setTenantContext(b.t)
	labels := map[int]string{1: "Critical", 2: "Poor", 3: "Adequate", 4: "Good", 5: "Excellent"}
	_, err := b.ctx.db.Exec(
		"INSERT INTO application_fit_scores (id, tenant_id, component_id, component_name, pillar_id, pillar_name, score, score_label, scored_at, scored_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), $9)",
		afs.id, testTenantID(), afs.componentID, afs.componentName, afs.pillarID, afs.pillarName, afs.score, labels[afs.score], "test-user",
	)
	require.NoError(b.t, err)
	b.ctx.trackID(afs.id)
}

func (b *testDataBuilder) createDomainCapabilityMetadata(dcm domainCapabilityMetadataData) {
	b.ctx.setTenantContext(b.t)
	_, err := b.ctx.db.Exec(
		"INSERT INTO domain_capability_metadata (tenant_id, capability_id, capability_name, capability_level, l1_capability_id, business_domain_id, business_domain_name) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		testTenantID(), dcm.capabilityID, dcm.capabilityName, dcm.level, dcm.l1CapabilityID, dcm.domainID, dcm.domainName,
	)
	require.NoError(b.t, err)
	b.ctx.trackID(dcm.capabilityID)
}

func (b *testDataBuilder) createEffectiveImportance(ei effectiveImportanceData) {
	b.ctx.setTenantContext(b.t)
	labels := map[int]string{1: "Very Low", 2: "Low", 3: "Medium", 4: "High", 5: "Very High"}
	_, err := b.ctx.db.Exec(
		"INSERT INTO effective_capability_importance (tenant_id, capability_id, pillar_id, business_domain_id, effective_importance, importance_label, source_capability_id, source_capability_name, is_inherited, computed_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())",
		testTenantID(), ei.capabilityID, ei.pillarID, ei.domainID, ei.importance, labels[ei.importance], ei.sourceCapabilityID, ei.sourceCapabilityName, ei.isInherited,
	)
	require.NoError(b.t, err)
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

type strategicFitTestHarness struct {
	testCtx     *strategicFitTestContext
	mockGateway *mockStrategyPillarsGateway
	handlers    *StrategicFitAnalysisHandlers
	scsManager  *scs.SessionManager
	cleanup     func()
}

func setupStrategicFitTestHarness(t *testing.T) *strategicFitTestHarness {
	testCtx, cleanup := setupStrategicFitTestDB(t)
	mockGateway := newMockPillarsGateway()
	handlers, scsManager := setupStrategicFitHandlers(testCtx.db, mockGateway)
	return &strategicFitTestHarness{
		testCtx:     testCtx,
		mockGateway: mockGateway,
		handlers:    handlers,
		scsManager:  scsManager,
		cleanup:     cleanup,
	}
}

func (h *strategicFitTestHarness) executeRequest(t *testing.T, pillarID string) (*httptest.ResponseRecorder, StrategicFitAnalysisResponse) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategic-fit-analysis/"+pillarID, nil)
	req = withTestTenant(req)
	req = withAuthenticatedSession(req, h.scsManager)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pillarId", pillarID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	router := chi.NewRouter()
	router.Use(h.scsManager.LoadAndSave)
	router.Get("/api/v1/strategic-fit-analysis/{pillarId}", h.handlers.GetStrategicFitAnalysis)
	router.ServeHTTP(w, req)

	var response StrategicFitAnalysisResponse
	if w.Code == http.StatusOK {
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
	}
	return w, response
}

func TestGetStrategicFitAnalysis_ErrorCases_Integration(t *testing.T) {
	tests := []struct {
		name               string
		pillarID           string
		pillarName         string
		fitScoringEnabled  bool
		addPillar          bool
		expectedStatusCode int
	}{
		{
			name:               "fit scoring disabled returns bad request",
			pillarID:           fmt.Sprintf("test-pillar-%d", time.Now().UnixNano()),
			pillarName:         "Transform",
			fitScoringEnabled:  false,
			addPillar:          true,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "pillar not found returns not found",
			pillarID:           uuid.New().String(),
			addPillar:          false,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "invalid pillar ID returns bad request",
			pillarID:           "invalid-uuid",
			addPillar:          false,
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := setupStrategicFitTestHarness(t)
			defer h.cleanup()

			if tc.addPillar {
				h.mockGateway.addPillar(tc.pillarID, tc.pillarName, tc.fitScoringEnabled)
			}

			w, _ := h.executeRequest(t, tc.pillarID)
			assert.Equal(t, tc.expectedStatusCode, w.Code)
		})
	}
}

func TestGetStrategicFitAnalysis_WithDirectRating_Integration(t *testing.T) {
	h := setupStrategicFitTestHarness(t)
	defer h.cleanup()

	pillarID := uuid.New().String()
	pillarName := "Always On"
	h.mockGateway.addPillar(pillarID, pillarName, true)

	capabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()

	newTestDataBuilder(h.testCtx, t).
		withCapability(capabilityID, "Customer Onboarding", "L1").
		withDomain(domainID, "Customer Management").
		withComponent(componentID, "CRM System").
		withRealization(uuid.New().String(), capabilityID, componentID).
		withDomainAssignment(uuid.New().String(), domainID, "Customer Management", capabilityID, "Customer Onboarding").
		withDomainCapabilityMetadata(capabilityID, "Customer Onboarding", "L1", capabilityID, domainID, "Customer Management").
		withStrategyImportance(uuid.New().String(), domainID, "Customer Management", capabilityID, "Customer Onboarding", pillarID, pillarName, 5).
		withEffectiveImportance(capabilityID, pillarID, domainID, 5, capabilityID, "Customer Onboarding", false).
		withApplicationFitScore(uuid.New().String(), componentID, "CRM System", pillarID, pillarName, 2).
		build()

	w, response := h.executeRequest(t, pillarID)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())
	assert.Equal(t, pillarID, response.PillarID)
	assert.Equal(t, pillarName, response.PillarName)
	assert.Equal(t, 1, response.Summary.ScoredRealizations)
	assert.Equal(t, 1, response.Summary.LiabilityCount)
	require.Len(t, response.Liabilities, 1)

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

func TestGetStrategicFitAnalysis_InheritedFromParent_Integration(t *testing.T) {
	h := setupStrategicFitTestHarness(t)
	defer h.cleanup()

	pillarID := uuid.New().String()
	pillarName := "Always On"
	h.mockGateway.addPillar(pillarID, pillarName, true)

	l1CapabilityID := uuid.New().String()
	l2CapabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()

	newTestDataBuilder(h.testCtx, t).
		withCapability(l1CapabilityID, "Payment Processing", "L1").
		withCapabilityParent(l2CapabilityID, "Card Payments", "L2", l1CapabilityID).
		withDomain(domainID, "Finance").
		withComponent(componentID, "Payment Gateway").
		withRealization(uuid.New().String(), l2CapabilityID, componentID).
		withDomainAssignment(uuid.New().String(), domainID, "Finance", l1CapabilityID, "Payment Processing").
		withDomainCapabilityMetadata(l2CapabilityID, "Card Payments", "L2", l1CapabilityID, domainID, "Finance").
		withStrategyImportance(uuid.New().String(), domainID, "Finance", l1CapabilityID, "Payment Processing", pillarID, pillarName, 4).
		withEffectiveImportance(l2CapabilityID, pillarID, domainID, 4, l1CapabilityID, "Payment Processing", true).
		withApplicationFitScore(uuid.New().String(), componentID, "Payment Gateway", pillarID, pillarName, 2).
		build()

	w, response := h.executeRequest(t, pillarID)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())
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

func TestGetStrategicFitAnalysis_MultipleCapabilitiesInSameChain_Integration(t *testing.T) {
	h := setupStrategicFitTestHarness(t)
	defer h.cleanup()

	pillarID := uuid.New().String()
	pillarName := "Always On"
	h.mockGateway.addPillar(pillarID, pillarName, true)

	l1CapabilityID := uuid.New().String()
	l2CapabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()

	newTestDataBuilder(h.testCtx, t).
		withCapability(l1CapabilityID, "Payment Processing", "L1").
		withCapabilityParent(l2CapabilityID, "Card Payments", "L2", l1CapabilityID).
		withDomain(domainID, "Finance").
		withComponent(componentID, "Payment Gateway").
		withRealization(uuid.New().String(), l1CapabilityID, componentID).
		withRealization(uuid.New().String(), l2CapabilityID, componentID).
		withDomainAssignment(uuid.New().String(), domainID, "Finance", l1CapabilityID, "Payment Processing").
		withDomainCapabilityMetadata(l1CapabilityID, "Payment Processing", "L1", l1CapabilityID, domainID, "Finance").
		withDomainCapabilityMetadata(l2CapabilityID, "Card Payments", "L2", l1CapabilityID, domainID, "Finance").
		withStrategyImportance(uuid.New().String(), domainID, "Finance", l1CapabilityID, "Payment Processing", pillarID, pillarName, 4).
		withStrategyImportance(uuid.New().String(), domainID, "Finance", l2CapabilityID, "Card Payments", pillarID, pillarName, 5).
		withEffectiveImportance(l1CapabilityID, pillarID, domainID, 4, l1CapabilityID, "Payment Processing", false).
		withEffectiveImportance(l2CapabilityID, pillarID, domainID, 5, l2CapabilityID, "Card Payments", false).
		withApplicationFitScore(uuid.New().String(), componentID, "Payment Gateway", pillarID, pillarName, 2).
		build()

	w, response := h.executeRequest(t, pillarID)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())
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

func TestGetStrategicFitAnalysis_NoRatingInHierarchy_NoGapEntry_Integration(t *testing.T) {
	h := setupStrategicFitTestHarness(t)
	defer h.cleanup()

	pillarID := uuid.New().String()
	pillarName := "Transform"
	h.mockGateway.addPillar(pillarID, pillarName, true)

	l1CapabilityID := uuid.New().String()
	l2CapabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()

	newTestDataBuilder(h.testCtx, t).
		withCapability(l1CapabilityID, "Support Operations", "L1").
		withCapabilityParent(l2CapabilityID, "Ticket Management", "L2", l1CapabilityID).
		withDomain(domainID, "Operations").
		withComponent(componentID, "Helpdesk System").
		withRealization(uuid.New().String(), l2CapabilityID, componentID).
		withDomainAssignment(uuid.New().String(), domainID, "Operations", l1CapabilityID, "Support Operations").
		withDomainCapabilityMetadata(l2CapabilityID, "Ticket Management", "L2", l1CapabilityID, domainID, "Operations").
		withApplicationFitScore(uuid.New().String(), componentID, "Helpdesk System", pillarID, pillarName, 4).
		build()

	w, response := h.executeRequest(t, pillarID)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())
	assert.Equal(t, 0, response.Summary.ScoredRealizations)
	assert.Empty(t, response.Liabilities)
	assert.Empty(t, response.Concerns)
	assert.Empty(t, response.Aligned)
}
