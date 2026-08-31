package readmodels

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAssessments struct {
	byPair       *TimeAssessmentDTO
	all          []TimeAssessmentDTO
	rollups      []TimeAssessmentRollupDTO
	capabilityID []string
	err          error
}

func (s *stubAssessments) GetByPair(_ context.Context, _, _ string) (*TimeAssessmentDTO, error) {
	return s.byPair, s.err
}

func (s *stubAssessments) GetByCapabilityIDs(_ context.Context, capabilityIDs []string) ([]TimeAssessmentDTO, error) {
	s.capabilityID = capabilityIDs
	return s.all, s.err
}

func (s *stubAssessments) GetAll(_ context.Context) ([]TimeAssessmentDTO, error) {
	return s.all, s.err
}

func (s *stubAssessments) GetRollupsByComponentIDs(_ context.Context, _ []string) ([]TimeAssessmentRollupDTO, error) {
	return s.rollups, s.err
}

type stubSuggestions struct {
	realizations []SuggestedRealization
	pair         *TimeSuggestionDTO
	err          error
}

func (s *stubSuggestions) All(_ context.Context) ([]SuggestedRealization, error) {
	return s.realizations, s.err
}

func (s *stubSuggestions) ForCapabilities(_ context.Context, _ []string) ([]SuggestedRealization, error) {
	return s.realizations, s.err
}

func (s *stubSuggestions) ForPair(_ context.Context, _, _ string) (*TimeSuggestionDTO, error) {
	return s.pair, s.err
}

type suggestionSpec struct {
	grade      string
	confidence string
}

func (s suggestionSpec) dto() TimeSuggestionDTO {
	return TimeSuggestionDTO{Grade: &s.grade, Confidence: s.confidence}
}

func (s suggestionSpec) forPair(pair RealizationPair) SuggestedRealization {
	return SuggestedRealization{Pair: pair, Suggestion: s.dto()}
}

var testPair = RealizationPair{CapabilityID: "cap-1", ComponentID: "comp-1"}

func recordedAssessment(pair RealizationPair, recorded string) TimeAssessmentDTO {
	assessedAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	return TimeAssessmentDTO{
		ID:           "ta-" + pair.ComponentID,
		CapabilityID: pair.CapabilityID,
		ComponentID:  pair.ComponentID,
		Grade:        &recorded,
		AssessedAt:   &assessedAt,
	}
}

func TestTimeAssessmentView_GetByPair_ComposesSuggestion(t *testing.T) {
	assessment := recordedAssessment(testPair, "Tolerate")
	suggestion := suggestionSpec{grade: "Migrate", confidence: "MEDIUM"}.dto()
	view := NewTimeAssessmentView(
		&stubAssessments{byPair: &assessment},
		&stubSuggestions{pair: &suggestion},
	)

	dto, err := view.GetByPair(context.Background(), testPair.CapabilityID, testPair.ComponentID)

	require.NoError(t, err)
	require.NotNil(t, dto)
	assert.Equal(t, "Tolerate", *dto.Grade)
	require.NotNil(t, dto.Suggestion)
	assert.Equal(t, "Migrate", *dto.Suggestion.Grade)
	assert.Equal(t, "MEDIUM", dto.Suggestion.Confidence)
}

func TestTimeAssessmentView_GetByPair_UnassessedStaysNotFound(t *testing.T) {
	suggestion := suggestionSpec{grade: "Migrate", confidence: "HIGH"}.dto()
	view := NewTimeAssessmentView(&stubAssessments{}, &stubSuggestions{pair: &suggestion})

	dto, err := view.GetByPair(context.Background(), testPair.CapabilityID, testPair.ComponentID)

	require.NoError(t, err)
	assert.Nil(t, dto)
}

func TestTimeAssessmentView_Collection_ComposesSuggestionOntoRecordedGrade(t *testing.T) {
	view := NewTimeAssessmentView(
		&stubAssessments{all: []TimeAssessmentDTO{recordedAssessment(testPair, "Tolerate")}},
		&stubSuggestions{realizations: []SuggestedRealization{
			suggestionSpec{grade: "Migrate", confidence: "HIGH"}.forPair(testPair),
		}},
	)

	assessments, err := view.GetAll(context.Background())

	require.NoError(t, err)
	require.Len(t, assessments, 1)
	assert.Equal(t, "Tolerate", *assessments[0].Grade)
	require.NotNil(t, assessments[0].Suggestion)
	assert.Equal(t, "Migrate", *assessments[0].Suggestion.Grade)
}

func TestTimeAssessmentView_Collection_ListsUnassessedRealizationCarryingASuggestion(t *testing.T) {
	view := NewTimeAssessmentView(
		&stubAssessments{},
		&stubSuggestions{realizations: []SuggestedRealization{{
			Pair:           testPair,
			CapabilityName: "Pricing",
			ComponentName:  "Seabook",
			Suggestion:     suggestionSpec{grade: "Eliminate", confidence: "HIGH"}.dto(),
		}}},
	)

	assessments, err := view.GetByCapabilityIDs(context.Background(), []string{testPair.CapabilityID})

	require.NoError(t, err)
	require.Len(t, assessments, 1)
	assert.Nil(t, assessments[0].Grade)
	assert.Nil(t, assessments[0].AssessedAt)
	assert.Equal(t, "Pricing", assessments[0].CapabilityName)
	assert.Equal(t, "Seabook", assessments[0].ComponentName)
	require.NotNil(t, assessments[0].Suggestion)
	assert.Equal(t, "Eliminate", *assessments[0].Suggestion.Grade)
}

func TestTimeAssessmentView_Collection_OmitsUnassessedRealizationWithoutASuggestedGrade(t *testing.T) {
	view := NewTimeAssessmentView(
		&stubAssessments{},
		&stubSuggestions{realizations: []SuggestedRealization{{
			Pair:       testPair,
			Suggestion: TimeSuggestionDTO{Confidence: "LOW"},
		}}},
	)

	assessments, err := view.GetAll(context.Background())

	require.NoError(t, err)
	assert.Empty(t, assessments)
}

func TestTimeAssessmentView_Collection_FailsWhenSuggestionsCannotBeComputed(t *testing.T) {
	view := NewTimeAssessmentView(
		&stubAssessments{all: []TimeAssessmentDTO{recordedAssessment(testPair, "Invest")}},
		&stubSuggestions{err: errors.New("pillars unavailable")},
	)

	_, err := view.GetAll(context.Background())

	assert.Error(t, err)
}

func TestTimeAssessmentView_RollupsPassThrough(t *testing.T) {
	rollups := []TimeAssessmentRollupDTO{{ComponentID: testPair.ComponentID}}
	view := NewTimeAssessmentView(&stubAssessments{rollups: rollups}, &stubSuggestions{})

	got, err := view.GetRollupsByComponentIDs(context.Background(), []string{testPair.ComponentID})

	require.NoError(t, err)
	assert.Equal(t, rollups, got)
}
