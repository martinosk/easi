//go:build integration

package readmodels

import (
	"testing"
	"time"

	mmPL "easi/backend/internal/metamodel/publishedlanguage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gapPillars() *mmPL.StrategyPillarsConfigDTO {
	return &mmPL.StrategyPillarsConfigDTO{
		Pillars: []mmPL.StrategyPillarDTO{
			{ID: "pillar-tech", Name: "Technical Quality", FitScoringEnabled: true, FitType: "TECHNICAL"},
			{ID: "pillar-func", Name: "Functional Fit", FitScoringEnabled: true, FitType: "FUNCTIONAL"},
		},
	}
}

func highGapScores() []pillarTestScore {
	return []pillarTestScore{
		{PillarID: "pillar-tech", Importance: 80, FitScore: 60},
		{PillarID: "pillar-func", Importance: 70, FitScore: 50},
	}
}

func (f *timeSuggestionTestFixture) recordAssessment(assessments *TimeAssessmentReadModel, capabilityID, componentID, grade string) {
	f.t.Helper()
	err := assessments.UpsertCurrent(f.ctx, UpsertTimeAssessmentParams{
		ID:            uuid.New().String(),
		CapabilityID:  capabilityID,
		ComponentID:   componentID,
		RealizationID: uuid.New().String(),
		Grade:         grade,
		AssessedBy:    "architect@example.com",
		AssessedAt:    time.Now().UTC(),
	})
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM architecturedirection.time_assessments WHERE tenant_id = 'default' AND capability_id = $1 AND component_id = $2",
			capabilityID, componentID)
	})
}

func TestTimeAssessmentView_ComposesSuggestionsOverRealData(t *testing.T) {
	f := newTimeSuggestionTestFixture(t, gapPillars())
	assessments := NewTimeAssessmentReadModel(f.readModel.db)
	view := NewTimeAssessmentView(assessments, f.readModel)

	assessedCapability, assessedComponent := uuid.New().String(), uuid.New().String()
	unassessedCapability, unassessedComponent := uuid.New().String(), uuid.New().String()

	f.seedSuggestionData(suggestionSeedData{
		CapabilityID:  assessedCapability,
		ComponentID:   assessedComponent,
		ComponentName: "Seabook",
		DomainID:      uuid.New().String(),
		PillarScores:  highGapScores(),
	})
	f.seedSuggestionData(suggestionSeedData{
		CapabilityID:  unassessedCapability,
		ComponentID:   unassessedComponent,
		ComponentName: "Phoenix",
		DomainID:      uuid.New().String(),
		PillarScores:  highGapScores(),
	})
	f.recordAssessment(assessments, assessedCapability, assessedComponent, "Tolerate")

	entries, err := view.GetByCapabilityIDs(f.ctx, []string{assessedCapability, unassessedCapability})
	require.NoError(t, err)

	byComponent := map[string]TimeAssessmentDTO{}
	for _, entry := range entries {
		byComponent[entry.ComponentID] = entry
	}

	assessed, found := byComponent[assessedComponent]
	require.True(t, found, "assessed realisation is listed")
	require.NotNil(t, assessed.Grade)
	assert.Equal(t, "Tolerate", *assessed.Grade)
	require.NotNil(t, assessed.Suggestion)
	require.NotNil(t, assessed.Suggestion.Grade)
	assert.Equal(t, "Eliminate", *assessed.Suggestion.Grade, "the suggestion may disagree with the recorded grade")

	unassessed, found := byComponent[unassessedComponent]
	require.True(t, found, "unassessed realisation carrying a suggestion is listed")
	assert.Nil(t, unassessed.Grade)
	assert.Nil(t, unassessed.AssessedAt)
	assert.Equal(t, "Phoenix", unassessed.ComponentName)
	require.NotNil(t, unassessed.Suggestion)
	require.NotNil(t, unassessed.Suggestion.Grade)
	assert.Equal(t, "Eliminate", *unassessed.Suggestion.Grade)
}

func TestTimeAssessmentView_PairReadCarriesTheSuggestion(t *testing.T) {
	f := newTimeSuggestionTestFixture(t, gapPillars())
	assessments := NewTimeAssessmentReadModel(f.readModel.db)
	view := NewTimeAssessmentView(assessments, f.readModel)

	capabilityID, componentID := uuid.New().String(), uuid.New().String()
	f.seedSuggestionData(suggestionSeedData{
		CapabilityID:  capabilityID,
		ComponentID:   componentID,
		ComponentName: "Seabook",
		DomainID:      uuid.New().String(),
		PillarScores:  highGapScores(),
	})

	unassessed, err := view.GetByPair(f.ctx, capabilityID, componentID)
	require.NoError(t, err)
	assert.Nil(t, unassessed, "an unassessed pair stays not found")

	f.recordAssessment(assessments, capabilityID, componentID, "Invest")

	dto, err := view.GetByPair(f.ctx, capabilityID, componentID)
	require.NoError(t, err)
	require.NotNil(t, dto)
	require.NotNil(t, dto.Grade)
	assert.Equal(t, "Invest", *dto.Grade)
	require.NotNil(t, dto.Suggestion)
	require.NotNil(t, dto.Suggestion.Grade)
	assert.Equal(t, "Eliminate", *dto.Suggestion.Grade)
	assert.Equal(t, "MEDIUM", dto.Suggestion.Confidence)
}
