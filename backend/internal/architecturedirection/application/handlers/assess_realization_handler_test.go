package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTimeAssessmentRepository struct {
	saved     []*aggregates.TimeAssessment
	loaded    *aggregates.TimeAssessment
	getErr    error
	saveErr   error
	getCalled bool
}

func (m *mockTimeAssessmentRepository) Save(_ context.Context, ta *aggregates.TimeAssessment) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, ta)
	return nil
}

func (m *mockTimeAssessmentRepository) GetByID(_ context.Context, _ string) (*aggregates.TimeAssessment, error) {
	m.getCalled = true
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.loaded, nil
}

type mockExistingTimeAssessmentLookup struct {
	id     string
	exists bool
	err    error
}

func (m *mockExistingTimeAssessmentLookup) FindAggregateIDForPair(_ context.Context, _, _ string) (string, bool, error) {
	if m.err != nil {
		return "", false, m.err
	}
	return m.id, m.exists, nil
}

func alwaysDirect(realizationID string) services.DirectRealizationLookup {
	return func(context.Context, string, string) (string, bool, error) {
		return realizationID, true, nil
	}
}

func neverDirect() services.DirectRealizationLookup {
	return func(context.Context, string, string) (string, bool, error) {
		return "", false, nil
	}
}

func failingDirectLookup(err error) services.DirectRealizationLookup {
	return func(context.Context, string, string) (string, bool, error) {
		return "", false, err
	}
}

func validAssessCmd() *commands.AssessRealization {
	return &commands.AssessRealization{
		CapabilityID: uuid.New().String(),
		ComponentID:  uuid.New().String(),
		Grade:        valueobjects.TimeGradeMigrate,
		Rationale:    "carve-out candidate",
		AssessedBy:   "architect@example.com",
	}
}

func TestAssessRealizationHandler_FirstAssessment_CreatesAggregateWithFreshID(t *testing.T) {
	repo := &mockTimeAssessmentRepository{}
	lookup := &mockExistingTimeAssessmentLookup{exists: false}

	handler := NewAssessRealizationHandler(repo, lookup, alwaysDirect(uuid.New().String()))
	cmd := validAssessCmd()
	result, err := handler.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, repo.saved[0].ID(), result.CreatedID)
	assert.NotEqual(t, cmd.CapabilityID, result.CreatedID)
	assert.False(t, repo.getCalled, "first assessment must not load an existing aggregate")
	assert.Equal(t, valueobjects.TimeGradeMigrate, repo.saved[0].Grade().Value())
	assert.Equal(t, "architect@example.com", repo.saved[0].AssessedBy())
}

func TestAssessRealizationHandler_Reassessment_LoadsExistingAndReplaces(t *testing.T) {
	capID := uuid.New().String()
	compID := uuid.New().String()
	existing := buildExistingTimeAssessment(t, capID, compID, valueobjects.TimeGradeTolerate)
	repo := &mockTimeAssessmentRepository{loaded: existing}
	lookup := &mockExistingTimeAssessmentLookup{id: existing.ID(), exists: true}

	handler := NewAssessRealizationHandler(repo, lookup, alwaysDirect(uuid.New().String()))
	cmd := &commands.AssessRealization{
		CapabilityID: capID,
		ComponentID:  compID,
		Grade:        valueobjects.TimeGradeEliminate,
		Rationale:    "reconsidered",
		AssessedBy:   "b@example.com",
	}
	result, err := handler.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.True(t, repo.getCalled)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, existing.ID(), result.CreatedID)
	assert.Equal(t, valueobjects.TimeGradeEliminate, repo.saved[0].Grade().Value())
}

func TestAssessRealizationHandler_NoDirectRealization_Fails(t *testing.T) {
	repo := &mockTimeAssessmentRepository{}
	handler := NewAssessRealizationHandler(repo, &mockExistingTimeAssessmentLookup{}, neverDirect())

	_, err := handler.Handle(context.Background(), validAssessCmd())

	assert.ErrorIs(t, err, services.ErrReferencedEntityNotFound)
	assert.Empty(t, repo.saved)
}

func TestAssessRealizationHandler_DirectRealizationCheckErrors_Fails(t *testing.T) {
	repo := &mockTimeAssessmentRepository{}
	handler := NewAssessRealizationHandler(repo, &mockExistingTimeAssessmentLookup{}, failingDirectLookup(errors.New("db down")))

	_, err := handler.Handle(context.Background(), validAssessCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestAssessRealizationHandler_InvalidInputs_FailWithoutSaving(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*commands.AssessRealization)
	}{
		{"invalid capability id", func(c *commands.AssessRealization) { c.CapabilityID = "not-a-uuid" }},
		{"invalid component id", func(c *commands.AssessRealization) { c.ComponentID = "not-a-uuid" }},
		{"invalid grade", func(c *commands.AssessRealization) { c.Grade = "invest" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockTimeAssessmentRepository{}
			handler := NewAssessRealizationHandler(repo, &mockExistingTimeAssessmentLookup{}, alwaysDirect(uuid.New().String()))
			cmd := validAssessCmd()
			tc.mutate(cmd)

			_, err := handler.Handle(context.Background(), cmd)

			assert.Error(t, err)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestAssessRealizationHandler_LookupError_Fails(t *testing.T) {
	repo := &mockTimeAssessmentRepository{}
	lookup := &mockExistingTimeAssessmentLookup{err: errors.New("db down")}

	handler := NewAssessRealizationHandler(repo, lookup, alwaysDirect(uuid.New().String()))
	_, err := handler.Handle(context.Background(), validAssessCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func buildExistingTimeAssessment(t *testing.T, capID, compID, grade string) *aggregates.TimeAssessment {
	t.Helper()
	capRef, err := valueobjects.NewPhysicalCapabilityRef(capID)
	require.NoError(t, err)
	compRef, err := valueobjects.NewApplicationRef(compID)
	require.NoError(t, err)
	gradeVO, err := valueobjects.NewTimeGrade(grade)
	require.NoError(t, err)
	rationale := mustNewNarrative(t, "first")
	ta, err := aggregates.NewTimeAssessment(aggregates.TimeAssessmentFacts{
		CapabilityID:  capRef,
		ComponentID:   compRef,
		RealizationID: uuid.New().String(),
		Grade:         gradeVO,
		Rationale:     rationale,
		AssessedBy:    "a@example.com",
	})
	require.NoError(t, err)
	ta.MarkChangesAsCommitted()
	return ta
}
