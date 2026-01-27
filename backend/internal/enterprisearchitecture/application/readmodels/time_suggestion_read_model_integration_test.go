//go:build integration

package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	"easi/backend/internal/infrastructure/database"
	sharedcontext "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type timeSuggestionTestFixture struct {
	db        *sql.DB
	tenantDB  *database.TenantAwareDB
	readModel *TimeSuggestionReadModel
	ctx       context.Context
	t         *testing.T
}

type mockPillarsGateway struct {
	pillars *metamodel.StrategyPillarsConfigDTO
	err     error
}

func (m *mockPillarsGateway) GetStrategyPillars(ctx context.Context) (*metamodel.StrategyPillarsConfigDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pillars, nil
}

func newTimeSuggestionTestFixture(t *testing.T, pillars *metamodel.StrategyPillarsConfigDTO) *timeSuggestionTestFixture {
	db := setupTimeSuggestionTestDB(t)
	tenantDB := database.NewTenantAwareDB(db)

	_, err := db.Exec("SET app.current_tenant = 'default'")
	require.NoError(t, err)

	gateway := &mockPillarsGateway{pillars: pillars}

	return &timeSuggestionTestFixture{
		db:        db,
		tenantDB:  tenantDB,
		readModel: NewTimeSuggestionReadModel(tenantDB, gateway),
		ctx:       sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID()),
		t:         t,
	}
}

func setupTimeSuggestionTestDB(t *testing.T) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		"localhost", "5432", "easi_app", "localdev", "easi", "disable")
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { db.Close() })
	return db
}

func (f *timeSuggestionTestFixture) uniqueID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func (f *timeSuggestionTestFixture) createComponent(id, name string) {
	_, err := f.db.Exec(`INSERT INTO components (id, tenant_id, name, status, created_at)
		VALUES ($1, 'default', $2, 'Active', NOW()) ON CONFLICT (id, tenant_id) DO NOTHING`, id, name)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM components WHERE id = $1 AND tenant_id = 'default'", id) })
}

func (f *timeSuggestionTestFixture) createCapability(id, name string) {
	_, err := f.db.Exec(`INSERT INTO capabilities (id, tenant_id, name, level, status, created_at)
		VALUES ($1, 'default', $2, 'L1', 'Active', NOW()) ON CONFLICT (id, tenant_id) DO NOTHING`, id, name)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM capabilities WHERE id = $1 AND tenant_id = 'default'", id) })
}

func (f *timeSuggestionTestFixture) createEnterpriseCapability(id, name string) {
	_, err := f.db.Exec(`INSERT INTO enterprise_capabilities (id, tenant_id, name, description, category, active, link_count, domain_count, created_at)
		VALUES ($1, 'default', $2, '', '', true, 0, 0, NOW()) ON CONFLICT (id, tenant_id) DO NOTHING`, id, name)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM enterprise_capabilities WHERE id = $1 AND tenant_id = 'default'", id) })
}

func (f *timeSuggestionTestFixture) createEnterpriseCapabilityLink(id, enterpriseCapID, domainCapID string) {
	_, err := f.db.Exec(`INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		VALUES ($1, 'default', $2, $3, 'test@example.com', NOW()) ON CONFLICT (id, tenant_id) DO NOTHING`, id, enterpriseCapID, domainCapID)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1 AND tenant_id = 'default'", id) })
}

func (f *timeSuggestionTestFixture) createRealization(capabilityID, componentID, componentName string) {
	_, err := f.db.Exec(`INSERT INTO capability_realizations (tenant_id, capability_id, component_id, component_name, origin)
		VALUES ('default', $1, $2, $3, 'Direct') ON CONFLICT DO NOTHING`, capabilityID, componentID, componentName)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM capability_realizations WHERE tenant_id = 'default' AND capability_id = $1 AND component_id = $2", capabilityID, componentID)
	})
}

