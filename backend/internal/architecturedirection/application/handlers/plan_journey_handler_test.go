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

type mockCapabilityJourneyRepository struct {
	saved   []*aggregates.CapabilityJourney
	loaded  *aggregates.CapabilityJourney
	getErr  error
	saveErr error
}

func (m *mockCapabilityJourneyRepository) Save(_ context.Context, j *aggregates.CapabilityJourney) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, j)
	return nil
}

func (m *mockCapabilityJourneyRepository) GetByID(_ context.Context, _ string) (*aggregates.CapabilityJourney, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.loaded, nil
}

type mockActiveJourneyLookup struct {
	id     string
	exists bool
	err    error
}

func (m *mockActiveJourneyLookup) FindActiveJourneyIDForCapability(_ context.Context, _ string) (string, bool, error) {
	if m.err != nil {
		return "", false, m.err
	}
	return m.id, m.exists, nil
}

func alwaysExists(_ context.Context, _ string) (bool, error) { return true, nil }
func neverExists(_ context.Context, _ string) (bool, error)  { return false, nil }

func allExistingRefs() JourneyReferenceChecks {
	return JourneyReferenceChecks{
		CapabilityExists:              alwaysExists,
		ComponentExists:               alwaysExists,
		DomainExists:                  alwaysExists,
		CapabilityEffectivelyInDomain: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
	}
}

func validPlanJourneyCmd() *commands.PlanJourney {
	return &commands.PlanJourney{
		CapabilityID:     uuid.New().String(),
		Kind:             valueobjects.JourneyKindMigration,
		FromComponentIDs: []string{uuid.New().String()},
		ToComponentID:    uuid.New().String(),
		Note:             "moving to a modern platform",
		PlannedBy:        "architect@example.com",
	}
}

func TestPlanJourneyHandler_ValidMigration_CreatesJourney(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs())

	cmd := validPlanJourneyCmd()
	result, err := h.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, repo.saved[0].ID(), result.CreatedID)
	assert.Equal(t, valueobjects.JourneyStatusPlanned, repo.saved[0].Status().Value())
}

func TestPlanJourneyHandler_ActiveJourneyExists_ReturnsErrorCarryingExistingID(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	existingID := uuid.New().String()
	lookup := &mockActiveJourneyLookup{id: existingID, exists: true}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs())

	_, err := h.Handle(context.Background(), validPlanJourneyCmd())

	require.Error(t, err)
	var conflict *services.ActiveJourneyError
	require.True(t, errors.As(err, &conflict))
	assert.Equal(t, existingID, conflict.ExistingJourneyID)
	assert.Empty(t, repo.saved)
}

func TestPlanJourneyHandler_CapabilityDoesNotExist_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	refs := allExistingRefs()
	refs.CapabilityExists = neverExists
	h := NewPlanJourneyHandler(repo, lookup, refs)

	_, err := h.Handle(context.Background(), validPlanJourneyCmd())

	assert.ErrorIs(t, err, services.ErrReferencedEntityNotFound)
	assert.Empty(t, repo.saved)
}

