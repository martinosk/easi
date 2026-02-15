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

	realizationID := uuid.New().String()
	_, err := f.db.Exec(`INSERT INTO ea_realization_cache (tenant_id, realization_id, capability_id, component_id, component_name, origin)
		VALUES ('default', $1, $2, $3, $4, 'Direct') ON CONFLICT DO NOTHING`,
		realizationID, data.CapabilityID, data.ComponentID, data.ComponentName)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM ea_realization_cache WHERE tenant_id = 'default' AND realization_id = $1", realizationID)
	})

	_, err = f.db.Exec(`INSERT INTO domain_capability_metadata (tenant_id, capability_id, capability_name, capability_level, l1_capability_id, business_domain_id, business_domain_name)
		VALUES ('default', $1, 'Cap Name', 'L1', $1, $2, 'Domain Name') ON CONFLICT DO NOTHING`,
		data.CapabilityID, data.DomainID)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM domain_capability_metadata WHERE tenant_id = 'default' AND capability_id = $1", data.CapabilityID)
	})

	for _, ps := range data.PillarScores {
		ps := ps
		_, err = f.db.Exec(`INSERT INTO ea_importance_cache (tenant_id, capability_id, business_domain_id, pillar_id, effective_importance)
			VALUES ('default', $1, $2, $3, $4) ON CONFLICT DO NOTHING`,
			data.CapabilityID, data.DomainID, ps.PillarID, ps.Importance)
		require.NoError(f.t, err)
		f.t.Cleanup(func() {
			f.db.Exec("DELETE FROM ea_importance_cache WHERE tenant_id = 'default' AND capability_id = $1 AND pillar_id = $2",
				data.CapabilityID, ps.PillarID)
		})

		_, err = f.db.Exec(`INSERT INTO ea_fit_score_cache (tenant_id, component_id, pillar_id, score, rationale)
			VALUES ('default', $1, $2, $3, '') ON CONFLICT DO NOTHING`,
			data.ComponentID, ps.PillarID, ps.FitScore)
		require.NoError(f.t, err)
		f.t.Cleanup(func() {
			f.db.Exec("DELETE FROM ea_fit_score_cache WHERE tenant_id = 'default' AND component_id = $1 AND pillar_id = $2",
				data.ComponentID, ps.PillarID)
		})
	}
}

func TestTimeSuggestionReadModel_GetAllSuggestions_Empty(t *testing.T) {
	pillars := &mmPL.StrategyPillarsConfigDTO{
		Pillars: []mmPL.StrategyPillarDTO{
			{ID: "p1", Name: "Technical", FitScoringEnabled: true, FitType: "TECHNICAL"},
		},
	}
	f := newTimeSuggestionTestFixture(t, pillars)

	suggestions, err := f.readModel.GetAllSuggestions(f.ctx)

	require.NoError(t, err)
	assert.Empty(t, suggestions)
}

func TestTimeSuggestionReadModel_GetAllSuggestions_WithData(t *testing.T) {
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

	suggestions, err := f.readModel.GetAllSuggestions(f.ctx)

	require.NoError(t, err)
	require.Len(t, suggestions, 1)

	suggestion := suggestions[0]
	assert.Equal(t, capabilityID, suggestion.CapabilityID)
	assert.Equal(t, "Cap Name", suggestion.CapabilityName)
	assert.Equal(t, componentID, suggestion.ComponentID)
	assert.Equal(t, "Test Component", suggestion.ComponentName)
	assert.NotNil(t, suggestion.TechnicalGap)
	assert.NotNil(t, suggestion.FunctionalGap)
	assert.Equal(t, 20.0, *suggestion.TechnicalGap)
	assert.Equal(t, 20.0, *suggestion.FunctionalGap)
	assert.NotNil(t, suggestion.SuggestedTime)
	assert.Equal(t, "ELIMINATE", *suggestion.SuggestedTime)
	assert.Equal(t, "MEDIUM", suggestion.Confidence)
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

	tests := []struct {
		name     string
		query    func(context.Context, string) ([]TimeSuggestionDTO, error)
		linkedID string
	}{
		{"GetByCapability", f.readModel.GetByCapability, capabilityID},
		{"GetByComponent", f.readModel.GetByComponent, componentID},
	}

	for _, tt := range tests {
		t.Run(tt.name+" filters correctly", func(t *testing.T) {
			linked, err := tt.query(f.ctx, tt.linkedID)
			require.NoError(t, err)
			assert.Len(t, linked, 1)

			unlinked, err := tt.query(f.ctx, uuid.New().String())
			require.NoError(t, err)
			assert.Empty(t, unlinked)
		})
	}
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

	suggestions, err := f.readModel.GetAllSuggestions(f.ctx)

	require.NoError(t, err)
	if len(suggestions) > 0 {
		assert.Equal(t, "LOW", suggestions[0].Confidence)
		assert.Nil(t, suggestions[0].SuggestedTime)
	}
}
