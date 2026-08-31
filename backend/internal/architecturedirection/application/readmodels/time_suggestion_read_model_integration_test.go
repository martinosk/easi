//go:build integration

package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"easi/backend/internal/infrastructure/database"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	sharedcontext "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type timeSuggestionTestFixture struct {
	db        *sql.DB
	readModel *TimeSuggestionReadModel
	ctx       context.Context
	t         *testing.T
}

type mockPillarsGateway struct {
	pillars *mmPL.StrategyPillarsConfigDTO
	err     error
}

func (m *mockPillarsGateway) GetStrategyPillars(ctx context.Context) (*mmPL.StrategyPillarsConfigDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pillars, nil
}

func (m *mockPillarsGateway) GetActivePillar(ctx context.Context, pillarID string) (*mmPL.StrategyPillarDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.pillars != nil {
		for _, p := range m.pillars.Pillars {
			if p.ID == pillarID && p.Active {
				return &p, nil
			}
		}
	}
	return nil, nil
}

func (m *mockPillarsGateway) InvalidateCache(tenantID string) {}

func newTimeSuggestionTestFixture(t *testing.T, pillars *mmPL.StrategyPillarsConfigDTO) *timeSuggestionTestFixture {
	db := setupTimeSuggestionTestDB(t)
	tenantDB := database.NewTenantAwareDB(db)

	_, err := db.Exec("SET app.current_tenant = 'default'")
	require.NoError(t, err)

	gateway := &mockPillarsGateway{pillars: pillars}

	return &timeSuggestionTestFixture{
		db:        db,
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

type pillarTestScore struct {
	PillarID   string
	Importance int
	FitScore   int
}

type suggestionSeedData struct {
	CapabilityID  string
	ComponentID   string
	ComponentName string
	DomainID      string
	PillarScores  []pillarTestScore
}

func (f *timeSuggestionTestFixture) seedSuggestionData(data suggestionSeedData) {
	f.t.Helper()

	f.seedRealization(data)
	f.seedCapabilityNode(data)
	f.seedComponentName(data)

	for _, ps := range data.PillarScores {
		for _, row := range pillarScopedRowsFor(data, ps) {
			f.seedPillarScopedRow(row)
		}
	}
}

func (f *timeSuggestionTestFixture) seedRealization(data suggestionSeedData) {
	realizationID := uuid.New().String()
	_, err := f.db.Exec(`INSERT INTO architecturedirection.realization_cache (tenant_id, realization_id, capability_id, component_id)
		VALUES ('default', $1, $2, $3) ON CONFLICT DO NOTHING`,
		realizationID, data.CapabilityID, data.ComponentID)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM architecturedirection.realization_cache WHERE tenant_id = 'default' AND realization_id = $1", realizationID)
	})
}

func (f *timeSuggestionTestFixture) seedCapabilityNode(data suggestionSeedData) {
	_, err := f.db.Exec(`INSERT INTO architecturedirection.capability_node_cache (tenant_id, capability_id, capability_name, capability_level, l1_capability_id, business_domain_id, business_domain_name, maturity_value)
		VALUES ('default', $1, 'Cap Name', 'L1', $1, $2, 'Domain Name', 12) ON CONFLICT DO NOTHING`,
		data.CapabilityID, data.DomainID)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM architecturedirection.capability_node_cache WHERE tenant_id = 'default' AND capability_id = $1", data.CapabilityID)
	})
}

func (f *timeSuggestionTestFixture) seedComponentName(data suggestionSeedData) {
	_, err := f.db.Exec(`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
		VALUES ('default', 'application', $1, $2) ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name`,
		data.ComponentID, data.ComponentName)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM architecturedirection.reference_name_cache WHERE tenant_id = 'default' AND entity_type = 'application' AND entity_id = $1", data.ComponentID)
	})
}

type pillarScopedRow struct {
	insert   string
	values   []any
	remove   string
	ownerID  string
	pillarID string
}

func pillarScopedRowsFor(data suggestionSeedData, ps pillarTestScore) []pillarScopedRow {
	return []pillarScopedRow{
		{
			insert: `INSERT INTO architecturedirection.ea_importance_cache (tenant_id, capability_id, business_domain_id, pillar_id, effective_importance)
				VALUES ('default', $1, $2, $3, $4) ON CONFLICT DO NOTHING`,
			values:   []any{data.CapabilityID, data.DomainID, ps.PillarID, ps.Importance},
			remove:   `DELETE FROM architecturedirection.ea_importance_cache WHERE tenant_id = 'default' AND capability_id = $1 AND pillar_id = $2`,
			ownerID:  data.CapabilityID,
			pillarID: ps.PillarID,
		},
		{
			insert: `INSERT INTO architecturedirection.ea_fit_score_cache (tenant_id, component_id, pillar_id, score, rationale)
				VALUES ('default', $1, $2, $3, '') ON CONFLICT DO NOTHING`,
			values:   []any{data.ComponentID, ps.PillarID, ps.FitScore},
			remove:   `DELETE FROM architecturedirection.ea_fit_score_cache WHERE tenant_id = 'default' AND component_id = $1 AND pillar_id = $2`,
			ownerID:  data.ComponentID,
			pillarID: ps.PillarID,
		},
	}
}