func TestPlanJourneyHandler_ComponentDoesNotExist_Fails(t *testing.T) {
	cases := []struct {
		name          string
		missingCompFn func(cmd *commands.PlanJourney) string
	}{
		{"to-component missing", func(cmd *commands.PlanJourney) string { return cmd.ToComponentID }},
		{"from-component missing", func(cmd *commands.PlanJourney) string { return cmd.FromComponentIDs[0] }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCapabilityJourneyRepository{}
			lookup := &mockActiveJourneyLookup{exists: false}
			cmd := validPlanJourneyCmd()
			missingComponentID := tc.missingCompFn(cmd)
			refs := allExistingRefs()
			refs.ComponentExists = func(_ context.Context, id string) (bool, error) { return id != missingComponentID, nil }
			h := NewPlanJourneyHandler(repo, lookup, refs)

			_, err := h.Handle(context.Background(), cmd)

			assert.ErrorIs(t, err, services.ErrReferencedEntityNotFound)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestPlanJourneyHandler_KindCardinalityViolation_Fails(t *testing.T) {
	cases := []struct {
		name    string
		kind    string
		sources int
	}{
		{"migration with zero from-apps rejected (rule 3)", valueobjects.JourneyKindMigration, 0},
		{"consolidation with one from-app rejected (rule 3)", valueobjects.JourneyKindConsolidation, 1},
		{"carve-out with two from-apps rejected (rule 3)", valueobjects.JourneyKindCarveOut, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCapabilityJourneyRepository{}
			lookup := &mockActiveJourneyLookup{exists: false}
			h := NewPlanJourneyHandler(repo, lookup, allExistingRefs())
			cmd := validPlanJourneyCmd()
			cmd.Kind = tc.kind
			ids := make([]string, tc.sources)
			for i := range ids {
				ids[i] = uuid.New().String()
			}
			cmd.FromComponentIDs = ids

			_, err := h.Handle(context.Background(), cmd)

			assert.ErrorIs(t, err, valueobjects.ErrInvalidSourceApplicationCount)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestPlanJourneyHandler_AggregateInvariantViolations_Fail(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(cmd *commands.PlanJourney)
		wantErr error
	}{
		{
			name:    "target among sources",
			mutate:  func(cmd *commands.PlanJourney) { cmd.FromComponentIDs = []string{cmd.ToComponentID} },
			wantErr: aggregates.ErrJourneyTargetAmongSources,
		},
		{
			name: "move without target domain",
			mutate: func(cmd *commands.PlanJourney) {
				cmd.Kind = valueobjects.JourneyKindMove
				cmd.FromComponentIDs = []string{}
			},
			wantErr: aggregates.ErrJourneyMoveRequiresTargetDomain,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCapabilityJourneyRepository{}
			lookup := &mockActiveJourneyLookup{exists: false}
			h := NewPlanJourneyHandler(repo, lookup, allExistingRefs())
			cmd := validPlanJourneyCmd()
			tc.mutate(cmd)

			_, err := h.Handle(context.Background(), cmd)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestPlanJourneyHandler_ValidMove_VerifiesDomainAndParent(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	cmd := validPlanJourneyCmd()
	cmd.Kind = valueobjects.JourneyKindMove
	cmd.FromComponentIDs = []string{}
	cmd.TargetDomainID = uuid.New().String()
	cmd.TargetParentID = uuid.New().String()
	cmd.ResultingName = "Freight invoicing"

	var checkedCapability, checkedDomain string
	refs := allExistingRefs()
	refs.CapabilityEffectivelyInDomain = func(_ context.Context, capabilityID, domainID string) (bool, error) {
		checkedCapability, checkedDomain = capabilityID, domainID
		return true, nil
	}
	h := NewPlanJourneyHandler(repo, lookup, refs)

	_, err := h.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, cmd.TargetParentID, checkedCapability)
	assert.Equal(t, cmd.TargetDomainID, checkedDomain)
}

func TestPlanJourneyHandler_MoveParentNotEffectivelyInDomain_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	cmd := validPlanJourneyCmd()
	cmd.Kind = valueobjects.JourneyKindMove
	cmd.FromComponentIDs = []string{}
	cmd.TargetDomainID = uuid.New().String()
	cmd.TargetParentID = uuid.New().String()
	cmd.ResultingName = "Freight invoicing"

	refs := allExistingRefs()
	refs.CapabilityEffectivelyInDomain = func(_ context.Context, _, _ string) (bool, error) { return false, nil }
	h := NewPlanJourneyHandler(repo, lookup, refs)

	_, err := h.Handle(context.Background(), cmd)

	assert.ErrorIs(t, err, services.ErrReferencedEntityNotFound)
	assert.Empty(t, repo.saved)
}

func TestPlanJourneyHandler_LookupError_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{err: errors.New("db down")}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs())

	_, err := h.Handle(context.Background(), validPlanJourneyCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestPlanJourneyHandler_InvalidCommandType_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs())

	_, err := h.Handle(context.Background(), &commands.StartJourney{})

	assert.Error(t, err)
}