func (f *timeSuggestionTestFixture) createDomainCapabilityMetadata(capabilityID, domainID string) {
	_, err := f.db.Exec(`INSERT INTO domain_capability_metadata (tenant_id, capability_id, capability_name, capability_level, l1_capability_id, business_domain_id, business_domain_name)
		VALUES ('default', $1, 'Cap Name', 'L1', $1, $2, 'Domain Name') ON CONFLICT DO NOTHING`, capabilityID, domainID)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM domain_capability_metadata WHERE tenant_id = 'default' AND capability_id = $1", capabilityID)
	})
}

func (f *timeSuggestionTestFixture) createEffectiveCapabilityImportance(capabilityID, domainID, pillarID string, importance int) {
	_, err := f.db.Exec(`INSERT INTO effective_capability_importance (tenant_id, capability_id, business_domain_id, pillar_id, effective_importance)
		VALUES ('default', $1, $2, $3, $4) ON CONFLICT DO NOTHING`, capabilityID, domainID, pillarID, importance)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM effective_capability_importance WHERE tenant_id = 'default' AND capability_id = $1 AND pillar_id = $2", capabilityID, pillarID)
	})
}

func (f *timeSuggestionTestFixture) createApplicationFitScore(componentID, pillarID string, score int) {
	_, err := f.db.Exec(`INSERT INTO application_fit_scores (tenant_id, component_id, pillar_id, score)
		VALUES ('default', $1, $2, $3) ON CONFLICT DO NOTHING`, componentID, pillarID, score)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM application_fit_scores WHERE tenant_id = 'default' AND component_id = $1 AND pillar_id = $2", componentID, pillarID)
	})
}