func (f *timeSuggestionTestFixture) seedPillarScopedRow(row pillarScopedRow) {
	_, err := f.db.Exec(row.insert, row.values...)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec(row.remove, row.ownerID, row.pillarID)
	})
}

func TestTimeSuggestionReadModel_All_Empty(t *testing.T) {
	pillars := &mmPL.StrategyPillarsConfigDTO{
		Pillars: []mmPL.StrategyPillarDTO{
			{ID: "p1", Name: "Technical", FitScoringEnabled: true, FitType: "TECHNICAL"},
		},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	suggestions, err := f.readModel.All(f.ctx)

	require.NoError(t, err)
	assert.Empty(t, suggestions)
}

func TestTimeSuggestionReadModel_All_WithData(t *testing.T) {
	pillars := &mmPL.StrategyPillarsConfigDTO{
		Pillars: []mmPL.StrategyPillarDTO{
			{ID: "pillar-tech", Name: "Technical Quality", FitScoringEnabled: true, FitType: "TECHNICAL"},
			{ID: "pillar-func", Name: "Functional Fit", FitScoringEnabled: true, FitType: "FUNCTIONAL"},
		},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	componentID := uuid.New().String()
	capabilityID := uuid.New().String()
	domainID := uuid.New().String()

	f.seedSuggestionData(suggestionSeedData{
		CapabilityID:  capabilityID,
		ComponentID:   componentID,
		ComponentName: "Test Component",
		DomainID:      domainID,
		PillarScores: []pillarTestScore{
			{PillarID: "pillar-tech", Importance: 80, FitScore: 60},
			{PillarID: "pillar-func", Importance: 70, FitScore: 50},
		},
	})

	suggestions, err := f.readModel.All(f.ctx)

	require.NoError(t, err)
	require.Len(t, suggestions, 1)

	realization := suggestions[0]
	assert.Equal(t, capabilityID, realization.Pair.CapabilityID)
	assert.Equal(t, "Cap Name", realization.CapabilityName)
	assert.Equal(t, componentID, realization.Pair.ComponentID)
	assert.Equal(t, "Test Component", realization.ComponentName)
	require.NotNil(t, realization.Suggestion.TechnicalGap)
	require.NotNil(t, realization.Suggestion.FunctionalGap)
	assert.Equal(t, 20.0, *realization.Suggestion.TechnicalGap)
	assert.Equal(t, 20.0, *realization.Suggestion.FunctionalGap)
	require.NotNil(t, realization.Suggestion.Grade)
	assert.Equal(t, "Eliminate", *realization.Suggestion.Grade)
	assert.Equal(t, "MEDIUM", realization.Suggestion.Confidence)
}

func TestTimeSuggestionReadModel_FilterMethods(t *testing.T) {
	pillars := &mmPL.StrategyPillarsConfigDTO{
		Pillars: []mmPL.StrategyPillarDTO{
			{ID: "pillar-tech", Name: "Technical", FitScoringEnabled: true, FitType: "TECHNICAL"},
		},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	capabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()

	f.seedSuggestionData(suggestionSeedData{
		CapabilityID:  capabilityID,
		ComponentID:   componentID,
		ComponentName: "Test Component",
		DomainID:      domainID,
		PillarScores:  []pillarTestScore{{PillarID: "pillar-tech", Importance: 80, FitScore: 70}},
	})

	t.Run("ForCapabilities filters correctly", func(t *testing.T) {
		linked, err := f.readModel.ForCapabilities(f.ctx, []string{capabilityID})
		require.NoError(t, err)
		assert.Len(t, linked, 1)

		unlinked, err := f.readModel.ForCapabilities(f.ctx, []string{uuid.New().String()})
		require.NoError(t, err)
		assert.Empty(t, unlinked)
	})

	t.Run("ForPair returns the pair's suggestion only", func(t *testing.T) {
		linked, err := f.readModel.ForPair(f.ctx, capabilityID, componentID)
		require.NoError(t, err)
		require.NotNil(t, linked)
		assert.Equal(t, "LOW", linked.Confidence)

		unlinked, err := f.readModel.ForPair(f.ctx, capabilityID, uuid.New().String())
		require.NoError(t, err)
		assert.Nil(t, unlinked)
	})
}

func TestTimeSuggestionReadModel_CalculatesInsufficientConfidenceWhenNoPillars(t *testing.T) {
	pillars := &mmPL.StrategyPillarsConfigDTO{
		Pillars: []mmPL.StrategyPillarDTO{},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	capabilityID := uuid.New().String()
	componentID := uuid.New().String()
	domainID := uuid.New().String()

	f.seedSuggestionData(suggestionSeedData{
		CapabilityID:  capabilityID,
		ComponentID:   componentID,
		ComponentName: "Test Component",
		DomainID:      domainID,
		PillarScores:  []pillarTestScore{{PillarID: "pillar-unknown", Importance: 80, FitScore: 70}},
	})

	suggestions, err := f.readModel.All(f.ctx)

	require.NoError(t, err)
	if len(suggestions) > 0 {
		assert.Equal(t, "LOW", suggestions[0].Suggestion.Confidence)
		assert.Nil(t, suggestions[0].Suggestion.Grade)
	}
}