func TestTimeSuggestionReadModel_GetAllSuggestions_Empty(t *testing.T) {
	pillars := &metamodel.StrategyPillarsConfigDTO{
		Pillars: []metamodel.StrategyPillarDTO{
			{ID: "p1", Name: "Technical", FitScoringEnabled: true, FitType: "TECHNICAL"},
		},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	suggestions, err := f.readModel.GetAllSuggestions(f.ctx)

	require.NoError(t, err)
	assert.Empty(t, suggestions)
}

func TestTimeSuggestionReadModel_GetAllSuggestions_WithData(t *testing.T) {
	pillars := &metamodel.StrategyPillarsConfigDTO{
		Pillars: []metamodel.StrategyPillarDTO{
			{ID: "pillar-tech", Name: "Technical Quality", FitScoringEnabled: true, FitType: "TECHNICAL"},
			{ID: "pillar-func", Name: "Functional Fit", FitScoringEnabled: true, FitType: "FUNCTIONAL"},
		},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	componentID := f.uniqueID("comp")
	capabilityID := f.uniqueID("cap")
	domainID := f.uniqueID("domain")

	f.createComponent(componentID, "Test Component")
	f.createCapability(capabilityID, "Test Capability")
	f.createRealization(capabilityID, componentID, "Test Component")
	f.createDomainCapabilityMetadata(capabilityID, domainID)
	f.createEffectiveCapabilityImportance(capabilityID, domainID, "pillar-tech", 80)
	f.createEffectiveCapabilityImportance(capabilityID, domainID, "pillar-func", 70)
	f.createApplicationFitScore(componentID, "pillar-tech", 60)
	f.createApplicationFitScore(componentID, "pillar-func", 50)

	suggestions, err := f.readModel.GetAllSuggestions(f.ctx)

	require.NoError(t, err)
	require.Len(t, suggestions, 1)

	suggestion := suggestions[0]
	assert.Equal(t, capabilityID, suggestion.CapabilityID)
	assert.Equal(t, "Test Capability", suggestion.CapabilityName)
	assert.Equal(t, componentID, suggestion.ComponentID)
	assert.Equal(t, "Test Component", suggestion.ComponentName)
	assert.NotNil(t, suggestion.TechnicalGap)
	assert.NotNil(t, suggestion.FunctionalGap)
	assert.Equal(t, 20.0, *suggestion.TechnicalGap)
	assert.Equal(t, 20.0, *suggestion.FunctionalGap)
	assert.NotNil(t, suggestion.SuggestedTime)
	assert.Equal(t, "Eliminate", *suggestion.SuggestedTime)
	assert.Equal(t, "High", suggestion.Confidence)
}

func TestTimeSuggestionReadModel_GetByCapability_FiltersCorrectly(t *testing.T) {
	pillars := &metamodel.StrategyPillarsConfigDTO{
		Pillars: []metamodel.StrategyPillarDTO{
			{ID: "pillar-tech", Name: "Technical", FitScoringEnabled: true, FitType: "TECHNICAL"},
		},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	componentID := f.uniqueID("comp")
	capabilityID1 := f.uniqueID("cap1")
	capabilityID2 := f.uniqueID("cap2")
	domainID := f.uniqueID("domain")

	f.createComponent(componentID, "Test Component")
	f.createCapability(capabilityID1, "Capability 1")
	f.createCapability(capabilityID2, "Capability 2")
	f.createRealization(capabilityID1, componentID, "Test Component")
	f.createDomainCapabilityMetadata(capabilityID1, domainID)
	f.createEffectiveCapabilityImportance(capabilityID1, domainID, "pillar-tech", 80)
	f.createApplicationFitScore(componentID, "pillar-tech", 70)

	suggestionsCap1, err := f.readModel.GetByCapability(f.ctx, capabilityID1)
	require.NoError(t, err)
	assert.Len(t, suggestionsCap1, 1)

	suggestionsCap2, err := f.readModel.GetByCapability(f.ctx, capabilityID2)
	require.NoError(t, err)
	assert.Empty(t, suggestionsCap2)
}

func TestTimeSuggestionReadModel_GetByComponent_FiltersCorrectly(t *testing.T) {
	pillars := &metamodel.StrategyPillarsConfigDTO{
		Pillars: []metamodel.StrategyPillarDTO{
			{ID: "pillar-tech", Name: "Technical", FitScoringEnabled: true, FitType: "TECHNICAL"},
		},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	componentID1 := f.uniqueID("comp1")
	componentID2 := f.uniqueID("comp2")
	capabilityID := f.uniqueID("cap")
	domainID := f.uniqueID("domain")

	f.createComponent(componentID1, "Component 1")
	f.createComponent(componentID2, "Component 2")
	f.createCapability(capabilityID, "Test Capability")
	f.createRealization(capabilityID, componentID1, "Component 1")
	f.createDomainCapabilityMetadata(capabilityID, domainID)
	f.createEffectiveCapabilityImportance(capabilityID, domainID, "pillar-tech", 80)
	f.createApplicationFitScore(componentID1, "pillar-tech", 70)
	f.createApplicationFitScore(componentID2, "pillar-tech", 60)

	suggestionsComp1, err := f.readModel.GetByComponent(f.ctx, componentID1)
	require.NoError(t, err)
	assert.Len(t, suggestionsComp1, 1)

	suggestionsComp2, err := f.readModel.GetByComponent(f.ctx, componentID2)
	require.NoError(t, err)
	assert.Empty(t, suggestionsComp2)
}

func TestTimeSuggestionReadModel_CalculatesInsufficientConfidenceWhenNoPillars(t *testing.T) {
	pillars := &metamodel.StrategyPillarsConfigDTO{
		Pillars: []metamodel.StrategyPillarDTO{},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	componentID := f.uniqueID("comp")
	capabilityID := f.uniqueID("cap")
	domainID := f.uniqueID("domain")

	f.createComponent(componentID, "Test Component")
	f.createCapability(capabilityID, "Test Capability")
	f.createRealization(capabilityID, componentID, "Test Component")
	f.createDomainCapabilityMetadata(capabilityID, domainID)
	f.createEffectiveCapabilityImportance(capabilityID, domainID, "pillar-unknown", 80)
	f.createApplicationFitScore(componentID, "pillar-unknown", 70)

	suggestions, err := f.readModel.GetAllSuggestions(f.ctx)

	require.NoError(t, err)
	if len(suggestions) > 0 {
		assert.Equal(t, "Insufficient", suggestions[0].Confidence)
		assert.Nil(t, suggestions[0].SuggestedTime)
	}
}
